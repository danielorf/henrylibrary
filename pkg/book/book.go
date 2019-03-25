package book

// BookDB is a wrapper around a storm.DB instance to allow
// access to DB in HTTP handlers

// Simplified DBBook struct for JSON post unmarshalling
type Book struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Binding string `json:"binding"`
	Source  string `json:"source"`
}
