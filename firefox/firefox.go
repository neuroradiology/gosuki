// TODO: unit test critical error should shutdown the browser
// TODO: shutdown procedure (also close reducer)
// TODO: handle flag management from this package
package firefox

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"git.blob42.xyz/gomark/gosuki/database"
	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/modules"
	"git.blob42.xyz/gomark/gosuki/mozilla"
	"git.blob42.xyz/gomark/gosuki/profiles"

	// "git.blob42.xyz/gomark/gosuki/profiles"
	"git.blob42.xyz/gomark/gosuki/tree"
	"git.blob42.xyz/gomark/gosuki/utils"
	"git.blob42.xyz/gomark/gosuki/watch"

	"github.com/fsnotify/fsnotify"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var (
	ErrInitFirefox = errors.New("could not start Firefox watcher")
	log            = logging.GetLogger("FF")
)

const (
	WatchMinJobInterval = 1500 * time.Millisecond
	TagsBranchName      = mozilla.TagsBranchName // name of the `tags` branch in the node tree
)

type sqlid = mozilla.Sqlid
type timestamp = int64

type AutoIncr struct {
	ID sqlid
}

type MozBookmark = mozilla.MozBookmark
type MozFolder = mozilla.MozFolder

type Firefox struct {
	*FirefoxConfig

	// sqlite con to places.sqlite
	places *database.DB

	// All elements stored in URLIndex
	URLIndexList []string

	// Map from moz_bookmarks tag ids to a tree node
	// tagMap is used as a quick lookup table into the node tree
	tagMap map[string]*tree.Node

	// map from moz_bookmarks folder id to a folder node in the tree
	folderMap map[sqlid]*tree.Node

	// internal folder map used for scanning
	folderScanMap map[sqlid]*MozFolder

	lastRunTime time.Time
}

// func (ff *Firefox) updateModifiedFolders(since timestamp) ([]*MozFolder, error) {
//     // Get list of modified folders
//     var folders = []*MozFolders
//     folderChangeQuery := map[string]interface
//
//     return nil, nil
// }

// scan all folders from moz_bookmarks and load them into the node tree
// takes a timestamp(int64) parameter to select folders based on last modified date
func (f *Firefox) scanFolders(since timestamp) ([]*MozFolder, error) {

	var folders []*MozFolder
	f.folderScanMap = make(map[sqlid]*MozFolder)

	folderQueryArgs := map[string]interface{}{
		"change_since": since,
	}

	boundQuery, args, err := f.places.Handle.BindNamed(mozilla.QFolders, folderQueryArgs)
	if err != nil {
		return nil, err
	}

	err = f.places.Handle.Select(&folders, boundQuery, args...)
	if err != nil {
		return nil, err
	}

	// store all folders in a hashmap for easier tree construction
	for _, folder := range folders {
		f.folderScanMap[folder.Id] = folder
	}

	for _, folder := range folders {
		// Ignore the `tags` virtual folder
		if folder.Id != 4 {
			f.addFolderNode(*folder)
		}
	}

	return folders, err
}

// load bookmarks and tags into the node tree then attach them to
// their assigned folder hierarchy
func (f *Firefox) loadBookmarksToTree(bookmarks []*MozBookmark) {

	for _, bkEntry := range bookmarks {
		// Create/Update URL node and apply tag node
		ok, urlNode := f.addURLNode(bkEntry.Url, bkEntry.Title, bkEntry.PlDesc)
		if !ok {
			log.Infof("url <%s> already in url index", bkEntry.Url)
		}

		/*
		 * Iterate through bookmark tags and synchronize new tags with
		 * the node tree.
		 */
		for _, tagName := range strings.Split(bkEntry.Tags, ",") {
			if tagName == "" {
				continue
			}
			seen, tagNode := f.addTagNode(tagName)
			if !seen {
				log.Infof("tag <%s> already in tag map", tagNode.Name)
			}

			// Add tag name to urlnode tags
			urlNode.Tags = utils.Extends(urlNode.Tags, tagNode.Name)

			// Add URL node as child of Tag node
			// Parent will be a folder or nothing?
			tree.AddChild(f.tagMap[tagNode.Name], urlNode)

			f.CurrentUrlCount++
		}

		// Link this URL node to its corresponding folder node if it exists.
		//TODO: add all parent folders in the tags list of this url node
		folderNode, fOk := f.folderMap[bkEntry.ParentId]
		// If we found the parent folder
		if fOk {
			tree.AddChild(folderNode, urlNode)
		}
	}
}

