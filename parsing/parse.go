package parsing

import (
	"regexp"
	"time"

	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/tree"
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


// ParseTags is a Hook that extracts tags like #tag from the bookmark name.
// It is stored as a tag in the bookmark metadata.
func ParseTags(node *Node) error {
	log.Debugf("running ParseTags hook on node: %s", node.Name)

    var regex = regexp.MustCompile(ReTags)

	matches := regex.FindAllStringSubmatch(node.Name, -1)
	for _, m := range matches {
		node.Tags = append(node.Tags, m[1])
	}
	//res := regex.FindAllStringSubmatch(bk.Metadata, -1)

	if len(node.Tags) > 0 {
		log.Debugf("[in title] found following tags: %s", node.Tags)
	}

	return nil
}

func S(value interface{}) string {
	return string(value.([]byte))

var Hooks = map[string]Hook{
	"tags_from_name": { 
		Name: "tags_from_name",
		Func: ParseTags,
	},
}

