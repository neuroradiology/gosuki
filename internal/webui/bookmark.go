//
//  Copyright (c) 2024 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
//  All rights reserved.
//
//  SPDX-License-Identifier: AGPL-3.0-or-later
//
//  This file is part of GoSuki.
//
//  GoSuki is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  GoSuki is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
//

package webui

import (
	"strings"

	"github.com/blob42/gosuki"
)

type Bookmarks []*gosuki.Bookmark

type UIBookmark struct {
	*gosuki.Bookmark
	DisplayURL string
}

func NewUIBookmark(b *gosuki.Bookmark) *UIBookmark {
	return &UIBookmark{
		Bookmark:   b,
		DisplayURL: strings.Clone(b.URL),
	}
}

func (marks Bookmarks) UIBookmarks() []*UIBookmark {
	res := []*UIBookmark{}
	for _, bk := range marks {
		res = append(res, NewUIBookmark(bk))
	}

	return res
}
