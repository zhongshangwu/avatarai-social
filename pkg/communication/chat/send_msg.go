package chat

import (
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
)

func (actor *ChatActor) SendMsg(sendMsgEvent *messages.SendMsgEvent) (*messages.Message, error) {
	// 一个通用的 IM 消息发送流程：
	// 1. 构建消息, 格式化消息格式
	// 2. 用户状态、权限和好友关系检查 (暂时忽略)
	// 3. 内容审核 (暂时忽略)
	// 4. 存储消息
	// 5. 消息分发, Websocket 或者 IM Push 通知 (暂时忽略)
	// 6. 后处理: 更新最后会话时间、最后活跃等等 (暂时忽略)
	message, err := BuildMessageFromSendMsgEvent(sendMsgEvent)
	if err != nil {
		return nil, err
	}

	// if message.SenderID != actor.User.Did {
	// 	return nil, errors.New("invalid sender id")
	// }

	dbMessage := messages.MessageToDB(message)
	if err := database.InsertMessage(actor.DB, dbMessage); err != nil {
		return nil, err
	}

	return message, nil
}