// scans bookmarks from places.sqlite and loads them into the node tree
func (f *Firefox) scanBookmarks() ([]*MozBookmark, error) {

	// scan folders and load them into node tree
	_, err := f.scanFolders(0)
	if err != nil {
		return nil, err
	}

	var bookmarks []*MozBookmark

	dotx, err := database.DotxQueryEmbedFS(mozilla.EmbeddedSqlQueries, mozilla.MozBookmarkQueryFile)
	if err != nil {
		return nil, err
	}
	err = dotx.Select(f.places.Handle, &bookmarks, mozilla.MozBookmarkQuery)

	// load bookmarks and tags into the node tree
	// then attach them to their assigned folder hierarchy

	return bookmarks, err
}

func (f *Firefox) scanModifiedBookmarks(since timestamp) ([]*MozBookmark, error) {
	// scan new/modifed folders and load them into node tree
	_, err := f.scanFolders(since)
	// tree.PrintTree(ff.NodeTree)
	if err != nil {
		return nil, err
	}

	var bookmarks []*MozBookmark

	dotx, err := database.DotxQueryEmbedFS(mozilla.EmbeddedSqlQueries,
		mozilla.MozChangedBookmarkQueryFile)

	if err != nil {
		return nil, err
	}

	queryArgs := map[string]interface{}{
		"change_since": since,
	}

	// query, args, err := dotx.NamedQuery(ff.places.Handle, mozilla.MozChangedBookmarkQuery, queryArgs)
	boundQuery, args, err := dotx.BindNamed(f.places.Handle, mozilla.MozChangedBookmarkQuery, queryArgs)
	if err != nil {
		return nil, err
	}

	err = f.places.Handle.Select(&bookmarks, boundQuery, args...)
	if err != nil {
		return nil, err
	}

	return bookmarks, err
}

// init is required to register the module as a plugin when it is imported
func init() {
	modules.RegisterBrowser(Firefox{FirefoxConfig: FFConfig})
	//TIP: cmd.RegisterModCommand(BrowserName, &cli.Command{
	// 	Name: "test",
	// })
	// cmd.RegisterModCommand(BrowserName, &cli.Command{
	// 	Name: "test2",
	// })
}

func NewFirefox() *Firefox {
	return &Firefox{
		FirefoxConfig: FFConfig,
		places:        &database.DB{},
		URLIndexList:  []string{},
		tagMap:        map[string]*tree.Node{},
		folderMap:     map[sqlid]*tree.Node{},
	}
}

func (f Firefox) ModInfo() modules.ModInfo {
	return modules.ModInfo{
		ID: modules.ModID(f.Name),
		//HACK: duplicate instance with init().RegisterBrowser ??
		New: func() modules.Module {
			return NewFirefox()
		},
	}
}

// Implement the profiles.ProfileManager interface
func (f *Firefox) GetProfiles() ([]*profiles.Profile, error) {
	return FirefoxProfileManager.GetProfiles()
}

func (f *Firefox) GetDefaultProfile() (*profiles.Profile, error) {
	return FirefoxProfileManager.GetDefaultProfile()
}

func (f *Firefox) GetProfilePath(p profiles.Profile) string {
	return filepath.Join(FirefoxProfileManager.ConfigDir, p.Path)
}

