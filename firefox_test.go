package main

import (
	"testing"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/index"
	"git.sp4ke.xyz/sp4ke/gomark/mozilla"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
)

// func Test_scanBookmarks(t *testing.T) {
// 	t.Run("")
// }

func Test_addUrlNode(t *testing.T) {

	ff := &FFBrowser{
		BaseBrowser: browsers.BaseBrowser{
			Name:     "firefox",
			Type:     browsers.TFirefox,
			BkFile:   mozilla.BookmarkFile,
			BaseDir:  mozilla.GetBookmarkDir(),
			NodeTree: &tree.Node{Name: "root", Parent: nil, Type: "root"},
			Stats:    &parsing.Stats{},
		},
		tagMap: make(map[sqlid]*tree.Node),
	}

	// Creates in memory Index (RB-Tree)
	ff.URLIndex = index.NewIndex()

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

	testNewUrl := "new urlNode: url does not exist in URLIndex"

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

	ff := &FFBrowser{
		BaseBrowser: browsers.BaseBrowser{
			Name:     "firefox",
			Type:     browsers.TFirefox,
			BkFile:   mozilla.BookmarkFile,
			BaseDir:  mozilla.GetBookmarkDir(),
			NodeTree: &tree.Node{Name: "root", Parent: nil, Type: "root"},
			Stats:    &parsing.Stats{},
		},
		tagMap: make(map[sqlid]*tree.Node),
	}

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
			if ff.Stats.CurrentNodeCount != 1 {
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
