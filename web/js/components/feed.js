// Feed流组件
const FeedComponent = {
    feedCursor: null,
    replyData: {},

    // 初始化
    init() {
        this.renderFeedContainer();
        this.bindEvents();
    },

    // 渲染Feed容器
    renderFeedContainer() {
        const container = document.getElementById('feed-container');
        if (!container) return;

        container.innerHTML = `
            <div id="feed-list">
                <div class="welcome-animation">
                    <div class="welcome-icon">
                        <i class="fas fa-newspaper"></i>
                    </div>
                    <div class="welcome-text">
                        暂无动态，发布第一条帖子吧！
                    </div>
                </div>
            </div>
            <button id="load-more-btn" class="load-more-btn hidden">
                <i class="fas fa-plus"></i> 加载更多
            </button>
        `;
    },

    // 绑定事件
    bindEvents() {
        // 加载更多按钮
        const loadMoreBtn = document.getElementById('load-more-btn');
        if (loadMoreBtn) {
            loadMoreBtn.addEventListener('click', () => this.loadMoreFeed());
        }
    },

    // 加载Feed数据
    async loadFeedData(refresh = false) {
        if (refresh) {
            this.feedCursor = null;
        }

        try {
            const feedData = await ApiService.getFeed(this.feedCursor, APP_CONSTANTS.DEFAULTS.FEED_LIMIT);
            console.log('Feed数据:', feedData);

            if (refresh) {
                this.clearFeedList();
            }

            this.displayFeedCards(feedData.feed || []);
            this.feedCursor = feedData.cursor;

            // 更新加载更多按钮
            const loadMoreBtn = document.getElementById('load-more-btn');
            if (this.feedCursor && feedData.feed && feedData.feed.length > 0) {
                loadMoreBtn.classList.remove('hidden');
            } else {
                loadMoreBtn.classList.add('hidden');
            }

        } catch (error) {
            console.error('加载Feed失败:', error);
            Utils.showMessage('加载动态失败: ' + error.message, 'error');
        }
    },

    // 加载更多Feed
    loadMoreFeed() {
        this.loadFeedData(false);
    },

    // 清空Feed列表
    clearFeedList() {
        const feedList = document.getElementById('feed-list');
        if (feedList) {
            feedList.innerHTML = '';
        }
    },

    // 显示Feed卡片
    displayFeedCards(cards) {
        const feedList = document.getElementById('feed-list');
        if (!feedList) return;

        if (cards.length === 0 && feedList.children.length === 0) {
            feedList.innerHTML = `
                <div class="welcome-animation">
                    <div class="welcome-icon">
                        <i class="fas fa-newspaper"></i>
                    </div>
                    <div class="welcome-text">
                        暂无动态，发布第一条帖子吧！
                    </div>
                </div>
            `;
            return;
        }

        cards.forEach(feedCard => {
            if (feedCard.type === 'moment') {
                const cardElement = this.createMomentCard(feedCard.card);
                feedList.appendChild(cardElement);
            }
        });
    },

    // 创建Moment卡片
    createMomentCard(moment) {
        const cardDiv = document.createElement('div');
        cardDiv.className = 'feed-card';
        cardDiv.setAttribute('data-moment-id', moment.id);
        cardDiv.setAttribute('data-moment-uri', moment.uri || Utils.buildMomentURI(moment.id, window.App.currentUser?.did));

        const authorAvatar = moment.author.avatar || '';
        const authorName = moment.author.displayName || moment.author.handle || 'Unknown';
        const authorHandle = moment.author.handle || '';
        const avatarContent = authorAvatar ?
            `<img src="${authorAvatar}" alt="${authorName}">` :
            authorName.charAt(0).toUpperCase();

        const createdTime = Utils.formatRelativeTime(moment.createdAt * 1000);

        // 处理回复信息
        let replyInfo = '';
        if (moment.reply && moment.reply.parent) {
            replyInfo = `
                <div class="feed-card-reply-info">
                    <i class="fas fa-reply"></i> 回复了一条帖子
                </div>
            `;
        }

        // 处理媒体内容
        let mediaContent = '';
        if (moment.embed) {
            if (moment.embed.images && moment.embed.images.length > 0) {
                mediaContent += '<div class="feed-card-images">';
                moment.embed.images.forEach(img => {
                    const imageUrl = img.thumb || `/api/blobs?id=${img.cid}`;
                    mediaContent += `<img src="${imageUrl}" alt="${img.alt || '图片'}" onclick="FeedComponent.showImageModal('${imageUrl}')">`;
                });
                mediaContent += '</div>';
            }

            if (moment.embed.video) {
                const videoUrl = moment.embed.video.video || `/api/blobs?id=${moment.embed.video.cid}`;
                mediaContent += `
                    <div class="feed-card-video">
                        <video controls>
                            <source src="${videoUrl}" type="video/mp4">
                            您的浏览器不支持视频播放。
                        </video>
                    </div>
                `;
            }

            if (moment.embed.external) {
                const ext = moment.embed.external;
                const thumbUrl = ext.thumbURL || (ext.thumbCid ? `/api/blobs?id=${ext.thumbCid}` : '');
                mediaContent += `
                    <div class="feed-card-external">
                        ${thumbUrl ? `<img src="${thumbUrl}" class="feed-card-external-thumb" alt="缩略图">` : ''}
                        <div class="feed-card-external-content">
                            <div class="feed-card-external-title">${Utils.escapeHtml(ext.title || ext.uri)}</div>
                            ${ext.description ? `<div class="feed-card-external-desc">${Utils.escapeHtml(ext.description)}</div>` : ''}
                            <a href="${ext.uri}" class="feed-card-external-url" target="_blank">${ext.uri}</a>
                        </div>
                    </div>
                `;
            }
        }

        // 处理标签
        let tagsContent = '';
        if (moment.tags && moment.tags.length > 0) {
            tagsContent = '<div class="feed-card-tags">';
            moment.tags.forEach(tag => {
                tagsContent += `<span class="feed-card-tag">#${tag}</span>`;
            });
            tagsContent += '</div>';
        }

        const momentUri = moment.uri || Utils.buildMomentURI(moment.id, window.App.currentUser?.did);

        cardDiv.innerHTML = `
            <div class="feed-card-header">
                <div class="feed-card-avatar">${avatarContent}</div>
                <div class="feed-card-author">
                    <div class="handle">${Utils.escapeHtml(authorName)}</div>
                    <div class="did">@${Utils.escapeHtml(authorHandle)}</div>
                </div>
                <div class="feed-card-time">${createdTime}</div>
            </div>
            ${replyInfo}
            <div class="feed-card-content">${Utils.escapeHtml(moment.text)}</div>
            ${mediaContent}
            ${tagsContent}
            <div class="feed-card-actions">
                <button class="action-btn reply-btn" onclick="FeedComponent.showReplyForm('${moment.id}', '${momentUri}')">
                    <i class="fas fa-reply"></i> 回复
                </button>
                <button class="action-btn thread-btn" onclick="FeedComponent.showThread('${momentUri}')">
                    <i class="fas fa-comments"></i> 查看对话
                </button>
            </div>
            <div class="reply-form-container" id="reply-form-${moment.id}" style="display: none;"></div>
        `;

        return cardDiv;
    },

    // 显示图片模态框
    showImageModal(imageUrl) {
        const modal = document.createElement('div');
        modal.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.8);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 10000;
            cursor: pointer;
        `;

        const img = document.createElement('img');
        img.src = imageUrl;
        img.style.cssText = `
            max-width: 90%;
            max-height: 90%;
            object-fit: contain;
            border-radius: 8px;
        `;

        modal.appendChild(img);
        document.body.appendChild(modal);

        modal.addEventListener('click', () => {
            document.body.removeChild(modal);
        });
    },

    // 显示回复表单
    showReplyForm(momentId, momentUri) {
        const container = document.getElementById(`reply-form-${momentId}`);
        if (!container) return;

        // 如果已经显示，则隐藏
        if (container.style.display !== 'none') {
            this.hideReplyForm(momentId);
            return;
        }

        // 初始化回复数据
        this.replyData[momentId] = {
            images: [],
            video: null,
            external: null
        };

        // 创建回复表单
        container.innerHTML = this.createReplyFormHTML(momentId, momentUri);
        container.style.display = 'block';

        // 绑定事件
        this.bindReplyFormEvents(momentId);

        // 聚焦到文本框
        const textarea = container.querySelector('.reply-textarea');
        if (textarea) {
            textarea.focus();
        }
    },

    // 隐藏回复表单
    hideReplyForm(momentId) {
        const container = document.getElementById(`reply-form-${momentId}`);
        if (container) {
            container.style.display = 'none';
            container.innerHTML = '';
        }
        // 清理回复数据
        delete this.replyData[momentId];
    },

    // 创建回复表单HTML
    createReplyFormHTML(momentId, momentUri) {
        return `
            <div class="reply-form">
                <div class="reply-to-info">
                    <i class="fas fa-reply"></i> 回复此帖子
                </div>

                <textarea
                    class="reply-textarea"
                    placeholder="输入你的回复..."
                    maxlength="3000"
                    id="reply-text-${momentId}"
                ></textarea>

                <!-- 媒体预览区域 -->
                <div id="reply-media-preview-${momentId}" class="media-preview hidden">
                    <div id="reply-images-preview-${momentId}" class="images-preview"></div>
                    <div id="reply-video-preview-${momentId}" class="video-preview"></div>
                    <div id="reply-external-preview-${momentId}" class="external-preview"></div>
                </div>

                <div class="reply-actions">
                    <div class="reply-media-buttons">
                        <button type="button" class="media-btn" onclick="FeedComponent.addReplyImages('${momentId}')" title="添加图片">
                            <i class="fas fa-image"></i>
                        </button>
                        <button type="button" class="media-btn" onclick="FeedComponent.addReplyVideo('${momentId}')" title="添加视频">
                            <i class="fas fa-video"></i>
                        </button>
                    </div>

                    <div class="reply-submit-actions">
                        <div class="reply-char-counter">
                            <span id="reply-char-count-${momentId}">0</span>/3000
                        </div>
                        <button type="button" class="btn btn-danger" onclick="FeedComponent.hideReplyForm('${momentId}')">
                            <i class="fas fa-times"></i> 取消
                        </button>
                        <button type="button" class="btn btn-success" onclick="FeedComponent.submitReply('${momentId}', '${momentUri}')">
                            <i class="fas fa-reply"></i> 发布回复
                        </button>
                    </div>
                </div>

                <!-- 隐藏的文件输入 -->
                <input type="file" id="reply-images-input-${momentId}" accept="image/*" multiple style="display: none;">
                <input type="file" id="reply-video-input-${momentId}" accept="video/*" style="display: none;">
            </div>
        `;
    },

    // 绑定回复表单事件
    bindReplyFormEvents(momentId) {
        // 字符计数
        const textarea = document.getElementById(`reply-text-${momentId}`);
        if (textarea) {
            textarea.addEventListener('input', () => this.updateReplyCharCount(momentId));
        }

        // 文件输入事件
        const imagesInput = document.getElementById(`reply-images-input-${momentId}`);
        if (imagesInput) {
            imagesInput.addEventListener('change', (e) => this.handleReplyImagesSelect(momentId, e));
        }

        const videoInput = document.getElementById(`reply-video-input-${momentId}`);
        if (videoInput) {
            videoInput.addEventListener('change', (e) => this.handleReplyVideoSelect(momentId, e));
        }
    },

    // 更新回复字符计数
    updateReplyCharCount(momentId) {
        const textarea = document.getElementById(`reply-text-${momentId}`);
        const charCount = document.getElementById(`reply-char-count-${momentId}`);
        const charCounter = charCount.parentElement;

        const length = textarea.value.length;
        charCount.textContent = length;

        if (length > 2800) {
            charCounter.className = 'reply-char-counter error';
        } else if (length > 2500) {
            charCounter.className = 'reply-char-counter warning';
        } else {
            charCounter.className = 'reply-char-counter';
        }
    },

    // 添加回复图片
    addReplyImages(momentId) {
        document.getElementById(`reply-images-input-${momentId}`).click();
    },

    // 添加回复视频
    addReplyVideo(momentId) {
        document.getElementById(`reply-video-input-${momentId}`).click();
    },

    // 处理回复图片选择
    async handleReplyImagesSelect(momentId, event) {
        const files = Array.from(event.target.files);
        const maxImages = APP_CONSTANTS.FILE_LIMITS.MAX_IMAGES;

        if (!this.replyData[momentId]) this.replyData[momentId] = { images: [], video: null, external: null };

        if (this.replyData[momentId].images.length + files.length > maxImages) {
            Utils.showMessage(`最多只能添加${maxImages}张图片`, 'warning');
            return;
        }

        for (const file of files) {
            if (!file.type.startsWith('image/')) {
                Utils.showMessage('只支持图片文件', 'error');
                continue;
            }

            if (file.size > APP_CONSTANTS.FILE_LIMITS.IMAGE_MAX_SIZE) {
                Utils.showMessage('图片大小不能超过10MB', 'error');
                continue;
            }

            try {
                const uploadResult = await ApiService.uploadFile(file);
                const imageData = {
                    cid: uploadResult.cid,
                    file: file,
                    url: uploadResult.url || `/api/blobs?id=${uploadResult.cid}`,
                    alt: file.name
                };
                this.replyData[momentId].images.push(imageData);
                this.updateReplyMediaPreview(momentId);
            } catch (error) {
                Utils.showMessage('图片上传失败: ' + error.message, 'error');
            }
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 处理回复视频选择
    async handleReplyVideoSelect(momentId, event) {
        const file = event.target.files[0];
        if (!file) return;

        if (!file.type.startsWith('video/')) {
            Utils.showMessage('只支持视频文件', 'error');
            return;
        }

        if (file.size > APP_CONSTANTS.FILE_LIMITS.VIDEO_MAX_SIZE) {
            Utils.showMessage('视频大小不能超过100MB', 'error');
            return;
        }

        try {
            const uploadResult = await ApiService.uploadFile(file);
            if (!this.replyData[momentId]) this.replyData[momentId] = { images: [], video: null, external: null };

            this.replyData[momentId].video = {
                cid: uploadResult.cid,
                file: file,
                url: uploadResult.url || `/api/blobs?id=${uploadResult.cid}`,
                alt: file.name
            };
            this.updateReplyMediaPreview(momentId);
        } catch (error) {
            Utils.showMessage('视频上传失败: ' + error.message, 'error');
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 更新回复媒体预览
    updateReplyMediaPreview(momentId) {
        const mediaPreview = document.getElementById(`reply-media-preview-${momentId}`);
        const imagesPreview = document.getElementById(`reply-images-preview-${momentId}`);
        const videoPreview = document.getElementById(`reply-video-preview-${momentId}`);

        if (!mediaPreview || !this.replyData[momentId]) return;

        // 清空预览
        imagesPreview.innerHTML = '';
        videoPreview.innerHTML = '';

        let hasMedia = false;
        const data = this.replyData[momentId];

        // 图片预览
        if (data.images.length > 0) {
            hasMedia = true;
            data.images.forEach((imageData, index) => {
                const imageItem = document.createElement('div');
                imageItem.className = 'image-preview-item';
                imageItem.innerHTML = `
                    <img src="${imageData.url}" alt="${imageData.alt}">
                    <button class="remove-btn" onclick="FeedComponent.removeReplyImage('${momentId}', ${index})">
                        <i class="fas fa-times"></i>
                    </button>
                `;
                imagesPreview.appendChild(imageItem);
            });
        }

        // 视频预览
        if (data.video) {
            hasMedia = true;
            const videoItem = document.createElement('div');
            videoItem.className = 'video-preview-item';
            videoItem.innerHTML = `
                <video controls>
                    <source src="${data.video.url}" type="${data.video.file.type}">
                    您的浏览器不支持视频播放。
                </video>
                <button class="remove-btn" onclick="FeedComponent.removeReplyVideo('${momentId}')">
                    <i class="fas fa-times"></i>
                </button>
            `;
            videoPreview.appendChild(videoItem);
        }

        // 显示或隐藏媒体预览区域
        if (hasMedia) {
            mediaPreview.classList.remove('hidden');
        } else {
            mediaPreview.classList.add('hidden');
        }
    },

    // 移除回复图片
    removeReplyImage(momentId, index) {
        if (this.replyData[momentId] && this.replyData[momentId].images) {
            this.replyData[momentId].images.splice(index, 1);
            this.updateReplyMediaPreview(momentId);
        }
    },

    // 移除回复视频
    removeReplyVideo(momentId) {
        if (this.replyData[momentId]) {
            this.replyData[momentId].video = null;
            this.updateReplyMediaPreview(momentId);
        }
    },

    // 提交回复
    async submitReply(momentId, momentUri) {
        const textarea = document.getElementById(`reply-text-${momentId}`);
        const text = textarea.value.trim();

        if (!text && (!this.replyData[momentId] || (this.replyData[momentId].images.length === 0 && !this.replyData[momentId].video))) {
            Utils.showMessage('请输入回复内容或添加媒体', 'warning');
            return;
        }

        if (text.length > APP_CONSTANTS.TEXT_LIMITS.POST_MAX_LENGTH) {
            Utils.showMessage('回复内容不能超过3000个字符', 'error');
            return;
        }

        try {
            // 构建回复数据
            const replyMomentData = {
                text: text,
                parentId: momentId,
                facets: Utils.parseFacets(text),
                langs: ['zh', 'en'],
                tags: Utils.extractTags(text)
            };

            // 添加媒体内容
            const data = this.replyData[momentId];
            if (data) {
                if (data.images.length > 0) {
                    replyMomentData.images = data.images.map(img => ({ cid: img.cid }));
                }

                if (data.video) {
                    replyMomentData.video = { cid: data.video.cid };
                }
            }

            // 发送创建回复请求
            const result = await ApiService.createMoment(replyMomentData);
            console.log('回复发布成功:', result);

            // 隐藏回复表单
            this.hideReplyForm(momentId);
            Utils.showMessage('回复发布成功！', 'success');

            // 刷新Feed流
            setTimeout(() => {
                this.loadFeedData(true);
            }, 1000);

        } catch (error) {
            console.error('发布回复失败:', error);
            Utils.showMessage('发布回复失败: ' + error.message, 'error');
        }
    },

    // 显示Thread对话
    async showThread(momentUri) {
        ModalsComponent.showThreadModal(momentUri);
    }
};

// 导出Feed组件
window.FeedComponent = FeedComponent;