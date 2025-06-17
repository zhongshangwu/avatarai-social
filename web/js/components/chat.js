// 聊天组件
const ChatComponent = {
    // 初始化
    init() {
        this.renderChatContainer();
        this.bindEvents();
        this.registerWebSocketHandlers();
    },

    // 渲染聊天容器
    renderChatContainer() {
        const container = document.getElementById('chat-container');
        if (!container) return;

        container.innerHTML = `
            <div id="connection-status" class="connection-status disconnected">
                <div class="status-indicator"></div>
                <span id="connection-text">未连接</span>
            </div>

            <div class="chat-container">
                <div id="chat-messages" class="chat-messages">
                    <div class="chat-message system">
                        <i class="fas fa-robot"></i> 欢迎使用 AvatarAI 智能对话！我是你的AI助手，有什么可以帮助你的吗？
                    </div>
                </div>

                <div class="chat-input-container">
                    <button id="chat-image-btn" class="chat-image-btn" title="发送图片">
                        <i class="fas fa-image"></i>
                    </button>
                    <textarea
                        id="chat-input"
                        class="chat-input"
                        placeholder="输入消息..."
                        rows="1"
                        disabled
                    ></textarea>
                    <button id="chat-interrupt-btn" class="chat-interrupt-btn hidden">
                        <i class="fas fa-stop"></i>
                    </button>
                    <button id="chat-send-btn" class="chat-send-btn" disabled>
                        <i class="fas fa-paper-plane"></i>
                    </button>
                </div>
            </div>

            <!-- 隐藏的文件输入 -->
            <input type="file" id="chat-image-input" accept="image/*" style="display: none;">
        `;
    },

    // 绑定事件
    bindEvents() {
        // 发送按钮
        const chatSendBtn = document.getElementById('chat-send-btn');
        if (chatSendBtn) {
            chatSendBtn.addEventListener('click', () => this.sendMessage());
        }

        // 输入框
        const chatInput = document.getElementById('chat-input');
        if (chatInput) {
            chatInput.addEventListener('keypress', (e) => this.handleKeyPress(e));
            chatInput.addEventListener('input', (e) => Utils.autoResizeTextarea(e.target));
            // 初始化输入框高度
            Utils.autoResizeTextarea(chatInput);
        }

        // 中断按钮
        const interruptBtn = document.getElementById('chat-interrupt-btn');
        if (interruptBtn) {
            interruptBtn.addEventListener('click', () => this.interruptAIResponse());
        }

        // 图片发送按钮
        const imageBtn = document.getElementById('chat-image-btn');
        if (imageBtn) {
            imageBtn.addEventListener('click', () => this.selectImage());
        }

        // 图片文件输入
        const imageInput = document.getElementById('chat-image-input');
        if (imageInput) {
            imageInput.addEventListener('change', (e) => this.handleImageSelect(e));
        }
    },

    // 注册WebSocket处理器
    registerWebSocketHandlers() {
        // 注册各种消息处理器
        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.MESSAGE_RECEIVED,
            (event) => this.handleMessageReceived(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.AGENT_MESSAGE_CREATED,
            (event) => this.handleAgentMessageCreated(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.AGENT_MESSAGE_IN_PROGRESS,
            (event) => this.handleAgentMessageInProgress(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.OUTPUT_ITEM_ADDED,
            (event) => this.handleOutputItemAdded(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.CONTENT_PART_ADDED,
            (event) => this.handleContentPartAdded(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.TEXT_DELTA,
            (event) => this.handleTextDelta(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.TEXT_DONE,
            (event) => this.handleTextDone(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.CONTENT_PART_DONE,
            (event) => this.handleContentPartDone(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.OUTPUT_ITEM_DONE,
            (event) => this.handleOutputItemDone(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.AGENT_MESSAGE_COMPLETED,
            (event) => this.handleAgentMessageCompleted(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.AGENT_MESSAGE_FAILED,
            (event) => this.handleAgentMessageFailed(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.AGENT_MESSAGE_INCOMPLETE,
            (event) => this.handleAgentMessageIncomplete(event));

        WebSocketService.registerHandler(APP_CONSTANTS.WS_EVENTS.ERROR,
            (event) => this.handleErrorEvent(event));
    },

    // 处理按键事件
    handleKeyPress(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            this.sendMessage();
        }
    },

    // 发送消息
    sendMessage() {
        const input = document.getElementById('chat-input');
        const message = input.value.trim();

        if (!message) {
            return;
        }

        // 显示用户消息
        this.addChatMessage(message, 'user');

        // 发送到服务器
        const success = WebSocketService.sendTextMessage(message);
        if (success) {
            input.value = '';
            Utils.autoResizeTextarea(input);
        } else {
            Utils.showMessage('WebSocket未连接，无法发送消息', 'error');
        }
    },

    // 选择图片
    selectImage() {
        document.getElementById('chat-image-input').click();
    },

    // 处理图片选择
    async handleImageSelect(event) {
        const file = event.target.files[0];
        if (!file) return;

        if (!file.type.startsWith('image/')) {
            Utils.showMessage('只支持图片文件', 'error');
            return;
        }

        if (file.size > APP_CONSTANTS.FILE_LIMITS.IMAGE_MAX_SIZE) {
            Utils.showMessage('图片大小不能超过10MB', 'error');
            return;
        }

        // 显示上传中的消息
        this.addChatMessage('[正在上传图片...]', 'user');

        try {
            const uploadResult = await ApiService.uploadFile(file);
            console.log('图片上传成功:', uploadResult);

            // 发送图片消息
            const success = WebSocketService.sendImageMessage(uploadResult.cid, file.name);

            if (success) {
                // 更新最后一条消息显示实际图片
                const messages = document.querySelectorAll('.chat-message.user');
                const lastMessage = messages[messages.length - 1];
                if (lastMessage && lastMessage.textContent.includes('[正在上传图片...]')) {
                    const imageUrl = uploadResult.url || `/api/blobs?id=${uploadResult.cid}`;
                    lastMessage.innerHTML = `<i class="fas fa-user"></i> <img src="${imageUrl}" alt="${file.name}">`;
                }
            } else {
                Utils.showMessage('WebSocket未连接，无法发送图片', 'error');
                // 移除上传中的消息
                const messages = document.querySelectorAll('.chat-message.user');
                const lastMessage = messages[messages.length - 1];
                if (lastMessage && lastMessage.textContent.includes('[正在上传图片...]')) {
                    lastMessage.remove();
                }
            }
        } catch (error) {
            console.error('图片上传失败:', error);
            Utils.showMessage('图片上传失败: ' + error.message, 'error');

            // 移除上传中的消息
            const messages = document.querySelectorAll('.chat-message.user');
            const lastMessage = messages[messages.length - 1];
            if (lastMessage && lastMessage.textContent.includes('[正在上传图片...]')) {
                lastMessage.remove();
            }
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 中断AI响应
    interruptAIResponse() {
        const success = WebSocketService.interruptAIResponse();
        if (success) {
            this.hideInterruptButton();
        }
    },

    // 显示中断按钮
    showInterruptButton() {
        const interruptBtn = document.getElementById('chat-interrupt-btn');
        if (interruptBtn) {
            interruptBtn.classList.remove('hidden');
        }
    },

    // 隐藏中断按钮
    hideInterruptButton() {
        const interruptBtn = document.getElementById('chat-interrupt-btn');
        if (interruptBtn) {
            interruptBtn.classList.add('hidden');
        }
    },

    // 添加聊天消息
    addChatMessage(text, type, messageId = null, isLoading = false, isHtml = false) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = `chat-message ${type}`;

        if (messageId) {
            messageDiv.setAttribute('data-message-id', messageId);
        }

        if (isLoading) {
            messageDiv.classList.add('loading');
        }

        let iconHtml = '';
        if (type === 'user') {
            iconHtml = '<i class="fas fa-user"></i> ';
        } else if (type === 'assistant') {
            iconHtml = '<i class="fas fa-robot"></i> ';
        } else {
            iconHtml = '<i class="fas fa-info-circle"></i> ';
        }

        if (isHtml) {
            messageDiv.innerHTML = iconHtml + text;
        } else {
            messageDiv.innerHTML = iconHtml + Utils.escapeHtml(text);
        }

        if (isLoading) {
            const loadingSpinner = document.createElement('span');
            loadingSpinner.className = 'loading-spinner';
            loadingSpinner.innerHTML = ' <i class="fas fa-spinner fa-spin"></i>';
            messageDiv.appendChild(loadingSpinner);
        }

        messagesContainer.appendChild(messageDiv);
        this.scrollToBottom();
    },

    // 添加历史聊天消息
    addHistoryChatMessage(content, type, isHtml = false, timestamp = null) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = `chat-message ${type}`;

        let iconHtml = '';
        if (type === 'user') {
            iconHtml = '<i class="fas fa-user"></i> ';
        } else if (type === 'assistant') {
            iconHtml = '<i class="fas fa-robot"></i> ';
        } else {
            iconHtml = '<i class="fas fa-info-circle"></i> ';
        }

        let timestampHtml = '';
        if (timestamp) {
            const date = new Date(timestamp);
            const timeStr = date.toLocaleTimeString('zh-CN', {
                hour: '2-digit',
                minute: '2-digit'
            });
            timestampHtml = `<span class="message-timestamp">${timeStr}</span>`;
        }

        if (isHtml) {
            messageDiv.innerHTML = iconHtml + content + timestampHtml;
        } else {
            messageDiv.innerHTML = iconHtml + Utils.escapeHtml(content) + timestampHtml;
        }

        messagesContainer.appendChild(messageDiv);
    },

    // 滚动到底部
    scrollToBottom() {
        const messagesContainer = document.getElementById('chat-messages');
        if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    },

    // 创建流式消息容器
    createStreamingMessage(messageId) {
        console.log('创建流式消息容器:', messageId);

        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message assistant loading';
        messageDiv.setAttribute('data-message-id', messageId);

        // 创建图标容器
        const iconContainer = document.createElement('span');
        iconContainer.className = 'message-icon';
        iconContainer.innerHTML = '<i class="fas fa-robot"></i> ';

        // 创建内容容器
        const contentContainer = document.createElement('span');
        contentContainer.className = 'message-content';

        // 创建加载指示器
        const loadingSpinner = document.createElement('span');
        loadingSpinner.className = 'loading-spinner';
        loadingSpinner.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 正在思考...';

        messageDiv.appendChild(iconContainer);
        messageDiv.appendChild(contentContainer);
        messageDiv.appendChild(loadingSpinner);

        messagesContainer.appendChild(messageDiv);
        this.scrollToBottom();

        console.log('流式消息容器已创建:', messageDiv);
    },

    // 加载聊天历史
    async loadHistory() {
        if (!window.App.accessToken) {
            console.log('未登录，无法加载历史消息');
            return;
        }

        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        // 显示加载提示
        const loadingDiv = document.createElement('div');
        loadingDiv.className = 'history-loading';
        loadingDiv.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 正在加载历史消息...';
        messagesContainer.appendChild(loadingDiv);

        try {
            const data = await ApiService.getChatHistory();
            console.log('历史消息响应:', data);

            // 移除加载提示
            loadingDiv.remove();

            if (data.messages && data.messages.length > 0) {
                this.displayHistoryMessages(data.messages);
                console.log(`成功加载 ${data.messages.length} 条历史消息`);
            } else {
                console.log('没有历史消息');
                this.addChatMessage('📝 暂无历史消息', 'system');
            }
        } catch (error) {
            console.error('加载历史消息失败:', error);
            // 移除加载提示
            loadingDiv.remove();
            this.addChatMessage(`❌ 加载历史消息失败: ${error.message}`, 'system');
        }
    },

    // 显示历史消息
    displayHistoryMessages(messages) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        // 清空现有消息（除了欢迎消息）
        const welcomeMessage = messagesContainer.querySelector('.chat-message.system');
        messagesContainer.innerHTML = '';
        if (welcomeMessage) {
            messagesContainer.appendChild(welcomeMessage);
        }

        if (!messages || messages.length === 0) {
            console.log('没有历史消息');
            return;
        }

        console.log(`加载 ${messages.length} 条历史消息`);

        // 按时间顺序排序（最旧的在前）
        const sortedMessages = messages.sort((a, b) => {
            const timeA = a.senderAt || a.createdAt || 0;
            const timeB = b.senderAt || b.createdAt || 0;
            return timeA - timeB;
        });

        sortedMessages.forEach((message, index) => {
            console.log(`处理历史消息 ${index + 1}:`, {
                id: message.id,
                msgType: message.msgType,
                senderId: message.senderId,
                senderAt: new Date(message.senderAt).toLocaleString(),
                content: message.content
            });
            this.displayHistoryMessage(message);
        });

        this.scrollToBottom();
    },

    // 显示单条历史消息
    displayHistoryMessage(message) {
        let content = '';
        let type = 'user';
        let isHtml = false;

        // 根据发送者确定消息类型
        if (message.senderId === window.App.currentUser?.did) {
            type = 'user';
        } else {
            type = 'assistant';
        }

        // 解析消息内容
        if (message.content) {
            switch (message.msgType) {
                case APP_CONSTANTS.MESSAGE_TYPES.TEXT:
                    content = message.content.text || '';
                    break;
                case APP_CONSTANTS.MESSAGE_TYPES.IMAGE:
                    if (message.content.imageUrl || message.content.imageCid) {
                        const imageUrl = message.content.imageUrl || `/api/blobs?id=${message.content.imageCid}`;
                        content = `<img src="${imageUrl}" alt="${message.content.alt || '图片'}">`;
                        isHtml = true;
                    } else {
                        content = `[图片] ${message.content.alt || ''}`;
                    }
                    break;
                case APP_CONSTANTS.MESSAGE_TYPES.AGENT_MESSAGE:
                    type = 'assistant';
                    if (message.content.message) {
                        const agentMessage = message.content.message;
                        console.log('处理AI消息:', agentMessage);

                        // 检查消息状态
                        if (agentMessage.status === 'incomplete') {
                            if (agentMessage.error) {
                                content = `❌ 错误: ${agentMessage.error.message || '处理失败'}`;
                                type = 'system';
                            } else if (agentMessage.interruptType === 2) {
                                content = '⚠️ 消息被中断';
                                type = 'system';
                            } else {
                                content = '⚠️ 消息不完整';
                                type = 'system';
                            }
                        } else if (agentMessage.status === 'failed') {
                            const errorMsg = agentMessage.error ? agentMessage.error.message : '处理失败';
                            content = `❌ 处理失败: ${errorMsg}`;
                            type = 'system';
                        } else if (agentMessage.altText && agentMessage.altText.trim()) {
                            content = agentMessage.altText;
                        } else if (agentMessage.messageItems && agentMessage.messageItems.length > 0) {
                            const textContents = [];
                            agentMessage.messageItems.forEach(item => {
                                if (item.content && Array.isArray(item.content)) {
                                    item.content.forEach(contentPart => {
                                        if (contentPart.type === 'output_text' && contentPart.text) {
                                            textContents.push(contentPart.text);
                                        }
                                    });
                                }
                            });
                            content = textContents.join('') || '🤖 [AI消息无内容]';
                        } else {
                            content = '🤖 [AI消息无内容]';
                        }
                    } else {
                        content = '🤖 [AI消息格式错误]';
                    }
                    break;
                default:
                    content = `[不支持的消息类型: ${message.msgType}]`;
                    type = 'system';
            }
        }

        if (content) {
            const timestamp = message.senderAt || message.createdAt;
            this.addHistoryChatMessage(content, type, isHtml, timestamp);
        }
    },

    // WebSocket消息处理器
    handleMessageReceived(event) {
        if (event.message && event.message.content) {
            const content = event.message.content;
            const msgType = event.message.msgType;

            switch (msgType) {
                case APP_CONSTANTS.MESSAGE_TYPES.TEXT:
                    if (content.text) {
                        this.addChatMessage(content.text, 'assistant');
                    }
                    break;
                case APP_CONSTANTS.MESSAGE_TYPES.IMAGE:
                    if (content.imageUrl || content.imageCid) {
                        const imageUrl = content.imageUrl || `/api/blobs?id=${content.imageCid}`;
                        const imageHtml = `<img src="${imageUrl}" alt="${content.alt || '图片'}">`;
                        this.addChatMessage(imageHtml, 'assistant', null, false, true);
                    }
                    break;
                default:
                    console.log('未处理的消息类型:', msgType);
            }
        }
    },

    handleAgentMessageCreated(event) {
        const agentMessage = event.agentMessage;
        console.log('AI消息创建:', agentMessage);

        // 保存当前AI消息ID，用于中断
        WebSocketService.currentAgentMessageId = agentMessage.id;
        this.showInterruptButton();

        // 创建一个占位符消息，用于后续更新
        this.createStreamingMessage(agentMessage.id);
    },

    handleAgentMessageInProgress(event) {
        const agentMessage = event.agentMessage;
        console.log('AI消息处理中:', agentMessage);

        // 更新加载状态
        const messageElement = document.querySelector(`[data-message-id="${agentMessage.id}"]`);
        if (messageElement) {
            const loadingSpinner = messageElement.querySelector('.loading-spinner');
            if (loadingSpinner) {
                loadingSpinner.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 正在思考...';
            }
        }
    },

    handleOutputItemAdded(event) {
        const { outputIndex, item } = event;
        console.log('输出项添加:', item);

        if (item.type === 'message') {
            const messageId = item.id;
            if (!document.querySelector(`[data-message-id="${messageId}"]`)) {
                this.createStreamingMessage(messageId);
            }
        }
    },

    handleContentPartAdded(event) {
        const { itemId, outputIndex, contentIndex, part } = event;
        console.log('内容部分添加:', part);

        if (part.type === 'output_text') {
            const messageElement = document.querySelector(`[data-message-id="${itemId}"]`);
            if (messageElement) {
                // 隐藏加载指示器
                const loadingSpinner = messageElement.querySelector('.loading-spinner');
                if (loadingSpinner) {
                    loadingSpinner.style.display = 'none';
                }

                // 在内容容器中创建文本容器
                const contentContainer = messageElement.querySelector('.message-content');
                if (contentContainer) {
                    const textContainer = document.createElement('span');
                    textContainer.setAttribute('data-content-index', contentIndex);
                    contentContainer.appendChild(textContainer);
                }
            }
        }
    },

    handleTextDelta(event) {
        const { itemId, outputIndex, contentIndex, delta } = event;
        console.log('收到文本增量:', { itemId, contentIndex, delta });

        const messageElement = document.querySelector(`[data-message-id="${itemId}"]`);
        if (messageElement) {
            // 隐藏加载指示器
            const loadingSpinner = messageElement.querySelector('.loading-spinner');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }

            // 获取或创建内容容器
            const contentContainer = messageElement.querySelector('.message-content');
            if (contentContainer) {
                let textContainer = contentContainer.querySelector(`[data-content-index="${contentIndex}"]`);
                if (!textContainer) {
                    textContainer = document.createElement('span');
                    textContainer.setAttribute('data-content-index', contentIndex);
                    contentContainer.appendChild(textContainer);
                }
                textContainer.textContent += delta;
            }

            this.scrollToBottom();
        }
    },

    handleTextDone(event) {
        const { itemId, outputIndex, contentIndex, text } = event;
        console.log('文本完成:', text);

        const messageElement = document.querySelector(`[data-message-id="${itemId}"]`);
        if (messageElement) {
            const loadingSpinner = messageElement.querySelector('.loading-spinner');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }

            const contentContainer = messageElement.querySelector('.message-content');
            if (contentContainer) {
                let textContainer = contentContainer.querySelector(`[data-content-index="${contentIndex}"]`);
                if (!textContainer) {
                    textContainer = document.createElement('span');
                    textContainer.setAttribute('data-content-index', contentIndex);
                    contentContainer.appendChild(textContainer);
                }
                textContainer.textContent = text;
            }
        }
    },

    handleContentPartDone(event) {
        const { itemId, outputIndex, contentIndex, part } = event;
        console.log('内容部分完成:', part);
    },

    handleOutputItemDone(event) {
        const { outputIndex, item } = event;
        console.log('输出项完成:', item);

        const messageElement = document.querySelector(`[data-message-id="${item.id}"]`);
        if (messageElement) {
            messageElement.classList.remove('loading');
        }
    },

    handleAgentMessageCompleted(event) {
        const agentMessage = event.agentMessage;
        console.log('AI消息完成:', agentMessage);

        // 清除当前AI消息ID并隐藏中断按钮
        WebSocketService.currentAgentMessageId = null;
        this.hideInterruptButton();

        // 移除所有加载状态
        const messageElement = document.querySelector(`[data-message-id="${agentMessage.id}"]`);
        if (messageElement) {
            messageElement.classList.remove('loading');
            messageElement.classList.add('completed');

            // 完全隐藏加载指示器
            const loadingSpinner = messageElement.querySelector('.loading-spinner');
            if (loadingSpinner) {
                loadingSpinner.remove();
            }
        }
    },

    handleAgentMessageFailed(event) {
        const agentMessage = event.agentMessage;
        console.log('AI消息失败:', agentMessage);

        // 清除当前AI消息ID并隐藏中断按钮
        WebSocketService.currentAgentMessageId = null;
        this.hideInterruptButton();

        const errorMsg = agentMessage.error ? agentMessage.error.message : '处理失败';
        this.addChatMessage(`错误: ${errorMsg}`, 'system');
    },

    handleAgentMessageIncomplete(event) {
        const agentMessage = event.agentMessage;
        console.log('AI消息不完整:', agentMessage);

        // 清除当前AI消息ID并隐藏中断按钮
        WebSocketService.currentAgentMessageId = null;
        this.hideInterruptButton();

        this.addChatMessage('响应被中断或不完整', 'system');
    },

    handleErrorEvent(event) {
        const errorMsg = event.message || '发生未知错误';
        this.addChatMessage(`错误: ${errorMsg}`, 'system');
    }
};

// 导出聊天组件
window.ChatComponent = ChatComponent;