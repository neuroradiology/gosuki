package hooks
// Global available hooks for browsers to use

import "git.blob42.xyz/gomark/gosuki/pkg/parsing"

var Predefined = map[string]Hook{
	"tags_from_name": { 
		Name: "tags_from_name",
		Func: parsing.ParseTags,
	},
}



func regHook(hook Hook) {
	Predefined[hook.Name] = hook
}
