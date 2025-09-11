package models

import "time"

type Room struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type Message struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type PageInfo struct {
	NextBeforeTS *time.Time `json:"next_before_ts,omitempty"`
	NextBeforeID *int64     `json:"next_before_id,omitempty"`
	HasMore      bool       `json:"has_more"`
}

type MessageResp struct {
	Items    []Message `json:"items"`
	PageInfo PageInfo  `json:"page_info"`
}
