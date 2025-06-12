package services

import (
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type MessageService struct {
	metaStore *repositories.MetaStore
	Converter *MessageConverter
}

func NewMessageService(metaStore *repositories.MetaStore) *MessageService {
	messageRepo := repositories.NewMessageRepository(metaStore)
	converter := NewMessageConverter(messageRepo)

	return &MessageService{
		metaStore: metaStore,
		Converter: converter,
	}
}
