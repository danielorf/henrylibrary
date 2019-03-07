package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/bmizerany/pat"
)

type Book struct {
	Title  string
	Author string
	Length int
}

func main() {
	mux := pat.New()
	mux.Get("/", http.HandlerFunc(index))

	log.Println("Listening...")
	http.ListenAndServe(":3000", mux)
}

func index(w http.ResponseWriter, r *http.Request) {

	books := []Book{
		{"Where the Wild Things Are", "Maurice Sendak", 45},
		{"Cat in the hat", "Doctor Seuss", 32},
	}
	render(w, "templates/index.html", books)
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
