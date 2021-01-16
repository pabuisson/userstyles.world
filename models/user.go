package models

import (
	"errors"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model `json:"-"`
	Username   string `gorm:"unique;not null"`
	Email      string `gorm:"unique;not null"`
	Password   string
}

func FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	user := new(User)

	if res := db.Where("email = ?", email).First(&user); res.Error != nil {
		return nil, res.Error
	}

	if user.ID == 0 {
		return user, errors.New("User not found.")
	}

	return user, nil
}
