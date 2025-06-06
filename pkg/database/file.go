package database

import (
	"time"

	"gorm.io/gorm"
)

func CreateUploadFile(db *gorm.DB, file *UploadFile) error {
	file.CreatedAt = time.Now().UnixMilli()
	return db.Create(file).Error
}

func GetUploadFileByID(db *gorm.DB, id string) (*UploadFile, error) {
	var file UploadFile
	if err := db.Where("id = ?", id).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func GetUploadFileByCID(db *gorm.DB, cid string) (*UploadFile, error) {
	var file UploadFile
	if err := db.Where("cid = ?", cid).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func GetUploadFilesByCreator(db *gorm.DB, createdBy string, limit int, offset int) ([]*UploadFile, error) {
	var files []*UploadFile
	query := db.Where("created_by = ?", createdBy).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func UpdateUploadFile(db *gorm.DB, id string, updates map[string]interface{}) error {
	return db.Model(&UploadFile{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteUploadFile(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&UploadFile{}).Error
}

func GetUploadFileStats(db *gorm.DB, createdBy string) (map[string]interface{}, error) {
	var stats struct {
		TotalCount int64 `json:"total_count"`
		TotalSize  int64 `json:"total_size"`
	}

	if err := db.Model(&UploadFile{}).
		Where("created_by = ?", createdBy).
		Select("COUNT(*) as total_count, COALESCE(SUM(size), 0) as total_size").
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_count": stats.TotalCount,
		"total_size":  stats.TotalSize,
	}, nil
}
