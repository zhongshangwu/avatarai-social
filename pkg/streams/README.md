# Stream 包

一个高性能、类型安全的 Go 流处理库，专为并发环境下的数据流传输而设计。

## 特性

- 🚀 **高性能**: 基于 Go channel 的零拷贝数据传输
- 🔒 **线程安全**: 使用原子操作和同步原语确保并发安全
- 🎯 **类型安全**: 基于泛型的强类型支持
- 🛡️ **优雅关闭**: 支持优雅关闭和强制关闭两种模式
- 📊 **状态管理**: 清晰的流状态跟踪和管理
- 🔄 **上下文感知**: 完整的 context.Context 支持
- 📦 **简洁 API**: 直观易用的接口设计

## 快速开始

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "time"

    "your-project/pkg/streams"
)

func main() {
    ctx := context.Background()

    // 创建一个缓冲区大小为 10 的流
    stream := streams.NewStream[string](ctx, 10)

    // 发送数据
    go func() {
        defer stream.CloseSend()

        for i := 0; i < 5; i++ {
            if err := stream.Send(fmt.Sprintf("message-%d", i)); err != nil {
                fmt.Printf("发送失败: %v\n", err)
                return
            }
        }
    }()

    // 接收数据
    for {
        result := stream.Recv()

        if result.HasData {
            fmt.Printf("收到数据: %s\n", result.Data)
            continue
        }

        if result.Completed {
            if result.Error != nil {
                fmt.Printf("流结束，错误: %v\n", result.Error)
            } else {
                fmt.Println("流正常结束")
            }
            break
        }
    }
}
```

### 错误处理

```go
func handleStreamWithError() {
    ctx := context.Background()
    stream := streams.NewStream[int](ctx, 5)

    go func() {
        defer stream.CloseSend()

        // 发送一些数据
        stream.Send(1)
        stream.Send(2)

        // 发送错误
        stream.SendError(errors.New("处理失败"))
    }()

    for {
        result := stream.Recv()

        if result.HasData {
            fmt.Printf("数据: %d\n", result.Data)
            continue
        }

        if result.Completed {
            if result.Error != nil {
                fmt.Printf("流异常结束: %v\n", result.Error)
            } else {
                fmt.Println("流正常结束")
            }
            break
        }
    }
}
```

### 上下文取消

```go
func handleContextCancellation() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    stream := streams.NewStream[string](ctx, 10)

    go func() {
        defer stream.CloseSend()

        for i := 0; i < 100; i++ {
            if err := stream.Send(fmt.Sprintf("data-%d", i)); err != nil {
                fmt.Printf("发送中断: %v\n", err)
                return
            }
            time.Sleep(100 * time.Millisecond)
        }
    }()

    for {
        result := stream.Recv()

        if result.HasData {
            fmt.Printf("收到: %s\n", result.Data)
            continue
        }

        if result.Completed {
            if result.Error == streams.ErrContextAlreadyDone {
                fmt.Println("上下文超时，流被取消")
            }
            break
        }
    }
}
```

## API 文档

### 类型定义

#### StreamResult[T]

表示接收操作的结果：

```go
type StreamResult[T any] struct {
    Data      T     // 接收到的数据
    HasData   bool  // 是否有数据
    Completed bool  // 流是否已完成
    Error     error // 错误信息（仅在 Completed=true 时有意义）
}
```

#### StreamState

流的状态枚举：

```go
type StreamState int32

const (
    StreamStateActive   StreamState = iota // 活跃状态
    StreamStateClosing                     // 正在关闭
    StreamStateClosed                      // 已关闭
)
```

### 核心方法

#### NewStream[T](ctx context.Context, maxSize int) *Stream[T]

创建一个新的流实例。

**参数:**
- `ctx`: 上下文，用于控制流的生命周期
- `maxSize`: 内部缓冲区大小

**返回:**
- 新的流实例

#### Send(item T) error

向流中发送数据。

**参数:**
- `item`: 要发送的数据

**返回:**
- `nil`: 发送成功
- `ErrChannelClosed`: 流已关闭
- `ErrContextAlreadyDone`: 上下文已取消

#### SendError(err error)

向流中发送错误信号，这会触发流的关闭。

**参数:**
- `err`: 错误信息

#### Recv() StreamResult[T]

从流中接收数据，这是一个阻塞操作。

**返回:**
- `StreamResult[T]`: 接收结果

**使用模式:**
```go
result := stream.Recv()

