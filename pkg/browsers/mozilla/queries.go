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
	QFolders = `
    SELECT id, title, parent FROM moz_bookmarks 
    WHERE type = 2 AND parent NOT IN (4, 0) AND lastModified > :change_since
    `
)
