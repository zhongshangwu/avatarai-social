package events

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 测试用事件类型
const (
	TestEventTypeA = "test.event.a"
	TestEventTypeB = "test.event.b"
	TestEventTypeC = "test.event.c"
)

// TestEvent 是一个用于测试的简单事件实现
type TestEvent struct {
	id        string
	eventType string
	payload   string
	source    string
	timestamp time.Time
	metadata  map[string]interface{}
}

func NewTestEvent(eventType string, payload string) *TestEvent {
	return &TestEvent{
		id:        fmt.Sprintf("test-%d", time.Now().UnixNano()),
		eventType: eventType,
		payload:   payload,
		source:    "test",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}
}

func (e *TestEvent) ID() string   { return e.id }
func (e *TestEvent) Type() string { return e.eventType }

// TestEventBus 基本功能测试
func TestEventBus_BasicFunctionality(t *testing.T) {
	// 创建事件总线
	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](100),
		BusWithWorkerCount[*TestEvent](2),
		BusWithErrorHandler[*TestEvent](func(err error) {
			t.Logf("Error handler called: %v", err)
		}),
	)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动事件总线
	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	// 用于等待事件处理完成
	var wg sync.WaitGroup
	wg.Add(1)

	// 记录接收到的事件
	var receivedEvent *TestEvent

	// 订阅事件
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		t.Logf("Received event: %s", event.Type())
		receivedEvent = event
		wg.Done()
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to subscribe to event: %v", err)
	}

	// 发布事件
	testEvent := NewTestEvent(TestEventTypeA, "test payload")
	if err := eventBus.Publish(ctx, testEvent); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// 等待事件处理完成
	waitWithTimeout(&wg, 2*time.Second)

	// 验证事件是否被正确接收
	if receivedEvent == nil {
		t.Fatal("Event was not received")
	}

	if receivedEvent.Type() != TestEventTypeA {
		t.Errorf("Expected event type %s, got %s", TestEventTypeA, receivedEvent.Type())
	}

	// 停止事件总线
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()

	if err := eventBus.Stop(shutdownCtx); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestEventBus_MultipleSubscriptions 测试多个订阅
