package main

import (
	"fmt"

	"gorm.io/gorm"
)

func db_init(db *gorm.DB) {
	db.Exec("DELETE FROM qr_codes")
	db.Exec("INSERT INTO qr_codes (content) VALUES ('JavaScript')")
	db.Exec("INSERT INTO qr_codes (content) VALUES ('TypeScript')")
	db.Exec("INSERT INTO qr_codes (content) VALUES ('Ruby')")
	db.Exec("INSERT INTO qr_codes (content) VALUES ('Golang')")

	fmt.Println("db_init")
}

// go run main.go db_init.go