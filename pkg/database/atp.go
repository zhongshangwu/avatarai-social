package database

import (
	"errors"

	"gorm.io/gorm"
)

func GetAtpRecords(db *gorm.DB, uris []string) ([]*AtpRecord, error) {
	var records []*AtpRecord
	if err := db.Where("uri IN ?", uris).Find(&records).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAsterNotFound
		}
		return nil, err
	}
	return records, nil
}

func GetAtpRecord(db *gorm.DB, uri string) (*AtpRecord, error) {
	var record AtpRecord
	if err := db.Where("uri = ?", uri).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAsterNotFound
		}
		return nil, err
	}
	return &record, nil
}

func InsertOrUpdateAtpRecord(db *gorm.DB, record *AtpRecord) error {
	existingRecord := &AtpRecord{}
	if err := db.Where("uri = ?", record.URI).First(existingRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(record).Error
		}
		return err
	}

	return db.Model(&AtpRecord{}).Where("uri = ?", record.URI).Updates(record).Error
}
