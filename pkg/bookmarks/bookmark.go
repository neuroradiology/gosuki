package bookmarks

// Bookmark type
type Bookmark struct {
	URL      string   `json:"url"`
	Metadata string   `json:"metadata"`
	Tags     []string `json:"tags"`
	Desc     string   `json:"desc"`
	//flags int
}
