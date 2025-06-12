package repositories

import "github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"

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
	query := r.metaStore.DB.Order("created_at DESC")

	if cursor != "" {
		query = query.Where("created_at < ?", cursor)
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
	query := r.metaStore.DB.Where("creator = ?", creator).Order("created_at DESC")

	if cursor != "" {
		query = query.Where("created_at < ?", cursor)
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
	// bytes := make([]byte, 16)
	// rand.Read(bytes)
	// return hex.EncodeToString(bytes)
	return helper.GenerateTID()
}

// GetMomentReplies 获取指定 moment 的直接回复
func (r *MomentRepository) GetMomentReplies(momentID string) ([]*Moment, error) {
	var replies []*Moment
	if err := r.metaStore.DB.Where("reply_parent_id = ? AND deleted = ?", momentID, false).
		Order("created_at ASC").Find(&replies).Error; err != nil {
		return nil, err
	}
	return replies, nil
}

// GetMomentAncestors 获取指定 moment 的所有祖先链（包括自己），使用递归 SQL
func (r *MomentRepository) GetMomentAncestors(momentID string, maxDepth int) ([]*Moment, error) {
	var ancestors []*Moment

	// 使用 WITH RECURSIVE 查询祖先链
	query := `
		WITH RECURSIVE ancestors AS (
			-- 基础情况：起始 moment
			SELECT id, uri, cid, text, facets, reply_root_id, reply_parent_id,
				   langs, tags, created_at, updated_at, indexed_at, creator, deleted, 0 as depth
			FROM moments
			WHERE id = ? AND deleted = false

			UNION ALL

			-- 递归情况：查找父级 moment
			SELECT m.id, m.uri, m.cid, m.text, m.facets, m.reply_root_id, m.reply_parent_id,
				   m.langs, m.tags, m.created_at, m.updated_at, m.indexed_at, m.creator, m.deleted, a.depth + 1
			FROM moments m
			INNER JOIN ancestors a ON m.id = a.reply_parent_id
			WHERE m.deleted = false AND a.depth < ?
		)
		SELECT * FROM ancestors ORDER BY depth DESC
	`

	if err := r.metaStore.DB.Raw(query, momentID, maxDepth).Scan(&ancestors).Error; err != nil {
		return nil, err
	}

	return ancestors, nil
}

// GetMomentDescendants 获取指定 moment 的所有后代回复，使用递归 SQL
func (r *MomentRepository) GetMomentDescendants(momentID string, maxDepth int) ([]*Moment, error) {
	var descendants []*Moment

	// 使用 WITH RECURSIVE 查询后代树
	query := `
		WITH RECURSIVE descendants AS (
			-- 基础情况：起始 moment
			SELECT id, uri, cid, text, facets, reply_root_id, reply_parent_id,
				   langs, tags, created_at, updated_at, indexed_at, creator, deleted, 0 as depth
			FROM moments
			WHERE id = ? AND deleted = false

			UNION ALL

			-- 递归情况：查找子级 moment
			SELECT m.id, m.uri, m.cid, m.text, m.facets, m.reply_root_id, m.reply_parent_id,
				   m.langs, m.tags, m.created_at, m.updated_at, m.indexed_at, m.creator, m.deleted, d.depth + 1
			FROM moments m
			INNER JOIN descendants d ON m.reply_parent_id = d.id
			WHERE m.deleted = false AND d.depth < ?
		)
		SELECT * FROM descendants ORDER BY depth ASC, created_at ASC
	`

	if err := r.metaStore.DB.Raw(query, momentID, maxDepth).Scan(&descendants).Error; err != nil {
		return nil, err
	}

	return descendants, nil
}

// GetMomentThread 获取指定 moment 的完整 thread（祖先链 + 后代树）
func (r *MomentRepository) GetMomentThread(momentID string, ancestorDepth, descendantDepth int) ([]*Moment, error) {
	// 获取祖先链（不包括自己）
	ancestors, err := r.GetMomentAncestors(momentID, ancestorDepth)
	if err != nil {
		return nil, err
	}

	// 获取后代树（包括自己）
	descendants, err := r.GetMomentDescendants(momentID, descendantDepth)
	if err != nil {
		return nil, err
	}

	// 合并结果，去重（祖先链的第一个就是目标 moment）
	allMoments := make([]*Moment, 0, len(ancestors)+len(descendants))

	// 添加祖先链（除了第一个，因为第一个是目标 moment）
	if len(ancestors) > 1 {
		allMoments = append(allMoments, ancestors[1:]...)
	}

	// 添加后代树（包括目标 moment）
	allMoments = append(allMoments, descendants...)

	return allMoments, nil
}

func (r *MomentRepository) GetRootMoment(momentID string) (*Moment, error) {
	var moment Moment
	if err := r.metaStore.DB.Where("id = ?", momentID).First(&moment).Error; err != nil {
		return nil, err
	}

	// 如果没有 reply_root_id，则当前 moment 就是根
	if moment.ReplyRootID == "" {
		return &moment, nil
	}

	// 获取根 moment
	var rootMoment Moment
	if err := r.metaStore.DB.Where("id = ?", moment.ReplyRootID).First(&rootMoment).Error; err != nil {
		return nil, err
	}

	return &rootMoment, nil
}

func (r *MomentRepository) CreateLike(like *Like) error {
	return r.metaStore.DB.Create(like).Error
}

func (r *MomentRepository) DeleteLike(likeURI string) error {
	return r.metaStore.DB.Where("uri = ?", likeURI).Delete(&Like{}).Error
}
