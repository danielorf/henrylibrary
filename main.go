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

// Book is a representation of a book object with database storage in mind
type Book struct {
	Pk           int    `storm:"id,increment"`
	Title        string `storm:"unique"`
	Author       string `storm:"index"`
	DateAdded    time.Time
	DateModified time.Time
}

// Simplified Book struct for JSON post
type BookSimp struct {
	Title  string `json:"title"`
	Author string `json:"author"`
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
		{"Where the Wild Things Are", "Maurice Sendak"},
		{"Cat in the Hat", "Doctor Seuss"},
		{"On the Road", "Jack Kerouac"},
		{"Dictionary", "Steve"},
		{"Quicksilver", "Neil Stephenson"},
	}
	for _, elem := range bookList {
		_ = bookDB.addBook(elem[0], elem[1])
	}

	// ------- (Temporary) DB Testing Start ------- //
	// https://github.com/asdine/storm#fetch-all-objects
	// https://godoc.org/github.com/asdine/storm#Query
	// https://zupzup.org/boltdb-with-storm/
	// https://github.com/zupzup/boltdb-storm-example/blob/master/main.go

	// Get sample DB entry
	var books []Book
	// err = db.All(&books, storm.Limit(1), storm.Reverse())
	err := bookDB.db.All(&books, storm.Reverse())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(books)

	// Testing out update and delete
	var book3 Book
	err = bookDB.db.One("Title", "Dictionary", &book3)
	// fmt.Println("----------------")
	tempInt := book3.Pk
	_ = bookDB.db.UpdateField(&Book{Pk: tempInt}, "Author", "Bob")
	err = bookDB.db.All(&books, storm.Reverse())
	// fmt.Println(books)
	// fmt.Println("----------------")
	err = bookDB.db.One("Title", "Cat in the Hat", &book3)
	err = bookDB.db.DeleteStruct(&book3)
	err = bookDB.db.All(&books, storm.Reverse())
	// fmt.Println(books)

	// ------- (Temporary) DB Testing End ------- //
}

func (bookDB *BookDB) addBook(title string, author string) error {
	book := Book{Title: title, Author: author, DateAdded: time.Now(), DateModified: time.Now()}
	err := bookDB.db.Save(&book)
	if err != nil {
		return fmt.Errorf("could not save book, %v", err)
	}
	return nil
}

func (bookDB *BookDB) listBooks(w http.ResponseWriter, r *http.Request) {
	var books []Book
	err := bookDB.db.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	render(w, "templates/index.html", books)
}

func (bookDB *BookDB) listBooksJSON(w http.ResponseWriter, r *http.Request) {
	var books []Book
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
	_ = bookDB.addBook(r.FormValue("title"), r.FormValue("author"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (bookDB *BookDB) addBooksJSON(w http.ResponseWriter, r *http.Request) {
	var booksSimp []BookSimp
	err := json.NewDecoder(r.Body).Decode(&booksSimp)
	if err != nil {
		println("whoops!  json decode failure.")
		errorz := struct {
			Error string `json:"error"`
		}{
			"failed to decode",
		}
		json.NewEncoder(w).Encode(errorz)
	} else {
		for _, elem := range booksSimp {
			err = bookDB.addBook(elem.Title, elem.Author)
			if err != nil {
				log.Println(err)
				break
			}
		}
	}
}

func (bookDB *BookDB) deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var delBook Book
	_ = bookDB.db.One("Pk", id, &delBook)
	_ = bookDB.db.DeleteStruct(&delBook)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (bookDB *BookDB) updateBookGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var upBook Book
	_ = bookDB.db.One("Pk", id, &upBook)
	render(w, "templates/updatebook.html", upBook)
}

func (bookDB *BookDB) updateBookPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	_ = bookDB.db.Update(&Book{Pk: id, Title: r.FormValue("title"), Author: r.FormValue("author"), DateModified: time.Now()})
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