func TestEventBus_MultipleSubscriptions(t *testing.T) {
	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](100),
		BusWithWorkerCount[*TestEvent](2),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	var counterA, counterB, counterC atomic.Int32
	var wg sync.WaitGroup
	wg.Add(3) // 期望3个事件处理器被调用

	// 订阅事件A
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		counterA.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event A: %v", err)
	}

	// 再次订阅事件A
	_, err = eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		counterB.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event A again: %v", err)
	}

	// 订阅通配符
	_, err = eventBus.Subscribe("*", func(ctx context.Context, event *TestEvent) error {
		counterC.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe with wildcard: %v", err)
	}

	// 发布事件
	testEvent := NewTestEvent(TestEventTypeA, "test multiple subscriptions")
	if err := eventBus.Publish(ctx, testEvent); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// 等待所有处理器被调用
	waitWithTimeout(&wg, 2*time.Second)

	// 验证计数器
	if counterA.Load() != 1 {
		t.Errorf("Expected counterA to be 1, got %d", counterA.Load())
	}
	if counterB.Load() != 1 {
		t.Errorf("Expected counterB to be 1, got %d", counterB.Load())
	}
	if counterC.Load() != 1 {
		t.Errorf("Expected counterC to be 1, got %d", counterC.Load())
	}

	// 停止事件总线
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestEventBus_Unsubscribe 测试取消订阅
func TestEventBus_Unsubscribe(t *testing.T) {
	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](100),
		BusWithWorkerCount[*TestEvent](2),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	var counter atomic.Int32

	// 订阅事件
	subID, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		counter.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event: %v", err)
	}

	// 发布第一个事件
	firstEvent := NewTestEvent(TestEventTypeA, "before unsubscribe")
	if err := eventBus.Publish(ctx, firstEvent); err != nil {
		t.Fatalf("Failed to publish first event: %v", err)
	}

	// 等待事件处理完成
	time.Sleep(500 * time.Millisecond)

	// 取消订阅
	if err := eventBus.Unsubscribe(subID); err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// 发布第二个事件
	secondEvent := NewTestEvent(TestEventTypeA, "after unsubscribe")
	if err := eventBus.Publish(ctx, secondEvent); err != nil {
		t.Fatalf("Failed to publish second event: %v", err)
	}

	// 等待一段时间，确保事件有机会被处理
	time.Sleep(500 * time.Millisecond)

	// 验证计数器
	if counter.Load() != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter.Load())
	}

	// 停止事件总线
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestEventBus_ErrorHandling 测试错误处理
func TestEventBus_ErrorHandling(t *testing.T) {
	var errorCaught atomic.Bool

	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](100),
		BusWithWorkerCount[*TestEvent](2),
		BusWithErrorHandler[*TestEvent](func(err error) {
			if err != nil && strings.Contains(err.Error(), "test error") {
				errorCaught.Store(true)
			}
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 订阅事件，处理器会返回错误
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		wg.Done()
		return errors.New("test error")
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event: %v", err)
	}

	// 发布事件
	testEvent := NewTestEvent(TestEventTypeA, "error test")
	if err := eventBus.Publish(ctx, testEvent); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// 等待事件处理完成
	waitWithTimeout(&wg, 2*time.Second)

	// 等待错误处理器有机会被调用
	time.Sleep(500 * time.Millisecond)

	// 验证错误是否被捕获
	if !errorCaught.Load() {
		t.Error("Expected error to be caught by error handler")
	}

	// 停止事件总线
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestEventBus_ConcurrentPublish 测试并发发布
func TestEventBus_ConcurrentPublish(t *testing.T) {
	eventBus := NewEventBus[*TestEvent](BusWithBufferSize[*TestEvent](1000), BusWithWorkerCount[*TestEvent](4))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	var counter atomic.Int32
	var wg sync.WaitGroup
	const numEvents = 100

	// 订阅事件
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		counter.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event: %v", err)
	}

	// 并发发布事件
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := NewTestEvent(TestEventTypeA, fmt.Sprintf("concurrent event %d", i))
			if err := eventBus.Publish(ctx, event); err != nil {
				t.Errorf("Failed to publish event %d: %v", i, err)
			}
		}(i)
	}

	// 等待所有发布完成
	wg.Wait()

	// 等待事件处理完成
	time.Sleep(1 * time.Second)

	// 验证计数器
	if counter.Load() != numEvents {
		t.Errorf("Expected counter to be %d, got %d", numEvents, counter.Load())
	}

	// 停止事件总线
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestEventBus_MultipleEventTypes 测试多种事件类型
func TestEventBus_MultipleEventTypes(t *testing.T) {
	eventBus := NewEventBus[*TestEvent](BusWithBufferSize[*TestEvent](100), BusWithWorkerCount[*TestEvent](2))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	var typeACounter, typeBCounter, typeCCounter atomic.Int32
	var wg sync.WaitGroup
	wg.Add(3) // 期望3个事件各被处理1次

	// 订阅事件A
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		typeACounter.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event A: %v", err)
	}

	// 订阅事件B
	_, err = eventBus.Subscribe(TestEventTypeB, func(ctx context.Context, event *TestEvent) error {
		typeBCounter.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event B: %v", err)
	}

	// 订阅事件C
	_, err = eventBus.Subscribe(TestEventTypeC, func(ctx context.Context, event *TestEvent) error {
		typeCCounter.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event C: %v", err)
	}

	// 发布3种不同的事件
	eventA := NewTestEvent(TestEventTypeA, "event A")
	eventB := NewTestEvent(TestEventTypeB, "event B")
	eventC := NewTestEvent(TestEventTypeC, "event C")

	if err := eventBus.Publish(ctx, eventA); err != nil {
		t.Fatalf("Failed to publish event A: %v", err)
	}
	if err := eventBus.Publish(ctx, eventB); err != nil {
		t.Fatalf("Failed to publish event B: %v", err)
	}
	if err := eventBus.Publish(ctx, eventC); err != nil {
		t.Fatalf("Failed to publish event C: %v", err)
	}

	// 等待所有事件处理完成
	waitWithTimeout(&wg, 2*time.Second)

	// 验证计数器
	if typeACounter.Load() != 1 {
		t.Errorf("Expected typeACounter to be 1, got %d", typeACounter.Load())
	}
	if typeBCounter.Load() != 1 {
		t.Errorf("Expected typeBCounter to be 1, got %d", typeBCounter.Load())
	}
	if typeCCounter.Load() != 1 {
		t.Errorf("Expected typeCCounter to be 1, got %d", typeCCounter.Load())
	}

	// 停止事件总线
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestEventBus_StopWhileProcessing 测试在处理事件时停止事件总线
func TestEventBus_StopWhileProcessing(t *testing.T) {
	eventBus := NewEventBus[*TestEvent](BusWithBufferSize[*TestEvent](100), BusWithWorkerCount[*TestEvent](2))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	// 创建一个阻塞的处理器
	processingStarted := make(chan struct{})
	processingDone := make(chan struct{})

	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		close(processingStarted) // 通知测试处理已开始

		// 模拟长时间处理
		select {
		case <-time.After(2 * time.Second):
			// 处理完成
		case <-ctx.Done():
			// 上下文被取消
		}

		close(processingDone) // 通知测试处理已完成
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event: %v", err)
	}

	// 发布事件
	testEvent := NewTestEvent(TestEventTypeA, "stop while processing")
	if err := eventBus.Publish(ctx, testEvent); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// 等待处理开始
	select {
	case <-processingStarted:
		// 处理已开始
	case <-time.After(1 * time.Second):
		t.Fatal("Event processing did not start in time")
	}

	// 停止事件总线，应该等待处理完成
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()

	stopDone := make(chan struct{})
	go func() {
		err := eventBus.Stop(shutdownCtx)
		if err != nil {
			t.Errorf("Failed to stop event bus: %v", err)
		}
		close(stopDone)
	}()

	// 验证停止操作是否等待处理完成
	select {
	case <-stopDone:
		t.Error("Event bus stopped before processing completed")
	case <-time.After(500 * time.Millisecond):
		// 正常，停止操作应该在等待
	}

	// 等待处理完成
	select {
	case <-processingDone:
		// 处理已完成
	case <-time.After(3 * time.Second):
		t.Fatal("Event processing did not complete in time")
	}

	// 现在停止操作应该完成
	select {
	case <-stopDone:
		// 停止操作已完成
	case <-time.After(1 * time.Second):
		t.Fatal("Event bus stop operation did not complete in time")
	}
}

// BenchmarkEventBus_Publish 基准测试：事件发布性能
func BenchmarkEventBus_Publish(b *testing.B) {
	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](10000),
		BusWithWorkerCount[*TestEvent](4),
	)
	ctx := context.Background()

	if err := eventBus.Start(ctx); err != nil {
		b.Fatalf("Failed to start event bus: %v", err)
	}

	// 添加一个简单的处理器
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		return nil
	})
	if err != nil {
		b.Fatalf("Failed to subscribe to event: %v", err)
	}

	// 准备测试事件
	testEvent := NewTestEvent(TestEventTypeA, "benchmark payload")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := eventBus.Publish(ctx, testEvent); err != nil {
			b.Fatalf("Failed to publish event: %v", err)
		}
	}
	b.StopTimer()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		b.Fatalf("Failed to stop event bus: %v", err)
	}
}

