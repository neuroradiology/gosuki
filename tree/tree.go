package tree

import (
	"fmt"

	"git.sp4ke.xyz/sp4ke/gomark/bookmarks"
	"git.sp4ke.xyz/sp4ke/gomark/index"
	"git.sp4ke.xyz/sp4ke/gomark/logging"
	"github.com/kr/pretty"

	"github.com/xlab/treeprint"
)

var log = logging.GetLogger("TREE")

type Bookmark = bookmarks.Bookmark

type NodeType int

const (
	RootNode NodeType = iota
	URLNode
	FolderNode
	TagNode
)

type Node struct {
	Name       string
	Type       NodeType // folder, tag, url
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

	for nodePtr.Name != "root" {
		nodePtr = nodePtr.Parent
	}

	return nodePtr
}

// Returns the ancestor of this node
func Ancestor(node *Node) *Node {
	if node.Parent == nil {
		return node
	} else {
		return Ancestor(node.Parent)
	}
}

func (node *Node) DirectChildOf(parent *Node) bool {
    return node.Parent == parent
}


// Finds a node and the tree starting at root
func FindNode(node *Node, root *Node) bool {

	if node == root {
		return true
	} else {
        for _, child := range root.Children {
            found := FindNode(node, child)
            if found { return true }
        }
    }

    return false
}

func FindNodeByName(name string, root *Node) bool {

	if name == root.Name {
		return true
	} else {
        for _, child := range root.Children {
            found := FindNodeByName(name, child)
            if found { return true }
        }
    }

    return false
}

// Insert *Node in nodeList if it does not already exists
func AddChild(parent *Node, child *Node) {
	log.Debugf("adding child %v: <%s>", child.Type, child.Name)

	if len(parent.Children) == 0 {
		parent.Children = []*Node{child}
		child.Parent = parent
		return
	}

	for _, n := range parent.Children {
		if child == n {
			// log.Errorf("<%s> Node already exists", child)
            log.Info(pretty.Sprintf("Node <%#v> already exists", child))
			return
		}
	}

	parent.Children = append(parent.Children, child)
	child.Parent = parent
}

// Returns all parent tags for URL nodes
func (node *Node) GetParentTags() []*Node {
	var parents []*Node
	var walk func(node *Node)
	var nodePtr *Node

	root := node.GetRoot()

	walk = func(n *Node) {
		nodePtr = n

		if nodePtr.Type == URLNode {
			return
		}

		if len(nodePtr.Children) == 0 {
			return
		}

		for _, v := range nodePtr.Children {
			if v.URL == node.URL &&
				nodePtr.Type == TagNode {
				parents = append(parents, nodePtr)
			}
			walk(v)
		}
	}

	walk(root)
	return parents
}

func PrintTree(root *Node) {
	fmt.Println("---")
	fmt.Println("PrintTree")
	var walk func(node *Node, tree treeprint.Tree)
	tree := treeprint.New()

	walk = func(node *Node, t treeprint.Tree) {

		if len(node.Children) > 0 {
			t = t.AddBranch(fmt.Sprintf("%#v <%s>", node.Type, node.Name))

			for _, child := range node.Children {
				walk(child, t)
			}
		} else {
			t.AddNode(fmt.Sprintf("%#v <%s>", node.Type, node.Name))
		}
	}

	walk(root, tree)
	fmt.Println(tree.String())
	fmt.Println("---")
}

// Rebuilds the memory url index after parsing all bookmarks.
// Keeps memory index in sync with last known state of browser bookmarks
func WalkBuildIndex(node *Node, index index.HashTree) {
	if node.Type == URLNode {
		index.Insert(node.URL, node)
		//log.Debugf("Inserted URL: %s and Hash: %v", node.URL, node.NameHash)
	}

	if len(node.Children) > 0 {
		for _, node := range node.Children {
			WalkBuildIndex(node, index)
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
