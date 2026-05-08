package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template
var db *sql.DB

func init() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	var err error
	db, err = InitDB("./bookclub.db")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/books", handleBooks)
	http.HandleFunc("/gallery", handleGallery)
	http.HandleFunc("/reviews", handleReviews)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	books, err := GetAllBooks(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reviews, err := GetAllReviews(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Берём первые 3 книги и 3 отзыва
	if len(books) > 3 {
		books = books[:3]
	}
	if len(reviews) > 3 {
		reviews = reviews[:3]
	}

	data := struct {
		Title   string
		Books   []Book
		Reviews []Review
	}{
		Title:   "Книжный клуб",
		Books:   books,
		Reviews: reviews,
	}

	templates.ExecuteTemplate(w, "index.html", data)
}

func handleBooks(w http.ResponseWriter, r *http.Request) {
	books, err := GetAllBooks(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Books []Book
	}{
		Title: "Все книги",
		Books: books,
	}

	templates.ExecuteTemplate(w, "books.html", data)
}

func handleGallery(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "gallery.html", nil)
}

func handleReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := GetAllReviews(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title   string
		Reviews []Review
	}{
		Title:   "Отзывы о книгах",
		Reviews: reviews,
	}

	templates.ExecuteTemplate(w, "reviews.html", data)
}