// BenchmarkEventBus_MultiSubscriber 基准测试：多订阅者性能
func BenchmarkEventBus_MultiSubscriber(b *testing.B) {
	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](10000),
		BusWithWorkerCount[*TestEvent](4),
	)
	ctx := context.Background()

	if err := eventBus.Start(ctx); err != nil {
		b.Fatalf("Failed to start event bus: %v", err)
	}

	// 添加10个处理器
	for i := 0; i < 10; i++ {
		_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
			return nil
		})
		if err != nil {
			b.Fatalf("Failed to subscribe to event: %v", err)
		}
	}

	// 准备测试事件
	testEvent := NewTestEvent(TestEventTypeA, "benchmark multi-subscriber")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := eventBus.Publish(ctx, testEvent); err != nil {
			b.Fatalf("Failed to publish event: %v", err)
		}
	}
	b.StopTimer()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		b.Fatalf("Failed to stop event bus: %v", err)
	}
}

// BenchmarkEventBus_ConcurrentPublish 基准测试：并发发布性能
func BenchmarkEventBus_ConcurrentPublish(b *testing.B) {
	eventBus := NewEventBus[*TestEvent](
		BusWithBufferSize[*TestEvent](10000),
		BusWithWorkerCount[*TestEvent](4),
	)
	ctx := context.Background()

	if err := eventBus.Start(ctx); err != nil {
		b.Fatalf("Failed to start event bus: %v", err)
	}

	// 添加一个简单的处理器
	_, err := eventBus.Subscribe(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		return nil
	})
	if err != nil {
		b.Fatalf("Failed to subscribe to event: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			event := NewTestEvent(TestEventTypeA, "concurrent benchmark")
			if err := eventBus.Publish(ctx, event); err != nil {
				b.Fatalf("Failed to publish event: %v", err)
			}
		}
	})
	b.StopTimer()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := eventBus.Stop(shutdownCtx); err != nil {
		b.Fatalf("Failed to stop event bus: %v", err)
	}
}

