package main

import (
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

type BookDB struct {
	db *storm.DB
}

// Book is a struct!
type Book struct {
	Pk           int    `storm:"id,increment"`
	Title        string `storm:"unique"`
	Author       string `storm:"index"`
	DateAdded    time.Time
	DateModified time.Time
}

// Books is a Global variable!  It will go away when we add a database.
var Books = []Book{
	{1, "Where the Wild Things Are", "Maurice Sendak", time.Now(), time.Now()},
	{2, "Cat in the Hat", "Doctor Seuss", time.Now(), time.Now()},
	{3, "Dictionary", "Steve", time.Now(), time.Now()},
	{4, "Quicksiver", "Neil Stevenson", time.Now(), time.Now()},
}

var IDCounter = len(Books)

func main() {
	dbFile := "test.db"
	err := os.Remove(dbFile)

	var bookDB BookDB
	bookDB.db, err = storm.Open(dbFile)
	if err != nil {
		fmt.Println(err)
	}
	defer bookDB.db.Close()

	// ------- (Temporary) DB Testing Start ------- //
	// https://github.com/asdine/storm#fetch-all-objects
	// https://godoc.org/github.com/asdine/storm#Query
	// https://zupzup.org/boltdb-with-storm/

	// Add sample data to BoltDB
	bookDB.fillSampleData()

	// Get sample DB entry
	var books []Book
	// err = db.All(&books, storm.Limit(1), storm.Reverse())
	err = bookDB.db.All(&books, storm.Reverse())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(books)

	// Testing out update and delete
	var book3 Book
	err = bookDB.db.One("Title", "Dictionary", &book3)
	fmt.Println("----------------")
	tempInt := book3.Pk
	_ = bookDB.db.UpdateField(&Book{Pk: tempInt}, "Author", "Bob")
	err = bookDB.db.All(&books, storm.Reverse())
	fmt.Println(books)
	fmt.Println("----------------")
	err = bookDB.db.One("Title", "Cat in the Hat", &book3)
	err = bookDB.db.DeleteStruct(&book3)
	err = bookDB.db.All(&books, storm.Reverse())
	fmt.Println(books)

	// ------- (Temporary) DB Testing End ------- //

	// gorilla/mux router and routes
	r := mux.NewRouter()
	r.HandleFunc("/addbook", bookDB.addBookGet).Methods("GET")
	r.HandleFunc("/addbook", bookDB.addBookPost).Methods("POST")
	r.HandleFunc("/deletebook/{id:[0-9]+}", bookDB.deleteBook).Methods("GET")
	r.HandleFunc("/", bookDB.listBooks)
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
}

func (bookDB *BookDB) addBook(title string, author string) error {
	book := Book{Title: title, Author: author, DateAdded: time.Now().Add(-10 * time.Minute), DateModified: time.Now()}
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

func (bookDB *BookDB) addBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func (bookDB *BookDB) addBookPost(w http.ResponseWriter, r *http.Request) {
	_ = bookDB.addBook(r.FormValue("title"), r.FormValue("author"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (bookDB *BookDB) deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var delBook Book
	_ = bookDB.db.One("Pk", id, &delBook)
	_ = bookDB.db.DeleteStruct(&delBook)

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
