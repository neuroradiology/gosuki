package parsing

import (
	"regexp"
	"time"

	"git.sp4ke.xyz/sp4ke/gomark/logging"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
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
	LastFullTreeParseTime time.Duration
	LastWatchRunTime      time.Duration
	LastNodeCount         int
	LastURLCount          int
	CurrentNodeCount      int
	CurrentUrlCount       int
}

func (s *Stats) Reset() {
    s.LastURLCount = s.CurrentUrlCount
	s.LastNodeCount = s.CurrentNodeCount
	s.CurrentNodeCount = 0
	s.CurrentUrlCount = 0
}

type Hook func(node *Node)

// Browser.Run hook function that extracts
// tags from url titles and descriptions
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
