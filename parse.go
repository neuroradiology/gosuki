package main

import (
	"regexp"
	"time"
)

const (
	// First group is tag
	// TODO: use named groups
	// [named groups](https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter2.markdown)

	// Regex matching tests:

	//#start test2 #test3 elol
	//#start word with #end
	//word in the #middle of sentence
	//tags with a #dot.caracter
	//this is a end of sentence #tag

	ReTags     = "\\B#(?P<tag>\\w+\\.?\\w+)"
	TagJoinSep = "|"
)

type ParserStats struct {
	lastParseTime    time.Duration
	lastNodeCount    int
	lastURLCount     int
	currentNodeCount int
	currentUrlCount  int
}

type ParseHook func(node *Node)

func ParseTags(node *Node) {

	var regex = regexp.MustCompile(ReTags)

	matches := regex.FindAllStringSubmatch(node.Name, -1)
	for _, m := range matches {
		node.Tags = append(node.Tags, m[1])
	}
	//res := regex.FindAllStringSubmatch(bk.Metadata, -1)

	if len(node.Tags) > 0 {
		log.Debugf("[Title] found following tags: %s", node.Tags)
	}
}

func _s(value interface{}) string {
	return string(value.([]byte))
}
