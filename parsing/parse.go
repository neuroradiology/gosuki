package parsing

import (
	"gomark/logging"
	"gomark/tree"
	"regexp"
	"time"
)

type Node = tree.Node

var log = logging.GetLogger("PARSE")

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

	ReTags = "\\B#(?P<tag>\\w+\\.?\\w+)"
)

type Stats struct {
	LastParseTime    time.Duration
	LastNodeCount    int
	LastURLCount     int
	CurrentNodeCount int
	CurrentUrlCount  int
}

type Hook func(node *Node)

func ParseTags(node *Node) {

	var regex = regexp.MustCompile(ReTags)

	matches := regex.FindAllStringSubmatch(node.Name, -1)
	for _, m := range matches {
		node.Tags = append(node.Tags, m[1])
	}
	//res := regex.FindAllStringSubmatch(bk.Metadata, -1)

	if len(node.Tags) > 0 {
		log.Debugf("[in title] found following tags: %s", node.Tags)
	}
}

func S(value interface{}) string {
	return string(value.([]byte))
}
