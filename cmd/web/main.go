package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

var (
    db   *sql.DB
    tmpl *template.Template
)

type Event struct {
    ID          int
    Title       string
    Description string
    EventDate   string
    Location    string
}

type Review struct {
    ID        int
    EventID   int
    Author    string
    Content   string
    Rating    int
    CreatedAt string
}

var funcMap = template.FuncMap{
    "isImage": func(filename string) bool {
        ext := strings.ToLower(filepath.Ext(filename))
        return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
    },
    "slice": func(s string, start, end int) string {
        if s == "" {
            return ""
        }
        runes := []rune(s)
        if start < 0 || start >= len(runes) {
            return ""
        }
        if end > len(runes) {
            end = len(runes)
        }
        return string(runes[start:end])
    },
    "formatDate": func(dateStr, format string) string {
        // Пробуем разные форматы даты
        layouts := []string{
            "2006-01-02 15:04",
            "2006-01-02",
            "02.01.2006 15:04",
        }
        
        var t time.Time
        var err error
        
        for _, layout := range layouts {
            t, err = time.Parse(layout, dateStr)
            if err == nil {
                break
            }
        }
        
        if err != nil {
            return dateStr
        }
        
        // Месяцы на русском
        months := []string{
            "января", "февраля", "марта", "апреля", "мая", "июня",
            "июля", "августа", "сентября", "октября", "ноября", "декабря",
        }
        
        switch format {
        case "day":
            return fmt.Sprintf("%d", t.Day())
        case "month":
            return months[t.Month()-1]
        case "time":
            return t.Format("15:04")
        case "full":
            return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
        case "datetime":
            return fmt.Sprintf("%d %s %d, %s", t.Day(), months[t.Month()-1], t.Year(), t.Format("15:04"))
        default:
            return dateStr
        }
    },
    "pluralize": func(one, two, five string, n int) string {
        n = n % 100
        if n >= 11 && n <= 19 {
            return five
        }
        n = n % 10
        if n == 1 {
            return one
        }
        if n >= 2 && n <= 4 {
            return two
        }
        return five
    },
}

