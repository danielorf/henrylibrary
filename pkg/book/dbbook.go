package book

import (
	"time"
)

// DBBook is a representation of a book object with database storage in mind
// Note: I have chosen not to include a Book struct because storm doesn't allow
// for query inside a nested struct.  This should be re-evaluated when DB changes
type DBBook struct {
	Pk int `storm:"id,increment"`
	// Book         Book `storm:"unique"`
	Title        string `storm:"unique"`
	Author       string `storm:"index"`
	Binding      string
	Source       string
	DateAdded    time.Time
	DateModified time.Time
	Error        string
}
