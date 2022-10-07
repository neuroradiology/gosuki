package tree

import "testing"

func Test_AddChild(t *testing.T) {
	rootNode := &Node{Name: "root", Parent: nil, Type: "root"}
	childNode := &Node{Name: "child", Type: "testing"}
	childNode2 := &Node{Type: "testing", Name: "second child"}
	childNode3 := &Node{Type: "testing", Name: "third child"}

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
