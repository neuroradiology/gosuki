package main

import (
	"regexp"
)

const (
	RE_TAGS = `\B#\w+`
)

type NodeType uint8

type Node struct {
	Type     string
	Name     string
	URL      string
	Parent   *Node
	Children []*Node
}

type ParserStats struct {
	lastNodeCount    int
	lastURLCount     int
	currentNodeCount int
	currentUrlCount  int
}

type ParseHook func(bk *Bookmark)

func WalkNode(node *Node) {
	log.Debugf("Node --> %s | %s", node.Name, node.Type)

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			go WalkNode(node)
		}
	}
}

func ParseTags(bk *Bookmark) {

	var regex = regexp.MustCompile(RE_TAGS)

	bk.Tags = regex.FindAllString(bk.Metadata, -1)

	if len(bk.Tags) > 0 {
		log.Debugf("[Title] found following tags: %s", bk.Tags)
	}

	//bk.tags = regex.FindAllString(bk.url, -1)
	//if len(tags) > 0 {
	//log.Debugf("[URL] found following tags: %s", tags)
	//}
}

func _s(value interface{}) string {
	return string(value.([]byte))
}

func findTagsInTitle(title []byte) {
	var regex = regexp.MustCompile(RE_TAGS)
	tags := regex.FindAll(title, -1)
	debugPrint("%s ---> found following tags: %s", title, tags)
}
