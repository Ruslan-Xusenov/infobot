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

	CREATE TABLE IF NOT EXISTS buttons (
		id SERIAL PRIMARY KEY,
		unique_name VARCHAR(50) UNIQUE NOT NULL,
		label VARCHAR(255) NOT NULL,
		order_num INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Insert default content if not exists
	INSERT INTO contents (button_name, text_content, media_type) VALUES 
	('bepul_shaxsiy_brend', 'Bepul shaxsiy brend haqida ma''lumot...', 'text'),
	('oz_hisobimizdan', 'O''z hisobimizdan sotib berish haqida ma''lumot...', 'text'),
	('biz_haqimizda', 'Biz haqimizda ma''lumot...', 'text'),
	('zapusk_mablag', 'Zapuskka qancha mablag'' ketishi haqida ma''lumot...', 'text'),
	('shaxsiy_brend_kerak', 'Shaxsiy brend nima uchun kerakligi haqida ma''lumot...', 'text'),
	('zapusk_sotmang', 'Zapusk qilib sotmaslik haqida ma''lumot...', 'text'),
	('telegram_botimiz', 'Telegram Botimiz: @UySotPro_Bot', 'text')
	ON CONFLICT (button_name) DO NOTHING;

	INSERT INTO buttons (unique_name, label, order_num) VALUES
	('bepul_shaxsiy_brend', '1✅.Bepul shaxsiy brend', 1),
	('oz_hisobimizdan', '2✅.O''z xisobimizdan sotib berish', 2),
	('biz_haqimizda', '3✅.Biz xaqimizda', 3),
	('zapusk_mablag', '4✅.Zapuskka qancha mablag'' ketadi', 4),
	('shaxsiy_brend_kerak', '5✅.Shaxsiy brend nima uchun kerak', 5),
	('zapusk_sotmang', '6✅.Zapusk qilib sotmang', 6),
	('telegram_botimiz', '7.Telegram Botimiz @UySotPro_Bot', 7)
	ON CONFLICT (unique_name) DO NOTHING;
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
