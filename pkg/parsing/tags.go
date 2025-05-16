//
// Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

// Tag related parsing functions
package parsing

import (
	"fmt"
	"regexp"

	"github.com/blob42/gosuki"
	"github.com/blob42/gosuki/pkg/logging"
	"github.com/blob42/gosuki/pkg/tree"
)

const (
	// First group is tag
	// TODO: use named groups
	// [named groups](https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter2.markdown)

	// Regex matching tests:

	//#start test2 #test3 elol
	//#start word with #end
	//word in the #middle of sentence
	//tags with a #dot.caracter
	//this is a end of sentence #tag
	ReTags = `\B#(?P<tag>\w+\.?\w+)`

	// #tag:notify
	ReNotify = `\b(?P<tag>\w+\.?\w+):notify`
)

var log = logging.GetLogger("PARSE")

// ParseTags is a [gosuki.Hook] that extracts tags like #tag from the title of the bookmark or node.
// It is stored as a tag in the metadata field of the bookmark or node.
func parseTags(item any) error {
	var regex = regexp.MustCompile(ReTags)
	switch v := item.(type) {
	case *tree.Node:
		// log.Debugf("running ParseTags hook on node: %s", v.URL)
		matches := regex.FindAllStringSubmatch(v.Title, -1)
		for _, m := range matches {
			if v.Tags == nil {
				v.Tags = []string{m[1]}
			} else {
				v.Tags = append(v.Tags, m[1])
			}
		}
		if len(v.Tags) > 0 {
			log.Debugf("[hook] found following tags: %s", v.Tags)
		}
	case *gosuki.Bookmark:
		// log.Debugf("running ParseTags hook on node: %s", v.URL)
		matches := regex.FindAllStringSubmatch(v.Title, -1)
		for _, m := range matches {
			if v.Tags == nil {
				v.Tags = []string{m[1]}
			} else {
				v.Tags = append(v.Tags, m[1])
			}
		}
		if len(v.Tags) > 0 {
			log.Debugf("[hook] found following tags: %s", v.Tags)
		}
	default:
		return fmt.Errorf("unsupported type")
	}
	return nil
}

func ParseNodeTags(n *tree.Node) error {
	return parseTags(n)
}

func ParseBkTags(b *gosuki.Bookmark) error {
	return parseTags(b)
}
