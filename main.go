package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/asdine/storm"
	"github.com/danielorf/henry_library/book"
	"github.com/gorilla/mux"
)

func main() {
	dbFile := "test.db"
	err := os.Remove(dbFile) // Clear out old test db file

	var bookDB book.BookDB

	bookDB.DB, err = storm.Open(dbFile, storm.Batch())
	if err != nil {
		fmt.Println(err)
	}
	defer bookDB.DB.Close()

	// Add sample data to BoltDB
	bookDB.FillSampleData()

	// gorilla/mux router and routes
	r := mux.NewRouter()
	r.HandleFunc("/addbook", bookDB.AddBookGet).Methods("GET")
	r.HandleFunc("/addbook", bookDB.AddBookPost).Methods("POST")
	r.HandleFunc("/updatebook/{id:[0-9]+}", bookDB.UpdateBookGet).Methods("GET")
	r.HandleFunc("/updatebook", bookDB.UpdateBookPost).Methods("POST")
	r.HandleFunc("/deletebook/{id:[0-9]+}", bookDB.DeleteBook).Methods("GET")
	r.HandleFunc("/api/v1/list", bookDB.ListBooksJSON).Methods("GET")
	r.HandleFunc("/api/v1/addbook", bookDB.AddBooksJSON).Methods("POST")
	r.HandleFunc("/", bookDB.ListBooks).Methods("GET")
	http.Handle("/", r)

	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}
