package events

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 定义一个具体的消息类型用于测试
type SimpleMessage struct {
	id      string
	msgType string
	content string
}

func (m SimpleMessage) ID() string {
	return m.id
}

func (m SimpleMessage) Type() string {
	return m.msgType
}

// 定义一个具体的 Actor 实现，包含自己的状态
type CounterActor struct {
	*BaseActor[SimpleMessage]

	mu    sync.RWMutex
	count int
}

// 创建 CounterActor 的工厂函数
func NewCounterActor(id string, options ...ActorOption[SimpleMessage]) *CounterActor {
	baseActor := NewActor(id, options...)

	actor := &CounterActor{
		BaseActor: baseActor,
		count:     0,
	}

	// 注册消息处理器
	actor.RegisterHandler("greeting", actor.handleGreeting)
	actor.RegisterHandler("reset", actor.handleReset)

	return actor
}

// 获取计数器值
func (a *CounterActor) GetCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.count
}

// 设置计数器值
func (a *CounterActor) SetCount(newCount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count = newCount
}

func (a *CounterActor) IncrCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count++
	return a.count
}

// 处理问候消息
func (a *CounterActor) handleGreeting(actorCtx ActorContext[SimpleMessage], msg SimpleMessage) error {
	// 获取当前计数
	currentCount := a.IncrCount()

	// 更新计数
	// a.SetCount(currentCount + 1)

	// 处理完后，将结果发布到 outbox
	response := SimpleMessage{
		id:      "response-" + msg.ID(),
		msgType: "greeting-response",
		content: fmt.Sprintf("已收到您的问候: %s (这是第 %d 次问候)", msg.content, currentCount),
	}

	return actorCtx.Actor.PublishToOutbox(actorCtx.Context, response)
}

// 处理重置消息
func (a *CounterActor) handleReset(actorCtx ActorContext[SimpleMessage], msg SimpleMessage) error {
	a.SetCount(0)

	response := SimpleMessage{
		id:      "response-" + msg.ID(),
		msgType: "reset-response",
		content: "计数器已重置",
	}

	return actorCtx.Actor.PublishToOutbox(actorCtx.Context, response)
}

// 测试 Actor 的基本功能
func TestActorBasicFunctionality(t *testing.T) {
	// 创建一个 Actor
	myActor := NewCounterActor(
		"counter-actor",
		ActorWithInboxCapacity[SimpleMessage](10),
		ActorWithOutboxCapacity[SimpleMessage](20),
		ActorWithGracePeriod[SimpleMessage](3*time.Second),
	)

	// 启动 Actor
	ctx := context.Background()
	err := myActor.Start(ctx)
	assert.NoError(t, err, "Actor 启动应该成功")

	// 测试：发送问候消息到 inbox
	err = myActor.Send(ctx, SimpleMessage{
		id:      "msg-1",
		msgType: "greeting",
		content: "你好，Actor!",
	})
	assert.NoError(t, err, "发送消息应该成功")

	// 等待消息处理完成
	time.Sleep(100 * time.Millisecond)

	// 测试：从 outbox 接收消息
	response, err := myActor.ReceiveFromOutbox(ctx)
	assert.NoError(t, err, "接收回复应该成功")
	assert.Equal(t, "greeting-response", response.Type(), "回复消息类型应该是 greeting-response")
	assert.Equal(t, "response-msg-1", response.ID(), "回复消息 ID 应该是 response-msg-1")
	assert.Contains(t, response.content, "已收到您的问候", "回复消息内容应该包含问候确认")
	assert.Contains(t, response.content, "这是第 1 次问候", "回复消息应该包含计数信息")

	// 测试：检查计数是否正确更新
	assert.Equal(t, 1, myActor.GetCount(), "计数应该增加到 1")

	// 测试：发送重置消息
	err = myActor.Send(ctx, SimpleMessage{
		id:      "msg-2",
		msgType: "reset",
		content: "重置计数器",
	})
	assert.NoError(t, err, "发送重置消息应该成功")

	// 等待消息处理完成
	time.Sleep(100 * time.Millisecond)

	// 测试：从 outbox 接收重置消息的回复
	response, err = myActor.ReceiveFromOutbox(ctx)
	assert.NoError(t, err, "接收重置回复应该成功")
	assert.Equal(t, "reset-response", response.Type(), "回复消息类型应该是 reset-response")
	assert.Equal(t, "response-msg-2", response.ID(), "回复消息 ID 应该是 response-msg-2")
	assert.Contains(t, response.content, "计数器已重置", "回复消息应该确认重置")

	// 测试：检查计数是否已重置
	assert.Equal(t, 0, myActor.GetCount(), "计数应该重置为 0")

	// 停止 Actor
	err = myActor.Stop()
	assert.NoError(t, err, "Actor 停止应该成功")
	assert.True(t, myActor.IsStopped(), "Actor 应该处于停止状态")
}

