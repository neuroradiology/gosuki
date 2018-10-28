package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func getBookmarks(c *gin.Context) {

	rows, err := CacheDB.Handle.QueryContext(c, "SELECT URL, metadata, tags FROM bookmarks")
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

		bookmark.Tags = strings.Split(tags, TagJoinSep)

		//log.Debugf("GET %s", tags)
		//log.Debugf("%v", bookmark)

		bookmarks = append(bookmarks, bookmark)
	}
	//log.Debugf("%v", bookmarks)

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": bookmarks,
	})
}