if result.HasData {
    // 处理数据: result.Data
} else if result.Completed {
    if result.Error != nil {
        // 处理错误: result.Error
    } else {
        // 流正常结束
    }
}
```

#### CloseSend()

开始优雅关闭流程。调用后不能再发送新数据，但接收方仍可以读取剩余数据。

#### ForceClose()

强制关闭流，可能会丢失缓冲区中的数据。

#### Closed() bool

检查流是否已关闭。

**返回:**
- `true`: 流已关闭或正在关闭
- `false`: 流仍然活跃

#### State() StreamState

获取流的当前状态。

**返回:**
- 当前流状态

## 错误类型

```go
var (
    ErrContextAlreadyDone = errors.New("stream context already done")
    ErrChannelClosed      = errors.New("send to closed channel")
)
```

## 最佳实践

### 1. 总是使用 defer 关闭发送端

```go
go func() {
    defer stream.CloseSend() // 确保流被正确关闭

    // 发送数据...
}()
```

### 2. 正确处理接收循环

```go
for {
    result := stream.Recv()

    if result.HasData {
        // 主要逻辑：处理数据
        processData(result.Data)
        continue
    }

    if result.Completed {
        // 流已结束，检查结束原因
        if result.Error != nil {
            // 异常结束
            handleError(result.Error)
        } else {
            // 正常结束
            handleSuccess()
        }
        break
    }
}
```

**关于 `continue` 的使用：**
- 虽然 `HasData` 和 `Completed` 是互斥的，但 `continue` 能明确表达处理意图
- 避免不必要的条件检查，提升性能
- 符合状态处理的标准模式，代码更清晰

**判断顺序详解：**
1. **优先判断 `HasData`** - 处理数据是主要逻辑，发生最频繁
2. **然后判断 `Completed`** - 确认流是否结束
3. **最后判断 `Error`** - 在流结束的前提下，区分正常/异常结束

**重要说明：**
- `HasData` 和 `Completed` 是互斥的，不会同时为 `true`
- 数据处理是主要逻辑，发生频率更高，先判断常见情况可以提高性能
- 使用 `continue` 明确表达"处理完数据立即进入下一轮循环"的意图

### 3. 使用上下文控制超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream := streams.NewStream[Data](ctx, 100)
```

### 4. 错误处理策略

```go
// 发送端错误处理
if err := stream.Send(data); err != nil {
    if err == streams.ErrChannelClosed {
        // 流已关闭，停止发送
        return
    }
    // 其他错误处理...
}

// 接收端错误处理
result := stream.Recv()
if result.Completed && result.Error != nil {
    switch result.Error {
    case streams.ErrContextAlreadyDone:
        // 上下文超时或取消
    default:
        // 其他业务错误
    }
}
```

## 性能考虑

### 缓冲区大小选择

- **小缓冲区 (1-10)**: 适用于低延迟场景，内存占用少
- **中等缓冲区 (10-100)**: 平衡延迟和吞吐量的通用选择
- **大缓冲区 (100+)**: 适用于高吞吐量场景，但会增加内存占用

### 并发模式

```go
// 单生产者-单消费者 (最高性能)
producer := func() { /* 发送数据 */ }
consumer := func() { /* 接收数据 */ }

// 多生产者-单消费者 (需要外部同步)
// 注意: Send() 方法是线程安全的

// 单生产者-多消费者 (需要外部协调)
// 注意: 每个数据只会被一个消费者接收
```

## 线程安全性

- ✅ `Send()` 方法是线程安全的，多个 goroutine 可以并发调用
- ✅ `SendError()` 方法是线程安全的
- ✅ `Recv()` 方法是线程安全的，但通常只有一个消费者
- ✅ `CloseSend()` 和 `ForceClose()` 是线程安全的
- ✅ `Closed()` 和 `State()` 是线程安全的

## 常见问题

### Q: 为什么 Recv() 是阻塞的？

A: 阻塞设计简化了使用模式，避免了忙等待。如果需要非阻塞行为，可以结合 `select` 语句使用：

```go
select {
case result := <-resultChan:
    // 处理结果
default:
    // 没有数据时的处理
}
```

### Q: 如何实现超时接收？

A: 使用上下文或 `time.After`：

```go
// 方法1: 使用上下文
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
stream := streams.NewStream[Data](ctx, 10)

// 方法2: 使用 select
select {
case result := <-receiveChan:
    // 处理结果
case <-time.After(5 * time.Second):
    // 超时处理
}
```

### Q: 流关闭后还能接收数据吗？

A: 可以。调用 `CloseSend()` 后，发送端关闭但接收端仍可以读取缓冲区中的剩余数据，直到收到 `Completed=true` 的结果。

### Q: 如何处理背压（backpressure）？

A: 当缓冲区满时，`Send()` 会阻塞。可以通过以下方式处理：

```go
// 方法1: 使用带超时的上下文
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

// 方法2: 检查流状态
if stream.Closed() {
    // 流已关闭，停止发送
    return
}

// 方法3: 增大缓冲区大小
stream := streams.NewStream[Data](ctx, 1000) // 更大的缓冲区
```

## 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。
