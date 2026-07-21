package repository

import "gorm.io/gorm"

type Repositories struct {
	User *UserRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User: NewUserRepository(db),
	}
}
