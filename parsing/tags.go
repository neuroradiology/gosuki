// Tag related parsing functions
package parsing

import (
	"regexp"
)

// ParseTags is a Hook that extracts tags like #tag from the bookmark name.
// It is stored as a tag in the bookmark metadata.
func ParseTags(node *Node) error {
	log.Debugf("running ParseTags hook on node: %s", node.Name)

    var regex = regexp.MustCompile(ReTags)

	matches := regex.FindAllStringSubmatch(node.Name, -1)
	for _, m := range matches {
		node.Tags = append(node.Tags, m[1])
	}
	//res := regex.FindAllStringSubmatch(bk.Metadata, -1)

	if len(node.Tags) > 0 {
		log.Debugf("[in title] found following tags: %s", node.Tags)
	}

	return nil
}
