package database

import (
	"errors"

	"gorm.io/gorm"
)

func GetAsterByCreatorDid(db *gorm.DB, did string) (*Avatar, error) {
	var aster Avatar
	if err := db.Where("creator_did = ? AND is_aster = ?", did, true).First(&aster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAsterNotFound
		}
		return nil, err
	}
	return &aster, nil
}

func CreateAster(db *gorm.DB, aster *Avatar) error {
	return db.Create(aster).Error
}
