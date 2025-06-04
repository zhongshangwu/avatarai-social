package database

import (
	"context"

	"gorm.io/gorm"
)

type MetaStore struct {
	DB *gorm.DB
}

func NewMetaStore(db *gorm.DB) *MetaStore {
	return &MetaStore{DB: db}
}

func (ms *MetaStore) Init() error {
	ms.DB.Set("gorm:table_options", "WITHOUT ROWID")
	return ms.DB.AutoMigrate(
		&OAuthAuthRequest{},
		&OAuthSession{},
		&OAuthCode{},
		&Session{},
		&Avatar{},
		// &AvatarIntegrate{},
		// &AvatarMCPServer{},
		// &AvatarBsky{},
		// &AvatarResponseAPI{},
		&Moment{},
		&MomentImage{},
		&MomentVideo{},
		&MomentExternal{},

		// atp
		&AtpRecord{},

		// messages
		&Room{},
		&UserRoomStatus{},
		&Message{},
		&Thread{},
		&AgentMessage{},
		&AgentMessageItem{},
	)
}

func (ms *MetaStore) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return ms.DB.WithContext(ctx).Transaction(fn)
}
