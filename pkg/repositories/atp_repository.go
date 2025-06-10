package repositories

import (
	"errors"

	"gorm.io/gorm"
)

type AtpRepository struct {
	metaStore *MetaStore
}

func NewAtpRepository(metastore *MetaStore) *AtpRepository {
	return &AtpRepository{
		metaStore: metastore,
	}
}

func (r *AtpRepository) GetAtpRecords(uris []string) ([]*AtpRecord, error) {
	var records []*AtpRecord
	if err := r.metaStore.DB.Where("uri IN ?", uris).Find(&records).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAsterNotFound
		}
		return nil, err
	}
	return records, nil
}

func (r *AtpRepository) GetAtpRecord(uri string) (*AtpRecord, error) {
	var record AtpRecord
	if err := r.metaStore.DB.Where("uri = ?", uri).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAsterNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *AtpRepository) InsertOrUpdateAtpRecord(record *AtpRecord) error {
	existingRecord := &AtpRecord{}
	if err := r.metaStore.DB.Where("uri = ?", record.URI).First(existingRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.metaStore.DB.Create(record).Error
		}
		return err
	}

	return r.metaStore.DB.Model(&AtpRecord{}).Where("uri = ?", record.URI).Updates(record).Error
}
