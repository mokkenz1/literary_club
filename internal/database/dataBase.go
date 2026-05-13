// Package database provides database connection and data manipulation utilities.
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Event Структуры данных
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

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./bookclub.db")
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Проверяем соединение
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Создаем таблицы
	if err = createTables(); err != nil {
		return err
	}

	// Добавляем тестовые данные, если БД пустая
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	if count == 0 {
		fmt.Println("📝 Добавление тестовых данных...")
		seedData()
	}

	fmt.Println("✅ База данных готова")
	return nil
}

func createTables() error {
	schema := `
    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        event_date TEXT NOT NULL,
        location TEXT,
        created_at TEXT DEFAULT (datetime('now'))
    );

    CREATE TABLE IF NOT EXISTS reviews (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_id INTEGER NOT NULL,
        author TEXT NOT NULL,
        content TEXT NOT NULL,
        rating INTEGER DEFAULT 5,
        created_at TEXT DEFAULT (datetime('now')),
        FOREIGN KEY (event_id) REFERENCES events(id)
    );
    `

	_, err := DB.Exec(schema)
	return err
}

func seedData() {
	// Мероприятия
	events := []struct {
		title       string
		description string
		date        string
		location    string
	}{}

	for _, e := range events {
		DB.Exec(
			"INSERT INTO events (title, description, event_date, location) VALUES (?, ?, ?, ?)",
			e.title, e.description, e.date, e.location,
		)
	}

	// Отзывы
	reviews := []struct {
		eventID int
		author  string
		content string
		rating  int
	}{
		{1, "Анна Петрова", "Прекрасное обсуждение! Оруэлл актуален как никогда.", 5},
		{1, "Михаил", "Хорошая встреча, но хотелось больше времени на дискуссию.", 4},
		{2, "Елена", "Стихи звучали волшебно. Спасибо организаторам!", 5},
	}

	for _, r := range reviews {
		DB.Exec(
			"INSERT INTO reviews (event_id, author, content, rating, created_at) VALUES (?, ?, ?, ?, ?)",
			r.eventID, r.author, r.content, r.rating, time.Now().Format("2006-01-02 15:04"),
		)
	}
}

func GetEvents() ([]Event, error) {
	rows, err := DB.Query(`
        SELECT id, title, description, event_date, location 
        FROM events 
        ORDER BY event_date DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.EventDate, &e.Location); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func GetReviews() ([]Review, error) {
	rows, err := DB.Query(`
        SELECT id, event_id, author, content, rating, created_at 
        FROM reviews 
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.EventID, &r.Author, &r.Content, &r.Rating, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func AddReview(eventID int, author, content string, rating int) error {
	_, err := DB.Exec(
		"INSERT INTO reviews (event_id, author, content, rating, created_at) VALUES (?, ?, ?, ?, ?)",
		eventID, author, content, rating, time.Now().Format("2006-01-02 15:04"),
	)
	return err
}
