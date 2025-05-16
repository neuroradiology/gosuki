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

package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/blob42/gosuki/internal/api"
	webui "github.com/blob42/gosuki/internal/webui"
	"github.com/blob42/gosuki/pkg/manager"
)

const (
	BindAddr = "0.0.0.0:2025"
)

type WebUIServer struct {
	http.Handler
}

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}

func (s *WebUIServer) Run(m manager.UnitManager) {
	server := &http.Server{
		Addr:         BindAddr,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      s.Handler,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				m.Panic(err)
			}
		}
	}()

	// Wait for stop signal
	<-m.ShouldStop()
	m.Done()
}

func NewWebUIServer(tuiMode bool) *WebUIServer {

	router := chi.NewRouter()
	if !tuiMode {
		router.Use(middleware.Logger)
	}
	router.Use(middleware.Recoverer)

	apiRoute := chi.NewRouter()
	apiRoute.Get("/bookmarks", api.GetAPIBookmarks)

	router.Mount("/api", apiRoute)

	router.Get("/greet", greet)
	router.Get("/bookmarks", webui.ListBookmarks)
	router.Get("/bookmarks/{tag}", webui.ListBookmarks)
	router.Get("/kill", func(w http.ResponseWriter, r *http.Request) {
		panic("quit")
	})

	staticContent, err := fs.Sub(webui.Static, "static")
	if err != nil {
		panic(err)
	}

	static := http.FileServer(http.FS(staticContent))
	router.Handle("/static/*", http.StripPrefix("/static", static))

	router.Get("/", webui.IndexView)
	router.Get("/test", webui.NamedView("test"))

	return &WebUIServer{router}
}
