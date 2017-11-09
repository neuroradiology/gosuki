package main

import (
	"log"
	"regexp"

	"github.com/buger/jsonparser"
)

const (
	RE_TAGS = `\B#\w+`
)

var parserStat = struct {
	lastNodeCount int
	currentCount  int
}{0, 0}

var nodeTypes = struct {
	Folder, Url string
}{"folder", "url"}

var nodePaths = struct {
	Type, Children, Url string
}{"type", "children", "url"}

type parseFunc func([]byte, []byte, jsonparser.ValueType, int) error

func parseChildren(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
	if err != nil {
		log.Panic(err)
	}

	parse(nil, childVal, dataType, offset)
}

func _s(value interface{}) string {
	return string(value.([]byte))
}

func findTagsInTitle(title []byte) {
	var regex = regexp.MustCompile(RE_TAGS)
	tags := regex.FindAll(title, -1)
	debugPrint("%s ---> found following tags: %s", title, tags)
}

func parse(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {
	parserStat.lastNodeCount++

	var nodeType, name, url, children []byte
	var childrenType jsonparser.ValueType

	// Paths to lookup in node payload
	paths := [][]string{
		[]string{"type"},
		[]string{"name"},
		[]string{"url"},
		[]string{"children"},
	}

	jsonparser.EachKey(node, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			nodeType = value
		case 1:
			name = value
		case 2:
			url = value
		case 3:
			children, childrenType = value, vt
		}
	}, paths...)

	// If node type is string ignore (needed for sync_transaction_version)
	if dataType == jsonparser.String {
		return nil
	}

	// if node is url(leaf), handle the url
	if _s(nodeType) == nodeTypes.Url {
		//debugPrint("%s", url)
		debugPrint("%s", node)
		findTagsInTitle(name)

	}

	// if node is a folder with children
	if childrenType == jsonparser.Array && len(children) > 2 { // if len(children) > len("[]")
		jsonparser.ArrayEach(node, parseChildren, nodePaths.Children)
	}

	return nil
}
