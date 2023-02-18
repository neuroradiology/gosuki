package database

import (
	"fmt"

	"git.blob42.xyz/gomark/gosuki/tree"
)

func NewBuffer(name string) (*DB, error) {
	bufferName := fmt.Sprintf("buffer_%s", name)
	buffer, err := NewDB(bufferName, "", DBTypeInMemoryDSN).Init()
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
		buffer.InsertOrUpdateBookmark(bk)
	}
}

func SyncTreeToBuffer(node *Node, buffer *DB) {
	if node.Type == tree.URLNode {
		bk := node.GetBookmark()
		buffer.InsertOrUpdateBookmark(bk)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			SyncTreeToBuffer(node, buffer)
		}
	}
}
