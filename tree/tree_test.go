package tree

import(
    "testing"
	"github.com/stretchr/testify/assert"
)

func Test_AddChild(t *testing.T) {
	rootNode := &Node{Name: "root", Parent: nil, Type:RootNode}
	childNode := &Node{Name: "child", Type: URLNode}
	childNode2 := &Node{Type: FolderNode, Name: "second child"}
	childNode3 := &Node{Type: TagNode, Name: "third child"}

	AddChild(rootNode, childNode)
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
		AddChild(rootNode, childNode2)
		AddChild(rootNode, childNode3)

		if len(rootNode.Children) < 3 {
			t.Error("child does not see brothers")
		}

		if len(rootNode.Children) > 3 {
			t.Errorf("child sees too many brothers, expected %v, got %v", 3, len(rootNode.Children))
		}
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


