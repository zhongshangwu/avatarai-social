import json
import asyncio
import websockets
import uuid
from datetime import datetime
import threading
import queue
import sys

class AIChatStreamClient:
    def __init__(self, server_url):
        """
        初始化 AI 聊天流客户端

        Args:
            server_url (str): WebSocket 服务器 URL
        """
        self.server_url = server_url
        self.websocket = None
        self.running = False

    async def connect(self):
        """建立 WebSocket 连接"""
        try:
            self.websocket = await websockets.connect(self.server_url)
            self.running = True
            print(f"已连接到服务器: {self.server_url}")
            return True
        except Exception as e:
            print(f"连接失败: {str(e)}")
            return False

    async def disconnect(self):
        """关闭 WebSocket 连接"""
        if self.websocket:
            await self.websocket.close()
            self.running = False
            print("已断开连接")

    async def send_message(self, text, user_id="user123"):
        """
        发送文本消息

        Args:
            text (str): 要发送的文本消息
            user_id (str): 用户 ID
        """
        if not self.websocket:
            print("未连接到服务器")
            return

        message = {
            "eventId": str(uuid.uuid4()),
            "eventType": "send_msg",
            "event": {
                "roomId": "123",
                "msgType": 9,
                "body": {
                    "messageItems": [
                        {
                            "type": "message",
                            "role": "user",
                            "content": [
                                {
                                    "type": "input_text",
                                    "text": text
                                }
                            ],
                            "status": "success"
                        }
                    ]
                },
                "senderId": user_id,
                "threadId": "123",
                "quoteId": "123",
                "senderAt": int(datetime.now().timestamp())
            }
        }

        try:
            await self.websocket.send(json.dumps(message))
            print(f"已发送消息: {text}")
        except Exception as e:
            print(f"发送消息失败: {str(e)}")

    async def send_interrupt(self):
        """发送中断请求"""
        if not self.websocket:
            print("未连接到服务器")
            return

        message = {
            "type": "interrupt",
            "data": {}
        }

        try:
            await self.websocket.send(json.dumps(message))
            print("已发送中断请求")
        except Exception as e:
            print(f"发送中断请求失败: {str(e)}")

    async def request_history(self, user_id="user123", limit=10, before=None):
        """
        请求聊天历史记录

        Args:
            user_id (str): 用户 ID
            limit (int): 返回的最大记录数
            before (str, optional): 分页标记
        """
        if not self.websocket:
            print("未连接到服务器")
            return

        request = {
            "type": "request_history",
            "data": {
                "userId": user_id,
                "limit": limit
            }
        }

        if before:
            request["data"]["before"] = before

        try:
            await self.websocket.send(json.dumps(request))
            print(f"已请求历史记录: 用户={user_id}, 限制={limit}")
        except Exception as e:
            print(f"请求历史记录失败: {str(e)}")

    async def listen_events(self):
        """监听并处理服务器发送的事件"""
        print("开始监听服务器事件...")
        if not self.websocket:
            print("未连接到服务器")
            return

        current_response = {
            "text": "",
            "items": []
        }

        try:
            while self.running:
                try:
                    message = await self.websocket.recv()
                    print(f"收到原始消息: {message[:100]}..." if len(message) > 100 else f"收到原始消息: {message}")

                    try:
                        event = json.loads(message)
                    except json.JSONDecodeError:
                        print(f"无法解析JSON消息: {message}")
                        continue

                    event_type = event.get("type", "")

                    # 打印事件类型和时间戳
                    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
                    print(f"[{timestamp}] 收到事件: {event_type}")

                    # 根据事件类型处理不同的事件
                    if event_type == "created":
                        print("响应已创建")
                        current_response = {"text": "", "items": []}

                    elif event_type == "text_delta":
                        delta = event.get("delta", "")
                        current_response["text"] += delta
                        print(f"文本增量: {delta}", end="", flush=True)

                    elif event_type == "text_done":
                        text = event.get("text", "")
                        print(f"\n文本完成: {text}")

                    elif event_type == "completed":
                        response = event.get("response", {})
                        print("\n响应已完成:")
                        print(f"  - 状态: {response.get('status', '')}")
                        print(f"  - 文本: {response.get('text', '')}")
                        if "usage" in response:
                            usage = response["usage"]
                            print(f"  - 使用情况: 输入={usage.get('inputTokens', 0)}, "
                                f"输出={usage.get('outputTokens', 0)}, "
                                f"总计={usage.get('totalTokens', 0)}")

                    elif event_type == "error":
                        code = event.get("code", "unknown")
                        error_message = event.get("message", "未知错误")
                        print(f"\n错误: [{code}] {error_message}")

                    # 处理没有type字段的消息
                    elif "eventType" in event:
                        print(f"收到服务器事件: {event.get('eventType')}")
                        print(f"事件内容: {json.dumps(event, ensure_ascii=False, indent=2)}")

                    else:
                        print(f"未知事件类型: {json.dumps(event, ensure_ascii=False)}")

                except websockets.exceptions.ConnectionClosed as e:
                    print(f"WebSocket连接已关闭: {e}")
                    self.running = False
                    break
                except Exception as e:
                    print(f"处理消息时出错: {str(e)}")

        except asyncio.CancelledError:
            print("监听任务已取消")
        except Exception as e:
            print(f"监听事件时出错: {str(e)}")
        finally:
            self.running = False

