package api

import (
	"io"
	"net/http"
	"os"
	"strings"

	"git.blob42.xyz/gomark/gosuki/bookmarks"
	"git.blob42.xyz/gomark/gosuki/internal/database"
	"git.blob42.xyz/gomark/gosuki/logging"

	"git.blob42.xyz/sp4ke/gum"
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
	engine *gin.Engine
	router *gin.RouterGroup
}

func (api *API) Shutdown() {}

func (api *API) Run(m gum.UnitManager) {
	api.router.GET("/urls", getBookmarks)

	// Run router
	// TODO: config params for api
	go func() {
		err := api.engine.Run(":4444")
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

	api := gin.Default()

	return &API{
		engine: api,
		router: api.Group("/api"),
	}

}
