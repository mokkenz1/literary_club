package handlers

import (
    "fmt"
    "html/template"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/mokkenz1/literaty_club/internal/database"
)

var tmpl *template.Template

// Данные для страниц
type HomeData struct {
    Title  string
    Active string
    Events []database.Event
}

type GalleryData struct {
    Title  string
    Active string
    Files  []string
}

type ReviewsData struct {
    Title   string
    Active  string
    Reviews []database.Review
    Events  []database.Event
}

func InitTemplates() {
    var err error
    tmpl, err = template.ParseGlob("front/templates/*.html")
    if err != nil {
        panic(fmt.Sprintf("Ошибка загрузки шаблонов: %v", err))
    }
    fmt.Println("✅ Шаблоны загружены")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        renderError(w, "Страница не найдена", http.StatusNotFound)
        return
    }

    events, err := database.GetEvents()
    if err != nil {
        renderError(w, "Ошибка загрузки данных", http.StatusInternalServerError)
        return
    }

    data := HomeData{
        Title:  "Книжный клуб",
        Active: "home",
        Events: events,
    }

    renderTemplate(w, "index.html", data)
}

func GalleryHandler(w http.ResponseWriter, r *http.Request) {
    files, _ := filepath.Glob("front/static/uploads/*")

    var mediaFiles []string
    for _, f := range files {
        ext := strings.ToLower(filepath.Ext(f))
        if isMediaFile(ext) {
            // Получаем имя файла для URL
            mediaFiles = append(mediaFiles, filepath.Base(f))
        }
    }

    data := GalleryData{
        Title:  "Галерея | Книжный клуб",
        Active: "gallery",
        Files:  mediaFiles,
    }

    renderTemplate(w, "gallery.html", data)
}

func ReviewsHandler(w http.ResponseWriter, r *http.Request) {
    reviews, err := database.GetReviews()
    if err != nil {
        renderError(w, "Ошибка загрузки отзывов", http.StatusInternalServerError)
        return
    }

    events, _ := database.GetEvents()

    data := ReviewsData{
        Title:   "Отзывы | Книжный клуб",
        Active:  "reviews",
        Reviews: reviews,
        Events:  events,
    }

    renderTemplate(w, "reviews.html", data)
}

func SubmitReviewHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/reviews", http.StatusSeeOther)
        return
    }

    if err := r.ParseForm(); err != nil {
        renderError(w, "Ошибка обработки формы", http.StatusBadRequest)
        return
    }

    author := strings.TrimSpace(r.FormValue("author"))
    content := strings.TrimSpace(r.FormValue("content"))
    ratingStr := r.FormValue("rating")
    eventIDStr := r.FormValue("event_id")

    if author == "" || content == "" {
        renderError(w, "Имя и отзыв обязательны для заполнения", http.StatusBadRequest)
        return
    }

    // Простая конвертация
    eventID := 1
    rating := 5
    fmt.Sscanf(eventIDStr, "%d", &eventID)
    fmt.Sscanf(ratingStr, "%d", &rating)

    if err := database.AddReview(eventID, author, content, rating); err != nil {
        renderError(w, "Ошибка сохранения отзыва", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/reviews", http.StatusSeeOther)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/gallery", http.StatusSeeOther)
        return
    }

    // Ограничение размера: 10 MB
    r.ParseMultipartForm(10 << 20)

    file, header, err := r.FormFile("file")
    if err != nil {
        renderError(w, "Ошибка загрузки файла", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Проверка расширения
    ext := strings.ToLower(filepath.Ext(header.Filename))
    if !isMediaFile(ext) {
        renderError(w, "Неподдерживаемый формат. Разрешены: jpg, png, gif, mp4", http.StatusBadRequest)
        return
    }

    // Создаем уникальное имя файла
    filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
    dst, err := os.Create(filepath.Join("front/static/uploads", filename))
    if err != nil {
        renderError(w, "Ошибка сохранения файла", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        renderError(w, "Ошибка записи файла", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/gallery", http.StatusSeeOther)
}

// Вспомогательные функции
func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
    if err := tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
        http.Error(w, "Ошибка рендеринга страницы", http.StatusInternalServerError)
        fmt.Printf("Ошибка шаблона %s: %v\n", tmplName, err)
    }
}

func renderError(w http.ResponseWriter, msg string, code int) {
    http.Error(w, msg, code)
}

func isMediaFile(ext string) bool {
    allowed := map[string]bool{
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
        ".gif":  true,
        ".mp4":  true,
        ".webm": true,
    }
    return allowed[ext]
}