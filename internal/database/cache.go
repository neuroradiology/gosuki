//
// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
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

package database

const (
	CacheName = "memcache"
	//MemcacheFmt = "file:%s?mode=memory&cache=shared"
	//BufferFmt   = "file:%s?mode=memory&cache=shared"
	DBTypeInMemoryDSN = "file:%s?mode=memory&cache=shared"
	DBTypeCacheDSN    = DBTypeInMemoryDSN
)

var (
	// Global in memory cache of gosuki database
	// Main in memory db, is synced with disc
	// `CacheDB` is a memory replica of disk db
	Cache = &CacheDB{}
)

type CacheDB struct {
	*DB
}

func GetCacheDB() *CacheDB {
	if !Cache.IsInitialized() {
		log.Fatal("cache is not initialized")
	}
	return Cache
}

func (c *CacheDB) IsInitialized() bool {
	return Cache.DB != nil && Cache.Handle != nil
}

func initCache() {
	log.Debug("initializing cacheDB")
	var err error

	// Initialize memory db with schema
	Cache.DB, err = NewDB(CacheName, "", DBTypeCacheDSN).Init()
	if err != nil {
		log.Fatal(err)
	}

	err = Cache.InitSchema()
	if err != nil {
		log.Fatal(err)
	}

	//TEST: sqlite table locked
	// Cache.Handle.SetMaxIdleConns(1)
}
