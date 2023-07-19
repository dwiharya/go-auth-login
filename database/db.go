// database/db.go
package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
)

// DBPool menyimpan koneksi pool untuk digunakan di seluruh aplikasi
// var DBPool *pgxpool.Pool

// User adalah struct untuk tabel users
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"-"`
}

func NewDBConnection() (*pgx.Conn, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to database successfully")
	return conn, nil
}

// CreateUsersTable membuat tabel users jika belum ada
func CreateUsersTable() error {
	conn, err := NewDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(50) NOT NULL,
            first_name VARCHAR(200) NOT NULL,
            last_name VARCHAR(200) NOT NULL,
            password VARCHAR(120) NOT NULL
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

// CreateUser menyimpan data pengguna baru ke dalam database
func CreateUser(username, firstName, lastName, password string) error {
	conn, err := NewDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
		INSERT INTO users (username, first_name, last_name, password)
		VALUES ($1, $2, $3, $4)
	`, username, firstName, lastName, password)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetUserByUsernameAndPassword mengambil data pengguna dari tabel "users" berdasarkan username dan password
func GetUserByUsernameAndPassword(username, password string) (*User, error) {
	conn, err := NewDBConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(context.Background())

	var user User
	err = conn.QueryRow(context.Background(), `
		SELECT id, username, first_name, last_name
		FROM users
		WHERE username = $1 AND password = $2
	`, username, password).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Tidak ditemukan pengguna dengan kombinasi username dan password yang benar
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// CreateLogoutHistoryTable membuat tabel logout_history jika belum ada
func CreateLogoutHistoryTable() error {
	conn, err := NewDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS logout_history (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL,
            logout_time TIMESTAMP NOT NULL
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create logout_history table: %w", err)
	}

	return nil
}

// SaveLogoutHistory menyimpan riwayat logout ke dalam database
func SaveLogoutHistory(userID int) error {
	conn, err := NewDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
        INSERT INTO logout_history (user_id, logout_time)
        VALUES ($1, NOW())
    `, userID)
	if err != nil {
		return fmt.Errorf("failed to save logout history: %w", err)
	}

	return nil
}
