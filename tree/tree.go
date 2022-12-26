package tree

import (
	"fmt"

	"git.sp4ke.xyz/sp4ke/gomark/bookmarks"
	"git.sp4ke.xyz/sp4ke/gomark/index"
	"git.sp4ke.xyz/sp4ke/gomark/logging"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
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
    if len(parent.Children) == 0 { return false }
    var found bool
    for _, child := range parent.Children {
        if node == child { found = true }
    }

    return found
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


// Inserts child node into parent node. Parent will point to child
// and child will point to parent EXCEPT when parent is a TAG node.
// If parent is a Tag node, child should not point back to parent
// as URL nodes should always point to folder parent nodes only.
func AddChild(parent *Node, child *Node) {
	log.Debugf("adding child %v: <%s>", child.Type, child.Name)


	if len(parent.Children) == 0 {
		parent.Children = []*Node{child}

        // Do not point back to TAG parent node from child
        if parent.Type != TagNode {
            child.Parent = parent
        }
		return
	}

	for _, n := range parent.Children {
		if child == n {
			// log.Errorf("<%s> Node already exists", child)
            log.Info(pretty.Sprintf("skipping node <%s>, already exists", child.Name))
			return
		}
	}

	parent.Children = append(parent.Children, child)
    if parent.Type != TagNode {
        child.Parent = parent
    }
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

// Get all possible tags for this url node
// The tags make sense only in the context of a URL node
// This will traverse the three breadth first to find all Parent folders and 
// add them as a tag. URL nodes should already be populated with the list of 
// tags that exist under the TAG tree. So we only need to find the parent folders
// and turn them into tags.
func (node *Node) getTags() []string {

    if node.Parent.Type ==  RootNode {
        return []string{}
    }

    if node.Parent.Type == FolderNode {
        node.Tags = utils.Extends(node.Tags, node.Parent.Name)
        return append(node.Parent.getTags(), node.Tags...)
    }

    return node.Tags
}

func (node *Node) GetBookmark() *Bookmark {

    if node.Type != URLNode {
        return nil
    }

	return &Bookmark{
		URL:      node.URL,
		Metadata: node.Name,
		Desc:     node.Desc,
		Tags:     node.getTags(),
	}
}
