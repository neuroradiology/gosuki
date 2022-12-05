package firefox

import (
	"os"
	"strings"
	"testing"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/index"
	"git.sp4ke.xyz/sp4ke/gomark/mozilla"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"github.com/chenhg5/collection"
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

func runPlacesTest(name string, t *testing.T, test func(t *testing.T)) {

	bkPath, err := ff.BookmarkPath()
	if err != nil {
		t.Error(err)
	}

	ff.places, err = database.NewDB("places", bkPath, database.DBTypeFileDSN,
		FFConfig.PlacesDSN).Init()

	t.Cleanup(func() {
		err = ff.places.Close()
		if err != nil {
			t.Error(err)
		}
	})

	if err != nil {
		t.Error(err)
	}

	t.Run(name, test)
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
func Test_scanBookmarks(t *testing.T) {

	// expected data from testdata/places.sqlite
	// data := struct {
	data := struct {
		tags    []string
		folders []string // list of tags

		bookmarkTags map[string][]string // list of folder names

	}{ // list of urls which are bookmarked
		tags: []string{"golang", "programming", "rust"},

		folders: []string{
			"menu",    // Bookmarks Menu
			"toolbar", // Bookmarks Toolbar
			"tags",    // Tags Virtual Folder
			"unfiled", // Other Bookmarks
			"mobile",  // Mobile Bookmarks
			"cooking",
			"Travel",
			"indian",
			"GomarkMenu",
		},

		bookmarkTags: map[string][]string{
			"https://based.cooking/":                              {"based"},
			"https://go.dev/":                                     {"golang", "programming"},
			"https://www.rust-lang.org/":                          {"programming", "rust", "systems"},
			"https://www.tasteofhome.com/article/indian-cooking/": {},
			"http://gomark.io/":                                   {"gomark"},
			"https://www.budapestinfo.hu/":                        {"budapest"},
			"https://www.fsf.org/":                                {"libre"},
		},
	}

	t.Log("loading firefox bookmarks")

	// First make sure bookmarks are scaned then verify they are loaded
	// in CacheDB

	runPlacesTest("find", t, func(t *testing.T) {
		bookmarks, err := scanBookmarks(ff.places.Handle)
		if err != nil {
			t.Error(err)
		}

		// 1- find all tags defined by user
		t.Run("all urls", func(t *testing.T) {
			var urls []string
			for _, bk := range bookmarks {
				urls = utils.Extends(urls, bk.Url)
			}

			var testUrls []string
			for url, _ := range data.bookmarkTags {
				testUrls = append(testUrls, url)
			}
			testUrls = collection.Collect(testUrls).Unique().ToStringArray()

			assert.ElementsMatch(t, urls, testUrls)
		})

		/*
		   2.find all folders
		*/
		t.Run("all folders", func(t *testing.T) {
			var folders []string
			for _, bk := range bookmarks {
				folderS := strings.Split(bk.Folders, ",")
				for _, f := range folderS {
					folders = utils.Extends(folders, f)
				}
			}
			assert.ElementsMatch(t, folders, data.folders)
			t.Error("should find the right bookmark folders for each bookmark")
		})

		/*
		   3. find all url bookmarks with their corresponding tags
		   - should get any user added bookmark (id > 12)
		*/
		t.Run("all tags", func(t *testing.T) {
			bkTags := map[string][]string{}

			for _, bk := range bookmarks {
				bkTags[bk.Url] = collection.Collect(strings.Split(bk.Tags, ",")).
					Unique().Filter(func(item, val interface{}) bool {
					// Filter out empty ("") strings
					if v, ok := val.(string); ok {
						if v == "" {
							return false
						}
					}
					return true
				}).ToStringArray()
			}

			assert.Equal(t, data.bookmarkTags, bkTags)
			// t.Error("urls with their matching tags")
		})

		t.Error("should find all bookmarks that have tags AND within folders")
	})
}

func Test_scanFolders(t *testing.T) {

	folders := []string{
		"menu",    // Bookmarks Menu
		"toolbar", // Bookmarks Toolbar
		"tags",    // Tags Virtual Folder
		"unfiled", // Other Bookmarks
		"mobile",  // Mobile Bookmarks
        "Mozilla Firefox",
		"cooking",
		"Travel",
		"indian",
		"GomarkMenu",
	}

	runPlacesTest("scan all folders", t, func(t *testing.T) {

		// query all folders
		scannedFolders, err := scanFolders(ff.places.Handle)
		if err != nil {
			t.Error(err)
		}

        // test that we loaded all folders
        folderS := []string{}
        for _, f := range scannedFolders {
            folderS = utils.Extends(folderS, f.Title)
        }
        assert.ElementsMatch(t, folders, folderS)

	})

	// test that folders are loaded into tree
	// print tree
	// test tree
}

func Test_FindChangedBookmarks(t *testing.T) {
	t.Error("should find all bookmarks modified/added since last change")
}
