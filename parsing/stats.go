package parsing

import (
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
	CurrentURLCount       int
}

func (s *Stats) Reset() {
    s.LastURLCount = s.CurrentURLCount
	s.LastNodeCount = s.CurrentNodeCount
	s.CurrentNodeCount = 0
	s.CurrentURLCount = 0
}


