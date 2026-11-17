package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID uint `gorm:"primaryKey" json:"id"`
	gorm.Model
	Name      string    `gorm:"size:200" json:"name"`
	Email     string    `gorm:"size:320;unique" json:"email"`
	Password  string    `gorm:"column:password;size:100" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}
