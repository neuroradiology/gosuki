// Package hooks packages is the main feature of Gosuki. It permits to register
// custom hooks that will be called during the parsing process of a bookmark
// file. Hooks can be used to extract tags, commands or any custom data from a
// bookmark title or description.
//
// They can effectively be used as a command line interface to the host system
// through the browser builtin Ctrl+D bookmark feature.
//
// TODO: document types of hooks
package hooks

import (
	"git.blob42.xyz/gomark/gosuki/pkg/tree"
)

// A Hook is a function that takes a *Node as input and is called on a bookmark
// node during the parsing process. It can be used to extract tags from bookmark
// titles and descriptions. It can also be called to handle commands and
// messages found in the various fields of a bookmark.
type Hook struct {
	// Unique name of the hook
	Name string

	// Function to call on a node
	Func func(*tree.Node) error
}

// Browser who implement this interface will be able to register custom
// hooks which are called during the main Run() to handle commands and
// messages found in tags and parsed data from browsers
type HookRunner interface {

	// Calls all registered hooks on a node
	CallHooks(*tree.Node) error
}


