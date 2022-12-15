package mozilla

import (
    "embed"
)


const(
    MozBookmarkQueryFile = "recursive_all_bookmarks.sql"
    MozBookmarkQuery = "recursive-all-bookmarks"
) 

var (
    //go:embed "recursive_all_bookmarks.sql"
    EmbeddedSqlQueries embed.FS
)

// sql queries
const (
	QGetBookmarkPlace = `
	SELECT *
	FROM moz_places
	WHERE id = ?
	`
	//TEST:
	QBookmarksChanged = `
    SELECT id,type,IFNULL(fk, -1) AS fk,parent,IFNULL(title, '') AS title from moz_bookmarks
    WHERE(lastModified > :last_runtime_utc
        AND lastModified < strftime('%s', 'now')*1000*1000
        AND NOT id IN (:not_root_tags)
    )
	`

	QFolders = `
    SELECT id, title, parent FROM moz_bookmarks 
        WHERE type = 2 AND parent NOT IN (4, 0)
    `

	//TEST:
	QgetBookmarks = `
    WITH bookmarks AS
	(SELECT moz_places.url AS url,
			moz_places.description as desc,
			moz_places.title as urlTitle,
			moz_bookmarks.parent AS tagId
		FROM moz_places LEFT OUTER JOIN moz_bookmarks
		ON moz_places.id = moz_bookmarks.fk
		WHERE moz_bookmarks.parent
		IN (SELECT id FROM moz_bookmarks WHERE parent = ? ))

	SELECT url, IFNULL(urlTitle, ''), IFNULL(desc,''),
			tagId, moz_bookmarks.title AS tagTitle

	FROM bookmarks LEFT OUTER JOIN moz_bookmarks
	ON tagId = moz_bookmarks.id
	ORDER BY url
    `

	//TEST:
	//TODO:
	QGetBookmarkFolders = `
		SELECT
		moz_places.id as placesId,
		moz_places.url as url,	
			moz_places.description as description,
			moz_bookmarks.title as title,
			moz_bookmarks.fk ISNULL as isFolder
			
		FROM moz_bookmarks LEFT OUTER JOIN moz_places
		ON moz_places.id = moz_bookmarks.fk
		WHERE moz_bookmarks.parent = 3
	`
)
