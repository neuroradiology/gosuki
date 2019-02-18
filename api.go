package main

import (
	"gomark/bookmarks"
	"gomark/database"
	"io"
	"net/http"
	"os"
	"strings"

	"git.sp4ke.com/sp4ke/gum"
	"github.com/gin-gonic/gin"
)

type Bookmark = bookmarks.Bookmark

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

		bookmark.Tags = strings.Split(tags, database.TagJoinSep)
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
	router *gin.Engine
}

func (api *API) Shutdown() {}

func (api *API) Run(m gum.UnitManager) {
	api.router.GET("/urls", getBookmarks)

	// Run router
	go func() {
		err := api.router.Run(":4444")
		if err != nil {
			panic(err)
		}

	}()

	// Wait for stop signal
	<-m.ShouldStop()

	api.Shutdown()
	m.Done()
}

func NewApi() *API {
	apiLogFile, _ := os.Create(".api.log")
	gin.DefaultWriter = io.MultiWriter(apiLogFile, os.Stdout)

	api := &API{
		router: gin.Default(),
	}

	return api
}
