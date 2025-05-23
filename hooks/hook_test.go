package hooks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blob42/gosuki"
	"github.com/blob42/gosuki/pkg/tree"
	"github.com/stretchr/testify/assert"
)

var Hooks = []NamedHook{
	Hook[*tree.Node]{
		name: "PreprocessNode",
		Func: func(n *tree.Node) error {
			if n.Title == "" {
				return fmt.Errorf("node title cannot be empty")
			}
			return nil
		},
		priority: 10,
	},
	Hook[*tree.Node]{
		name: "UpdateNodeMetadata",
		Func: func(n *tree.Node) error {
			// Example hash calculation (simplified)
			n.NameHash = uint64(len(n.Title + n.Desc))
			return nil
		},
		priority: 5,
	},
	Hook[*tree.Node]{
		name: "ValidateNodeTags",
		Func: func(n *tree.Node) error {
			seen := make(map[string]bool)
			for _, tag := range n.Tags {
				if seen[tag] {
					return fmt.Errorf("duplicate tag: %s", tag)
				}
				seen[tag] = true
			}
			return nil
		},
		priority: 8,
	},
	Hook[*tree.Node]{
		name: "LogNodeCreation",
		Func: func(n *tree.Node) error {
			fmt.Printf("Node created: %s\n", n.Title)
			return nil
		},
		priority: 15,
	},
	Hook[*tree.Node]{
		name: "CheckNodeParent",
		Func: func(n *tree.Node) error {
			if n.Parent == nil && n.Children != nil {
				return fmt.Errorf("parent must be set for nodes with children")
			}
			return nil
		},
		priority: 7,
	},
	Hook[*gosuki.Bookmark]{
		name: "ValidateBookmarkURL",
		Func: func(b *gosuki.Bookmark) error {
			if !strings.HasPrefix(b.URL, "http") {
				return fmt.Errorf("invalid URL format")
			}
			return nil
		},
		priority: 2,
	},
	Hook[*gosuki.Bookmark]{
		name: "EnforceBookmarkTags",
		Func: func(b *gosuki.Bookmark) error {
			if len(b.Tags) == 0 {
				return fmt.Errorf("bookmark must have at least one tag")
			}
			return nil
		},
		priority: 3,
	},
	Hook[*gosuki.Bookmark]{
		name: "CheckBookmarkDesc",
		Func: func(b *gosuki.Bookmark) error {
			if b.Desc == "" {
				return fmt.Errorf("bookmark description cannot be empty")
			}
			return nil
		},
		priority: 4,
	},
	Hook[*gosuki.Bookmark]{
		name: "SanitizeBookmarkTitle",
		Func: func(b *gosuki.Bookmark) error {
			b.Title = strings.ReplaceAll(b.Title, "<", "")
			b.Title = strings.ReplaceAll(b.Title, ">", "")
			return nil
		},
		priority: 6,
	},
	Hook[*gosuki.Bookmark]{
		name: "SetDefaultModule",
		Func: func(b *gosuki.Bookmark) error {
			if b.Module == "" {
				b.Module = "default"
			}
			return nil
		},
		priority: 1,
	},
}

func TestHookPriority(t *testing.T) {
	assert.Equal(t, "PreprocessNode", Hooks[0].Name())

	SortByPriority(Hooks)
	assert.Equal(t, "SetDefaultModule", Hooks[0].Name())
}