# 非阻塞的输入函数
def get_input(prompt, input_queue):
    try:
        text = input(prompt)
        input_queue.put(text)
    except EOFError:
        input_queue.put(None)
    except Exception as e:
        input_queue.put(f"ERROR: {str(e)}")

async def async_input(prompt):
    """异步输入函数，不会阻塞事件循环"""
    input_queue = queue.Queue()
    threading.Thread(target=get_input, args=(prompt, input_queue), daemon=True).start()

    # 每隔0.1秒检查一次队列
    while True:
        if not input_queue.empty():
            return input_queue.get()
        await asyncio.sleep(0.1)

async def interactive_session(server_url):
    """
    交互式会话

    Args:
        server_url (str): WebSocket 服务器 URL
    """
    client = AIChatStreamClient(server_url)
    if not await client.connect():
        return

    # 启动事件监听
    listen_task = asyncio.create_task(client.listen_events())
    print("已启动监听任务")

    try:
        while client.running:
            # 显示菜单
            print("\n--- AI 聊天客户端 ---")
            print("1. 发送消息")
            print("2. 发送中断请求")
            print("3. 请求历史记录")
            print("4. 退出")

            choice = await async_input("请选择操作: ")
            if choice == "1":
                text = await async_input("请输入消息: ")
                await client.send_message(text)
            elif choice == "2":
                await client.send_interrupt()
            elif choice == "3":
                user_id = await async_input("用户ID (默认 user123): ") or "user123"
                limit_str = await async_input("记录数量 (默认 10): ") or "10"
                try:
                    limit = int(limit_str)
                except ValueError:
                    limit = 10
                await client.request_history(user_id, limit)
            elif choice == "4":
                break
            else:
                print("无效的选择，请重试")

    finally:
        # 取消监听任务并断开连接
        client.running = False
        listen_task.cancel()
        try:
            await listen_task
        except asyncio.CancelledError:
            pass
        await client.disconnect()

async def simple_demo(server_url):
    """
    简单演示

    Args:
        server_url (str): WebSocket 服务器 URL
    """
    client = AIChatStreamClient(server_url)
    if not await client.connect():
        return

    # 启动事件监听
    listen_task = asyncio.create_task(client.listen_events())
    print("已启动监听任务")

    try:
        # 发送一条消息
        await client.send_message("你好，请介绍一下自己")

        # 等待 10 秒，让服务器有时间响应
        await asyncio.sleep(10)

        # 发送另一条消息
        await client.send_message("你能帮我写一首诗吗？")

        # 等待 10 秒
        await asyncio.sleep(10)

        # 发送中断请求
        await client.send_interrupt()

        # 等待 2 秒
        await asyncio.sleep(2)

        # 请求历史记录
        await client.request_history()

        # 等待 5 秒
        await asyncio.sleep(5)

    finally:
        # 取消监听任务并断开连接
        client.running = False
        listen_task.cancel()
        try:
            await listen_task
        except asyncio.CancelledError:
            pass
        await client.disconnect()

if __name__ == "__main__":
    # 设置服务器 URL
    SERVER_URL = "ws://localhost:8082/api/demo/chat-stream"

    try:
        # 使用异步输入选择运行模式
        async def main():
            print("1. 交互式模式")
            print("2. 演示模式")
            mode = await async_input("请选择运行模式: ")

            if mode == "2":
                await simple_demo(SERVER_URL)
            else:
                await interactive_session(SERVER_URL)

        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n程序已被用户中断")
    except Exception as e:
        print(f"程序运行出错: {str(e)}")
