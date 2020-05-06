package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"encoding/json"
	"errors"
)

func main() {

	handler := http.NewServeMux()

	handler.HandleFunc("/hello/", Logger(helloHandler))
	handler.HandleFunc("/book/", bookHandler)
	handler.HandleFunc("/books/", booksHandler)

	s := http.Server{
		Addr:    ":8080",
		Handler: handler,
		//Port:       "3306",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Start")

	log.Fatal(s.ListenAndServe())
}

type Resp struct {
	Message interface{}
	Error   string
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := strings.Replace(r.URL.Path, "/hello/", "", 1)

	resp := Resp{
		Message: fmt.Sprintf("hello %s. Glad to see you again", name),
		Error:   "",
	}

	respJson, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)

	w.Write(respJson)
}

func Logger(next http.HandlerFunc) http.HandlerFunc { // промежуточная функция
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		log.Printf("server [net/http] method [%s] connection from [%v]", r.Method, r.RemoteAddr)

		next.ServeHTTP(w, r)
	}
}

func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !validate(pair[0], pair[1]) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func validate(username, password string) bool {
	if username == "test" && password == "test" { //Basic dGVzdDp0ZXN0
		return true
	}
	return false
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGetBook(w, r)
	} else if r.Method == http.MethodPost {
		handleAddBook(w, r)
	} else if r.Method == http.MethodPost {
		handleAddBook(w, r)
	} else if r.Method == http.MethodDelete {
		handleDeleteBook(w, r)
	} else if r.Method == http.MethodPut {
		handleUpdateBook(w, r)
	}
}

func handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	decoder := json.NewDecoder(r.Body)

	var book Book

	var resp Resp

	err := decoder.Decode(&book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()

		respJson, _ := json.Marshal(resp)

		w.Write(respJson)

		return
	}

	book.Id = id

	err = bookStore.UpdateBook(book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()

		respJson, _ := json.Marshal(resp)

		w.Write(respJson)

		return
	}

	resp.Message = book

	w.WriteHeader(http.StatusOK)

	respJson, _ := json.Marshal(resp)

	w.Write(respJson)
}

func handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	var resp Resp

	err := bookStore.DeleteBook(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		resp.Error = err.Error()

		respJson, _ := json.Marshal(resp)

		w.Write(respJson)

		return
	}

	bookHandler(w, r)

}

func handleAddBook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var book Book

	var resp Resp

	err := decoder.Decode(&book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()

		respJson, _ := json.Marshal(resp)

		w.Write(respJson)
		return
	}
	//err = bookStore.AddBooks(book)
	//
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	resp.Error = err.Error()
	//
	//	respJson, _ := json.Marshal(resp)
	//
	//	w.Write(respJson)
	//	return
	//}

	bookHandler(w, r)

}

//
func booksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		handleGetBook(w, r)
	}

}

type Book struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Name   string `json:"name"`
}

type BookStore struct {
	books []Book
}

func handleGetBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	book := bookStore.FindBookById(id)

	var resp Resp

	if book == nil {
		w.WriteHeader(http.StatusNotFound)
		resp.Error = fmt.Sprintf("")

		respJson, _ := json.Marshal(resp)

		w.Write(respJson)

		return
	}

	resp.Message = book

	w.WriteHeader(http.StatusOK)

	respJson, _ := json.Marshal(resp)

	w.Write(respJson)
}

var bookStore = BookStore{
	books: make([]Book, 0),
}

func (s BookStore) FindBookById(id string) *Book {
	for _, book := range s.books {
		if book.Id == id {
			return &book
		}
	}
	return nil
}

func (s BookStore) GetBooks() []Book {
	return s.books
}

func (s *BookStore) AddBook(book Book) error {
	bk := s.FindBookById(book.Id)
	if bk != nil {
		return errors.New(fmt.Sprintf("Book with id %s already exists", book.Id))

	}
	s.books = append(s.books, book)

	return nil
}

func (s *BookStore) SetBook(book Book) error {
	for i, bk := range s.books {
		if bk.Id == book.Id {

			s.books[i] = book

			return nil
		}
	}

	return errors.New(fmt.Sprintf("There is no book with id %s", book.Id))
}

//func (s BookStore) AddBooks(book Book) error {
//	for _, bk := range s.books {
//		if bk.Id == id {
//			return
//		}
//	}
//	s.books = append(s.books, book)
//	return s.books
//}
//
//func (s *BookStore) AddBook(book Book) error{
//	bk := s.FindBookById(book.Id)
//	if bk != nil {
//		return errors.New(fmt.Sprintf("Book with id %s already exists", book.Id))
//
//	}
//	s.books = append(s.books, book)
//
//	return nil
//}
//
//func (s *BookStore) SetBook(book Book) error{
//	for i, bk := range s.books {
//		if bk.Id == book.Id {
//
//			s.books[i] = book
//
//			return nil
//		}
//	}
//
//	return errors.New(fmt.Sprintf("There is no book with id %s", book.Id))
//}

func (s *BookStore) UpdateBook(book Book) error {
	for i, bk := range s.books {
		if bk.Id == book.Id {
			s.books[i] = book
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Book with id %s not found", book.Id))
}

func (s *BookStore) DeleteBook(id string) error {
	for i, bk := range s.books {
		if bk.Id == id {
			s.books = append(s.books[:i], s.books[i+1:]...)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Book with id %s not found", id))

}
