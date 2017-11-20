package main

import (
	"regexp"
)

const (
	RE_TAGS = `\B#\w+`
)

type ParserStats struct {
	lastNodeCount    int
	lastURLCount     int
	currentNodeCount int
	currentUrlCount  int
}

type ParseHook func(bk *Bookmark)

func ParseTags(bk *Bookmark) {

	var regex = regexp.MustCompile(RE_TAGS)
	tags := regex.FindAllString(bk.metadata, -1)
	if len(tags) > 0 {
		debugPrint("[Title] found following tags: %s", tags)
	}

	tags = regex.FindAllString(bk.url, -1)
	if len(tags) > 0 {
		debugPrint("[URL] found following tags: %s", tags)
	}
}

func _s(value interface{}) string {
	return string(value.([]byte))
}

func findTagsInTitle(title []byte) {
	var regex = regexp.MustCompile(RE_TAGS)
	tags := regex.FindAll(title, -1)
	debugPrint("%s ---> found following tags: %s", title, tags)
}
