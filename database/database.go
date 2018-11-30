//TODO: missing defer close() on sqlite funcs
//TODO: handle `modified` time
package database

import (
	"database/sql"
	"fmt"
	"gomark/logging"
	"gomark/tools"
	"gomark/tree"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/sp4ke/hashmap"
)

var (
	_sql3conns           []*sqlite3.SQLiteConn // Only used for backup hook
	backupHookRegistered bool                  // set to true once the backup hook is registered
)

type Index = *hashmap.RBTree
type Node = tree.Node

var log = logging.GetLogger("DB")

const (
	DBFileName    = "gomarks.db"
	DBMemcacheFmt = "file:%s?mode=memory&cache=shared"
	DBBufferFmt   = "file:%s?mode=memory&cache=shared"
	DBCacheName   = "memcache"

	DBBackupMode  = "sqlite_hook_backup"
	DBUpdateMode  = "sqlite_hook_update"
	DBDefaultMode = "sqlite3"
)

//  Database schemas used for the creation of new databases
const (
	// metadata: name or title of resource
	// modified: time.Now().Unix()
	QCreateLocalDbSchema = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL UNIQUE,
		metadata text default '',
		tags text default '',
		desc text default '',
		modified integer default ?, 
		flags integer default 0
	)`

	QCreateMemDbSchema = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL UNIQUE,
		metadata text default '',
		tags text default '',
		desc text default '',
		modified integer default (strftime('%s')),
		flags integer default 0
	)`
)

// DB encapsulates an sql.DB struct. All interactions with memory/buffer and
// disk databases are done through the DB object
type DB struct {
	Name       string
	Path       string
	Handle     *sqlx.DB
	EngineMode string
	AttachedTo []string
}

func New(name string, path string) *DB {
	return &DB{
		Name:       name,
		Path:       path,
		Handle:     nil,
		EngineMode: DBDefaultMode,
	}
}

func NewRO(name string, path string) *DB {
	var err error
	expandedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		log.Error(err)
	}

	pathRO := fmt.Sprintf("file:%s?_journal_mode=WAL", expandedPath)

	db := New(name, pathRO)

	if db.Handle != nil {
		if err != nil {
			log.Errorf("%s: already initialized", db)
		}
		return db
	}

	// Create the sqlite connection
	db.Handle, err = sqlx.Open(db.EngineMode, db.Path)
	log.Debugf("<%s> opened at <%s> with mode <%s>", db.Name, db.Path,
		db.EngineMode)
	if err != nil {
		log.Critical(err)
	}

	return db

}

