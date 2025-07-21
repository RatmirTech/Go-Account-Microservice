package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Создание пользователя
func (p *Postgres) CreateUser(ctx context.Context, email, passwordHash string) error {
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO users (email, password_hash) 
		VALUES ($1, $2)
	`, email, passwordHash)
	return err
}

// Получить пользователя по email
func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := p.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at 
		FROM users 
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Получить пользователя по ID
func (p *Postgres) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	var u User
	err := p.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Сохранение refresh-токена
func (p *Postgres) SaveRefreshToken(ctx context.Context, id uuid.UUID, userID int64, token, deviceID, ip, ua string, ttl time.Duration) error {
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (
			id, user_id, token, device_id, ip, user_agent, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW() + $7 * INTERVAL '1 second')
	`, id, userID, token, deviceID, ip, ua, int(ttl.Seconds()))
	return err
}

// Получить refresh-токен по ID
func (p *Postgres) GetRefreshTokenByID(ctx context.Context, id string) (*RefreshToken, error) {
	var rt RefreshToken
	err := p.Pool.QueryRow(ctx, `
		SELECT id, user_id, token, device_id, expires_at 
		FROM refresh_tokens 
		WHERE id = $1
	`, id).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.DeviceID, &rt.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// Получить refresh-токен по значению и device_id
func (p *Postgres) GetRefreshToken(ctx context.Context, tokenValue, deviceID string) (*RefreshToken, error) {
	var rt RefreshToken
	err := p.Pool.QueryRow(ctx, `
		SELECT id, user_id, token, device_id, expires_at 
		FROM refresh_tokens 
		WHERE token = $1 AND device_id = $2
	`, tokenValue, deviceID).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.DeviceID, &rt.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// Обновление refresh-токена
func (p *Postgres) UpdateRefreshToken(ctx context.Context, oldToken, newID, newToken string, ttl time.Duration) error {
	_, err := p.Pool.Exec(ctx, `
		UPDATE refresh_tokens 
		SET id = $1, token = $2, expires_at = NOW() + $3 * INTERVAL '1 second'
		WHERE token = $4
	`, newID, newToken, int(ttl.Seconds()), oldToken)
	return err
}

// Удаление refresh-токена по user_id и device_id
func (p *Postgres) DeleteRefreshToken(ctx context.Context, userID int64, deviceID string) error {
	_, err := p.Pool.Exec(ctx, `
		DELETE FROM refresh_tokens 
		WHERE user_id = $1 AND device_id = $2
	`, userID, deviceID)
	return err
}
