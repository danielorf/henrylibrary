package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Book is a struct!
type Book struct {
	ID     int
	Title  string
	Author string
	Length int
}

// Books is a Global variable!  It will go away when we add a database.
var Books = []Book{
	{1, "Where the Wild Things Are", "Maurice Sendak", 45},
	{2, "Cat in the Hat", "Doctor Seuss", 33},
	{3, "Dictionary", "Steve", 666},
	{4, "Quicksiver", "Neil Stevenson", 100005},
}

var IDCounter = len(Books)

func main() {
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

func listBooks(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/index.html", Books)
}

func addBookGet(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/addbook.html", nil)
}

func addBookPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Submitted...")
	IDCounter++
	bookLength, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("length")))

	msg := &Book{
		ID:     IDCounter,
		Title:  r.FormValue("title"),
		Author: r.FormValue("author"),
		Length: bookLength,
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
		if id == Books[i].ID {
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
