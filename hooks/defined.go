package hooks
// Global available hooks for browsers to use

import "git.blob42.xyz/gomark/gosuki/parsing"

var Predefined = map[string]Hook{
	"tags_from_name": { 
		Name: "tags_from_name",
		Func: parsing.ParseTags,
	},
}