// TEST:
// TODO: implement watching of multiple profiles
// NOTE: should be done at core gosuki level where multiple instances
//			are spawned for each profile
// Implements browser.Initializer interface
func (f *Firefox) Init(ctx *modules.Context) error {
	log.Infof("initializing <%s>", f.Name)
	bookmarkPath, err := f.BookmarkPath()
	if err != nil {
		return err
	}

	log.Debugf("bookmark path is: %s", bookmarkPath)

	watchedPath, err := filepath.EvalSymlinks(f.BkDir)
	if err != nil {
		return err
	}

	// Setup watcher
	w := &watch.Watch{
		Path:       watchedPath,
		EventTypes: []fsnotify.Op{fsnotify.Write},
		EventNames: []string{filepath.Join(watchedPath, "places.sqlite-wal")},
		ResetWatch: false,
	}

	modules.SetupWatchersWithReducer(f.BrowserConfig, modules.ReducerChanLen, w)

	/*
	 *Run reducer to avoid duplicate jobs when a batch of events is received
	 */
	// TODO!: make a new copy of places for every new event change

	// Add a reducer to the watcher
	go watch.ReduceEvents(WatchMinJobInterval, f)

	return nil
}

func (f *Firefox) Watch() *watch.WatchDescriptor {
	return f.GetWatcher()
}

func (f Firefox) Config() *modules.BrowserConfig {
	return f.BrowserConfig
}

// Firefox custom logic for preloading the bookmarks when the browser module
// starts. Implements modules.Loader interface.
func (f *Firefox) Load() error {
	pc, err := f.initPlacesCopy()
	if err != nil {
		return err
	}

	defer pc.Clean()

	// load all bookmarks
	start := time.Now()
	bookmarks, err := f.scanBookmarks()
	if err != nil {
		return err
	}
	f.loadBookmarksToTree(bookmarks)

	f.LastFullTreeParseTime = time.Since(start)
	f.lastRunTime = time.Now().UTC()

	log.Debugf("parsed %d bookmarks and %d nodes in %s",
		f.CurrentUrlCount,
		f.CurrentNodeCount,
		f.LastFullTreeParseTime)
	f.Reset()

	// Sync the URLIndex to the buffer
	// We do not use the NodeTree here as firefox tags are represented
	// as a flat tree which is not efficient, we use the go hashmap instead

	database.SyncURLIndexToBuffer(f.URLIndexList, f.URLIndex, f.BufferDB)

	// Handle empty cache
	if empty, err := database.Cache.DB.IsEmpty(); empty {
		if err != nil {
			return err
		}
		log.Info("cache empty: loading buffer to Cachedb")

		f.BufferDB.CopyTo(database.Cache.DB)

		log.Debugf("syncing <%s> to disk", database.Cache.DB.Name)
	} else {
		f.BufferDB.SyncTo(database.Cache.DB)
	}

	database.Cache.DB.SyncToDisk(database.GetDBFullPath())

	//DEBUG:
	// tree.PrintTree(f.NodeTree)

	// Close the copy places.sqlite
	err = f.places.Close()

	return err
}

// Implements modules.Runner interface
func (ff *Firefox) Run() {
	startRun := time.Now()

	pc, err := ff.initPlacesCopy()
	if err != nil {
		log.Error(err)
	}
	defer pc.Clean()

	log.Debugf("Checking changes since <%d> %s",
		ff.lastRunTime.UTC().UnixNano()/1000,
		ff.lastRunTime.Local().Format("Mon Jan 2 15:04:05 MST 2006"))

	scanSince := ff.lastRunTime.UTC().UnixNano() / 1000

	bookmarks, err := ff.scanModifiedBookmarks(scanSince)
	if err != nil {
		log.Error(err)
	}
	ff.loadBookmarksToTree(bookmarks)
	// tree.PrintTree(ff.NodeTree)

	database.SyncURLIndexToBuffer(ff.URLIndexList, ff.URLIndex, ff.BufferDB)
	ff.BufferDB.SyncTo(database.Cache.DB)
	database.Cache.DB.SyncToDisk(database.GetDBFullPath())

	ff.LastWatchRunTime = time.Since(startRun)
	ff.lastRunTime = time.Now().UTC()
}

