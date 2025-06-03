package database

import "gorm.io/gorm"

func InsertMessage(db *gorm.DB, message *Message) error {
	return db.Create(message).Error
}

func InsertAgentMessage(db *gorm.DB, message *AIChatMessage) error {
	return db.Create(message).Error
}
