//
// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
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
	"embed"
)

const (
	MozBookmarkQueryFile        = "recursive_all_bookmarks.sql"
	MozBookmarkQuery            = "recursive-all-bookmarks"
	MozChangedBookmarkQueryFile = "recursive_modified_bookmarks.sql"
	MozChangedBookmarkQuery     = "recursive-modified-bookmarks"
)

var (
	//go:embed "recursive_all_bookmarks.sql"
	//go:embed "recursive_modified_bookmarks.sql"
	EmbeddedSQLQueries embed.FS
)

// sql queries
const (
	QGetBookmarkPlace = `
	SELECT *
	FROM moz_places
	WHERE id = ?
	`
	//TEST:
	QBookmarksChanged = `
    SELECT id,type,IFNULL(fk, -1) AS fk,parent,IFNULL(title, '') AS title from moz_bookmarks
    WHERE(lastModified > :last_runtime_utc
        AND lastModified < strftime('%s', 'now')*1000*1000
        AND NOT id IN (:not_root_tags)
    )
	`

	QFolders = `
    SELECT id, title, parent FROM moz_bookmarks 
    WHERE type = 2 AND parent NOT IN (4, 0) AND lastModified > :change_since
    `

	//TEST:
	QgetBookmarks = `
    WITH bookmarks AS
	(SELECT moz_places.url AS url,
			moz_places.description as desc,
			moz_places.title as urlTitle,
			moz_bookmarks.parent AS tagId
		FROM moz_places LEFT OUTER JOIN moz_bookmarks
		ON moz_places.id = moz_bookmarks.fk
		WHERE moz_bookmarks.parent
		IN (SELECT id FROM moz_bookmarks WHERE parent = ? ))

	SELECT url, IFNULL(urlTitle, ''), IFNULL(desc,''),
			tagId, moz_bookmarks.title AS tagTitle

	FROM bookmarks LEFT OUTER JOIN moz_bookmarks
	ON tagId = moz_bookmarks.id
	ORDER BY url
    `

	//TEST:
	//TODO:
	QGetBookmarkFolders = `
		SELECT
		moz_places.id as placesId,
		moz_places.url as url,	
			moz_places.description as description,
			moz_bookmarks.title as title,
			moz_bookmarks.fk ISNULL as isFolder
			
		FROM moz_bookmarks LEFT OUTER JOIN moz_places
		ON moz_places.id = moz_bookmarks.fk
		WHERE moz_bookmarks.parent = 3
	`
)
