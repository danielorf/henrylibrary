package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/asdine/storm"
	"github.com/danielorf/henry_library/pkg/book"
	"github.com/gorilla/mux"
)

// App is a wrapper around a storm.DB instance to allow
// access to DB in HTTP handlers
type App struct {
	Router *mux.Router
	DB     *storm.DB
}

func (a *App) Initialize(dbname string) {
	err := os.Remove(dbname) // Clear out old test db file

	a.DB, err = storm.Open(dbname, storm.Batch())
	if err != nil {
		fmt.Println(err)
	}
	// defer a.DB.Close()  Moved to Run()

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/addbook", a.AddBookGet).Methods("GET")
	a.Router.HandleFunc("/addbook", a.AddBookPost).Methods("POST")
	a.Router.HandleFunc("/updatebook/{id:[0-9]+}", a.UpdateBookGet).Methods("GET")
	a.Router.HandleFunc("/updatebook", a.UpdateBookPost).Methods("POST")
	a.Router.HandleFunc("/deletebook/{id:[0-9]+}", a.DeleteBook).Methods("GET")
	a.Router.HandleFunc("/api/v1/list", a.ListBooksJSON).Methods("GET")
	a.Router.HandleFunc("/api/v1/addbook", a.AddBooksJSON).Methods("POST")
	a.Router.HandleFunc("/listbooktable", a.ListBooksTable).Methods("GET")
	a.Router.HandleFunc("/", a.ListBooks).Methods("GET")
}

func (a *App) Run(addr string) {
	defer a.DB.Close()
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) addBook(title string, author string, binding string, source string) error {
	// book := DBBook{Title: title, Author: author, DateAdded: time.Now(), DateModified: time.Now()}
	book := book.DBBook{Title: title, Author: author, Binding: binding, Source: source, DateAdded: time.Now(), DateModified: time.Now()}
	err := a.DB.Save(&book)
	if err != nil {
		return fmt.Errorf("could not save %s by %s, %v", title, author, err)
	}
	return nil
}

func (a *App) ListBooks(w http.ResponseWriter, r *http.Request) {
	var books []book.DBBook
	err := a.DB.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	render(w, "templates/index.html", books)
}

func (a *App) ListBooksTable(w http.ResponseWriter, r *http.Request) {
	// render(w, "templates/index.html", books)
	http.ServeFile(w, r, "templates/vuetable.html")
}

func (a *App) ListBooksJSON(w http.ResponseWriter, r *http.Request) {
	log.Println("CORS enabled for testing, remove in production")
	(w).Header().Set("Access-Control-Allow-Origin", "*")

	var books []book.DBBook
	err := a.DB.All(&books)
	if err != nil {
		fmt.Println(fmt.Errorf("could not fetch books, %v", err))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (a *App) AddBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func (a *App) AddBookPost(w http.ResponseWriter, r *http.Request) {
	err := a.addBook(r.FormValue("title"), r.FormValue("author"), r.FormValue("binding"), r.FormValue("source"))
	if err != nil {
		fmt.Println(err)
		errBook := book.DBBook{Title: r.FormValue("title"), Author: r.FormValue("author"), Error: err.Error()}
		render(w, "templates/addbook.html", errBook)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (a *App) AddBooksJSON(w http.ResponseWriter, r *http.Request) {
	var booksSimp []book.Book
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
			err = a.addBook(elem.Title, elem.Author, elem.Binding, elem.Source)
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

func (a *App) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var delBook book.DBBook
	_ = a.DB.One("Pk", id, &delBook)
	_ = a.DB.DeleteStruct(&delBook)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (a *App) UpdateBookGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var upBook book.DBBook
	_ = a.DB.One("Pk", id, &upBook)
	render(w, "templates/updatebook.html", upBook)
}

func (a *App) UpdateBookPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	err := a.DB.Update(&book.DBBook{Pk: id, Title: r.FormValue("title"), Author: r.FormValue("author"), Binding: r.FormValue("binding"), Source: r.FormValue("source"), DateModified: time.Now()})
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

func (a *App) FillSampleData() {
	bookList := [][]string{
		{"Where the Wild Things Are", "Maurice Sendak", "boardback", "G&G"},
		{"Cat in the Hat", "Doctor Seuss", "paperback", "Santa"},
		{"On the Road", "Jack Kerouac", "paperback", "A time traveller"},
		{"Dictionary", "Steve", "hardback", "Library"},
		{"Quicksilver", "Neil Stephenson", "paperback", "Different time traveller"},
	}
	for _, elem := range bookList {
		_ = a.addBook(elem[0], elem[1], elem[2], elem[3])
	}

	// ------- (Temporary) DB Testing Start ------- //
	// https://github.com/asdine/storm#fetch-all-objects
	// https://godoc.org/github.com/asdine/storm#Query
	// https://zupzup.org/boltdb-with-storm/
	// https://github.com/zupzup/boltdb-storm-example/blob/master/main.go

	// Get sample DB entry
	var books []book.DBBook
	// err = db.All(&books, storm.Limit(1), storm.Reverse())
	err := a.DB.All(&books, storm.Reverse())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(books)

	// Testing out update and delete
	var book3 book.DBBook
	err = a.DB.One("Title", "Dictionary", &book3)
	// err = a.db.One("Pk", 2, &book3)
	// fmt.Println("----------------")
	// fmt.Println(book3)
	// fmt.Println("----------------")
	tempInt := book3.Pk
	_ = a.DB.UpdateField(&book.DBBook{Pk: tempInt}, "Author", "Bob")
	err = a.DB.All(&books, storm.Reverse())
	// fmt.Println(books)
	// fmt.Println("----------------")
	err = a.DB.One("Title", "Cat in the Hat", &book3)
	// err = a.db.DeleteStruct(&book3)
	// err = a.db.All(&books, storm.Reverse())
	// fmt.Println(books)

	// ------- (Temporary) DB Testing End ------- //
}