func main() {
    var err error

    workDir, _ := os.Getwd()
    rootDir := "."
    if strings.HasSuffix(workDir, "cmd/web") {
        rootDir = "../.."
    }

    fmt.Println("📂 Корень проекта:", rootDir)

    // База данных
    dbPath := filepath.Join(rootDir, "bookclub.db")
    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal("Ошибка БД:", err)
    }
    defer db.Close()

    initDB()

    // Шаблоны
    tmplPath := filepath.Join(rootDir, "front", "templates", "*.html")
    fmt.Println("📄 Шаблоны:", tmplPath)
    
    tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob(tmplPath))
    fmt.Println("✅ Шаблоны загружены")

    // Папка для загрузок
    uploadsDir := filepath.Join(rootDir, "front", "static", "uploads")
    os.MkdirAll(uploadsDir, 0755)

    // Маршруты
    http.HandleFunc("/", handleHome)
    http.HandleFunc("/gallery", handleGallery)
    http.HandleFunc("/about", handleAbout)
    http.HandleFunc("/submit-review", handleSubmitReview)
    http.HandleFunc("/upload", handleUpload)

    // Статика
    staticDir := filepath.Join(rootDir, "front", "static")
    fs := http.FileServer(http.Dir(staticDir))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    port := "8080"
    fmt.Printf("\n🚀 Сервер запущен: http://localhost:%s\n", port)
    fmt.Println("📖 Главная:  /")
    fmt.Println("🖼️  Галерея: /gallery")
    fmt.Println("👥 О нас:   /about\n")
    
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initDB() {
    db.Exec(`CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT, description TEXT,
        event_date TEXT, location TEXT
    )`)
    
    db.Exec(`CREATE TABLE IF NOT EXISTS reviews (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_id INTEGER, author TEXT,
        content TEXT, rating INTEGER, created_at TEXT
    )`)

    var eventCount int
    db.QueryRow("SELECT COUNT(*) FROM events").Scan(&eventCount)
    
    if eventCount == 0 {
        fmt.Println("📝 Добавление тестовых мероприятий...")
        db.Exec(`INSERT INTO events (title, description, event_date, location) VALUES 
            ('Обсуждение "1984"', 'Поговорим об антиутопии Джорджа Оруэлла. Актуальна ли она сегодня? Какие параллели можно провести с современным миром?', '2026-05-15 18:00', 'Кафе "Книжный червь"'),
            ('Вечер поэзии', 'Читаем и обсуждаем стихи Ахматовой, Цветаевой, Мандельштама. Приносите любимые сборники!', '2026-05-22 18:00', 'Библиотека им. Пушкина'),
            ('Мастер-класс рецензий', 'Учимся писать рецензии: структура, стиль, аргументация. Разбираем примеры и практикуемся.', '2026-05-29 17:00', 'Коворкинг "Страницы"')`)
    }

    var reviewCount int
    db.QueryRow("SELECT COUNT(*) FROM reviews").Scan(&reviewCount)
    
    if reviewCount == 0 {
        fmt.Println("📝 Добавление тестовых отзывов...")
        db.Exec(`INSERT INTO reviews (event_id, author, content, rating, created_at) VALUES 
            (1, 'Анна', 'Прекрасная встреча! Обсуждение было глубоким и интересным. Оруэлл актуален как никогда.', 5, '2026-05-16 10:00'),
            (2, 'Михаил', 'Очень душевный вечер. Стихи звучали потрясающе в такой атмосфере.', 5, '2026-05-23 11:00'),
            (1, 'Елена', 'Понравилась организация и атмосфера. Приду ещё!', 4, '2026-05-16 15:00')`)
    }
    
    fmt.Println("✅ База данных готова")
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    rows, err := db.Query("SELECT id, title, description, event_date, location FROM events ORDER BY event_date ASC")
    if err != nil {
        http.Error(w, "Ошибка загрузки данных", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var events []Event
    for rows.Next() {
        var e Event
        if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.EventDate, &e.Location); err != nil {
            continue
        }
        events = append(events, e)
    }
    
    if events == nil {
        events = []Event{}
    }

    data := struct {
        Title  string
        Active string
        Events []Event
    }{
        "Главная | Книжный клуб",
        "home",
        events,
    }

    tmpl.ExecuteTemplate(w, "index.html", data)
}

func handleGallery(w http.ResponseWriter, r *http.Request) {
    rootDir := getRootDir()
    files, _ := filepath.Glob(filepath.Join(rootDir, "front", "static", "photos", "*"))
    
    var names []string
    for _, f := range files {
        ext := strings.ToLower(filepath.Ext(f))
        if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".mp4" {
            names = append(names, filepath.Base(f))
        }
    }
    
    if names == nil {
        names = []string{}
    }

    data := struct {
        Title  string
        Active string
        Files  []string
    }{
        "Галерея | Книжный клуб",
        "gallery",
        names,
    }

    tmpl.ExecuteTemplate(w, "gallery.html", data)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
    revRows, err := db.Query(`
        SELECT r.id, r.event_id, r.author, r.content, r.rating, r.created_at 
        FROM reviews r 
        ORDER BY r.created_at DESC
    `)
    
    var reviews []Review
    if err == nil {
        defer revRows.Close()
        for revRows.Next() {
            var r Review
            if err := revRows.Scan(&r.ID, &r.EventID, &r.Author, &r.Content, &r.Rating, &r.CreatedAt); err != nil {
                continue
            }
            reviews = append(reviews, r)
        }
    }
    
    if reviews == nil {
        reviews = []Review{}
    }

    evRows, err := db.Query("SELECT id, title, event_date FROM events")
    
    var events []Event
    if err == nil {
        defer evRows.Close()
        for evRows.Next() {
            var e Event
            if err := evRows.Scan(&e.ID, &e.Title, &e.EventDate); err != nil {
                continue
            }
            events = append(events, e)
        }
    }
    
    if events == nil {
        events = []Event{}
    }

    data := struct {
        Title   string
        Active  string
        Reviews []Review
        Events  []Event
    }{
        "О нас | Книжный клуб",
        "about",
        reviews,
        events,
    }

    tmpl.ExecuteTemplate(w, "about.html", data)
}

func handleSubmitReview(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/about", http.StatusSeeOther)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
        return
    }

    author := strings.TrimSpace(r.FormValue("author"))
    content := strings.TrimSpace(r.FormValue("content"))
    eventID := r.FormValue("event_id")
    rating := r.FormValue("rating")

    if author == "" || content == "" {
        http.Error(w, "Пожалуйста, заполните имя и отзыв", http.StatusBadRequest)
        return
    }

    _, err := db.Exec(
        "INSERT INTO reviews (event_id, author, content, rating, created_at) VALUES (?, ?, ?, ?, ?)",
        eventID, author, content, rating, time.Now().Format("2006-01-02 15:04"),
    )

    if err != nil {
        http.Error(w, "Ошибка сохранения отзыва", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/about", http.StatusSeeOther)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/gallery", http.StatusSeeOther)
        return
    }

    r.ParseMultipartForm(10 << 20)
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Ошибка загрузки", http.StatusBadRequest)
        return
    }
    defer file.Close()

    ext := strings.ToLower(filepath.Ext(header.Filename))
    if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".mp4" {
        http.Error(w, "Неверный формат", http.StatusBadRequest)
        return
    }

    rootDir := getRootDir()
    filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
    dst, err := os.Create(filepath.Join(rootDir, "front", "static", "uploads", filename))
    if err != nil {
        http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    io.Copy(dst, file)

    http.Redirect(w, r, "/gallery", http.StatusSeeOther)
}

func getRootDir() string {
    workDir, _ := os.Getwd()
    if strings.HasSuffix(workDir, "cmd/web") {
        return "../.."
    }
    return "."
}