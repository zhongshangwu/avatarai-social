package repositories

import (
	"time"
)

type FileRepository struct {
	metaStore *MetaStore
}

func NewFileRepository(metastore *MetaStore) *FileRepository {
	return &FileRepository{
		metaStore: metastore,
	}
}

func (r *FileRepository) CreateUploadFile(file *UploadFile) error {
	file.CreatedAt = time.Now().UnixMilli()
	return r.metaStore.DB.Create(file).Error
}

func (r *FileRepository) GetUploadFileByID(id string) (*UploadFile, error) {
	var file UploadFile
	if err := r.metaStore.DB.Where("id = ?", id).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) GetUploadFileByBlobCID(cid string) (*UploadFile, error) {
	var file UploadFile
	if err := r.metaStore.DB.Where("blob_cid = ?", cid).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) GetUploadFilesByCreator(createdBy string, limit int, offset int) ([]*UploadFile, error) {
	var files []*UploadFile
	query := r.metaStore.DB.Where("created_by = ?", createdBy).Order("created_at DESC")

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

func (r *FileRepository) UpdateUploadFile(id string, updates map[string]interface{}) error {
	return r.metaStore.DB.Model(&UploadFile{}).Where("id = ?", id).Updates(updates).Error
}

func (r *FileRepository) DeleteUploadFile(id string) error {
	return r.metaStore.DB.Where("id = ?", id).Delete(&UploadFile{}).Error
}

func (r *FileRepository) GetUploadFileStats(createdBy string) (map[string]interface{}, error) {
	var stats struct {
		TotalCount int64 `json:"total_count"`
		TotalSize  int64 `json:"total_size"`
	}

	if err := r.metaStore.DB.Model(&UploadFile{}).
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
