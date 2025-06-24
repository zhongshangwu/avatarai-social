// 模态框组件
const ModalsComponent = {
    currentModal: null,

    // 初始化
    init() {
        // 确保模态框容器存在
        if (!document.getElementById('modals-container')) {
            const container = document.createElement('div');
            container.id = 'modals-container';
            document.body.appendChild(container);
        }
    },

    // 显示通用模态框
    showModal(title, content, options = {}) {
        // 默认选项
        const defaultOptions = {
            size: 'medium', // small, medium, large
            showFooter: true,
            confirmText: '确认',
            cancelText: '取消',
            onConfirm: null,
            onCancel: null
        };

        const opts = { ...defaultOptions, ...options };

        // 创建模态框
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';

        const sizeClass = opts.size === 'large' ? 'modal-large' :
                         opts.size === 'small' ? 'modal-small' : 'modal-medium';

        modal.innerHTML = `
            <div class="modal-content ${sizeClass}">
                <div class="modal-header">
                    <h3>${title}</h3>
                    <button class="modal-close" onclick="ModalsComponent.hideModal()">&times;</button>
                </div>
                <div class="modal-body">
                    ${content}
                </div>
                ${opts.showFooter ? `
                    <div class="modal-footer">
                        <button class="btn btn-secondary" onclick="ModalsComponent.hideModal()">${opts.cancelText}</button>
                        ${opts.onConfirm ? `<button class="btn btn-primary" onclick="ModalsComponent.handleConfirm()">${opts.confirmText}</button>` : ''}
                    </div>
                ` : ''}
            </div>
        `;

        // 存储当前模态框和回调
        this.currentModal = modal;
        this.currentOptions = opts;

        // 添加到容器
        document.getElementById('modals-container').appendChild(modal);

        // 点击模态框外部关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.hideModal();
            }
        });

        // ESC键关闭
        const escHandler = (e) => {
            if (e.key === 'Escape') {
                this.hideModal();
                document.removeEventListener('keydown', escHandler);
            }
        };
        document.addEventListener('keydown', escHandler);

        return modal;
    },

    // 隐藏模态框
    hideModal() {
        if (this.currentModal) {
            this.currentModal.remove();
            this.currentModal = null;

            // 执行取消回调
            if (this.currentOptions?.onCancel) {
                this.currentOptions.onCancel();
            }

            this.currentOptions = null;
        }
    },

    // 处理确认
    async handleConfirm() {
        if (this.currentOptions?.onConfirm) {
            const result = await this.currentOptions.onConfirm();
            // 只有当回调返回 true 或 undefined 时才关闭模态框
            if (result !== false) {
                this.hideModal();
            }
        } else {
            this.hideModal();
        }
    },

    // 显示确认对话框
    showConfirm(title, message, options = {}) {
        return new Promise((resolve) => {
            const opts = {
                ...options,
                onConfirm: () => {
                    resolve(true);
                    return true;
                },
                onCancel: () => {
                    resolve(false);
                }
            };

            this.showModal(title, `<p>${message}</p>`, opts);
        });
    },

    // 显示Thread模态框
    async showThreadModal(momentUri) {
        // 创建模态框
        const modal = document.createElement('div');
        modal.className = 'thread-modal';
        modal.innerHTML = `
            <div class="thread-modal-content">
                <div class="thread-modal-header">
                    <h3><i class="fas fa-comments"></i> 对话线程</h3>
                    <button class="thread-modal-close" onclick="ModalsComponent.closeThreadModal()">&times;</button>
                </div>
                <div class="thread-content">
                    <div class="thread-loading">
                        <div class="spinner"></div>
                        正在加载对话...
                    </div>
                </div>
            </div>
        `;

        document.getElementById('modals-container').appendChild(modal);

        // 点击模态框外部关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeThreadModal();
            }
        });

        try {
            // 获取thread数据
            const threadData = await ApiService.getThread(momentUri, 10);
            console.log('Thread数据:', threadData);

            // 渲染thread内容
            this.renderThreadContent(modal, threadData);

        } catch (error) {
            console.error('加载对话失败:', error);
            const threadContent = modal.querySelector('.thread-content');
            threadContent.innerHTML = `
                <div class="thread-error">
                    <i class="fas fa-exclamation-triangle"></i>
                    加载对话失败: ${error.message}
                </div>
            `;
        }
    },

    // 关闭Thread模态框
    closeThreadModal() {
        const modal = document.querySelector('.thread-modal');
        if (modal) {
            modal.remove();
        }
    },

    // 渲染Thread内容
    renderThreadContent(modal, threadData) {
        const threadContent = modal.querySelector('.thread-content');

        if (!threadData.moment) {
            threadContent.innerHTML = `
                <div class="thread-error">
                    <i class="fas fa-exclamation-triangle"></i>
                    没有找到对话内容
                </div>
            `;
            return;
        }

        threadContent.innerHTML = '';

        // 渲染主帖子
        const mainCard = this.createThreadCard(threadData.moment, 0);
        threadContent.appendChild(mainCard);

        // 递归渲染回复
        if (threadData.replies && threadData.replies.length > 0) {
            this.renderThreadReplies(threadContent, threadData.replies, 1);
        }
    },

    // 递归渲染Thread回复
    renderThreadReplies(container, replies, depth) {
        replies.forEach(reply => {
            if (reply.moment) {
                const replyCard = this.createThreadCard(reply.moment, depth);
                container.appendChild(replyCard);

                // 递归渲染子回复
                if (reply.replies && reply.replies.length > 0) {
                    this.renderThreadReplies(container, reply.replies, depth + 1);
                }
            }
        });
    },

    // 创建Thread卡片
    createThreadCard(moment, depth) {
        const cardDiv = document.createElement('div');
        let className = 'thread-card';
        if (depth === 1) {
            className += ' is-reply';
        } else if (depth > 1) {
            className += ' is-nested-reply';
        }
        cardDiv.className = className;

        const authorAvatar = moment.author.avatar || '';
        const authorName = moment.author.displayName || moment.author.handle || 'Unknown';
        const authorHandle = moment.author.handle || '';
        const avatarContent = authorAvatar ?
            `<img src="${authorAvatar}" alt="${authorName}" style="width: 32px; height: 32px; border-radius: 50%; object-fit: cover;">` :
            `<div style="width: 32px; height: 32px; border-radius: 50%; background: ${Utils.getAvatarColor(authorHandle)}; color: white; display: flex; align-items: center; justify-content: center; font-weight: bold;">${authorName.charAt(0).toUpperCase()}</div>`;

        const createdTime = Utils.formatTime(moment.createdAt * 1000);

        // 处理媒体内容（简化版）
        let mediaContent = '';
        if (moment.embed) {
            if (moment.embed.images && moment.embed.images.length > 0) {
                mediaContent += '<div style="margin: 8px 0; display: flex; gap: 8px; flex-wrap: wrap;">';
                moment.embed.images.forEach(img => {
                    const imageUrl = img.thumb || `/api/blobs?id=${img.cid}`;
                    mediaContent += `<img src="${imageUrl}" alt="${img.alt || '图片'}" style="max-width: 120px; max-height: 120px; border-radius: 6px; cursor: pointer;" onclick="FeedComponent.showImageModal('${imageUrl}')">`;
                });
                mediaContent += '</div>';
            }

            if (moment.embed.video) {
                const videoUrl = moment.embed.video.video || `/api/blobs?id=${moment.embed.video.cid}`;
                mediaContent += `
                    <div style="margin: 8px 0;">
                        <video controls style="max-width: 300px; border-radius: 6px;">
                            <source src="${videoUrl}" type="video/mp4">
                            您的浏览器不支持视频播放。
                        </video>
                    </div>
                `;
            }
        }

        cardDiv.innerHTML = `
            <div style="display: flex; align-items: flex-start; gap: 12px;">
                ${avatarContent}
                <div style="flex: 1; min-width: 0;">
                    <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
                        <strong style="color: #374151;">${Utils.escapeHtml(authorName)}</strong>
                        <span style="color: var(--text-light); font-size: 0.9rem;">@${Utils.escapeHtml(authorHandle)}</span>
                        <span style="color: var(--text-light); font-size: 0.8rem;">·</span>
                        <span style="color: var(--text-light); font-size: 0.8rem;">${createdTime}</span>
                    </div>
                    <div style="color: #374151; line-height: 1.5; white-space: pre-wrap;">${Utils.escapeHtml(moment.text)}</div>
                    ${mediaContent}
                </div>
            </div>
        `;

        return cardDiv;
    }
};

// 导出模态框组件
window.ModalsComponent = ModalsComponent;