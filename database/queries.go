package database

import "database/sql"

func GetUserByTelegramID(tgID int64) (*User, error) {
	var user User
	err := DB.Get(&user, "SELECT * FROM users WHERE telegram_id = $1", tgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func CreateUser(tgID int64, firstName, username, phone string) error {
	_, err := DB.Exec(`
		INSERT INTO users (telegram_id, first_name, username, phone_number, status) 
		VALUES ($1, $2, $3, $4, 'active')
		ON CONFLICT (telegram_id) DO UPDATE SET phone_number = $4, status = 'active'
	`, tgID, firstName, username, phone)
	return err
}

func UpdateUserStatus(tgID int64, status string) error {
	_, err := DB.Exec("UPDATE users SET status = $1 WHERE telegram_id = $2", status, tgID)
	return err
}

func UpdateUserSecondaryPhone(tgID int64, phone string) error {
	_, err := DB.Exec("UPDATE users SET secondary_phone = $1 WHERE telegram_id = $2", phone, tgID)
	return err
}

func GetContent(buttonName string) (*Content, error) {
	var content Content
	err := DB.Get(&content, "SELECT * FROM contents WHERE button_name = $1", buttonName)
	return &content, err
}

func UpdateContent(buttonName, textContent, mediaFileID, mediaType string) error {
	_, err := DB.Exec(`
		INSERT INTO contents (button_name, text_content, media_file_id, media_type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (button_name) DO UPDATE 
		SET text_content = $2, media_file_id = $3, media_type = $4
	`, buttonName, textContent, mediaFileID, mediaType)
	return err
}

func GetAllChannels() ([]Channel, error) {
	var channels []Channel
	err := DB.Select(&channels, "SELECT * FROM channels")
	return channels, err
}

func AddChannel(channelID int64, url, name string) error {
	_, err := DB.Exec(`
		INSERT INTO channels (channel_id, url, name) VALUES ($1, $2, $3)
		ON CONFLICT (channel_id) DO UPDATE SET url = $2, name = $3
	`, channelID, url, name)
	return err
}

func DeleteChannel(channelID int64) error {
	_, err := DB.Exec("DELETE FROM channels WHERE channel_id = $1", channelID)
	return err
}

// Analytics
type Stats struct {
	TotalUsers     int
	ActiveUsers    int
	BlockedUsers   int
	JoinedToday    int
	JoinedWeek     int
	JoinedMonth    int
	Joined3Months  int
}

func GetStats() (*Stats, error) {
	var stats Stats
	err := DB.Get(&stats, `
		SELECT 
			COUNT(*) as totalusers,
			COUNT(*) FILTER (WHERE status = 'active') as activeusers,
			COUNT(*) FILTER (WHERE status = 'blocked') as blockedusers,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '1 day') as joinedtoday,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days') as joinedweek,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '1 month') as joinedmonth,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '3 months') as joined3months
		FROM users
	`)
	return &stats, err
}

func GetAllUsersTelegramIDs() ([]int64, error) {
	var ids []int64
	err := DB.Select(&ids, "SELECT telegram_id FROM users WHERE status = 'active'")
	return ids, err
}

func IsAdmin(tgID int64) bool {
	var id int64
	err := DB.Get(&id, "SELECT telegram_id FROM admins WHERE telegram_id = $1", tgID)
	return err == nil
}

func AddAdmin(tgID int64) error {
	_, err := DB.Exec("INSERT INTO admins (telegram_id) VALUES ($1) ON CONFLICT DO NOTHING", tgID)
	return err
}

func RemoveAdmin(tgID int64) error {
	_, err := DB.Exec("DELETE FROM admins WHERE telegram_id = $1", tgID)
	return err
}

func GetAllAdmins() ([]int64, error) {
	var ids []int64
	err := DB.Select(&ids, "SELECT telegram_id FROM admins")
	return ids, err
}

func GetAllUsers() ([]User, error) {
	var users []User
	err := DB.Select(&users, "SELECT * FROM users ORDER BY created_at DESC")
	return users, err
}