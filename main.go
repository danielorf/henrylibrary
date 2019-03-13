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
	db, err := storm.Open(dbFile)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// ------- (Temporary) DB Testing Start ------- //
	// https://github.com/asdine/storm#fetch-all-objects
	// https://godoc.org/github.com/asdine/storm#Query
	// https://zupzup.org/boltdb-with-storm/

	// Add sample data to BoltDB
	sampleData(db)

	// Get sample DB entry
	var books []Book
	// err = db.All(&books, storm.Limit(1), storm.Reverse())
	err = db.All(&books, storm.Reverse())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(books)

	// Testing out update and delete
	var book3 Book
	err = db.One("Title", "Dictionary", &book3)
	fmt.Println("----------------")
	tempInt := book3.Pk
	_ = db.UpdateField(&Book{Pk: tempInt}, "Author", "Bob")
	err = db.All(&books, storm.Reverse())
	fmt.Println(books)
	fmt.Println("----------------")
	err = db.One("Title", "Cat in the Hat", &book3)
	err = db.DeleteStruct(&book3)
	err = db.All(&books, storm.Reverse())
	fmt.Println(books)

	// ------- (Temporary) DB Testing End ------- //

	// gorilla/mux router and routes
	r := mux.NewRouter()
	r.HandleFunc("/confirmation", confirmation)
	r.HandleFunc("/addbook", addBookGet).Methods("GET")
	r.HandleFunc("/addbook", addBookPost).Methods("POST")
	r.HandleFunc("/deletebook/{id:[0-9]+}", deleteBook).Methods("GET")
	r.HandleFunc("/", listBooks)
	http.Handle("/", r)

	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}

func sampleData(db *storm.DB) {

	bookList := [][]string{
		{"Where the Wild Things Are", "Maurice Sendak"},
		{"Cat in the Hat", "Doctor Seuss"},
		{"Dictionary", "Steve"},
		{"Quicksilver", "Neil Stephenson"},
	}
	for _, elem := range bookList {
		_ = addBook(db, elem[0], elem[1])
	}
}

func addBook(db *storm.DB, title string, author string) error {
	book := Book{Title: title, Author: author, DateAdded: time.Now().Add(-10 * time.Minute), DateModified: time.Now()}
	err := db.Save(&book)
	if err != nil {
		return fmt.Errorf("could not save book, %v", err)
	}
	return nil
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/index.html", Books)
}

func addBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func addBookPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Submitted...")
	IDCounter++
	// bookLength, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("length")))

	msg := &Book{
		Pk:           IDCounter,
		Title:        r.FormValue("title"),
		Author:       r.FormValue("author"),
		DateAdded:    time.Now(),
		DateModified: time.Now(),
	}

	Books = append(Books, *msg)

	log.Println(msg)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	// w.WriteHeader(http.StatusOK)
	fmt.Println(vars["id"])
	id, _ := strconv.Atoi(vars["id"])

	for i := 0; i < len(Books); i++ {
		// Iterate through Book objects and remove Book if ID matches deleted ID
		if id == Books[i].Pk {
			if i < len(Books)-1 {
				Books = append(Books[:i], Books[i+1:]...)
				break
			} else {
				Books = Books[:i]
				break
			}
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func confirmation(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/confirmation.html", nil)
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
