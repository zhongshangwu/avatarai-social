package database

import (
	"gorm.io/gorm"
)

func InsertOAuthAuthRequest(db *gorm.DB, oar *OAuthAuthRequest) error {
	return db.Create(oar).Error
}

func GetOAuthAuthRequest(db *gorm.DB, state string) (*OAuthAuthRequest, error) {
	var oauthAuthRequest OAuthAuthRequest
	if err := db.Where("state = ?", state).First(&oauthAuthRequest).Error; err != nil {
		return nil, err
	}
	return &oauthAuthRequest, nil
}

func DeleteOAuthAuthRequest(db *gorm.DB, state string) error {
	return db.Where("state = ?", state).Delete(&OAuthAuthRequest{}).Error
}

func SaveOAuthSession(db *gorm.DB, session *OAuthSession) error {
	return db.Create(session).Error
}

func GetOAuthSessionByDID(db *gorm.DB, did string) (*OAuthSession, error) {
	var session OAuthSession
	if err := db.Where("did = ?", did).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func GetOauthSessionByID(db *gorm.DB, id uint) (*OAuthSession, error) {
	var session OAuthSession
	if err := db.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func UpdateOAuthSession(db *gorm.DB, session *OAuthSession) error {
	updates := map[string]interface{}{
		"access_token":          session.AccessToken,
		"refresh_token":         session.RefreshToken,
		"dpop_authserver_nonce": session.DpopAuthserverNonce,
	}

	return db.Model(&OAuthSession{}).
		Where("did = ?", session.Did).
		Updates(updates).
		Error
}

func DeleteOAuthSessionByDID(db *gorm.DB, did string) error {
	return db.Where("did = ?", did).Delete(&OAuthSession{}).Error
}

func GetOrCreateAvatar(db *gorm.DB, did string, handle string, pdsURL string) (*Avatar, error) {
	var avatar Avatar
	err := db.Where(Avatar{Did: did}).Assign(Avatar{Handle: handle, PdsUrl: pdsURL}).FirstOrCreate(&avatar).Error
	if err != nil {
		return nil, err
	}

	if avatar.Handle != handle || avatar.PdsUrl != pdsURL {
		updates := map[string]interface{}{
			"handle":  handle,
			"pds_url": pdsURL,
		}
		if err := db.Model(&avatar).Updates(updates).Error; err != nil {
			return nil, err
		}
		avatar.Handle = handle
		avatar.PdsUrl = pdsURL
	}

	return &avatar, nil
}

func SaveSession(db *gorm.DB, session *Session) error {
	return db.Create(session).Error
}

func GetSessionByID(db *gorm.DB, id string) (*Session, error) {
	var session Session
	if err := db.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func UpdateOAuthSessionDpopPdsNonce(db *gorm.DB, did string, newNonce string) error {
	updates := map[string]interface{}{
		"dpop_pds_nonce": newNonce,
	}

	return db.Model(&OAuthSession{}).
		Where("did = ?", did).
		Updates(updates).
		Error
}

func SaveOAuthCode(db *gorm.DB, oauthCode *OAuthCode) error {
	return db.Create(oauthCode).Error
}

func GetOAuthCode(db *gorm.DB, code string) (*OAuthCode, error) {
	var oauthCode OAuthCode
	if err := db.Where("code = ?", code).First(&oauthCode).Error; err != nil {
		return nil, err
	}
	return &oauthCode, nil
}
