package book

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/asdine/storm"
	"github.com/gorilla/mux"
)

// BookDB is a wrapper around a storm.DB instance to allow
// access to DB in HTTP handlers
type BookDB struct {
	DB *storm.DB
}

func (bookDB *BookDB) FillSampleData() {
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
	err := bookDB.DB.All(&books, storm.Reverse())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(books)

	// Testing out update and delete
	var book3 DBBook
	err = bookDB.DB.One("Title", "Dictionary", &book3)
	// err = bookDB.db.One("Pk", 2, &book3)
	// fmt.Println("----------------")
	// fmt.Println(book3)
	// fmt.Println("----------------")
	tempInt := book3.Pk
	_ = bookDB.DB.UpdateField(&DBBook{Pk: tempInt}, "Author", "Bob")
	err = bookDB.DB.All(&books, storm.Reverse())
	// fmt.Println(books)
	// fmt.Println("----------------")
	err = bookDB.DB.One("Title", "Cat in the Hat", &book3)
	// err = bookDB.db.DeleteStruct(&book3)
	// err = bookDB.db.All(&books, storm.Reverse())
	// fmt.Println(books)

	// ------- (Temporary) DB Testing End ------- //
}

func (bookDB *BookDB) addBook(title string, author string, binding string, source string) error {
	// book := DBBook{Title: title, Author: author, DateAdded: time.Now(), DateModified: time.Now()}
	book := DBBook{Title: title, Author: author, Binding: binding, Source: source, DateAdded: time.Now(), DateModified: time.Now()}
	err := bookDB.DB.Save(&book)
	if err != nil {
		return fmt.Errorf("could not save %s by %s, %v", title, author, err)
	}
	return nil
}

func (bookDB *BookDB) ListBooks(w http.ResponseWriter, r *http.Request) {
	var books []DBBook
	err := bookDB.DB.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	render(w, "templates/index.html", books)
}

func (bookDB *BookDB) ListBooksJSON(w http.ResponseWriter, r *http.Request) {
	var books []DBBook
	err := bookDB.DB.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (bookDB *BookDB) AddBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func (bookDB *BookDB) AddBookPost(w http.ResponseWriter, r *http.Request) {
	err := bookDB.addBook(r.FormValue("title"), r.FormValue("author"), r.FormValue("binding"), r.FormValue("source"))
	if err != nil {
		fmt.Println(err)
		errBook := DBBook{Title: r.FormValue("title"), Author: r.FormValue("author"), Error: err.Error()}
		render(w, "templates/addbook.html", errBook)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (bookDB *BookDB) AddBooksJSON(w http.ResponseWriter, r *http.Request) {
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

func (bookDB *BookDB) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var delBook DBBook
	_ = bookDB.DB.One("Pk", id, &delBook)
	_ = bookDB.DB.DeleteStruct(&delBook)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (bookDB *BookDB) UpdateBookGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var upBook DBBook
	_ = bookDB.DB.One("Pk", id, &upBook)
	render(w, "templates/updatebook.html", upBook)
}

func (bookDB *BookDB) UpdateBookPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	err := bookDB.DB.Update(&DBBook{Pk: id, Title: r.FormValue("title"), Author: r.FormValue("author"), Binding: r.FormValue("binding"), Source: r.FormValue("source"), DateModified: time.Now()})
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
