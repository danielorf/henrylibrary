package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Book is a struct!
type Book struct {
	Title  string
	Author string
	Length int
}

// Books is a Global variable!  It will go away when we add a database.
var Books = []Book{
	{"Where the Wild Things Are", "Maurice Sendak", 45},
	{"Cat in the Hat", "Doctor Seuss", 33},
	{"Dictionary", "Steve", 666},
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", listBooks)
	r.HandleFunc("/confirmation", confirmation)
	r.HandleFunc("/addbook", addBookGet).Methods("GET")
	r.HandleFunc("/addbook", addBookPost).Methods("POST")
	// http.Handle("/", r)

	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/index.html", Books)
}

func addBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func addBookPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Submitted...")

	bookLength, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("length")))

	msg := &Book{
		Title:  r.FormValue("title"),
		Author: r.FormValue("author"),
		Length: bookLength,
	}

	Books = append(Books, *msg)

	log.Println(msg)

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
