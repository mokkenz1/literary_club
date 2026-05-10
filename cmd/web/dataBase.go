package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// ===== СТРУКТУРЫ =====

// User - пользователь книжного клуба
type User struct {
	ID       int
	Username string
	Email    string
	Password string
	Role     string
}

// Book - книга (упрощённая версия)
type Book struct {
	ID          int
	Title       string
	Author      string
	Year        int
	Genre       string
	Description string
	CoverURL    string
}

// Review - отзыв на книгу
type Review struct {
	ID        int
	Username  string // имя пользователя (не ID, для простоты)
	BookTitle string // название книги (не ID, для простоты)
	Rating    int
	Text      string
}

// ===== ФУНКЦИИ БАЗЫ ДАННЫХ =====

// InitDB - инициализация базы данных
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ База данных подключена:", dbPath)

	if err = createTables(db); err != nil {
		return nil, err
	}

	if err = seedData(db); err != nil {
		return nil, err
	}

	return db, nil
}

// createTables - создаёт таблицы
func createTables(db *sql.DB) error {
	// Таблица пользователей
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'user'
	);`

	if _, err := db.Exec(usersTable); err != nil {
		return err
	}

	// Таблица книг
	booksTable := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		year INTEGER,
		genre TEXT,
		description TEXT,
		cover_url TEXT
	);`

	if _, err := db.Exec(booksTable); err != nil {
		return err
	}

	// Таблица отзывов
	reviewsTable := `
	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		book_title TEXT NOT NULL,
		rating INTEGER CHECK(rating >= 1 AND rating <= 5),
		text TEXT NOT NULL
	);`

	if _, err := db.Exec(reviewsTable); err != nil {
		return err
	}

	log.Println("✅ Таблицы созданы: users, books, reviews")
	return nil
}

// seedData - заполняет тестовыми данными
func seedData(db *sql.DB) error {
	// Проверяем, есть ли книги
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("📚 Данные уже есть, пропускаем заполнение")
		return nil
	}

	log.Println("📝 Заполняем тестовыми данными...")

	// 1. Добавляем пользователей
	users := []User{
		{Username: "alexey", Email: "alexey@bookclub.com", Password: "hash123", Role: "admin"},
		{Username: "maria", Email: "maria@bookclub.com", Password: "hash456", Role: "user"},
		{Username: "dmitry", Email: "dmitry@bookclub.com", Password: "hash789", Role: "user"},
	}

	for _, u := range users {
		_, err := db.Exec(`
			INSERT INTO users (username, email, password, role)
			VALUES (?, ?, ?, ?)`,
			u.Username, u.Email, u.Password, u.Role)
		if err != nil {
			return err
		}
	}

	// 2. Добавляем книги
	books := []Book{
		{Title: "Мастер и Маргарита", Author: "Михаил Булгаков", Year: 1967, Genre: "Роман", Description: "Сатирический роман о дьяволе в Москве", CoverURL: ""},
		{Title: "1984", Author: "Джордж Оруэлл", Year: 1949, Genre: "Антиутопия", Description: "Тоталитарное будущее", CoverURL: ""},
		{Title: "Преступление и наказание", Author: "Фёдор Достоевский", Year: 1866, Genre: "Роман", Description: "Психологический роман о студенте Раскольникове", CoverURL: ""},
		{Title: "Гарри Поттер и философский камень", Author: "Дж.К. Роулинг", Year: 1997, Genre: "Фэнтези", Description: "Начало приключений мальчика-волшебника", CoverURL: ""},
		{Title: "Три товарища", Author: "Эрих Мария Ремарк", Year: 1936, Genre: "Роман", Description: "Трогательная история о дружбе и любви", CoverURL: ""},
	}

	for _, b := range books {
		_, err := db.Exec(`
			INSERT INTO books (title, author, year, genre, description, cover_url)
			VALUES (?, ?, ?, ?, ?, ?)`,
			b.Title, b.Author, b.Year, b.Genre, b.Description, b.CoverURL)
		if err != nil {
			return err
		}
	}

	// 3. Добавляем отзывы
	reviews := []Review{
		{Username: "alexey", BookTitle: "Мастер и Маргарита", Rating: 5, Text: "Гениальная книга! Перечитывал несколько раз."},
		{Username: "maria", BookTitle: "Мастер и Маргарита", Rating: 5, Text: "Любимый роман. Булгаков создал нечто уникальное."},
		{Username: "maria", BookTitle: "1984", Rating: 4, Text: "Пугающе актуально. Стоит прочитать каждому."},
		{Username: "dmitry", BookTitle: "Преступление и наказание", Rating: 5, Text: "Тяжелая, но очень глубокая книга."},
		{Username: "alexey", BookTitle: "Гарри Поттер и философский камень", Rating: 4, Text: "Вернулся в детство. Отличное фэнтези."},
		{Username: "dmitry", BookTitle: "Три товарища", Rating: 5, Text: "Ремарк - гений. Трогательно до слез."},
	}

	for _, r := range reviews {
		_, err := db.Exec(`
			INSERT INTO reviews (username, book_title, rating, text)
			VALUES (?, ?, ?, ?)`,
			r.Username, r.BookTitle, r.Rating, r.Text)
		if err != nil {
			return err
		}
	}

	log.Printf("✅ Добавлено: %d пользователей, %d книг, %d отзывов",
		len(users), len(books), len(reviews))

	return nil
}

// ===== ЗАПРОСЫ К БАЗЕ =====

// GetAllBooks - получить все книги
func GetAllBooks(db *sql.DB) ([]Book, error) {
	rows, err := db.Query("SELECT id, title, author, year, genre, description, cover_url FROM books ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Genre, &b.Description, &b.CoverURL)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

// GetBookByID - получить книгу по ID
func GetBookByID(db *sql.DB, id int) (*Book, error) {
	var b Book
	err := db.QueryRow(`
		SELECT id, title, author, year, genre, description, cover_url
		FROM books WHERE id = ?`, id).
		Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Genre, &b.Description, &b.CoverURL)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetAllReviews - получить все отзывы
func GetAllReviews(db *sql.DB) ([]Review, error) {
	rows, err := db.Query("SELECT id, username, book_title, rating, text FROM reviews ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		err := rows.Scan(&r.ID, &r.Username, &r.BookTitle, &r.Rating, &r.Text)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}

// AddReview - добавить новый отзыв
func AddReview(db *sql.DB, username, bookTitle string, rating int, text string) error {
	_, err := db.Exec(`
		INSERT INTO reviews (username, book_title, rating, text)
		VALUES (?, ?, ?, ?)`,
		username, bookTitle, rating, text)
	return err
}