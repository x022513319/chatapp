package store

import (
	"backend/internal/models"
	"context"
)

func (s *Store) ListRooms(ctx context.Context) ([]models.Room, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, name, is_public FROM rooms ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Room
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.IsPublic); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
