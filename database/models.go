package database

import (
	"database/sql"
	"time"
)

type User struct {
	ID             int            `db:"id"`
	TelegramID     int64          `db:"telegram_id"`
	FirstName      string         `db:"first_name"`
	Username       string         `db:"username"`
	PhoneNumber    string         `db:"phone_number"`
	SecondaryPhone sql.NullString `db:"secondary_phone"`
	Status         string         `db:"status"` // "active" or "blocked"
	CreatedAt      time.Time      `db:"created_at"`
}

type Content struct {
	ButtonName  string `db:"button_name"`
	TextContent string `db:"text_content"`
	MediaFileID string `db:"media_file_id"`
	MediaType   string `db:"media_type"` // "text", "photo", "video", "voice", "document"
}

type Channel struct {
	ID        int    `db:"id"`
	ChannelID int64  `db:"channel_id"`
	URL       string `db:"url"`
	Name      string `db:"name"`
}