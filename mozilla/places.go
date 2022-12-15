package mozilla

import (

	"time"
)

// Constants representing the meaning if IDs defined in the table
// moz_bookmarks.id 
 const (
	_           = iota // 0
	RootID           // 1
	MenuID           // 2 Main bookmarks menu
	ToolbarID        // 3 Bk tookbar that can be toggled under URL zone
	TagsID           // 4 Hidden menu used for tags, stored as a flat one level menu
	OtherID          // 5 Most bookmarks are automatically stored here
	MobileID         // 6 Mobile bookmarks stored here by default
)

type Sqlid int64

var RootFolders = map[Sqlid]string{
    RootID: RootName,
    MenuID: "Bookmarks Menu",
    ToolbarID: "Bookmarks Toolbar",
    TagsID: TagsBranchName,
    OtherID: "Other Bookmarks",
    MobileID: "Mobile Bookmarks",
}

const (
    // Name of the root node
    RootName = `ROOT`

    // Name of the `Tags` node parent to all tag nodes
    TagsBranchName = `TAGS`
)

type MozFolder struct {
	Id    Sqlid
    Parent Sqlid
	Title string
}

// placeId  title  parentFolderId  folders url plDesc lastModified
// Type used for scanning from `recursive-all-bookmarks.sql`
type MozBookmark struct {
	PlId           Sqlid `db:"plId"`
	Title          string
	Tags           string
	Folders        string
	ParentId       Sqlid  `db:"parentFolderId"`
	ParentFolder   string `db:"parentFolder"`
	Url            string
	PlDesc         string `db:"plDesc"`
	BkLastModified Sqlid  `db:"lastModified"`
}

// Type is used for scanning from `merged-places-bookmarks.sql`
// plId  plUrl plDescription bkId  bkTitle bkLastModified  isFolder  isTag  isBk  bkParent
type MergedPlaceBookmark struct {
	PlId    Sqlid  `db:"plId"`
	PlUrl   string `db:"plUrl"`
	PlDesc  string `db:"plDescription"`
	BkId    Sqlid  `db:"bkId"`
	BkTitle string `db:"bkTitle"`

	//firefox stores timestamps in milliseconds as integer
	//sqlite3 strftime('%s', ...) returns seconds
	//This field stores the timestamp as raw milliseconds
	BkLastModified Sqlid `db:"bkLastModified"`

	//NOTE: parsing into time.Time not working, I need to have an sqlite column of
	//time Datetime [see](https://github.com/mattn/go-sqlite3/issues/748)!!
	//Our query converts to the format scannable by go-sqlite3 SQLiteTimestampFormats
	//This field stores the timestamp parsable as time.Time
	// BkLastModifiedDateTime time.Time `db:"bkLastModifiedDateTime"`

	IsFolder bool  `db:"isFolder"`
	IsTag    bool  `db:"isTag"`
	IsBk     bool  `db:"isBk"`
	BkParent Sqlid `db:"bkParent"`
}

func (pb *MergedPlaceBookmark) Datetime() time.Time {
	return time.Unix(int64(pb.BkLastModified/(1000*1000)),
		int64(pb.BkLastModified%(1000*1000))*1000).UTC()
}