// Implement moduls.Shutdowner
func (f *Firefox) Shutdown() {
	log.Debugf("shutting down ... ")

	if f.places != nil {

		err := f.places.Close()
		if err != nil {
			log.Critical(err)
		}
	}
}

func (f *Firefox) getPathToPlacesCopy() string {
	return path.Join(utils.TMPDIR, f.BkFile)
}

// HACK: addUrl and addTag share a lot of code, find a way to reuse shared code
// and only pass extra details about tag/url along in some data structure
// PROBLEM: tag nodes use IDs and URL nodes use URL as hashes
func (f *Firefox) addURLNode(url, title, desc string) (bool, *tree.Node) {

	var urlNode *tree.Node
	iUrlNode, exists := f.URLIndex.Get(url)
	if !exists {
		urlNode := &tree.Node{
			Name: title,
			Type: tree.URLNode,
			URL:  url,
			Desc: desc,
		}

		log.Debugf("inserting url %s in url index", url)
		f.URLIndex.Insert(url, urlNode)
		f.URLIndexList = append(f.URLIndexList, url)
		f.CurrentNodeCount++

		return true, urlNode
	} else {
		urlNode = iUrlNode.(*tree.Node)
	}

	return false, urlNode
}

// adds a new tagNode if it is not yet in the tagMap
// returns true if tag added or false if already existing
// returns the created tagNode
func (f *Firefox) addTagNode(tagName string) (bool, *tree.Node) {
	// Check if "tags" branch exists or create it
	var branchOk bool
	var tagsBranch *tree.Node
	for _, c := range f.NodeTree.Children {
		if c.Name == TagsBranchName {
			branchOk = true
			tagsBranch = c
		}
	}

	if !branchOk {
		tagsBranch = &tree.Node{
			Name: TagsBranchName,
		}
		tree.AddChild(f.NodeTree, tagsBranch)
	}

	// node, exists :=
	node, exists := f.tagMap[tagName]
	if exists {
		return false, node
	}

	tagNode := &tree.Node{
		Name:   tagName,
		Type:   tree.TagNode,
		Parent: f.NodeTree, // root node
	}

	tree.AddChild(tagsBranch, tagNode)
	f.tagMap[tagName] = tagNode
	f.CurrentNodeCount++

	return true, tagNode
}

// add a folder node to the parsed node tree under the specified folder parent
// returns true if a new folder is created and false if folder already exists
func (f *Firefox) addFolderNode(folder MozFolder) (bool, *tree.Node) {

	// use hashmap.RBTree to keep an index of scanned folders pointing
	// to their corresponding nodes in the tree

	folderNode, seen := f.folderMap[folder.Id]

	if seen {
		// Update folder name if changed

		//TODO!: trigger bookmark tag change in gosuki.db
		if folderNode.Name != folder.Title &&
			// Ignore root folders since we use our custom names
			!utils.InList([]int{2, 3, 5, 6}, int(folder.Id)) {
			log.Debugf("folder node <%s> updated to <%s>", folderNode.Name, folder.Title)
			folderNode.Name = folder.Title
		}

		return false, folderNode
	}

	folderNode = &tree.Node{
		// Name: folder.Title,
		Type: tree.FolderNode,
	}

	// keeping the same folder structure as Firefox

	// If this folders' is a Firefox root folder use the appropriate title
	// then add it to the root node
	if utils.InList([]int{2, 3, 5, 6}, int(folder.Id)) {
		folderNode.Name = mozilla.RootFolderNames[folder.Id]
		tree.AddChild(f.NodeTree, folderNode)
	} else {
		folderNode.Name = folder.Title
	}

	// check if folder's parent is already in the tree
	fParent, ok := f.folderMap[folder.Parent]

	// if we already saw folder's parent add it underneath
	if ok {
		tree.AddChild(fParent, folderNode)

		// if we never saw this folders' parent
	} else if folder.Parent != 1 { // recursively build the parent of this folder
		_, ok := f.folderScanMap[folder.Parent]
		if ok {
			_, newParentNode := f.addFolderNode(*f.folderScanMap[folder.Parent])
			tree.AddChild(newParentNode, folderNode)

		}
	}

	// Store a pointer to this folder
	f.folderMap[folder.Id] = folderNode
	f.CurrentNodeCount++

	return true, folderNode
}

