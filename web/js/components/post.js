// 帖子发布组件
const PostComponent = {
    // 帖子数据
    postData: {
        images: [],
        video: null,
        external: null
    },

    // 初始化
    init() {
        this.renderPostContainer();
        this.bindEvents();
    },

    // 渲染帖子容器
    renderPostContainer() {
        const container = document.getElementById('post-container');
        if (!container) return;

        container.innerHTML = `
            <form id="post-form" class="post-form">
                <div class="form-group">
                    <label for="post-content">
                        <i class="fas fa-pen"></i> 分享你的想法
                    </label>
                    <textarea
                        id="post-content"
                        class="post-textarea"
                        placeholder="今天有什么想分享的吗？#标签 @mention https://link.com"
                        maxlength="3000"
                    ></textarea>
                </div>

                <!-- 媒体预览区域 -->
                <div id="media-preview" class="media-preview hidden">
                    <div id="images-preview" class="images-preview"></div>
                    <div id="video-preview" class="video-preview"></div>
                    <div id="external-preview" class="external-preview"></div>
                </div>

                <!-- 外部链接输入 -->
                <div id="external-input" class="form-group hidden">
                    <label for="external-url">
                        <i class="fas fa-link"></i> 外部链接
                    </label>
                    <input
                        id="external-url"
                        type="url"
                        placeholder="输入外部链接URL"
                    >
                    <button type="button" id="add-external-btn" class="btn">
                        <i class="fas fa-plus"></i> 添加链接
                    </button>
                    <button type="button" id="cancel-external-btn" class="btn btn-danger">
                        <i class="fas fa-times"></i> 取消
                    </button>
                </div>

                <!-- 帖子工具栏 -->
                <div class="post-toolbar">
                    <div class="media-buttons">
                        <button type="button" id="add-images-btn" class="media-btn" title="添加图片">
                            <i class="fas fa-image"></i>
                        </button>
                        <button type="button" id="add-video-btn" class="media-btn" title="添加视频">
                            <i class="fas fa-video"></i>
                        </button>
                        <button type="button" id="add-link-btn" class="media-btn" title="添加外部链接">
                            <i class="fas fa-link"></i>
                        </button>
                    </div>

                    <div class="post-actions">
                        <div class="char-counter">
                            <span id="char-count">0</span>/3000
                        </div>
                        <button type="submit" class="btn btn-success" id="post-btn">
                            <i class="fas fa-share"></i> 发布
                        </button>
                    </div>
                </div>
            </form>

            <!-- 隐藏的文件输入 -->
            <input type="file" id="images-input" accept="image/*" multiple style="display: none;">
            <input type="file" id="video-input" accept="video/*" style="display: none;">
        `;
    },

    // 绑定事件
    bindEvents() {
        // 字符计数
        const postContent = document.getElementById('post-content');
        if (postContent) {
            postContent.addEventListener('input', () => this.updateCharCount());
        }

        // 媒体按钮事件
        const addImagesBtn = document.getElementById('add-images-btn');
        if (addImagesBtn) {
            addImagesBtn.addEventListener('click', () => {
                document.getElementById('images-input').click();
            });
        }

        const addVideoBtn = document.getElementById('add-video-btn');
        if (addVideoBtn) {
            addVideoBtn.addEventListener('click', () => {
                document.getElementById('video-input').click();
            });
        }

        const addLinkBtn = document.getElementById('add-link-btn');
        if (addLinkBtn) {
            addLinkBtn.addEventListener('click', () => this.toggleExternalInput());
        }

        // 文件输入事件
        const imagesInput = document.getElementById('images-input');
        if (imagesInput) {
            imagesInput.addEventListener('change', (e) => this.handleImagesSelect(e));
        }

        const videoInput = document.getElementById('video-input');
        if (videoInput) {
            videoInput.addEventListener('change', (e) => this.handleVideoSelect(e));
        }

        // 外部链接事件
        const addExternalBtn = document.getElementById('add-external-btn');
        if (addExternalBtn) {
            addExternalBtn.addEventListener('click', () => this.handleExternalAdd());
        }

        const cancelExternalBtn = document.getElementById('cancel-external-btn');
        if (cancelExternalBtn) {
            cancelExternalBtn.addEventListener('click', () => this.cancelExternalInput());
        }

        // 帖子表单提交
        const postForm = document.getElementById('post-form');
        if (postForm) {
            postForm.addEventListener('submit', (e) => this.handlePostSubmit(e));
        }
    },

    // 更新字符计数
    updateCharCount() {
        const postContent = document.getElementById('post-content');
        const charCount = document.getElementById('char-count');
        const charCounter = charCount.parentElement;

        const length = postContent.value.length;
        charCount.textContent = length;

        if (length > 2800) {
            charCounter.className = 'char-counter error';
        } else if (length > 2500) {
            charCounter.className = 'char-counter warning';
        } else {
            charCounter.className = 'char-counter';
        }
    },

    // 切换外部链接输入
    toggleExternalInput() {
        const externalInput = document.getElementById('external-input');
        const addLinkBtn = document.getElementById('add-link-btn');

        if (externalInput.classList.contains('hidden')) {
            externalInput.classList.remove('hidden');
            addLinkBtn.classList.add('active');
            document.getElementById('external-url').focus();
        } else {
            this.cancelExternalInput();
        }
    },

    // 取消外部链接输入
    cancelExternalInput() {
        const externalInput = document.getElementById('external-input');
        const addLinkBtn = document.getElementById('add-link-btn');

        externalInput.classList.add('hidden');
        addLinkBtn.classList.remove('active');
        document.getElementById('external-url').value = '';
    },

    // 处理图片选择
    async handleImagesSelect(event) {
        const files = Array.from(event.target.files);
        const maxImages = APP_CONSTANTS.FILE_LIMITS.MAX_IMAGES;

        if (this.postData.images.length + files.length > maxImages) {
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
                this.postData.images.push(imageData);
                this.updateMediaPreview();
            } catch (error) {
                Utils.showMessage('图片上传失败: ' + error.message, 'error');
            }
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 处理视频选择
    async handleVideoSelect(event) {
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
            this.postData.video = {
                cid: uploadResult.cid,
                file: file,
                url: uploadResult.url || `/api/blobs?id=${uploadResult.cid}`,
                alt: file.name
            };
            this.updateMediaPreview();
        } catch (error) {
            Utils.showMessage('视频上传失败: ' + error.message, 'error');
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 处理外部链接添加
    async handleExternalAdd() {
        const urlInput = document.getElementById('external-url');
        const url = urlInput.value.trim();

        if (!url) {
            Utils.showMessage('请输入外部链接URL', 'warning');
            return;
        }

        try {
            new URL(url); // 验证URL格式
        } catch {
            Utils.showMessage('请输入有效的URL', 'error');
            return;
        }

        try {
            // 获取链接元数据（暂时使用基本信息）
            this.postData.external = {
                uri: url,
                title: url,
                description: '',
                thumbCid: ''
            };
            this.updateMediaPreview();
            this.cancelExternalInput();
        } catch (error) {
            // 即使获取元数据失败，也添加基本链接
            this.postData.external = {
                uri: url,
                title: url,
                description: '',
                thumbCid: ''
            };
            this.updateMediaPreview();
            this.cancelExternalInput();
        }
    },

    // 更新媒体预览
    updateMediaPreview() {
        const mediaPreview = document.getElementById('media-preview');
        const imagesPreview = document.getElementById('images-preview');
        const videoPreview = document.getElementById('video-preview');
        const externalPreview = document.getElementById('external-preview');

        // 清空预览
        imagesPreview.innerHTML = '';
        videoPreview.innerHTML = '';
        externalPreview.innerHTML = '';

        let hasMedia = false;

        // 图片预览
        if (this.postData.images.length > 0) {
            hasMedia = true;
            this.postData.images.forEach((imageData, index) => {
                const imageItem = document.createElement('div');
                imageItem.className = 'image-preview-item';
                imageItem.innerHTML = `
                    <img src="${imageData.url}" alt="${imageData.alt}">
                    <button class="remove-btn" onclick="PostComponent.removeImage(${index})">
                        <i class="fas fa-times"></i>
                    </button>
                `;
                imagesPreview.appendChild(imageItem);
            });
        }

        // 视频预览
        if (this.postData.video) {
            hasMedia = true;
            const videoItem = document.createElement('div');
            videoItem.className = 'video-preview-item';
            videoItem.innerHTML = `
                <video controls>
                    <source src="${this.postData.video.url}" type="${this.postData.video.file.type}">
                    您的浏览器不支持视频播放。
                </video>
                <button class="remove-btn" onclick="PostComponent.removeVideo()">
                    <i class="fas fa-times"></i>
                </button>
            `;
            videoPreview.appendChild(videoItem);
        }

        // 外部链接预览
        if (this.postData.external) {
            hasMedia = true;
            const externalItem = document.createElement('div');
            externalItem.className = 'external-preview-item';
            externalItem.innerHTML = `
                ${this.postData.external.thumbCid ? `<img src="/api/blobs?id=${this.postData.external.thumbCid}" class="external-thumb" alt="缩略图">` : '<div class="external-thumb"></div>'}
                <div class="external-info">
                    <div class="external-title">${Utils.escapeHtml(this.postData.external.title)}</div>
                    <div class="external-desc">${Utils.escapeHtml(this.postData.external.description)}</div>
                    <a href="${this.postData.external.uri}" class="external-url" target="_blank">${this.postData.external.uri}</a>
                </div>
                <button class="remove-btn" onclick="PostComponent.removeExternal()">
                    <i class="fas fa-times"></i>
                </button>
            `;
            externalPreview.appendChild(externalItem);
        }

        // 显示或隐藏媒体预览区域
        if (hasMedia) {
            mediaPreview.classList.remove('hidden');
        } else {
            mediaPreview.classList.add('hidden');
        }
    },

    // 移除图片
    removeImage(index) {
        this.postData.images.splice(index, 1);
        this.updateMediaPreview();
    },

    // 移除视频
    removeVideo() {
        this.postData.video = null;
        this.updateMediaPreview();
    },

    // 移除外部链接
    removeExternal() {
        this.postData.external = null;
        this.updateMediaPreview();
    },

    // 处理帖子提交
    async handlePostSubmit(event) {
        event.preventDefault();

        const postContent = document.getElementById('post-content');
        const postBtn = document.getElementById('post-btn');
        const text = postContent.value.trim();

        if (!text && this.postData.images.length === 0 && !this.postData.video && !this.postData.external) {
            Utils.showMessage('请输入文本内容或添加媒体', 'warning');
            return;
        }

        if (text.length > APP_CONSTANTS.TEXT_LIMITS.POST_MAX_LENGTH) {
            Utils.showMessage('文本内容不能超过3000个字符', 'error');
            return;
        }

        // 禁用发布按钮
        postBtn.disabled = true;
        postBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 发布中...';

        try {
            // 构建帖子数据
            const momentData = {
                text: text,
                facets: Utils.parseFacets(text),
                langs: ['zh', 'en'],
                tags: Utils.extractTags(text)
            };

            // 添加图片
            if (this.postData.images.length > 0) {
                momentData.images = this.postData.images.map(img => ({ cid: img.cid }));
            }

            // 添加视频
            if (this.postData.video) {
                momentData.video = { cid: this.postData.video.cid };
            }

            // 添加外部链接
            if (this.postData.external) {
                momentData.external = this.postData.external;
            }

            // 发送创建帖子请求
            const result = await ApiService.createMoment(momentData);
            console.log('帖子发布成功:', result);

            // 清空表单
            this.resetPostForm();
            Utils.showMessage('帖子发布成功！', 'success');

            // 刷新Feed流
            setTimeout(() => {
                FeedComponent.loadFeedData(true);
            }, 1000);

        } catch (error) {
            console.error('发布帖子失败:', error);
            Utils.showMessage('发布失败: ' + error.message, 'error');
        } finally {
            // 恢复发布按钮
            postBtn.disabled = false;
            postBtn.innerHTML = '<i class="fas fa-share"></i> 发布';
        }
    },

    // 重置帖子表单
    resetPostForm() {
        document.getElementById('post-content').value = '';
        this.postData.images = [];
        this.postData.video = null;
        this.postData.external = null;
        this.updateCharCount();
        this.updateMediaPreview();
        this.cancelExternalInput();
    }
};

// 导出帖子组件
window.PostComponent = PostComponent;