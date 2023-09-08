package parsing

import (
	"time"

	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/tree"
)

type Node = tree.Node

var log = logging.GetLogger("PARSE")

const (

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


