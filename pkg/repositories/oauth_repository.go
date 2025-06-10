package repositories

type OAuthRepository struct {
	metaStore *MetaStore
}

func NewOAuthRepository(metastore *MetaStore) *OAuthRepository {
	return &OAuthRepository{
		metaStore: metastore,
	}
}

// OAuth Auth Request 相关操作
func (r *OAuthRepository) InsertOAuthAuthRequest(oar *OAuthAuthRequest) error {
	return r.metaStore.DB.Create(oar).Error
}

func (r *OAuthRepository) GetOAuthAuthRequest(state string) (*OAuthAuthRequest, error) {
	var oauthAuthRequest OAuthAuthRequest
	if err := r.metaStore.DB.Where("state = ?", state).First(&oauthAuthRequest).Error; err != nil {
		return nil, err
	}
	return &oauthAuthRequest, nil
}

func (r *OAuthRepository) DeleteOAuthAuthRequest(state string) error {
	return r.metaStore.DB.Where("state = ?", state).Delete(&OAuthAuthRequest{}).Error
}

// OAuth Session 相关操作
func (r *OAuthRepository) SaveOAuthSession(session *OAuthSession) error {
	return r.metaStore.DB.Create(session).Error
}

func (r *OAuthRepository) GetOAuthSessionByDID(did string) (*OAuthSession, error) {
	var session OAuthSession
	if err := r.metaStore.DB.Where("did = ?", did).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *OAuthRepository) GetOAuthSessionByID(id uint) (*OAuthSession, error) {
	var session OAuthSession
	if err := r.metaStore.DB.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *OAuthRepository) UpdateOAuthSession(session *OAuthSession) error {
	updates := map[string]interface{}{
		"access_token":          session.AccessToken,
		"refresh_token":         session.RefreshToken,
		"dpop_authserver_nonce": session.DpopAuthserverNonce,
	}

	return r.metaStore.DB.Model(&OAuthSession{}).
		Where("did = ?", session.Did).
		Updates(updates).
		Error
}

func (r *OAuthRepository) UpdateOAuthSessionDpopPdsNonce(did string, newNonce string) error {
	updates := map[string]interface{}{
		"dpop_pds_nonce": newNonce,
	}

	return r.metaStore.DB.Model(&OAuthSession{}).
		Where("did = ?", did).
		Updates(updates).
		Error
}

func (r *OAuthRepository) DeleteOAuthSessionByDID(did string) error {
	return r.metaStore.DB.Where("did = ?", did).Delete(&OAuthSession{}).Error
}

// OAuth Code 相关操作
func (r *OAuthRepository) SaveOAuthCode(oauthCode *OAuthCode) error {
	return r.metaStore.DB.Create(oauthCode).Error
}

func (r *OAuthRepository) GetOAuthCode(code string) (*OAuthCode, error) {
	var oauthCode OAuthCode
	if err := r.metaStore.DB.Where("code = ?", code).First(&oauthCode).Error; err != nil {
		return nil, err
	}
	return &oauthCode, nil
}

func (r *OAuthRepository) UpdateOAuthCodeUsed(code string, used bool) error {
	return r.metaStore.DB.Model(&OAuthCode{}).
		Where("code = ?", code).
		Update("used", used).
		Error
}
