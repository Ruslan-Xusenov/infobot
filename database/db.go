package database

import (
	"fmt"
	"log"

	"github.com/company/infobot/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	DB = db
	log.Println("Connected to PostgreSQL successfully.")

	return createTables(cfg)
}

func createTables(cfg *config.Config) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		first_name VARCHAR(255),
		username VARCHAR(255),
		phone_number VARCHAR(50),
		secondary_phone VARCHAR(50),
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	ALTER TABLE users ADD COLUMN IF NOT EXISTS secondary_phone VARCHAR(50);

	CREATE TABLE IF NOT EXISTS contents (
		button_name VARCHAR(50) PRIMARY KEY,
		text_content TEXT,
		media_file_id VARCHAR(255),
		media_type VARCHAR(20)
	);

	CREATE TABLE IF NOT EXISTS channels (
		id SERIAL PRIMARY KEY,
		channel_id BIGINT UNIQUE NOT NULL,
		url VARCHAR(255),
		name VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS admins (
		telegram_id BIGINT PRIMARY KEY
	);

	-- Insert default content if not exists
	INSERT INTO contents (button_name, text_content, media_type) VALUES 
	('biz_kimmiz', 'Bizning kompaniya haqida qisqacha ma''lumot...', 'text'),
	('sotuv_bolimi', 'Sotuv bo''limi kontaktlari va ma''lumotlar...', 'text'),
	('shaxsiy_brend', 'Shaxsiy brend haqida...', 'text'),
	('zapusk', 'Zapusk loyihalari...', 'text'),
	('boglanish', 'Biz bilan bog''lanish uchun kontaktlar: ...', 'text')
	ON CONFLICT (button_name) DO NOTHING;
	`

	_, err := DB.Exec(schema)
	if err != nil {
		log.Printf("Error creating tables: %v\n", err)
		return err
	}

	// Make sure the main admin is in the database
	_, err = DB.Exec("INSERT INTO admins (telegram_id) VALUES ($1) ON CONFLICT (telegram_id) DO NOTHING", cfg.AdminID)
	if err != nil {
		log.Printf("Error inserting main admin: %v\n", err)
	}

	return nil
}
