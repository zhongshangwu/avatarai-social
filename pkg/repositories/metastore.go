package repositories

import (
	"context"

	"gorm.io/gorm"
)

type MetaStore struct {
	DB *gorm.DB

	// Repositories
	UserRepo    *UserRepository
	OAuthRepo   *OAuthRepository
	MessageRepo *MessageRepository
	MomentRepo  *MomentRepository
	AtpRepo     *AtpRepository
	FileRepo    *FileRepository
}

func NewMetaStore(db *gorm.DB) *MetaStore {
	metaStore := &MetaStore{DB: db}

	// 初始化所有 repositories
	metaStore.UserRepo = NewUserRepository(metaStore)
	metaStore.OAuthRepo = NewOAuthRepository(metaStore)
	metaStore.MessageRepo = NewMessageRepository(metaStore)
	metaStore.MomentRepo = NewMomentRepository(metaStore)
	metaStore.AtpRepo = NewAtpRepository(metaStore)
	metaStore.FileRepo = NewFileRepository(metaStore)

	return metaStore
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

		// files
		&UploadFile{},
	)
}

func (ms *MetaStore) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return ms.DB.WithContext(ctx).Transaction(fn)
}
