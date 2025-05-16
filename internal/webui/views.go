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
	"math"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/blob42/gosuki/internal/api"
	db "github.com/blob42/gosuki/internal/database"
	"github.com/go-chi/chi/v5"

	"github.com/kr/pretty"

	"github.com/blob42/gosuki"
)

var (
	templates *template.Template
	// templates = template.Must(template.ParseFS(
	// 	Templates,
	// 	"templates/*.html",
	// 	"templates/**/*.html",
	// ))
	views = template.Must(template.ParseFS(
		Views,
		"views/*.html",
	))
	previousQuery = ""
)

type QueryParams struct {
	Query       string
	Tag         string
	Fuzzy       bool
	NoHighlight bool
	*db.PaginationParams
}

func DefaultQueryParams() QueryParams {
	return QueryParams{PaginationParams: db.DefaultPagination()}
}

type MarksContext struct {
	Bookmarks []*UIBookmark
	Total     int // total number of results for query (excluding pagination)
	Pages     int
	QueryParams
}

// order of query param handling is important
// changing the order breaks the api
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

	if api.IsFuzzy(r) {
		res.Fuzzy = true
	}

	res.PaginationParams = api.GetPaginationParams(r)

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

func highlightQuery(r *http.Request, marks []*UIBookmark) error {
	if query := r.URL.Query().Get("query"); query != "" {
		//compile regex from query
		regex, err := regexp.Compile(`(?i)` + query)
		if err != nil {
			return errors.New("invalid regex pattern")
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

func preprocessQuery(r *http.Request) *http.Request {
	userQuery := r.URL.Query().Get("query")
	if userQuery != previousQuery {
		r = r.WithContext(context.WithValue(r.Context(), api.ResetPage{}, true))
	}
	return r
}

func ListBookmarks(w http.ResponseWriter, r *http.Request) {
	var bookmarks []*gosuki.Bookmark
	var err error
	var total uint

	r = preprocessQuery(r)

	bookmarks, total, err = api.GetBookmarks(r)

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

	// queryParams := fillQueryParms(r)
	// fmt.Printf("%#v\n", queryParams.PaginationParams)

	templates.ExecuteTemplate(w, "bookmarks.html",
		MarksContext{
			Total:       int(total),
			Bookmarks:   uiBookmarks,
			QueryParams: fillQueryParms(r),
		})
}

func IndexView(w http.ResponseWriter, r *http.Request) {
	var bookmarks []*gosuki.Bookmark
	var total uint

	v, err := templates.ParseFS(
		Views,
		"views/index.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "parsing template: %s", err)
		return
	}
	bookmarks, total, err = api.GetBookmarks(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "getting bookmarks: %s", err)
		return
	}
	uiBookmarks := Bookmarks(bookmarks).UIBookmarks()
	highlightQuery(r, uiBookmarks)

	queryParams := fillQueryParms(r)

	v.Execute(w, MarksContext{
		Total:       int(total),
		Pages:       int(math.Ceil(float64(total) / float64(queryParams.Size))),
		Bookmarks:   uiBookmarks,
		QueryParams: queryParams,
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

func init() {
	var err error
	templates, err = template.New("base.html").Funcs(map[string]any{
		"ceil": func(x float64) int {
			return int(math.Ceil(x))
		},
		"div": func(x, y int) float64 {
			return float64(x) / float64(y)
		},
		"sub": func(x, y int) int {
			return x - y
		},
		// returns a slice of integers from i to j inclusive
		"seq": func(i, j int) []int {
			if i > j {
				return []int{}
			}
			seq := make([]int, j-i+1)
			for k := range seq {
				seq[k] = i + k
			}
			return seq
		},
		"add": func(x, y int) int {
			return x + y
		},
		"int": func(x float64) int {
			return int(x)
		},
		"head": func(arr []int, to int) []int {
			return arr[:to]
		},
		"tail": func(arr []int, last int) []int {
			return arr[len(arr)-last:]
		},
	}).ParseFS(Templates,
		"templates/*.html",
		"templates/**/*.html",
	)

	if err != nil {
		panic(err)
	}

}
