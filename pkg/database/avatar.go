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

func GetAvatarByHandle(db *gorm.DB, handle string) (*Avatar, error) {
	var avatar Avatar
	if err := db.Where("handle = ?", handle).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func GetAvatarByID(db *gorm.DB, id uint) (*Avatar, error) {
	var avatar Avatar
	if err := db.Where("id = ?", id).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func GetAvatarByDID(db *gorm.DB, did string) (*Avatar, error) {
	var avatar Avatar
	if err := db.Where("did = ?", did).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func GetAvatarsByDIDs(db *gorm.DB, dids []string) ([]*Avatar, error) {
	var avatars []*Avatar
	if err := db.Where("did IN ?", dids).Find(&avatars).Error; err != nil {
		return nil, err
	}
	return avatars, nil
}
