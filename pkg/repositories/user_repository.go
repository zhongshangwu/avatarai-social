package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	metaStore *MetaStore
}

func NewUserRepository(metastore *MetaStore) *UserRepository {
	return &UserRepository{
		metaStore: metastore,
	}
}

// Avatar 相关操作
func (r *UserRepository) GetAsterByCreatorDid(did string) (*Avatar, error) {
	var aster Avatar
	if err := r.metaStore.DB.Where("creator_did = ? AND is_aster = ?", did, true).First(&aster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAsterNotFound
		}
		return nil, err
	}
	return &aster, nil
}

func (r *UserRepository) CreateAster(aster *Avatar) error {
	return r.metaStore.DB.Create(aster).Error
}

func (r *UserRepository) GetAvatarByHandle(handle string) (*Avatar, error) {
	var avatar Avatar
	if err := r.metaStore.DB.Where("handle = ?", handle).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *UserRepository) GetAvatarByID(id uint) (*Avatar, error) {
	var avatar Avatar
	if err := r.metaStore.DB.Where("id = ?", id).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *UserRepository) GetAvatarByDID(did string) (*Avatar, error) {
	var avatar Avatar
	if err := r.metaStore.DB.Where("did = ?", did).First(&avatar).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *UserRepository) GetAvatarsByDIDs(dids []string) ([]*Avatar, error) {
	var avatars []*Avatar
	if err := r.metaStore.DB.Where("did IN ?", dids).Find(&avatars).Error; err != nil {
		return nil, err
	}
	return avatars, nil
}

func (r *UserRepository) GetUsersByDIDs(dids []string) ([]*Avatar, error) {
	var avatars []*Avatar
	if err := r.metaStore.DB.Where("did IN ?", dids).Find(&avatars).Error; err != nil {
		return nil, err
	}
	return avatars, nil
}

func (r *UserRepository) GetOrCreateAvatar(did string, handle string, pdsURL string) (*Avatar, error) {
	var avatar Avatar
	err := r.metaStore.DB.Where(Avatar{Did: did}).Assign(Avatar{Handle: handle, PdsUrl: pdsURL}).FirstOrCreate(&avatar).Error
	if err != nil {
		return nil, err
	}

	if avatar.Handle != handle || avatar.PdsUrl != pdsURL {
		updates := map[string]interface{}{
			"handle":  handle,
			"pds_url": pdsURL,
		}
		if err := r.metaStore.DB.Model(&avatar).Updates(updates).Error; err != nil {
			return nil, err
		}
		avatar.Handle = handle
		avatar.PdsUrl = pdsURL
	}

	return &avatar, nil
}

func (r *UserRepository) UpdateAvatar(did string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.metaStore.DB.Model(&Avatar{}).Where("did = ?", did).Updates(updates).Error
}

// Session 相关操作
func (r *UserRepository) SaveSession(session *Session) error {
	return r.metaStore.DB.Create(session).Error
}

func (r *UserRepository) GetSessionByID(id string) (*Session, error) {
	var session Session
	if err := r.metaStore.DB.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UserRepository) DeleteSessionByID(id string) error {
	return r.metaStore.DB.Where("id = ?", id).Delete(&Session{}).Error
}

func (r *UserRepository) UpdateSession(id string, updates map[string]interface{}) error {
	return r.metaStore.DB.Model(&Session{}).Where("id = ?", id).Updates(updates).Error
}