// TestDispatcher_BasicFunctionality 测试分发器基本功能
func TestDispatcher_BasicFunctionality(t *testing.T) {
	dispatcher := NewDefaultDispatcher[*TestEvent]()
	ctx := context.Background()

	var receivedEvent *TestEvent
	var wg sync.WaitGroup
	wg.Add(1)

	// 注册处理器
	subID, err := dispatcher.Register(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		receivedEvent = event
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// 分发事件
	testEvent := NewTestEvent(TestEventTypeA, "dispatcher test")
	if err := dispatcher.Dispatch(ctx, testEvent); err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	// 等待事件处理完成
	waitWithTimeout(&wg, 2*time.Second)

	// 验证事件是否被正确接收
	if receivedEvent == nil {
		t.Fatal("Event was not received")
	}

	if receivedEvent.Type() != TestEventTypeA {
		t.Errorf("Expected event type %s, got %s", TestEventTypeA, receivedEvent.Type())
	}

	// 测试取消注册
	if err := dispatcher.Unregister(subID); err != nil {
		t.Fatalf("Failed to unregister handler: %v", err)
	}

	// 验证取消注册后不再接收事件
	receivedEvent = nil
	if err := dispatcher.Dispatch(ctx, testEvent); err != nil {
		t.Fatalf("Failed to dispatch event after unregister: %v", err)
	}

	// 等待一段时间，确保事件有机会被处理
	time.Sleep(500 * time.Millisecond)

	// 验证事件未被接收
	if receivedEvent != nil {
		t.Error("Event was received after unregister")
	}
}

// TestDispatcher_WildcardSubscription 测试通配符订阅
func TestDispatcher_WildcardSubscription(t *testing.T) {
	dispatcher := NewDefaultDispatcher[*TestEvent]()
	ctx := context.Background()

	var receivedEvents atomic.Int32
	var wg sync.WaitGroup
	wg.Add(3) // 期望接收3个事件

	// 注册通配符处理器
	_, err := dispatcher.Register("*", func(ctx context.Context, event *TestEvent) error {
		receivedEvents.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register wildcard handler: %v", err)
	}

	// 分发3个不同类型的事件
	eventA := NewTestEvent(TestEventTypeA, "wildcard test A")
	eventB := NewTestEvent(TestEventTypeB, "wildcard test B")
	eventC := NewTestEvent(TestEventTypeC, "wildcard test C")

	if err := dispatcher.Dispatch(ctx, eventA); err != nil {
		t.Fatalf("Failed to dispatch event A: %v", err)
	}
	if err := dispatcher.Dispatch(ctx, eventB); err != nil {
		t.Fatalf("Failed to dispatch event B: %v", err)
	}
	if err := dispatcher.Dispatch(ctx, eventC); err != nil {
		t.Fatalf("Failed to dispatch event C: %v", err)
	}

	// 等待所有事件处理完成
	waitWithTimeout(&wg, 2*time.Second)

	// 验证接收到的事件数量
	if receivedEvents.Load() != 3 {
		t.Errorf("Expected to receive 3 events, got %d", receivedEvents.Load())
	}
}

// TestDispatcher_MultipleHandlers 测试多个处理器
func TestDispatcher_MultipleHandlers(t *testing.T) {
	dispatcher := NewDefaultDispatcher[*TestEvent]()
	ctx := context.Background()

	var handlerACount, handlerBCount atomic.Int32
	var wg sync.WaitGroup
	wg.Add(2) // 期望2个处理器都被调用

	// 注册第一个处理器
	_, err := dispatcher.Register(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		handlerACount.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register handler A: %v", err)
	}

	// 注册第二个处理器
	_, err = dispatcher.Register(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		handlerBCount.Add(1)
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register handler B: %v", err)
	}

	// 分发事件
	testEvent := NewTestEvent(TestEventTypeA, "multiple handlers test")
	if err := dispatcher.Dispatch(ctx, testEvent); err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	// 等待所有处理器被调用
	waitWithTimeout(&wg, 2*time.Second)

	// 验证计数器
	if handlerACount.Load() != 1 {
		t.Errorf("Expected handlerACount to be 1, got %d", handlerACount.Load())
	}
	if handlerBCount.Load() != 1 {
		t.Errorf("Expected handlerBCount to be 1, got %d", handlerBCount.Load())
	}
}

// TestDispatcher_ErrorHandling 测试分发器错误处理
func TestDispatcher_ErrorHandling(t *testing.T) {
	dispatcher := NewDefaultDispatcher[*TestEvent]()
	ctx := context.Background()

	// 注册两个处理器，一个会返回错误
	_, err := dispatcher.Register(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		return errors.New("handler A error")
	})
	if err != nil {
		t.Fatalf("Failed to register handler A: %v", err)
	}

	var handlerBCalled atomic.Bool
	_, err = dispatcher.Register(TestEventTypeA, func(ctx context.Context, event *TestEvent) error {
		handlerBCalled.Store(true)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register handler B: %v", err)
	}

	// 分发事件
	testEvent := NewTestEvent(TestEventTypeA, "error handling test")
	err = dispatcher.Dispatch(ctx, testEvent)

	// 验证错误
	if err == nil {
		t.Fatal("Expected error from dispatch, got nil")
	}

	// 验证第二个处理器仍被调用
	if !handlerBCalled.Load() {
		t.Error("Expected handler B to be called despite error in handler A")
	}
}

// TestIntegration_ActorWithEventBus 测试Actor与EventBus的集成
// func TestIntegration_ActorWithEventBus(t *testing.T) {
// 	// 创建事件总线
// 	eventBus := NewEventBus(BusWithBufferSize(100), BusWithWorkerCount(2))
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	if err := eventBus.Start(ctx); err != nil {
// 		t.Fatalf("Failed to start event bus: %v", err)
// 	}

// 	// 创建两个Actor
// 	actorA := NewActor[Event]("actor-A")
// 	actorB := NewActor[Event]("actor-B")

// 	// 注册Actor A的处理器
// 	var actorAReceived atomic.Bool
// 	var wgA sync.WaitGroup
// 	wgA.Add(1)

// 	err := actorA.RegisterHandler(TestEventTypeA, func(actorCtx ActorContext[Event], event Event) error {
// 		t.Logf("Actor A received event: %s", event.Type())
// 		actorAReceived.Store(true)
// 		wgA.Done()

// 		// Actor A收到事件后发布新事件给Actor B
// 		replyEvent := NewTestEvent(TestEventTypeB, "reply from actor A")
// 		return eventBus.Publish(actorCtx.Context, replyEvent)
// 	})
// 	if err != nil {
// 		t.Fatalf("Failed to register handler with actor A: %v", err)
// 	}

// 	// 注册Actor B的处理器
// 	var actorBReceived atomic.Bool
// 	var wgB sync.WaitGroup
// 	wgB.Add(1)

// 	err = actorB.RegisterHandler(TestEventTypeB, func(actorCtx ActorContext[Event], event Event) error {
// 		t.Logf("Actor B received event: %s", event.Type())
// 		actorBReceived.Store(true)
// 		wgB.Done()
// 		return nil
// 	})
// 	if err != nil {
// 		t.Fatalf("Failed to register handler with actor B: %v", err)
// 	}

// 	// 启动Actor
// 	if err := actorA.Start(ctx); err != nil {
// 		t.Fatalf("Failed to start actor A: %v", err)
// 	}
// 	if err := actorB.Start(ctx); err != nil {
// 		t.Fatalf("Failed to start actor B: %v", err)
// 	}

// 	// 将Actor注册到事件总线
// 	_, err = RegisterActorWithEventBus(eventBus, TestEventTypeA, actorA)
// 	if err != nil {
// 		t.Fatalf("Failed to register actor A with event bus: %v", err)
// 	}
// 	_, err = RegisterActorWithEventBus(eventBus, TestEventTypeB, actorB)
// 	if err != nil {
// 		t.Fatalf("Failed to register actor B with event bus: %v", err)
// 	}

// 	// 发布事件给Actor A
// 	initialEvent := NewTestEvent(TestEventTypeA, "initial event")
// 	if err := eventBus.Publish(ctx, initialEvent); err != nil {
// 		t.Fatalf("Failed to publish event: %v", err)
// 	}

// 	// 等待Actor A处理事件
// 	waitWithTimeout(&wgA, 2*time.Second)

// 	// 等待Actor B处理事件
// 	waitWithTimeout(&wgB, 2*time.Second)

// 	// 验证两个Actor都收到了事件
// 	if !actorAReceived.Load() {
// 		t.Error("Expected actor A to receive event")
// 	}
// 	if !actorBReceived.Load() {
// 		t.Error("Expected actor B to receive event")
// 	}

// 	// 停止Actor和事件总线
// 	if err := actorA.Stop(); err != nil {
// 		t.Fatalf("Failed to stop actor A: %v", err)
// 	}
// 	if err := actorB.Stop(); err != nil {
// 		t.Fatalf("Failed to stop actor B: %v", err)
// 	}

// 	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
// 	defer shutdownCancel()
// 	if err := eventBus.Stop(shutdownCtx); err != nil {
// 		t.Fatalf("Failed to stop event bus: %v", err)
// 	}
// }

// 辅助函数：带超时的等待
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 等待完成
		return
	case <-time.After(timeout):
		// 超时
		return
	}
}
