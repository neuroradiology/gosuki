package parsing

// A Hook is a function that takes a *Node as input and is called on a bookmark
// node during the parsing process. It can be used to extract tags from bookmark
// titles and descriptions. It can also be called to handle commands and
// messages found in the various fields of a bookmark.
type Hook struct {
	// Unique name of the hook
	Name string

	// Function to call on a node
	Func func(*Node) error
}

// Browser who implement this interface will be able to register custom
// hooks which are called during the main Run() to handle commands and
// messages found in tags and parsed data from browsers
type HookRunner interface {

	// Calls all registered hooks on a node
	CallHooks(*Node) error
}

