package chat

import (
	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
)

type Session interface {
	Memory() memory.Memory
	Close()
}
