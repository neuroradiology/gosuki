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

package api

import (
	"io"
	"net/http"
	"os"
	"strings"

	"git.blob42.xyz/gomark/gosuki/internal/database"
	"git.blob42.xyz/gomark/gosuki/internal/logging"
	"git.blob42.xyz/gomark/gosuki/pkg/bookmarks"

	"git.blob42.xyz/gomark/gosuki/pkg/manager"
	"github.com/gin-gonic/gin"
)

var log = logging.GetLogger("API")

type Bookmark = bookmarks.Bookmark

func getBookmarks(c *gin.Context) {

	rows, err := database.Cache.DB.Handle.QueryContext(c, "SELECT URL, metadata, tags FROM bookmarks")
	if err != nil {
		log.Error(err)
	}
	var bookmarks []Bookmark

	var tags string
	for rows.Next() {
		bookmark := Bookmark{}
		err = rows.Scan(&bookmark.URL, &bookmark.Metadata, &tags)
		if err != nil {
			log.Error(err)
		}

		bookmark.Tags = strings.Split(tags, database.TagSep)
		//log.Debugf("GET %s", tags)
		//log.Debugf("%v", bookmark)

		bookmarks = append(bookmarks, bookmark)
	}
	//log.Debugf("%v", bookmarks)

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": bookmarks,
	})
}

type API struct {
	engine *gin.Engine
	router *gin.RouterGroup
}

func (api *API) Run(m manager.UnitManager) {
	api.router.GET("/urls", getBookmarks)

	// Run router
	// TODO: config params for api
	go func() {
		err := api.engine.Run(":4444")
		if err != nil {
			m.Panic(err)
		}

	}()

	// Wait for stop signal
	<-m.ShouldStop()
	m.Done()
}

func NewApi() *API {
	apiLogFile, _ := os.Create(".api.log")
	gin.DefaultWriter = io.MultiWriter(apiLogFile, os.Stdout)

	api := gin.Default()

	return &API{
		engine: api,
		router: api.Group("/api"),
	}

}