// TODO: retire this function after scanBookmarks() is implemented
// load all bookmarks from `places.sqlite` and store them in BaseBrowser.NodeTree
// this method is used the first time gosuki is started or to extract bookmarks
// using a command
func loadBookmarks(f *Firefox) {
	log.Debugf("root tree children len is %d", len(f.NodeTree.Children))
	//QGetTags := "SELECT id,title from moz_bookmarks WHERE parent = %d"
	//

	rows, err := f.places.Handle.Query(mozilla.QgetBookmarks, mozilla.TagsID)
	if err != nil {
		log.Fatal(err)
	}

	// Locked database is critical
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrBusy {
			log.Critical(err)
			f.Shutdown()
			return
		}
	}
	if err != nil {
		log.Errorf("%s: %s", f.places.Name, err)
		return
	}

	// Rebuilding node tree
	// Note: the node tree is built only for compatilibity with tree based
	// bookmark parsing and might later be useful for debug/UI features.
	// For efficiency reading after the initial Load() from
	// places.sqlite should be done using a loop instad of tree traversal.

	/*
	 *This pass is used only for fetching bookmarks from firefox.
	 *Checking against the URLIndex should not be done here
	 */
	for rows.Next() {
		var url, title, tagTitle, desc string
		var tagId sqlid

		err = rows.Scan(&url, &title, &desc, &tagId, &tagTitle)
		// log.Debugf("%s|%s|%s|%d|%s", url, title, desc, tagId, tagTitle)
		if err != nil {
			log.Error(err)
		}

		/*
		 * If this is the first time we see this tag
		 * add it to the tagMap and create its node
		 */
		ok, tagNode := f.addTagNode(tagTitle)
		if !ok {
			log.Infof("tag <%s> already in tag map", tagNode.Name)
		}

		// Add the url to the tag
		// NOTE: this call is responsible for updating URLIndexList
		ok, urlNode := f.addURLNode(url, title, desc)
		if !ok {
			log.Infof("url <%s> already in url index", url)
		}

		// Add tag name to urlnode tags
		urlNode.Tags = append(urlNode.Tags, tagNode.Name)

		// Set tag as parent to urlnode
		tree.AddChild(f.tagMap[tagTitle], urlNode)

		f.CurrentUrlCount++
	}

	log.Debugf("root tree children len is %d", len(f.NodeTree.Children))
}

// Copies places.sqlite to a tmp dir to read a VFS lock sqlite db
func (f *Firefox) initPlacesCopy() (mozilla.PlaceCopyJob, error) {
	// create a new copy job
	pc := mozilla.NewPlaceCopyJob()

	err := utils.CopyFilesToTmpFolder(path.Join(f.BkDir, f.BkFile+"*"), pc.Path())
	if err != nil {
		return pc, fmt.Errorf("could not copy places.sqlite to tmp folder: %s", err)
	}

	opts := FFConfig.PlacesDSN

	f.places, err = database.NewDB("places",
		// using the copied places file instead of the original to avoid
		// sqlite vfs lock errors
		path.Join(pc.Path(), f.BkFile),
		database.DBTypeFileDSN, opts).Init()

	if err != nil {
		return pc, err
	}

	return pc, nil
}
