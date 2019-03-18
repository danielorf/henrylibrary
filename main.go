package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asdine/storm"
	"github.com/gorilla/mux"
)

// BookDB is a wrapper around a storm.DB instance to allow
// access to DB in HTTP handlers
type BookDB struct {
	db *storm.DB
}

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

// Simplified DBBook struct for JSON post unmarshalling
type Book struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Binding string `json:"binding"`
	Source  string `json:"source"`
}

func main() {
	dbFile := "test.db"
	err := os.Remove(dbFile) // Clear out old test db file

	var bookDB BookDB
	bookDB.db, err = storm.Open(dbFile, storm.Batch())
	if err != nil {
		fmt.Println(err)
	}
	defer bookDB.db.Close()

	// Add sample data to BoltDB
	bookDB.fillSampleData()

	// gorilla/mux router and routes
	r := mux.NewRouter()
	r.HandleFunc("/addbook", bookDB.addBookGet).Methods("GET")
	r.HandleFunc("/addbook", bookDB.addBookPost).Methods("POST")
	r.HandleFunc("/updatebook/{id:[0-9]+}", bookDB.updateBookGet).Methods("GET")
	r.HandleFunc("/updatebook", bookDB.updateBookPost).Methods("POST")
	r.HandleFunc("/deletebook/{id:[0-9]+}", bookDB.deleteBook).Methods("GET")
	r.HandleFunc("/api/v1/list", bookDB.listBooksJSON).Methods("GET")
	r.HandleFunc("/api/v1/addbook", bookDB.addBooksJSON).Methods("POST")
	r.HandleFunc("/", bookDB.listBooks).Methods("GET")
	http.Handle("/", r)

	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}

func (bookDB *BookDB) fillSampleData() {
	bookList := [][]string{
		{"Where the Wild Things Are", "Maurice Sendak", "", ""},
		{"Cat in the Hat", "Doctor Seuss", "", ""},
		{"On the Road", "Jack Kerouac", "", ""},
		{"Dictionary", "Steve", "", ""},
		{"Quicksilver", "Neil Stephenson", "", ""},
	}
	for _, elem := range bookList {
		_ = bookDB.addBook(elem[0], elem[1], elem[2], elem[3])
	}

	// ------- (Temporary) DB Testing Start ------- //
	// https://github.com/asdine/storm#fetch-all-objects
	// https://godoc.org/github.com/asdine/storm#Query
	// https://zupzup.org/boltdb-with-storm/
	// https://github.com/zupzup/boltdb-storm-example/blob/master/main.go

	// Get sample DB entry
	var books []DBBook
	// err = db.All(&books, storm.Limit(1), storm.Reverse())
	err := bookDB.db.All(&books, storm.Reverse())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(books)

	// Testing out update and delete
	var book3 DBBook
	err = bookDB.db.One("Title", "Dictionary", &book3)
	// err = bookDB.db.One("Pk", 2, &book3)
	// fmt.Println("----------------")
	// fmt.Println(book3)
	// fmt.Println("----------------")
	tempInt := book3.Pk
	_ = bookDB.db.UpdateField(&DBBook{Pk: tempInt}, "Author", "Bob")
	err = bookDB.db.All(&books, storm.Reverse())
	// fmt.Println(books)
	// fmt.Println("----------------")
	err = bookDB.db.One("Title", "Cat in the Hat", &book3)
	// err = bookDB.db.DeleteStruct(&book3)
	// err = bookDB.db.All(&books, storm.Reverse())
	// fmt.Println(books)

	// ------- (Temporary) DB Testing End ------- //
}

func (bookDB *BookDB) addBook(title string, author string, binding string, source string) error {
	// book := DBBook{Title: title, Author: author, DateAdded: time.Now(), DateModified: time.Now()}
	book := DBBook{Title: title, Author: author, Binding: binding, Source: source, DateAdded: time.Now(), DateModified: time.Now()}
	err := bookDB.db.Save(&book)
	if err != nil {
		return fmt.Errorf("could not save %s by %s, %v", title, author, err)
	}
	return nil
}

func (bookDB *BookDB) listBooks(w http.ResponseWriter, r *http.Request) {
	var books []DBBook
	err := bookDB.db.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	render(w, "templates/index.html", books)
}

func (bookDB *BookDB) listBooksJSON(w http.ResponseWriter, r *http.Request) {
	var books []DBBook
	err := bookDB.db.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (bookDB *BookDB) addBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func (bookDB *BookDB) addBookPost(w http.ResponseWriter, r *http.Request) {
	err := bookDB.addBook(r.FormValue("title"), r.FormValue("author"), r.FormValue("binding"), r.FormValue("source"))
	if err != nil {
		fmt.Println(err)
		errBook := DBBook{Title: r.FormValue("title"), Author: r.FormValue("author"), Error: err.Error()}
		render(w, "templates/addbook.html", errBook)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (bookDB *BookDB) addBooksJSON(w http.ResponseWriter, r *http.Request) {
	var booksSimp []Book
	err := json.NewDecoder(r.Body).Decode(&booksSimp)
	if err != nil {
		println("whoops!  json decode failure.")
		errorJSON := struct {
			Error string `json:"error"`
		}{
			"failed to decode",
		}
		json.NewEncoder(w).Encode(errorJSON)
	} else {
		for _, elem := range booksSimp {
			err = bookDB.addBook(elem.Title, elem.Author, elem.Binding, elem.Source)
			if err != nil {
				log.Println(err)
				errorJSON := struct {
					Error string `json:"error"`
				}{
					err.Error(),
				}
				json.NewEncoder(w).Encode(errorJSON)
			}
		}
	}
}

func (bookDB *BookDB) deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var delBook DBBook
	_ = bookDB.db.One("Pk", id, &delBook)
	_ = bookDB.db.DeleteStruct(&delBook)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (bookDB *BookDB) updateBookGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var upBook DBBook
	_ = bookDB.db.One("Pk", id, &upBook)
	render(w, "templates/updatebook.html", upBook)
}

func (bookDB *BookDB) updateBookPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	err := bookDB.db.Update(&DBBook{Pk: id, Title: r.FormValue("title"), Author: r.FormValue("author"), Binding: r.FormValue("binding"), Source: r.FormValue("source"), DateModified: time.Now()})
	if err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