// 测试 Actor 在停止后的行为
func TestActorAfterStopped(t *testing.T) {
	myActor := NewCounterActor("stopped-actor")
	ctx := context.Background()

	// 启动然后立即停止
	myActor.Start(ctx)
	myActor.Stop()

	// 测试：向已停止的 Actor 发送消息应该失败
	err := myActor.Send(ctx, SimpleMessage{
		id:      "msg-3",
		msgType: "greeting",
		content: "这条消息不应该被处理",
	})
	assert.Error(t, err, "向已停止的 Actor 发送消息应该失败")
	assert.Equal(t, ErrActorStopped, err, "错误应该是 ErrActorStopped")

	// 测试：从已停止的 Actor 接收消息应该失败
	_, err = myActor.ReceiveFromOutbox(ctx)
	assert.Error(t, err, "从已停止的 Actor 接收消息应该失败")
	assert.Equal(t, ErrActorStopped, err, "错误应该是 ErrActorStopped")
}

// 测试 Actor 的超时行为
func TestActorTimeout(t *testing.T) {
	myActor := NewCounterActor("timeout-actor")
	ctx := context.Background()
	myActor.Start(ctx)

	// 设置一个非常短的超时
	veryShortTimeout := 1 * time.Nanosecond

	// 测试：使用极短超时从空的 outbox 接收消息应该超时
	_, err := myActor.ReceiveFromOutboxWithTimeout(ctx, veryShortTimeout)
	assert.Error(t, err, "接收应该超时")
	assert.Equal(t, ErrActorTimeout, err, "错误应该是 ErrActorTimeout")

	myActor.Stop()
}

func TestActorConcurrency(t *testing.T) {
	myActor := NewCounterActor("concurrent-actor",
		ActorWithInboxCapacity[SimpleMessage](200),  // 增加容量
		ActorWithOutboxCapacity[SimpleMessage](200), // 增加容量
	)
	ctx := context.Background()
	myActor.Start(ctx)

	// 并发发送多条消息
	messageCount := 100
	var wg sync.WaitGroup
	wg.Add(messageCount)

	for i := 0; i < messageCount; i++ {
		go func(idx int) {
			defer wg.Done()
			err := myActor.Send(ctx, SimpleMessage{
				id:      fmt.Sprintf("concurrent-msg-%d", idx),
				msgType: "greeting",
				content: fmt.Sprintf("并发消息 #%d", idx),
			})
			assert.NoError(t, err, "并发发送消息应该成功")
		}(i)
	}

	// 等待所有消息发送完成
	wg.Wait()

	// 增加等待时间，确保所有消息都被处理
	time.Sleep(2 * time.Second) // 从 500ms 增加到 1s

	// 测试：计数应该等于消息数量
	assert.Equal(t, messageCount, myActor.GetCount(), "计数应该等于发送的消息数量")

	// 测试：应该能从 outbox 接收到所有回复
	receivedCount := 0
	for i := 0; i < messageCount; i++ {
		response, err := myActor.ReceiveFromOutboxWithTimeout(ctx, 100*time.Millisecond)
		if err == nil {
			receivedCount++
			assert.Equal(t, "greeting-response", response.Type(), "回复消息类型应该是 greeting-response")
		}
	}

	assert.Equal(t, messageCount, receivedCount, "应该收到与发送消息数量相等的回复")

	myActor.Stop()
}

// 测试未注册的消息类型
func TestUnregisteredMessageType(t *testing.T) {
	// 创建一个自定义错误处理器来捕获错误
	var lastError error
	errorHandler := func(err error) {
		lastError = err
	}

	myActor := NewCounterActor(
		"unregistered-type-actor",
		ActorWithErrorHandler[SimpleMessage](errorHandler),
	)
	ctx := context.Background()
	myActor.Start(ctx)

	// 发送一个未注册类型的消息
	err := myActor.Send(ctx, SimpleMessage{
		id:      "unknown-type-msg",
		msgType: "unknown-type", // 这个类型没有注册处理器
		content: "这条消息没有对应的处理器",
	})
	assert.NoError(t, err, "发送消息应该成功，即使类型未注册")

	// 等待错误处理器被调用
	time.Sleep(100 * time.Millisecond)

	// 检查错误处理器是否捕获到了正确的错误
	assert.NotNil(t, lastError, "错误处理器应该捕获到错误")
	assert.Contains(t, lastError.Error(), "event handler not found", "错误应该指示处理器未找到")

	myActor.Stop()
}

// 测试中间件功能
func TestActorMiddleware(t *testing.T) {
	// 创建一个记录调用的中间件
	callCounter := 0
	middleware := func(next Handler[SimpleMessage]) Handler[SimpleMessage] {
		return func(actorCtx ActorContext[SimpleMessage], msg SimpleMessage) error {
			callCounter++
			return next(actorCtx, msg)
		}
	}

	myActor := NewCounterActor(
		"middleware-actor",
		ActorWithMiddleware(middleware),
	)
	ctx := context.Background()
	myActor.Start(ctx)

	// 发送消息
	err := myActor.Send(ctx, SimpleMessage{
		id:      "middleware-test-msg",
		msgType: "greeting",
		content: "测试中间件",
	})
	assert.NoError(t, err, "发送消息应该成功")

	// 等待消息处理完成
	time.Sleep(100 * time.Millisecond)

	// 检查中间件是否被调用
	assert.Equal(t, 1, callCounter, "中间件应该被调用一次")

	myActor.Stop()
}
