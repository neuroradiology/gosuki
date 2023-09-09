package database

import (
	"fmt"

	"github.com/teris-io/shortid"

	"git.blob42.xyz/gomark/gosuki/pkg/tree"
)



func NewBuffer(name string) (*DB, error) {
	// add random id to buf name
	randID := shortid.MustGenerate()
	bufName := fmt.Sprintf("buffer_%s_%s", name, randID)
	// bufName := fmt.Sprintf("buffer_%s", name)
	log.Debugf("creating buffer %s", bufName)
	buffer, err := NewDB(bufName, "", DBTypeInMemoryDSN).Init()
	if err != nil {
		return nil, fmt.Errorf("could not create buffer %w", err)
	}

	err = buffer.InitSchema()
	if err != nil {
		return nil, fmt.Errorf("could initialize buffer schema %w", err)
	}

	return buffer, nil
}

func SyncURLIndexToBuffer(urls []string, index Index, buffer *DB) {
	for _, url := range urls {
		iNode, exists := index.Get(url)
		if !exists {
			log.Warningf("url does not exist in index: %s", url)
			break
		}
		node := iNode.(*Node)
		bk := node.GetBookmark()
		buffer.UpsertBookmark(bk)
	}
}

func SyncTreeToBuffer(node *Node, buffer *DB) {
	if node.Type == tree.URLNode {
		bk := node.GetBookmark()
		buffer.UpsertBookmark(bk)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			SyncTreeToBuffer(node, buffer)
		}
	}
}
