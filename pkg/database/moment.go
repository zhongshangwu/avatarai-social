package database

import (
	"encoding/base64"
	"encoding/json"

	"gorm.io/gorm"
)

func CreateMoment(db *gorm.DB, moment *Moment) error {
	return db.Create(moment).Error
}

func GetMomentByID(db *gorm.DB, id uint) (*Moment, error) {
	var moment Moment
	if err := db.Where("id = ?", id).First(&moment).Error; err != nil {
		return nil, err
	}
	return &moment, nil
}

func GetMomentByURI(db *gorm.DB, uri string) (*Moment, error) {
	var moment Moment
	if err := db.Where("uri = ?", uri).First(&moment).Error; err != nil {
		return nil, err
	}
	return &moment, nil
}

func GetMomentByCreator(db *gorm.DB, creator string) ([]*Moment, error) {
	var moments []*Moment
	if err := db.Where("creator = ?", creator).Find(&moments).Error; err != nil {
		return nil, err
	}
	return moments, nil
}

// CreateMomentImage 创建 MomentImage 记录
func CreateMomentImage(db *gorm.DB, momentImage *MomentImage) error {
	return db.Create(momentImage).Error
}

// CreateMomentVideo 创建 MomentVideo 记录
func CreateMomentVideo(db *gorm.DB, momentVideo *MomentVideo) error {
	return db.Create(momentVideo).Error
}

// CreateMomentExternal 创建 MomentExternal 记录
func CreateMomentExternal(db *gorm.DB, momentExternal *MomentExternal) error {
	return db.Create(momentExternal).Error
}

// --- Potentially add query functions for related data later ---
// func GetMomentImages(db *gorm.DB, momentURI string) ([]*MomentImage, error)
// func GetMomentVideo(db *gorm.DB, momentURI string) (*MomentVideo, error)
// func GetMomentExternal(db *gorm.DB, momentURI string) (*MomentExternal, error)

// 假设这些函数已存在或需要实现

// GetLatestMoments 获取最新的 Moments 列表
func GetLatestMoments(db *gorm.DB, limit int, cursor string) ([]*Moment, error) {
	// 实现获取最新 moments 的逻辑
	var moments []*Moment
	query := db

	if cursor != "" {
		// 解析游标数据
		var cursorData map[string]interface{}
		cursorBytes, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			json.Unmarshal(cursorBytes, &cursorData)

			if t, ok := cursorData["t"].(string); ok {
				query = query.Where("sort_at < ?", t)
			}
		}
	}

	// 按时间倒序排列，限制数量
	if err := query.Order("sort_at DESC").Limit(limit).Find(&moments).Error; err != nil {
		return nil, err
	}

	return moments, nil
}

func GetBlockedDIDs(db *gorm.DB, viewerDID string) ([]string, error) {
	return []string{}, nil
}
