//
// Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

package mozilla

import (
	"os"
	"path"
	"time"

	"git.blob42.xyz/gosuki/gosuki/internal/utils"
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

// Represnets the root folder names as shown on Firefox
var RootFolderTitles = map[Sqlid]string{
    MenuID: "Bookmarks Menu",
    ToolbarID: "Bookmarks Toolbar",
    OtherID: "Other Bookmarks",
    MobileID: "Mobile Bookmarks",
}

// Some root folders names in the tree
var RootFolderNames = map[Sqlid]string{
    RootID: RootName,
    TagsID: TagsBranchName,
    MenuID: "menu",
    ToolbarID: "toolbar",
    OtherID: "other",
    MobileID: "mobile",
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

var CopyJobs []PlaceCopyJob

type PlaceCopyJob struct {
    Id string
}

func NewPlaceCopyJob() PlaceCopyJob {
    pc := PlaceCopyJob{
        Id: utils.GenStringID(5),
    }

    err := pc.makePath()
    if err != nil {
      log.Fatal(err)
    }

    CopyJobs = append(CopyJobs, pc)

    return pc
}

func (pc PlaceCopyJob) makePath() error {
    // make sure TMPDIR is not empty
    if len(utils.TMPDIR) == 0 {
        log.Error("missing tmp dir")
        return nil
    }

    return os.Mkdir(path.Join(utils.TMPDIR, pc.Id), 0750)
}

func (pc PlaceCopyJob) Path() string {
    return path.Join(utils.TMPDIR, pc.Id)
}

func (pc PlaceCopyJob) Clean() error {
    log.Debugf("cleaning <%s>", pc.Path())
    return os.RemoveAll(pc.Path())
}
