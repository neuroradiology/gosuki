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
