package tree

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_AddChild(t *testing.T) {
	rootNode := &Node{Name: "root", Parent: nil, Type:RootNode}
	childNode := &Node{Name: "url child", Type: URLNode}
    urlNode := &Node{Name: "url", Type: URLNode}
	folderNode := &Node{Type: FolderNode, Name: "folder child"}
    subFolderNode := &Node{Type: FolderNode, Name: "sub-folder"}
	tagNode := &Node{Type: TagNode, Name: "tag child"}

    AddChild(folderNode, subFolderNode)
    AddChild(subFolderNode, urlNode)
	AddChild(rootNode, childNode)

    t.Run("skip duplicate children", func(t *testing.T){
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

    t.Run("nested url node", func(t *testing.T){
        assert.Equal(t, rootNode, urlNode.Parent.Parent.Parent)
    })
}

func TestFindNode(t *testing.T){
	rootNode := &Node{Name: "root", Parent: nil, Type:RootNode}
	childNode := &Node{Name: "child", Type: URLNode}
	childNode2 := &Node{Type: FolderNode, Name: "second child"}
	childNode3 := &Node{Type: TagNode, Name: "third child"}

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

    t.Run("find nodes by name", func(t *testing.T){
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
	rootNode := &Node{Name: "root", Parent: nil, Type:RootNode}

	urlNode := &Node{Name: "child", Type: URLNode}

	tagNode1 := &Node{Type: TagNode, Name: "tag1"}
	tagNode2 := &Node{Type: TagNode, Name: "tag2"}
    folderNode := &Node{Type: FolderNode, Name: "folder1"}

    AddChild(rootNode, tagNode1)
    AddChild(rootNode, tagNode2)
    AddChild(rootNode, folderNode)

    AddChild(folderNode, urlNode)

    AddChild(tagNode1, urlNode)
    AddChild(tagNode2, urlNode)
    
    tags := urlNode.getTags()
    assert.ElementsMatch(t, tags, []string{"tag1", "tag2", "folder1"}, "node tags mismatch")
}
