package main

import (
	"html/template"
	"log"            
	"net/http"     
)

var templates *template.Template

func init() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleIndex)       // Главная страница
	http.HandleFunc("/gallery", handleGallery) // Галерея
	http.HandleFunc("/reviews", handleReviews) // Отзывы

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGallery(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "gallery.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReviews(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "reviews.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}