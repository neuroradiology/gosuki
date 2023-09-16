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
	"regexp"
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
	ReTags = "\\B#(?P<tag>\\w+\\.?\\w+)"

	// #tag:notify
	ReNotify = "\\B#(?P<tag>\\w+\\.?\\w+):notify"
)

// ParseTags is a Hook that extracts tags like #tag from the bookmark name.
// It is stored as a tag in the bookmark metadata.
func ParseTags(node *Node) error {
	log.Debugf("running ParseTags hook on node: %s", node.Name)

	var regex = regexp.MustCompile(ReTags)

	matches := regex.FindAllStringSubmatch(node.Name, -1)
	for _, m := range matches {
		if node.Tags == nil {
			node.Tags = []string{m[1]}
		} else {
			node.Tags = append(node.Tags, m[1])
		}
	}
	//res := regex.FindAllStringSubmatch(bk.Metadata, -1)

	if len(node.Tags) > 0 {
		log.Debugf("[in title] found following tags: %s", node.Tags)
	}

	return nil
}
