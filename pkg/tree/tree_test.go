package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AddChild(t *testing.T) {
	rootNode := &Node{Title: "root", Parent: nil, Type: RootNode}
	childNode := &Node{Title: "url child", Type: URLNode}
	urlNode := &Node{Title: "url", Type: URLNode}
	folderNode := &Node{Type: FolderNode, Title: "folder child"}
	subFolderNode := &Node{Type: FolderNode, Title: "sub-folder"}
	tagNode := &Node{Type: TagNode, Title: "tag child"}

	AddChild(folderNode, subFolderNode)
	AddChild(subFolderNode, urlNode)
	AddChild(rootNode, childNode)

	t.Run("skip duplicate children", func(t *testing.T) {
		AddChild(rootNode, childNode)
		assert.Equal(t, 1, len(rootNode.Children))

	})

	t.Run("[first child] parent has the child", func(t *testing.T) {
		found := false

		for _, child := range rootNode.Children {
			t.Log(child)
			if child == childNode {
				found = true
			}
		}

		if !found {
			t.Errorf("child not found")
		}
	})

	t.Run("[first child] child sees the parent", func(t *testing.T) {
		if childNode.Parent != rootNode {
			t.Error("child does not see the parent")
		}
	})

	t.Run("[new child] child sees brothers", func(t *testing.T) {
		AddChild(rootNode, folderNode)
		AddChild(rootNode, tagNode)

		if len(rootNode.Children) < 3 {
			t.Error("child does not see brothers")
		}

		if len(rootNode.Children) > 3 {
			t.Errorf("child sees too many brothers, expected %v, got %v", 3, len(rootNode.Children))
		}
	})

	t.Run("nested url node", func(t *testing.T) {
		assert.Equal(t, rootNode, urlNode.Parent.Parent.Parent)
	})

	parents := urlNode.GetFolderParents()
	var parentFolderNames []string
	for _, p := range parents {
		parentFolderNames = append(parentFolderNames, p.Title)
	}

	assert.ElementsMatch(t, parentFolderNames, []string{"folder child", "sub-folder"})

}

func TestFindNode(t *testing.T) {
	rootNode := &Node{Title: "root", Parent: nil, Type: RootNode}
	childNode := &Node{Title: "child", Type: URLNode}
	childNode2 := &Node{Type: FolderNode, Title: "second child"}
	childNode3 := &Node{Type: TagNode, Title: "third child"}

	AddChild(rootNode, childNode)
	AddChild(rootNode, childNode2)
	AddChild(childNode2, childNode3)

	result := FindNode(childNode3, rootNode)
	assert.True(t, result)

	result = FindNode(childNode2, rootNode)
	assert.True(t, result)

	result = FindNode(childNode, rootNode)
	assert.True(t, result)

	result = FindNode(rootNode, rootNode)
	assert.True(t, result)

	t.Run("find nodes by name", func(t *testing.T) {
		result := FindNodeByName("third child", rootNode)
		assert.True(t, result)

		result = FindNodeByName("second child", rootNode)
		assert.True(t, result)

		result = FindNodeByName("child", rootNode)
		assert.True(t, result)

		result = FindNodeByName("root", rootNode)
		assert.True(t, result)

		assert.False(t, FindNodeByName("not existing", rootNode))
	})
}

// implicitly tests getParentTags and getParentFolders
func TestGetTags(t *testing.T) {
	rootNode := &Node{Title: "root", Parent: nil, Type: RootNode}

	urlNode := &Node{Title: "child", Type: URLNode}

	tagNode1 := &Node{Type: TagNode, Title: "tag1"}
	tagNode2 := &Node{Type: TagNode, Title: "tag2"}
	folderNode := &Node{Type: FolderNode, Title: "folder1"}

	// camel case
	cmFolderNode := &Node{Type: FolderNode, Title: "SomeFolder"}

	// space in folder name
	spFolderNode := &Node{Type: FolderNode, Title: "Folder With Space"}

	AddChild(rootNode, tagNode1)
	AddChild(rootNode, tagNode2)
	AddChild(rootNode, folderNode)
	AddChild(rootNode, cmFolderNode)
	AddChild(rootNode, spFolderNode)

	AddChild(folderNode, urlNode)
	AddChild(cmFolderNode, urlNode)
	AddChild(spFolderNode, urlNode)

	AddChild(tagNode1, urlNode)
	AddChild(tagNode2, urlNode)

	// DEBUG:
	// PrintTree(rootNode)

	tags := urlNode.getTags()
	assert.ElementsMatch(t, []string{"tag1", "tag2", "folder1", "SomeFolder", "Folder With Space"}, tags, "node tags mismatch")
}

func Test_GetRoot(t *testing.T) {
	root := &Node{Type: RootNode, Title: "root"}
	fold := &Node{Type: FolderNode, Title: "folder"}
	AddChild(root, fold)
	subfold := &Node{Type: FolderNode, Title: "sub-folder"}
	AddChild(fold, subfold)
	url := &Node{Type: URLNode, Title: "url"}
	AddChild(subfold, url)

	foundRoot := url.GetRoot()
	assert.Equal(t, root, foundRoot)
}