// Initialize a sqlite database with Gomark Schema
func (db *DB) Init() {

	// `cacheDB` is a memory replica of disk db

	var err error

	if db.Handle != nil {
		log.Errorf("%s: already initialized", db)
		return
	}

	// Create the memory cache db
	db.Handle, err = sqlx.Open("sqlite3", db.Path)
	//log.Debugf("db <%s> opend at at <%s>", db.Name, db.Path)
	log.Debugf("<%s> opened at <%s>", db.Name, db.Path)
	if err != nil {
		log.Critical(err)
	}

	// Populate db schema
	tx, err := db.Handle.Begin()
	if err != nil {
		log.Error(err)
	}

	stmt, err := tx.Prepare(QCreateMemDbSchema)
	if err != nil {
		log.Error(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

	log.Debugf("<%s> initialized", db.Name)
}

func (db *DB) Attach(attached *DB) {

	stmtStr := fmt.Sprintf("ATTACH DATABASE '%s' AS '%s'",
		attached.Path,
		attached.Name)
	_, err := db.Handle.Exec(stmtStr)

	if err != nil {
		log.Error(err)
	}

	db.AttachedTo = append(db.AttachedTo, attached.Name)

	/////////////////
	// For debug only
	/////////////////
	//var idx int
	//var dt string
	//var name string

	//rows, err := db.Handle.Query("PRAGMA database_list;")
	//logPanic(err)
	//for rows.Next() {
	//err = rows.Scan(&idx, &dt, &name)
	//logPanic(err)
	//log.Debugf("pragmalist: %s", dt)
	//}
}

func (db *DB) Close() error {
	log.Debugf("Closing DB <%s>", db.Name)
	err := db.Handle.Close()
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Count() int {
	var count int

	row := db.Handle.QueryRow("select count(*) from bookmarks")
	err := row.Scan(&count)
	if err != nil {
		log.Error(err)
	}

	return count
}

func (db *DB) Print() error {

	var url, tags string

	rows, err := db.Handle.Query("select url,tags from bookmarks")

	for rows.Next() {
		err = rows.Scan(&url, &tags)
		if err != nil {
			return err
		}
		log.Debugf("url:%s  tags:%s", url, tags)
	}

	return nil
}

func (db *DB) IsEmpty() (bool, error) {
	var count int

	row := db.Handle.QueryRow("select count(*) from bookmarks")

	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return false, nil
	}

	return true, nil
}

// For ever row in `src` try to insert it into `dst`.
// If if fails then try to update it. It means `src` is synced to `dst`
func (src *DB) SyncTo(dst *DB) {
	var sqlite3Err sqlite3.Error
	var existingUrls []*SBookmark

	log.Debugf("syncing <%s> to <%s>", src.Name, dst.Name)

	getSourceTable, err := src.Handle.Prepare(`SELECT * FROM bookmarks`)
	defer getSourceTable.Close()
	if err != nil {
		log.Error(err)
	}

	getDstTags, err := dst.Handle.Prepare(
		`SELECT tags FROM bookmarks WHERE url=? LIMIT 1`,
	)
	defer getDstTags.Close()
	if err != nil {
		log.Error(err)
	}

	tryInsertDstRow, err := dst.Handle.Prepare(
		`INSERT INTO
		bookmarks(url, metadata, tags, desc, flags)
		VALUES (?, ?, ?, ?, ?)`,
	)
	defer tryInsertDstRow.Close()
	if err != nil {
		log.Error(err)
	}

	updateDstRow, err := dst.Handle.Prepare(
		`UPDATE bookmarks
		SET (metadata, tags, desc, modified, flags) = (?,?,?,strftime('%s'),?)
		WHERE url=?
		`,
	)

	defer updateDstRow.Close()
	if err != nil {
		log.Error(err)
	}

	srcTable, err := getSourceTable.Query()
	if err != nil {
		log.Error(err)
	}

	// Lock destination db
	dstTx, err := dst.Handle.Begin()
	if err != nil {
		log.Error(err)
	}

	// Start syncing all entries from source table
	for srcTable.Next() {

		// Fetch on row
		scan, err := ScanBookmarkRow(srcTable)
		if err != nil {
			log.Error(err)
		}

		// Try to insert to row in dst table
		_, err = dstTx.Stmt(tryInsertDstRow).Exec(
			scan.Url,
			scan.metadata,
			scan.tags,
			scan.desc,
			scan.flags,
		)

		if err != nil {
			sqlite3Err = err.(sqlite3.Error)
		}

		if err != nil && sqlite3Err.Code != sqlite3.ErrConstraint {
			log.Error(err)
		}

		// Record already exists in dst table, we need to use update
		// instead.
		if err != nil && sqlite3Err.Code == sqlite3.ErrConstraint {
			existingUrls = append(existingUrls, scan)
		}
	}

	err = dstTx.Commit()
	if err != nil {
		log.Error(err)
	}

	// Start a new transaction to update the existing urls
	dstTx, err = dst.Handle.Begin() // Lock dst db
	if err != nil {
		log.Error(err)
	}

	// Traverse existing urls and try an update this time
	for _, scan := range existingUrls {
		var tags string

		//log.Debugf("updating existing %s", scan.Url)

		row := getDstTags.QueryRow(
			scan.Url,
		)
		row.Scan(&tags)

		//log.Debugf("src tags: %v", scan.tags)
		//log.Debugf("dst tags: %v", dstTags)
		srcTags := strings.Split(scan.tags, TagJoinSep)
		dstTags := strings.Split(tags, TagJoinSep)
		tagMap := make(map[string]bool)
		for _, v := range srcTags {
			tagMap[v] = true
		}
		for _, v := range dstTags {
			tagMap[v] = true
		}

		var newTags []string //merged tags
		for k, _ := range tagMap {
			newTags = append(newTags, k)
		}

		_, err = dstTx.Stmt(updateDstRow).Exec(
			scan.metadata,
			strings.Join(newTags, TagJoinSep),
			scan.desc,
			0, //flags
			scan.Url,
		)

		if err != nil {
			log.Errorf("%s: %s", err, scan.Url)
		}

	}

	err = dstTx.Commit()
	if err != nil {
		log.Error(err)
	}

	// If we are syncing to memcache, sync cache to disk
	if dst.Name == DBCacheName {
		err = dst.SyncToDisk(GetDBFullPath())
		if err != nil {
			log.Error(err)
		}
	}

}

// Copy from src DB to dst DB
// Source DB os overwritten
func (src *DB) CopyTo(dst *DB) {

	log.Debugf("Copying <%s>(%d) to <%s>(%d)", src.Name,
		src.Count(),
		dst.Name,
		dst.Count())

	srcDb, err := sqlx.Open(DBBackupMode, src.Path)
	defer func() {
		srcDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	if err != nil {
		log.Error(err)
	}

	srcDb.Ping()

	dstDb, err := sqlx.Open(DBBackupMode, dst.Path)
	defer func() {
		dstDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	if err != nil {
		log.Error(err)
	}
	dstDb.Ping()

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	if err != nil {
		log.Error(err)
	}

	_, err = bk.Step(-1)
	if err != nil {
		log.Error(err)
	}

	bk.Finish()
}

func (src *DB) SyncToDisk(dbpath string) error {
	log.Debugf("Syncing <%s> to <%s>", src.Name, dbpath)

	//log.Debugf("[flush] openeing <%s>", src.path)
	srcDb, err := sqlx.Open(DBBackupMode, src.Path)
	defer flushSqliteCon(srcDb)
	if err != nil {
		return err
	}
	srcDb.Ping()

	//log.Debugf("[flush] opening <%s>", DB_FILENAME)

	dbUri := fmt.Sprintf("file:%s", dbpath)
	bkDb, err := sqlx.Open(DBBackupMode, dbUri)
	defer flushSqliteCon(bkDb)
	if err != nil {
		return err
	}
	bkDb.Ping()

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	if err != nil {
		return err
	}

	_, err = bk.Step(-1)
	if err != nil {
		return err
	}

	bk.Finish()

	return nil
}

func (dst *DB) SyncFromDisk(dbpath string) error {

	log.Debugf("Syncing <%s> to <%s>", dbpath, dst.Name)

	dbUri := fmt.Sprintf("file:%s", dbpath)
	srcDb, err := sqlx.Open(DBBackupMode, dbUri)
	defer flushSqliteCon(srcDb)
	if err != nil {
		return err
	}
	srcDb.Ping()

	//log.Debugf("[flush] opening <%s>", DB_FILENAME)
	bkDb, err := sqlx.Open(DBBackupMode, dst.Path)
	defer flushSqliteCon(bkDb)
	if err != nil {
		return err
	}
	bkDb.Ping()

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	if err != nil {
		return err
	}

	_, err = bk.Step(-1)
	if err != nil {
		return err
	}

	bk.Finish()

	return nil
}

// Print debug a single row (does not run rows.next())
func DebugPrintRow(rows *sql.Rows) {
	cols, _ := rows.Columns()
	count := len(cols)
	values := make([]interface{}, count)
	valuesPtrs := make([]interface{}, count)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	for _, col := range cols {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\t")
	}

	fmt.Fprintf(w, "\n")

	for i, _ := range cols {
		valuesPtrs[i] = &values[i]
	}
	rows.Scan(valuesPtrs...)

	finalValues := make(map[string]interface{})
	for i, col := range cols {
		var v interface{}
		val := values[i]
		b, ok := val.([]byte)
		if ok {
			v = string(b)
		} else {
			v = val
		}

		finalValues[col] = v
	}

	for _, col := range cols {
		fmt.Fprintf(w, "%v\t", finalValues[col])
	}
	fmt.Fprintf(w, "\n")
	w.Flush()
}

// Print debug Rows results
func DebugPrintRows(rows *sql.Rows) {
	cols, _ := rows.Columns()
	count := len(cols)
	values := make([]interface{}, count)
	valuesPtrs := make([]interface{}, count)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	for _, col := range cols {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\t")
	}

	fmt.Fprintf(w, "\n")

	for rows.Next() {
		for i, _ := range cols {
			valuesPtrs[i] = &values[i]
		}
		rows.Scan(valuesPtrs...)

		finalValues := make(map[string]interface{})
		for i, col := range cols {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}

			finalValues[col] = v
		}

		for _, col := range cols {
			fmt.Fprintf(w, "%v\t", finalValues[col])
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
}

// Struct represetning the schema of `bookmarks` db.
// The order in the struct respects the columns order
type SBookmark struct {
	id       int
	Url      string
	metadata string
	tags     string
	desc     string
	modified int64
	flags    int
}

// Scans a row into `SBookmark` schema
func ScanBookmarkRow(row *sql.Rows) (*SBookmark, error) {
	scan := new(SBookmark)
	err := row.Scan(
		&scan.id,
		&scan.Url,
		&scan.metadata,
		&scan.tags,
		&scan.desc,
		&scan.modified,
		&scan.flags,
	)

	if err != nil {
		return nil, err
	}

	return scan, nil
}

func SyncURLIndexToBuffer(urls []string, index Index, buffer *DB) {
	for _, url := range urls {
		iNode, exists := index.Get(url)
		if !exists {
			log.Warningf("url does not exist in index: %s", url)
			break
		}
		node := iNode.(*Node)
		bk := node.GetBookmark()
		buffer.InsertOrUpdateBookmark(bk)
	}
}

func SyncTreeToBuffer(node *Node, buffer *DB) {
	if node.Type == "url" {
		bk := node.GetBookmark()
		buffer.InsertOrUpdateBookmark(bk)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			SyncTreeToBuffer(node, buffer)
		}
	}
}

func GetDBFullPath() string {
	dbdir := tools.GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, DBFileName)
	return dbpath
}

func flushSqliteCon(con *sqlx.DB) {
	con.Close()
	_sql3conns = _sql3conns[:len(_sql3conns)-1]
	log.Debugf("Flushed sqlite conns -> %v", _sql3conns)
}

func registerSqliteHooks() {
	// sqlite backup hook
	log.Debugf("backup_hook: registering driver %s", DBBackupMode)
	// Register the hook
	sql.Register(DBBackupMode,
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				//log.Debugf("[ConnectHook] registering new connection")
				_sql3conns = append(_sql3conns, conn)
				//log.Debugf("%v", _sql3conns)
				return nil
			},
		})

}

func init() {
	registerSqliteHooks()
}
