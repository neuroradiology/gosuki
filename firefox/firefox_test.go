package firefox

import (
	"fmt"
	"os"
	"testing"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/index"
	"git.sp4ke.xyz/sp4ke/gomark/mozilla"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"github.com/stretchr/testify/assert"
)

// func Test_scanBookmarks(t *testing.T) {
// 	t.Run("")
// }

var ff Firefox

func TestMain(m *testing.M) {
	ff = Firefox{
		FirefoxConfig: &FirefoxConfig{
			BrowserConfig: &browsers.BrowserConfig{
				Name:     "firefox",
				Type:     browsers.TFirefox,
				BkFile:   mozilla.PlacesFile,
				BkDir:    "testdata",
				BufferDB: &database.DB{},
				URLIndex: index.NewIndex(),
				NodeTree: &tree.Node{Name: "root", Parent: nil, Type: tree.RootNode},
				Stats:    &parsing.Stats{},
			},
		},
		tagMap: map[sqlid]*tree.Node{},
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}

func Test_addUrlNode(t *testing.T) {

	testUrl := struct {
		url   string
		id    sqlid
		title string
		desc  string
	}{
		url:   "http://test-url.gomark",
		id:    24,
		title: "test url",
		desc:  "desc of test url",
	}

	// fetch url changes into places and bookmarks
	// for each urlId/place
	// if urlNode does not exists create it
	// if urlNode exists find fetch it
	// if urlNode exists put tag node as parent to this url

	testNewUrl := "new urlNode: url is not yet in URLIndex"

	t.Run(testNewUrl, func(t *testing.T) {
		ok, urlNode := ff.addUrlNode(testUrl.url, testUrl.title, testUrl.desc)
		if !ok {
			t.Fatalf("expected %v, got %v", true, false)
		}
		if urlNode == nil {
			t.Fatal("url node was not returned", testNewUrl)
		}

		_, ok = ff.URLIndex.Get(testUrl.url)
		if !ok {
			t.Fatal("url was not added to url index")
		}

		if !utils.Inlist(ff.URLIndexList, testUrl.url) {
			t.Fatal("url was not added to url index list")
		}

	})

	testUrlExists := "return existing urlNode found in URLIndex"
	t.Run(testUrlExists, func(t *testing.T) {
		_, origNode := ff.addUrlNode(testUrl.url, testUrl.title, testUrl.desc)
		ok, urlNode := ff.addUrlNode(testUrl.url, testUrl.title, testUrl.desc)
		if ok {
			t.Fatalf("expected %v, got %v", false, true)
		}

		if urlNode == nil {
			t.Fatal("existing url node was not returned from index")
		}

		if urlNode != origNode {
			t.Fatal("existing node does not match retrieved node from url index")
		}

		_, ok = ff.URLIndex.Get(testUrl.url)
		if !ok {
			t.Fatal("url was not added to url index")
		}

		if !utils.Inlist(ff.URLIndexList, testUrl.url) {
			t.Fatal("url was not added to url index list")
		}

	})

}

func Test_addTagNode(t *testing.T) {

	testTag := struct {
		tagname string
		tagType string
		id      sqlid
	}{
		tagname: "#test_tag",
		tagType: "tag",
		id:      42,
	}

	// Should return true with the new node
	testName := "add new tag to root tree"
	t.Run(testName, func(t *testing.T) {
		ok, tagNode := ff.addTagNode(testTag.id, testTag.tagname)
		if !ok {
			t.Errorf("[%s] expected %v ,got %v", testName, true, false)
		}
		if tagNode == nil {
			t.Fatalf("[%s] tag node was not returned", testName)
		}
		if tagNode.Parent != ff.NodeTree {
			t.Errorf("[%s] wrong parent root for tag", testName)
		}
		t.Run("should be in tagMap", func(t *testing.T) {
			node, ok := ff.tagMap[testTag.id]
			if !ok {
				t.Error("tag node was not found in tagMap")
			}

			if node != tagNode {
				t.Error("tag node different from the one added to tagMap")
			}
		})

		t.Run("increment node count", func(t *testing.T) {
			if ff.CurrentNodeCount != 1 {
				t.Errorf("wrong node count")
			}
		})
	})

	// This should return false with the existing node and not add a new one
	testName = "add existing tag to root tree"
	t.Run(testName, func(t *testing.T) {
		ff.addTagNode(testTag.id, testTag.tagname)
		ok, tagNode := ff.addTagNode(testTag.id, testTag.tagname)
		if tagNode == nil {
			t.Fatalf("[%s] tag node was not returned", testName)
		}
		if tagNode.Parent != ff.NodeTree {
			t.Errorf("[%s] wrong parent root for tag", testName)
		}
		if ok {
			t.Errorf("[%s] expected %v ,got %v", testName, false, true)
		}
	})
}

func Test_fetchUrlChanges(t *testing.T) {
	t.Error("split into small units")
}

func Test_PlaceBookmarkTimeParsing(t *testing.T) {
	assert := assert.New(t)
	pb := MergedPlaceBookmark{
		BkLastModified: 1663878015759000,
	}

	res := pb.datetime().Format("2006-01-02 15:04:05.000000")
	assert.Equal(res, "2022-09-22 20:20:15.759000", "wrong time in scanned bookmark")
}

// TODO!: integration test loading firefox bookmarks
func Test_loadBookmarks(t *testing.T) {

	// expected data from testdata/places.sqlite
	data := struct {
		tags    []string
		folders []string // list of tags

		urlBookmarks []string // list of folder names

	}{ // list of urls which are bookmarked
		tags: []string{"golang", "programming", "rust"},

		folders: []string{
			"menu",
			"toolbar", "tags",
			"unfiled",
			"mobile",
			"Mozilla Firefox",
			"cooking",
			"indian",
			"GomarkMenu",
		},

		urlBookmarks: []string{
			"https://based.cooking/",
			"https://go.dev/",
			"https://support.mozilla.org/en-US/kb/customize-firefox-controls-buttons-and-toolbars?utm_source=firefox-browser&utm_medium=default-bookmarks&utm_campaign=customize",
			"https://support.mozilla.org/en-US/products/firefox",
			"https://www.mozilla.org/en-US/about/",
			"https://www.mozilla.org/en-US/contribute/",
			"https://www.mozilla.org/en-US/firefox/central/",
			"https://www.rust-lang.org/",
			"https://www.tasteofhome.com/article/indian-cooking/",
		},
	}

	// expected tags are in testdata/places.sqlite
	database.DefaultDBPath = "testdata"

	t.Log("loading firefox bookmarks")
    fmt.Sprintf("%#v", data)

	// create a Firefox{} instance

	// find the following entries in CacheDB

	// 1- find all tags defined by user


	/*
		2.find all folders
		- Should ignore Mozilla folders, any folder with (id < 13 && type == 2)
		- Should get any user defined folder with bkId > 12
	*/

	/*
	   3. find all url bookmarks with their tags
	   - should get any user added bookmark (id > 12)
	*/

	// teardown
	// remove gomarks.db
}


func Test_FindAllBookmarks(t *testing.T) {
    t.Error("should find all bookmarked urls")
}

func Test_FindAllTags(t *testing.T) {
    t.Error("should find all tags")
}

func Test_FindBookmarkTags(t *testing.T) {
    t.Error("should find the right tags for each bookmark")
}

func Test_FindBookmarkFolders(t *testing.T) {
    t.Error("should find the right bookmark folders for each bookmark")
}

func Test_FindBookmarkTagsWithFolders(t *testing.T) {
    t.Error("should find all bookmarks that have tags AND within folders")
}

func Test_FindChangedBookmarks(t *testing.T) {
    t.Error("should find all bookmarks that since last change")
}

