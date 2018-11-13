package tree

import (
	"fmt"
	"gomark/bookmarks"
	"gomark/index"
	"gomark/logging"

	"github.com/xlab/treeprint"
)

var log = logging.GetLogger("")

type Bookmark = bookmarks.Bookmark

type Node struct {
	Name       string
	Type       string // folder, tag, url
	URL        string
	Tags       []string
	Desc       string
	HasChanged bool
	NameHash   uint64 // hash of the metadata
	Parent     *Node
	Children   []*Node
}

func (node *Node) GetRoot() *Node {
	nodePtr := node

	for {
		if nodePtr.Name == "root" {
			break
		}

		nodePtr = nodePtr.Parent
	}

	return nodePtr
}

// Insert *Node in nodeList if it does not already exists
func Insert(nodeList []*Node, node *Node) []*Node {
	for _, n := range nodeList {
		if node == n {
			log.Error("Node already exists")
			return nodeList
		} else {
			nodeList = append(nodeList, node)
		}
	}
	return nodeList
}

// Returns all parent tags for URL nodes
func (node *Node) GetParentTags() []*Node {
	var parents []*Node
	var walk func(node *Node)
	var nodePtr *Node

	root := node.GetRoot()

	walk = func(n *Node) {
		nodePtr = n

		if nodePtr.Type == "url" {
			return
		}

		if len(nodePtr.Children) == 0 {
			return
		}

		for _, v := range nodePtr.Children {
			if v.URL == node.URL &&
				nodePtr.Type == "tag" {
				parents = append(parents, nodePtr)
			}
			walk(v)
		}
	}

	walk(root)
	return parents
}

func PrintTree(root *Node) {
	var walk func(node *Node, tree treeprint.Tree)
	tree := treeprint.New()

	walk = func(node *Node, t treeprint.Tree) {

		if len(node.Children) > 0 {
			t = t.AddBranch(fmt.Sprintf("%s <%s>", node.Type, node.Name))

			for _, child := range node.Children {
				go walk(child, t)
			}
		} else {
			t.AddNode(fmt.Sprintf("%s <%s>", node.Type, node.URL))
		}
	}

	walk(root, tree)
	fmt.Println(tree.String())
}

// Rebuilds the memory url index after parsing all bookmarks.
// Keeps memory index in sync with last known state of browser bookmarks
func WalkBuildIndex(node *Node, index index.HashTree) {
	if node.Type == "url" {
		index.Insert(node.URL, node)
		//log.Debugf("Inserted URL: %s and Hash: %v", node.URL, node.NameHash)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			go WalkBuildIndex(node, index)
		}

	}
}

func (node *Node) GetBookmark() *Bookmark {
	return &Bookmark{
		URL:      node.URL,
		Metadata: node.Name,
		Desc:     node.Desc,
		Tags:     node.Tags,
	}
}
