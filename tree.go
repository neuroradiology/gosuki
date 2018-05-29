package main

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

// Rebuilds the memory url index after parsing all bookmarks.
// Keeps memory index in sync with last known state of browser bookmarks
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

func syncTreeToBuffer(node *Node, buffer *DB) {

	if node.Type == "url" {
		bk := node.GetBookmark()
		bk.InsertOrUpdateInDB(buffer)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			syncTreeToBuffer(node, buffer)
		}
	}
}
