package firefox

import (
	"testing"

	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"github.com/gchaincl/dotsql"
	"github.com/swithek/dotsqlx"
)

func Test_loadQueries(t *testing.T) {

    exists, err := utils.CheckFileExists("testdata/places.sqlite")
    if err != nil {
      t.Fatal(err)
    }

    if !exists {
        t.Fatal("places file does not exist")
    }
    

	db := database.NewDB("test_places", "testdata/places.sqlite", database.DBTypeFileDSN)
    defer func(){
        err := db.Close()
        if err != nil {
            t.Fatal()
        }
    }()
    

    _, err = db.Init()
    if err != nil {
      t.Error(err)
    }
    

	dot, err := dotsql.LoadFromFile("queries.sql")
	if err != nil {
		t.Fatal(err)
	}

    dotx := dotsqlx.Wrap(dot)
    _, err = dotx.Raw("merge-places-bookmarks")
    if err != nil {
      t.Fatal(err)
    }



    // Loading of bookmarks and their folders algorithm:
    //     1. [ ] execute merged places_bookmarks table query
    //         [*] scan the query into a bookmark_places struct
    //     3- go through bookmarks and
    //         - add tag nodes
    //         - add url nodes
    //         ?- add hierarchy relationship ?
    //     4- Sync URLIndex to the the buffer DB

    t.Run("Scan a bookmark", func(t *testing.T){

        rowsx, err := dotx.Queryx(db.Handle, "merge-places-bookmarks")
        if err != nil {
          t.Fatal(err)
        }

        for rowsx.Next() {
            var placebk PlaceBookmark

            err = rowsx.StructScan(&placebk)
            if err != nil {
              t.Error(err)
            }

            t.Log(placebk)
        }
    })
}
