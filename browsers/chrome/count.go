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

// Chrome browser module.
//
// Chrome bookmarks are stored in a json file normally called Bookmarks.
// The bookmarks file is updated atomically by chrome for each change to the
// bookmark entries by the user.
//
// Changes are detected by watching the parent directory for fsnotify.Create
// events on the bookmark file. On linux this is done by using fsnotify.
package chrome

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/buger/jsonparser"
)

var (
	urlCount atomic.Uint32
	wg       sync.WaitGroup
)

type parseFunc func([]byte, []byte, jsonparser.ValueType, int) error

func parseChildren(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
	if err != nil {
		log.Error(err)
	}

	wg.Add(1)
	go func(node []byte, dataType jsonparser.ValueType, offset int) {
		defer wg.Done()
		parse(nil, childVal, dataType, offset)
	}(childVal, dataType, offset)
}
func parse(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {

	// If node type is string ignore (needed for sync_transaction_version)
	if dataType == jsonparser.String {
		return nil
	}

	var nodeType, url, children []byte
	var childrenType jsonparser.ValueType

	// Paths to lookup in node payload
	paths := [][]string{
		{"type"},
		{"url"},
		{"children"},
	}

	jsonparser.EachKey(node, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			nodeType = value
		case 1:
			url = value
		case 2:
			children, childrenType = value, vt
		}
	}, paths...)

	// if node is url ===> leaf
	if string(nodeType) == jsonNodePaths.URL {
		fmt.Fprintf(io.Discard, "node is url >> %v", string(url))
		urlCount.Add(1)
		return nil
	}

	//log.Printf("childrenType %v\nchildrenVal %v", childrenType, len(childrenVal))

	// if node is a folder with children
	if childrenType == jsonparser.Array && len(children) > 2 { // if len(children) > len("[]")
		// log.Debugf("found children array of %d items", len(children))
		jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)
	}

	return nil
}

func preCountCountUrls(bkFile string) uint {
	urlCount.Store(0)
	data, _ := os.ReadFile(bkFile)

	rootsData, _, _, _ := jsonparser.Get(data, "roots")

	// for each node in roots, parse node
	jsonparser.ObjectEach(rootsData, parse)
	wg.Wait()
	return uint(urlCount.Load())
}
