package main

import (
	"regexp"
)

const (
	// First group is tag
	// TODO: use named groups
	// [named groups](https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter2.markdown)

	ReTags     = "\\B#(?P<tag>\\w+)"
	TagJoinSep = "|"
)

type Node struct {
	Name     string
	Type     string
	URL      string
	Tags     []string
	NameHash uint64 // hash of the metadata
	Parent   *Node
	Children []*Node
}

type ParserStats struct {
	lastNodeCount    int
	lastURLCount     int
	currentNodeCount int
	currentUrlCount  int
}

type ParseHook func(node *Node)

// Debuggin bookmark node tree
// TODO: Better usage of node trees
func WalkNode(node *Node) {
	log.Debugf("Node --> %s | %s | children: %d | parent: %v", node.Name, node.Type, len(node.Children), node.Parent)

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			go WalkNode(node)
		}
	}
}

func WalkBuildIndex(node *Node, b *BaseBrowser) {

	if node.Type == "url" {
		b.URLIndex.Insert(node.URL, node)
		//log.Debugf("Inserted URL: %s and Hash: %v", node.URL, node.NameHash)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			go WalkBuildIndex(node, b)
		}

	}
}

func ParseTags(node *Node) {

	var regex = regexp.MustCompile(ReTags)

	matches := regex.FindAllStringSubmatch(node.Name, -1)
	for _, m := range matches {
		node.Tags = append(node.Tags, _s(m[1]))
	}
	//res := regex.FindAllStringSubmatch(bk.Metadata, -1)

	if len(node.Tags) > 0 {
		log.Debugf("[Title] found following tags: %s", node.Tags)
	}

}

func _s(value interface{}) string {
	return string(value.([]byte))
}
