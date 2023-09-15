package database

import (
	"slices"
	"strings"

	"git.blob42.xyz/gomark/gosuki/internal/utils"
)

type Tags struct {
	delim string
	tags []string
}

// Reads tags from a slice of strings
func NewTags(tags []string, delim string) *Tags {
 	return &Tags{delim: delim, tags: tags}
}

func (t *Tags) Add(tag string) {
	t.tags = append(t.tags, tag)
}

func (t *Tags) Extend(tags []string) *Tags {
	t.tags = utils.Extends(t.tags, tags...)
	return t
}

// Sanitize the list of tags before saving them to the DB
func (t *Tags) PreSanitize() *Tags {
    t.tags = utils.ReplaceInList(t.tags, TagSep, "--")
	return t
}

// String representation of the tags.
// It can wrap the tags with the delim if wrap is true. This is done for
// compatibility with Buku DB format.
func (t Tags) String(wrap bool) string {
	if wrap {
		return delimWrap(strings.Join(t.tags, t.delim), t.delim)
	}
	return strings.Join(t.tags, t.delim)
}

// String representation of the tags. It wraps the tags with the delim.
func (t Tags) StringWrap() string {
	return delimWrap(strings.Join(t.tags, t.delim), t.delim)
}

// Builds a list of tags from a string as a Tags struct.
// It also removes empty tags
func TagsFromString(s, delim string) *Tags {
	tagslice := strings.Split(s, delim)
	tags := slices.DeleteFunc(tagslice, func (s string) bool {
		return s == ""
	})
	return &Tags{delim: delim, tags: tags}
}

/// Returns a string wrapped with the delim
func delimWrap(token string, delim string) string {
	if token == "" || strings.TrimSpace(token) == "" {
		return delim
	}

	if token[0] != delim[0] {
		token = delim + token
	}

	if token[len(token)-1] != delim[0] {
		token = token + delim
	}

	return token
}

