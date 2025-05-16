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
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	db "github.com/blob42/gosuki/internal/database"
	"github.com/go-chi/chi/v5"

	"github.com/kr/pretty"

	"github.com/blob42/gosuki"
)

type QueryParams struct {
	Query       string
	Tag         string
	Fuzzy       bool
	NoHighlight bool
}

func DefaultQueryParams() QueryParams {
	return QueryParams{}
}

type reqIsFuzzy struct{}

type MarksContext struct {
	Bookmarks []*UIBookmark
	QueryParams
}

var (
	templates = template.Must(template.ParseFS(
		Templates,
		"templates/*.html",
		"templates/**/*.html",
	))
	views = template.Must(template.ParseFS(
		Views,
		"views/*.html",
	))
)

func isFuzzy(r *http.Request) bool {
	fuzzy := r.Context().Value(reqIsFuzzy{})

	if v, ok := fuzzy.(bool); ok && v {
		return true
	}

	return false
}

// Find and add fuzzy search parameter to the request context
func trackFuzzySearch(r *http.Request) *http.Request {
	var fuzzy bool

	query := r.URL.Query().Get("query")

	if fuzzyParam := r.URL.Query().Get("fuzzy"); fuzzyParam != "" {
		fuzzy = true
	}

	// Check if the first character of query is `~`
	if len(query) > 0 && query[0] == '~' {
		fuzzy = true
	}

	rCtx := context.WithValue(r.Context(), reqIsFuzzy{}, fuzzy)
	return r.WithContext(rCtx)
}

func fillQueryParms(r *http.Request) QueryParams {
	res := DefaultQueryParams()

	if tag := chi.URLParam(r, "tag"); tag != "" {
		res.Tag = tag
	} else if tag := r.URL.Query().Get("tag"); tag != "" {
		res.Tag = tag
	}

	if query := r.URL.Query().Get("query"); query != "" {
		res.Query = query
	}

	if hi := r.URL.Query().Get("no-hl"); hi == "on" {
		res.NoHighlight = true
	}

	if isFuzzy(r) {
		res.Fuzzy = true
	}

	return res
}

// View handler that takes a name and parses the view file with same name.
// The view name must match the URL path.
// /test -> test.html
func NamedView(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := templates.ParseFS(
			Views,
			fmt.Sprintf("views/%s.html", name),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "parsing template: %s", err)
			return
		}
		v.Execute(w, struct {
			Foo string
			MarksContext
		}{Foo: "test"})
	}
}

// WIP: refactor query -> QueryBookmarks / ListBookmarks
func getBookmarks(r *http.Request) ([]*gosuki.Bookmark, error) {
	var bookmarks []*gosuki.Bookmark
	var err error

	urlQuery := r.URL.Query()

	if tag := chi.URLParam(r, "tag"); tag != "" {
		urlQuery.Add("tag", tag)
	}

	// Handle query by tag/query
	query := urlQuery.Get("query")
	tag := urlQuery.Get("tag")

	// Check if the first character of query is `~`
	if len(query) > 0 && query[0] == '~' {
		query = query[1:] // Trim the first character
	}

	if query != "" && tag != "" {
		bookmarks, err = db.QueryBookmarksByTag(r.Context(), query, tag, isFuzzy(r))
	} else if tag != "" {
		bookmarks, err = db.BookmarksByTag(r.Context(), tag)
	} else if query != "" {
		bookmarks, err = db.QueryBookmarks(r.Context(), query, isFuzzy(r))
	} else {
		bookmarks, err = db.ListBookmarks(r.Context())
	}

	return bookmarks, err
}

func highlightQuery(r *http.Request, marks []*UIBookmark) error {
	if query := r.URL.Query().Get("query"); query != "" {
		//compile regex from query
		regex, err := regexp.Compile(`(?i)` + query)
		if err != nil {
			return errors.New("Invalid regex pattern")
		}

		// highlight match
		for _, bk := range marks {
			bk.Title = regex.ReplaceAllString(bk.Title, fmt.Sprintf(
				"<em>%s</em>",
				query,
			))
			bk.DisplayURL = regex.ReplaceAllString(bk.URL, fmt.Sprintf(
				"<em>%s</em>",
				query,
			))
			bk.Desc = regex.ReplaceAllString(bk.Desc, fmt.Sprintf(
				"<em>%s</em>",
				query,
			))
		}
	}
	return nil
}

func ListBookmarks(w http.ResponseWriter, r *http.Request) {
	var bookmarks []*gosuki.Bookmark
	var err error

	r = trackFuzzySearch(r)
	bookmarks, err = getBookmarks(r)

	if err != nil {
		http.Error(w, fmt.Sprintf(
			"fetching bookmarks: %s",
			err,
		), http.StatusInternalServerError)
		return
	}

	uiBookmarks := Bookmarks(bookmarks).UIBookmarks()
	err = highlightQuery(r, uiBookmarks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	templates.ExecuteTemplate(w, "bookmarks.html",
		MarksContext{
			Bookmarks:   uiBookmarks,
			QueryParams: fillQueryParms(r),
		})
}

func IndexView(w http.ResponseWriter, r *http.Request) {
	var bookmarks []*gosuki.Bookmark

	v, err := templates.ParseFS(
		Views,
		"views/index.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "parsing template: %s", err)
		return
	}
	r = trackFuzzySearch(r)
	bookmarks, err = getBookmarks(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "getting bookmarks: %s", err)
		return
	}
	uiBookmarks := Bookmarks(bookmarks).UIBookmarks()
	highlightQuery(r, uiBookmarks)

	v.Execute(w, MarksContext{
		Bookmarks:   uiBookmarks,
		QueryParams: fillQueryParms(r),
	})
}

func Testview(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	pretty.Println(path)
	pretty.Println(path + ".html")

	v, err := templates.ParseFS(
		Views,
		fmt.Sprintf("views/%s.html", path),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "parsing template: %s", err)
		return
	}
	v.Execute(w, struct {
		Foo string
		MarksContext
	}{Foo: "This is FOO"})
}
