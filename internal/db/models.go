package db

import (
	"time"
)

type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type RefreshToken struct {
	ID        string
	UserID    int64
	Token     string
	DeviceID  string
	ExpiresAt time.Time
}
