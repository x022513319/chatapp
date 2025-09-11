package store

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"time"
)

func (s *Store) ListMessages(ctx context.Context, roomID int64, limit int, beforeTS *time.Time, beforeID *int64) ([]models.Message, error) {
	var rows *sql.Rows
	var err error
	if beforeTS == nil || beforeID == nil {
		rows, err = s.DB.QueryContext(ctx, `
			SELECT id, room_id, user_id, content, created_at
			FROM messages
			WHERE room_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`, roomID, limit)
	} else {
		rows, err = s.DB.QueryContext(ctx, `
			SELECT id, room_id, user_id, content, created_at
			FROM messages
			WHERE room_id = $1
			  AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`, roomID, *beforeTS, *beforeID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (s *Store) HasOlderMessages(ctx context.Context, roomID int64, ts time.Time, id int64) (bool, error) {
	var has bool
	err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM messages
			WHERE room_id = $1 AND (created_at, id) < ($2, $3)
		)
	`, roomID, ts, id).Scan(&has)
	return has, err
}

func (s *Store) CreateMessage(ctx context.Context, roomID, userID int64, content string) (models.Message, error) {
	var m models.Message
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO messages (room_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, room_id, user_id, content, created_at
	`, roomID, userID, content).
		Scan(&m.ID, &m.RoomID, &m.UserID, &m.Content, &m.CreatedAt)
	return m, err
}
