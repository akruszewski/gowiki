package page

import (
	"time"
)

// Types builded on top on Page requires Save method which handles
// saving page document and commits log message.
type Pager interface {
	Save() error
}

// Entries are stored in page structure and annotates changes in document.
type LogEntry struct {
	ID      string    `json:"id"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

// Log type stores entries with changes in document.
type Log []LogEntry

// Page is singe wiki page consisting of Title and Document, where Updated is
// date of last change with message which describes it. In log all page updates
// are stored.
type Page struct {
	Title    string    `json:"title"`
	Document string    `json:"document"`
	Updated  time.Time `json:"updated"`
	Message  string    `json:"message"`
	Log      Log       `json:"log"`
}
