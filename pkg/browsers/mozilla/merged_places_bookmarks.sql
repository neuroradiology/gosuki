--
-- Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
--
-- All rights reserved.
--
-- SPDX-License-Identifier: AGPL-3.0-or-later
--
-- This file is part of GoSuki.
--
-- GoSuki is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
--
-- GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more details.
--
-- You should have received a copy of the GNU Affero General Public License along with gosuki.  If not, see <http://www.gnu.org/licenses/>. 

-- name: merged-places-bookmarks
SELECT 
    moz_bookmarks.id as bkId,
    (moz_bookmarks.fk ISNULL and moz_bookmarks.parent not in (4,0)) as isFolder, -- folder = not son of root(0) or tag(4)
    moz_bookmarks.parent == 4 as isTag,
    moz_places.id IS NOT NULL as isBk,
    moz_bookmarks.parent as bkParent,
    ifnull(moz_places.id, -1) as plId,
    ifnull(moz_places.url, "") as plUrl,
    ifnull(moz_places.description, "") as plDescription,


    ifnull(moz_bookmarks.title, "") as bkTitle,
    moz_bookmarks.lastModified as bkLastModified
    -- datetime(moz_bookmarks.lastModified / (1000*1000), 'unixepoch') as bkLastModifiedDateTime

    FROM moz_bookmarks 
    LEFT OUTER JOIN moz_places
    ON moz_places.id = moz_bookmarks.fk
    ORDER BY plId
