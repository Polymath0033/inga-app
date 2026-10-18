package links

/*
	{
	  "url": "https://example.com",
	  "title": "Example Page",
	  "description": "Optional description",
	  "collection": "Favorites",
	  "category": "Work",
	  "tags": ["javascript", "tutorial"],
	  "reminder": {
	    "date": "2026-10-20",
	    "time": "15:30"
	  },
	  "bookmarked": true
	}
*/
type SaveLinkRequest struct {
	Url         string   `json:"url" binding:"required,url"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category" binding:"required"`
	RemindAt    string   `json:"remind_at" binding:"required"`
	Tags        []string `json:"tags" binding:"required"`
	Favicon     string   `json:"favicon"`
	Collection  string   `json:"collection"`
}

type SaveLinkResponse struct {
	ID int32 `json:"id"`
}
