// WebSocket 服务
const WebSocketService = {
    socket: null,
    isConnecting: false,
    currentAgentMessageId: null,
    messageHandlers: new Map(),

    // 连接WebSocket
    connect() {
        if (!window.App?.accessToken || !window.App?.currentUser || this.isConnecting) {
            return;
        }

        this.isConnecting = true;
        this.updateConnectionStatus('connecting', '连接中...');

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/chat/stream`;

        this.socket = new WebSocket(wsUrl);

        this.socket.onopen = (event) => {
            this.isConnecting = false;
            this.updateConnectionStatus('connected', '已连接');
            this.enableChatInput();
            console.log('WebSocket 连接已建立');
        };

        this.socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleMessage(data);
            } catch (error) {
                console.error('解析 WebSocket 消息失败:', error);
            }
        };

        this.socket.onclose = (event) => {
            this.isConnecting = false;
            this.updateConnectionStatus('disconnected', '连接已断开');
            this.disableChatInput();
            console.log('WebSocket 连接已关闭:', event.code, event.reason);

            // 如果不是主动关闭，尝试重连
            if (event.code !== 1000 && window.App?.currentUser) {
                setTimeout(() => {
                    if (window.App?.currentUser) {
                        this.connect();
                    }
                }, 3000);
            }
        };

        this.socket.onerror = (error) => {
            this.isConnecting = false;
            this.updateConnectionStatus('disconnected', '连接错误');
            this.disableChatInput();
            console.error('WebSocket 错误:', error);
        };
    },

    // 断开连接
    disconnect() {
        if (this.socket) {
            this.socket.close(1000, 'User logout');
            this.socket = null;
        }
        this.updateConnectionStatus('disconnected', '未连接');
        this.disableChatInput();
    },

    // 发送消息
    send(data) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
            return true;
        }
        return false;
    },

    // 发送文本消息
    sendTextMessage(text) {
        const chatEvent = {
            eventId: Utils.generateUUID(),
            eventType: APP_CONSTANTS.WS_EVENTS.MESSAGE_SEND,
            event: {
                roomId: APP_CONSTANTS.DEFAULTS.ROOM_ID,
                msgType: APP_CONSTANTS.MESSAGE_TYPES.TEXT,
                body: { text: text },
                receiverId: 'assistant',
                senderId: window.App?.currentUser?.did,
                threadId: APP_CONSTANTS.DEFAULTS.THREAD_ID,
                senderAt: Date.now()
            }
        };

        return this.send(chatEvent);
    },

    // 发送图片消息
    sendImageMessage(imageCid, fileName) {
        const chatEvent = {
            eventId: Utils.generateUUID(),
            eventType: APP_CONSTANTS.WS_EVENTS.MESSAGE_SEND,
            event: {
                roomId: APP_CONSTANTS.DEFAULTS.ROOM_ID,
                msgType: APP_CONSTANTS.MESSAGE_TYPES.IMAGE,
                body: {
                    imageCid: imageCid,
                    width: 0,
                    height: 0,
                    alt: fileName
                },
                receiverId: 'assistant',
                senderId: window.App?.currentUser?.did,
                threadId: APP_CONSTANTS.DEFAULTS.THREAD_ID,
                senderAt: Date.now()
            }
        };

        return this.send(chatEvent);
    },

    // 中断AI响应
    interruptAIResponse() {
        if (!this.currentAgentMessageId) {
            return false;
        }

        const interruptEvent = {
            eventId: Utils.generateUUID(),
            eventType: APP_CONSTANTS.WS_EVENTS.AGENT_MESSAGE_INTERRUPT,
            event: {
                agentMessageId: this.currentAgentMessageId
            }
        };

        const success = this.send(interruptEvent);
        if (success) {
            console.log('发送中断请求:', this.currentAgentMessageId);
        }
        return success;
    },

    // 处理接收到的消息
    handleMessage(data) {
        console.log('收到聊天消息:', data);

        // 调用注册的处理器
        const handler = this.messageHandlers.get(data.eventType);
        if (handler) {
            handler(data.event);
        } else {
            console.log('未处理的事件类型:', data.eventType);
        }
    },

    // 注册消息处理器
    registerHandler(eventType, handler) {
        this.messageHandlers.set(eventType, handler);
    },

    // 更新连接状态
    updateConnectionStatus(status, text) {
        const statusElement = document.getElementById('connection-status');
        const textElement = document.getElementById('connection-text');

        if (statusElement && textElement) {
            statusElement.className = `connection-status ${status}`;
            textElement.textContent = text;
        }
    },

    // 启用聊天输入
    enableChatInput() {
        const chatInput = document.getElementById('chat-input');
        const chatSendBtn = document.getElementById('chat-send-btn');

        if (chatInput) chatInput.disabled = false;
        if (chatSendBtn) chatSendBtn.disabled = false;
    },

    // 禁用聊天输入
    disableChatInput() {
        const chatInput = document.getElementById('chat-input');
        const chatSendBtn = document.getElementById('chat-send-btn');

        if (chatInput) chatInput.disabled = true;
        if (chatSendBtn) chatSendBtn.disabled = true;
    }
};

// 导出WebSocket服务
window.WebSocketService = WebSocketService;