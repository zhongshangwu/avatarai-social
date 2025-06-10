package repositories

import (
	"crypto/rand"
	"encoding/hex"
)

type MomentRepository struct {
	metaStore *MetaStore
}

func NewMomentRepository(metastore *MetaStore) *MomentRepository {
	return &MomentRepository{
		metaStore: metastore,
	}
}

func (r *MomentRepository) CreateMoment(moment *Moment) error {
	return r.metaStore.DB.Create(moment).Error
}

func (r *MomentRepository) GetMomentByURI(uri string) (*Moment, error) {
	var moment Moment
	if err := r.metaStore.DB.Where("uri = ?", uri).First(&moment).Error; err != nil {
		return nil, err
	}
	return &moment, nil
}

func (r *MomentRepository) GetMomentByID(id string) (*Moment, error) {
	var moment Moment
	if err := r.metaStore.DB.Where("id = ?", id).First(&moment).Error; err != nil {
		return nil, err
	}
	return &moment, nil
}

func (r *MomentRepository) GetLatestMomentURIs(limit int, cursor string) ([]string, error) {
	var moments []*Moment
	query := r.metaStore.DB.Order("sort_at DESC")

	if cursor != "" {
		query = query.Where("sort_at < ?", cursor)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&moments).Error; err != nil {
		return nil, err
	}

	uris := make([]string, 0, len(moments))
	for _, moment := range moments {
		uris = append(uris, moment.URI)
	}
	return uris, nil
}

func (r *MomentRepository) GetMomentsByCreator(creator string, limit int, cursor string) ([]*Moment, error) {
	var moments []*Moment
	query := r.metaStore.DB.Where("creator = ?", creator).Order("sort_at DESC")

	if cursor != "" {
		query = query.Where("sort_at < ?", cursor)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&moments).Error; err != nil {
		return nil, err
	}
	return moments, nil
}

func (r *MomentRepository) UpdateMoment(uri string, updates map[string]interface{}) error {
	return r.metaStore.DB.Model(&Moment{}).Where("uri = ?", uri).Updates(updates).Error
}

func (r *MomentRepository) DeleteMoment(uri string) error {
	return r.metaStore.DB.Where("uri = ?", uri).Delete(&Moment{}).Error
}

// MomentImage 相关操作
func (r *MomentRepository) CreateMomentImage(image *MomentImage) error {
	return r.metaStore.DB.Create(image).Error
}

func (r *MomentRepository) GetMomentImages(momentID string) ([]*MomentImage, error) {
	var images []*MomentImage
	if err := r.metaStore.DB.Where("moment_id = ?", momentID).Order("position ASC").Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *MomentRepository) GetMomentImagesByMomentIDs(momentIDs []string) (map[string][]*MomentImage, error) {
	var images []*MomentImage
	if err := r.metaStore.DB.Where("moment_id IN ?", momentIDs).Order("position ASC").Find(&images).Error; err != nil {
		return nil, err
	}
	imagesMap := make(map[string][]*MomentImage)
	for _, image := range images {
		imagesMap[image.MomentID] = append(imagesMap[image.MomentID], image)
	}
	return imagesMap, nil
}

func (r *MomentRepository) DeleteMomentImages(momentURI string) error {
	return r.metaStore.DB.Where("moment_uri = ?", momentURI).Delete(&MomentImage{}).Error
}

func (r *MomentRepository) CreateMomentVideo(video *MomentVideo) error {
	return r.metaStore.DB.Create(video).Error
}

func (r *MomentRepository) GetMomentVideo(momentURI string) (*MomentVideo, error) {
	var video MomentVideo
	if err := r.metaStore.DB.Where("moment_uri = ?", momentURI).First(&video).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *MomentRepository) GetMomentVideoByMomentIDs(momentIDs []string) (map[string]*MomentVideo, error) {
	var videos []*MomentVideo
	if err := r.metaStore.DB.Where("moment_id IN ?", momentIDs).Find(&videos).Error; err != nil {
		return nil, err
	}
	videosMap := make(map[string]*MomentVideo)
	for _, video := range videos {
		videosMap[video.MomentID] = video
	}
	return videosMap, nil
}

func (r *MomentRepository) DeleteMomentVideo(momentURI string) error {
	return r.metaStore.DB.Where("moment_uri = ?", momentURI).Delete(&MomentVideo{}).Error
}

func (r *MomentRepository) CreateMomentExternal(external *MomentExternal) error {
	return r.metaStore.DB.Create(external).Error
}

func (r *MomentRepository) GetMomentExternal(momentURI string) (*MomentExternal, error) {
	var external MomentExternal
	if err := r.metaStore.DB.Where("moment_uri = ?", momentURI).First(&external).Error; err != nil {
		return nil, err
	}
	return &external, nil
}

func (r *MomentRepository) GetMomentExternalByMomentIDs(momentIDs []string) (map[string]*MomentExternal, error) {
	var externals []*MomentExternal
	if err := r.metaStore.DB.Where("moment_id IN ?", momentIDs).Find(&externals).Error; err != nil {
		return nil, err
	}
	externalsMap := make(map[string]*MomentExternal)
	for _, external := range externals {
		externalsMap[external.MomentID] = external
	}
	return externalsMap, nil
}

func (r *MomentRepository) DeleteMomentExternal(momentURI string) error {
	return r.metaStore.DB.Where("moment_uri = ?", momentURI).Delete(&MomentExternal{}).Error
}

func (r *MomentRepository) GetMomentsByURIs(uris []string) ([]*Moment, error) {
	var moments []*Moment
	if err := r.metaStore.DB.Where("uri IN ?", uris).Find(&moments).Error; err != nil {
		return nil, err
	}
	return moments, nil
}

func (r *MomentRepository) GetBlockedDIDs(viewerDID string) ([]string, error) {
	return []string{}, nil
}

func (r *MomentRepository) GenerateMomentID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
