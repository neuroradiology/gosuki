//
//  Copyright (c) 2024 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
//  All rights reserved.
//
//  SPDX-License-Identifier: AGPL-3.0-or-later
//
//  This file is part of GoSuki.
//
//  GoSuki is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  GoSuki is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
//

package server

import (
	"encoding/json"
	"net/http"

	"github.com/blob42/gosuki"
	db "github.com/blob42/gosuki/internal/database"
	"github.com/blob42/gosuki/pkg/logging"
)

var log = logging.GetLogger("API")

type Bookmark = gosuki.Bookmark
type RawBookmark = db.RawBookmark

// TODO!: consolidate with getbookmarks from ui/views.go
// TODO: pagination
func apiGetBookmarks(w http.ResponseWriter, r *http.Request) {
	var bookmarks []*Bookmark
	var err error
	if query := r.URL.Query().Get("query"); len(query) > 0 {
		bookmarks, err = db.QueryBookmarks(r.Context(), query, false)
	} else {
		bookmarks, err = db.ListBookmarks(r.Context())
	}
	if err != nil {
		log.Error(err)
	}
	err = json.NewEncoder(w).Encode(bookmarks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
