package main

import (
	"github.com/harranali/authority"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

func main() {
	db, err := gorm.Open(sqlite.Open("../../test.db"))
	if err != nil {
		log.Fatal(err)
	}
	auth := authority.New(authority.Options{
		TablesPrefix: "authority_",
		DB:           db,
	})

	// TODO: make roles and perms
	_ = auth.CreatePermission("permission-1")
}
