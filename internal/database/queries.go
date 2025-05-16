// Copyright (c) 2024 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with gosuki.  If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/blob42/gosuki"
)

type RawBookmarks []*RawBookmark

type RawBookmark struct {
	ID  uint64
	URL string `db:"URL"`
	// Usually used for the bookmark title
	Metadata string
	Tags     string
	Desc     string
	// Last modified
	Modified uint64
	// Not used, keep for buku compatibility
	Flags  int
	Module string
}

const (
	QQueryBookmarks      = `URL like '%%%s%%' OR metadata like '%%%s%%' OR tags like '%%%s%%'`
	QQueryBookmarksFuzzy = `fuzzy('%s', URL) OR fuzzy('%s', metadata) OR fuzzy('%s', tags)`

	QQueryBookmarksByTag      = `(URL LIKE '%%%s%%' OR metadata LIKE '%%%s%%') AND tags LIKE '%%%s%%'`
	QQueryBookmarksByTagFuzzy = `(fuzzy('%s', URL) OR fuzzy('%s', metadata)) AND tags LIKE '%%%s%%'`
)

func (raws RawBookmarks) AsBookmarks() []*gosuki.Bookmark {
	res := []*Bookmark{}
	for _, raw := range raws {
		tags := TagsFromString(raw.Tags, TagSep)
		res = append(res, &Bookmark{
			URL:   raw.URL,
			Title: raw.Metadata,
			Tags:  tags.Get(),
			Desc:  raw.Desc,
		})
	}

	return res
}

func QueryBookmarksByTag(
	ctx context.Context,
	query,
	tag string,
	fuzzy bool,
) ([]*Bookmark, error) {
	query = strings.TrimSpace(query)
	tag = strings.TrimSpace(tag)

	if tag == "" || query == "" {
		return nil, errors.New("cannot use empty query or tags")
	}

	sqlPrelude := `SELECT URL, METADATA, tags FROM bookmarks WHERE `
	sqlQuery := sqlPrelude + QQueryBookmarksByTag
	if fuzzy {
		sqlQuery = sqlPrelude + QQueryBookmarksByTagFuzzy
	}

	var rawBooks RawBookmarks
	err := DiskDB.Handle.SelectContext(ctx, &rawBooks, fmt.Sprintf(sqlQuery, query, query, tag))
	if err != nil {
		return nil, err
	}

	return rawBooks.AsBookmarks(), nil
}

func QueryBookmarks(ctx context.Context, query string, fuzzy bool) ([]*Bookmark, error) {

	sqlPrelude := `SELECT URL, METADATA, tags FROM bookmarks WHERE `

	sqlQuery := sqlPrelude + QQueryBookmarks
	if fuzzy {
		sqlQuery = sqlPrelude + QQueryBookmarksFuzzy
	}

	rawBooks := RawBookmarks{}
	err := DiskDB.Handle.SelectContext(ctx, &rawBooks,
		fmt.Sprintf(sqlQuery, query, query, query))
	if err != nil {
		return nil, err
	}

	return rawBooks.AsBookmarks(), nil
}

func BookmarksByTag(ctx context.Context, tag string) ([]*Bookmark, error) {
	query := "SELECT * FROM bookmarks WHERE"
	tagsCondition := ""
	if len(tag) > 0 {
		tagsCondition = fmt.Sprintf(" tags LIKE '%%%s%%'", tag)
	} else {
		return nil, errors.New("empty tag provided")
	}

	query = query + " (" + tagsCondition + ")"

	rawBooks := RawBookmarks{}
	err := DiskDB.Handle.SelectContext(ctx, &rawBooks, query)
	if err != nil {
		return nil, err
	}

	return rawBooks.AsBookmarks(), nil
}

func ListBookmarks(ctx context.Context) ([]*Bookmark, error) {
	rawBooks := RawBookmarks{}
	err := DiskDB.Handle.SelectContext(
		ctx,
		&rawBooks,
		"SELECT * FROM bookmarks",
	)
	if err != nil {
		return nil, err
	}

	return rawBooks.AsBookmarks(), nil
}
