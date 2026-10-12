package config

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	for _, variable := range []struct {
		name  string
		value string
	}{
		{name: "DB_USER", value: user},
		{name: "DB_PASSWORD", value: password},
		{name: "DB_HOST", value: host},
		{name: "DB_PORT", value: port},
		{name: "DB_NAME", value: name},
	} {
		if variable.value == "" {
			return nil, fmt.Errorf("%s is empty; set it in .env or the environment", variable.name)
		}
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		port,
		name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Database connected successfully")

	return db, nil
}
