// 用户资料组件
const UserProfileComponent = {
    profileEditData: {
        avatarFile: null,
        avatarCID: null,
        bannerFile: null,
        bannerCID: null
    },

    // 初始化
    init() {
        this.renderUserInfo();
    },

    // 渲染用户信息容器
    renderUserInfo() {
        const container = document.getElementById('user-info-container');
        if (!container) return;

        container.innerHTML = `
            <div class="user-profile-section">
                <div class="user-avatar-container">
                    <div class="user-avatar" id="user-avatar">
                        <i class="fas fa-user"></i>
                    </div>
                    <button class="edit-avatar-btn" onclick="UserProfileComponent.showEditProfileModal()">
                        <i class="fas fa-camera"></i>
                    </button>
                </div>
                <div class="user-basic-info">
                    <div class="user-display-name" id="user-display-name">用户名</div>
                    <div class="user-handle-display" id="user-handle-display">@handle</div>
                    <div class="user-description" id="user-description">还没有个人简介</div>
                </div>
                <button class="edit-profile-btn" onclick="UserProfileComponent.showEditProfileModal()">
                    <i class="fas fa-edit"></i> 编辑资料
                </button>
            </div>
            <div class="user-info">
                <p>
                    <strong><i class="fas fa-fingerprint"></i> DID:</strong>
                    <span id="user-did"></span>
                </p>
                <p>
                    <strong><i class="fas fa-at"></i> Handle:</strong>
                    <span id="user-handle-full"></span>
                </p>
                <p>
                    <strong><i class="fas fa-calendar"></i> 创建时间:</strong>
                    <span id="user-created-at">-</span>
                </p>
            </div>
        `;
    },

    // 更新用户显示
    updateUserDisplay(user) {
        // 基本信息
        const userDid = document.getElementById('user-did');
        const userHandleFull = document.getElementById('user-handle-full');
        const userHandle = document.getElementById('user-handle');

        if (userDid) userDid.textContent = user.did;
        if (userHandleFull) userHandleFull.textContent = '@' + user.handle;
        if (userHandle) userHandle.textContent = '@' + user.handle;

        // 显示名称
        const displayName = user.displayName || user.handle || '用户';
        const userDisplayName = document.getElementById('user-display-name');
        const userHandleDisplay = document.getElementById('user-handle-display');

        if (userDisplayName) userDisplayName.textContent = displayName;
        if (userHandleDisplay) userHandleDisplay.textContent = '@' + user.handle;

        // 个人简介
        const description = user.description || '还没有个人简介';
        const userDescription = document.getElementById('user-description');
        if (userDescription) userDescription.textContent = description;

        // 头像
        const userAvatar = document.getElementById('user-avatar');
        if (userAvatar) {
            if (user.avatar) {
                userAvatar.innerHTML = `<img src="${user.avatar}" alt="${displayName}">`;
            } else {
                userAvatar.innerHTML = `<i class="fas fa-user"></i>`;
                userAvatar.style.background = Utils.getAvatarColor(user.handle);
            }
        }

        // 创建时间
        const userCreatedAt = document.getElementById('user-created-at');
        if (userCreatedAt && user.createdAt) {
            const createdDate = new Date(user.createdAt * 1000);
            userCreatedAt.textContent = createdDate.toLocaleDateString('zh-CN');
        }
    },

    // 显示编辑资料模态框
    showEditProfileModal() {
        const modal = this.createEditProfileModal();
        document.getElementById('modals-container').appendChild(modal);

        // 填充当前用户信息
        if (window.App.currentUser) {
            const user = window.App.currentUser;
            document.getElementById('edit-display-name').value = user.displayName || '';
            document.getElementById('edit-description').value = user.description || '';

            // 更新字符计数
            this.updateProfileCharCount('display-name', APP_CONSTANTS.TEXT_LIMITS.DISPLAY_NAME_MAX_LENGTH);
            this.updateProfileCharCount('description', APP_CONSTANTS.TEXT_LIMITS.DESCRIPTION_MAX_LENGTH);

            // 显示当前头像
            const avatarPreview = document.getElementById('avatar-preview');
            if (user.avatar) {
                avatarPreview.innerHTML = `<img src="${user.avatar}" alt="头像">`;
            } else {
                avatarPreview.innerHTML = '<i class="fas fa-user"></i>';
                avatarPreview.style.background = Utils.getAvatarColor(user.handle);
            }

            // 显示当前背景图
            const bannerPreview = document.getElementById('banner-preview');
            if (user.banner) {
                bannerPreview.innerHTML = `<img src="${user.banner}" alt="背景图">`;
            } else {
                bannerPreview.innerHTML = '<i class="fas fa-mountain"></i><span>点击或拖拽上传背景图</span>';
            }
        }

        this.bindProfileEditEvents();
    },

    // 创建编辑资料模态框
    createEditProfileModal() {
        const modal = document.createElement('div');
        modal.id = 'edit-profile-modal';
        modal.className = 'profile-modal';
        modal.innerHTML = `
            <div class="profile-modal-content">
                <div class="profile-modal-header">
                    <h3><i class="fas fa-user-edit"></i> 编辑个人资料</h3>
                    <button class="profile-modal-close" onclick="UserProfileComponent.hideEditProfileModal()">&times;</button>
                </div>
                <div class="profile-modal-body">
                    <form id="edit-profile-form" enctype="multipart/form-data">
                        <!-- Avatar Section -->
                        <div class="profile-section">
                            <label class="profile-label">
                                <i class="fas fa-image"></i> 头像
                            </label>
                            <div class="avatar-edit-container">
                                <div class="avatar-preview" id="avatar-preview">
                                    <i class="fas fa-user"></i>
                                </div>
                                <div class="avatar-actions">
                                    <input type="file" id="avatar-input" accept="image/*" style="display: none;">
                                    <button type="button" class="btn" onclick="UserProfileComponent.selectAvatar()">
                                        <i class="fas fa-upload"></i> 选择头像
                                    </button>
                                    <button type="button" class="btn btn-danger" onclick="UserProfileComponent.removeAvatar()">
                                        <i class="fas fa-trash"></i> 移除
                                    </button>
                                </div>
                            </div>
                        </div>

                        <!-- Banner Section -->
                        <div class="profile-section">
                            <label class="profile-label">
                                <i class="fas fa-image"></i> 背景图
                            </label>
                            <div class="banner-edit-container">
                                <div class="banner-preview" id="banner-preview">
                                    <i class="fas fa-mountain"></i>
                                    <span>点击或拖拽上传背景图</span>
                                </div>
                                <div class="banner-actions">
                                    <input type="file" id="banner-input" accept="image/*" style="display: none;">
                                    <button type="button" class="btn" onclick="UserProfileComponent.selectBanner()">
                                        <i class="fas fa-upload"></i> 选择背景
                                    </button>
                                    <button type="button" class="btn btn-danger" onclick="UserProfileComponent.removeBanner()">
                                        <i class="fas fa-trash"></i> 移除
                                    </button>
                                </div>
                            </div>
                        </div>

                        <!-- Display Name -->
                        <div class="profile-section">
                            <label for="edit-display-name" class="profile-label">
                                <i class="fas fa-user-tag"></i> 显示名称
                            </label>
                            <input
                                type="text"
                                id="edit-display-name"
                                maxlength="50"
                                placeholder="输入你的显示名称"
                                class="profile-input"
                            >
                            <div class="char-counter">
                                <span id="display-name-count">0</span>/50
                            </div>
                        </div>

                        <!-- Description -->
                        <div class="profile-section">
                            <label for="edit-description" class="profile-label">
                                <i class="fas fa-align-left"></i> 个人简介
                            </label>
                            <textarea
                                id="edit-description"
                                maxlength="300"
                                rows="4"
                                placeholder="介绍一下你自己..."
                                class="profile-textarea"
                            ></textarea>
                            <div class="char-counter">
                                <span id="description-count">0</span>/300
                            </div>
                        </div>

                        <!-- Upload Progress -->
                        <div id="profile-upload-progress" class="upload-progress hidden">
                            <div class="upload-progress-bar">
                                <div class="upload-progress-fill" id="profile-progress-fill"></div>
                            </div>
                            <div class="upload-progress-text" id="profile-progress-text">上传中...</div>
                        </div>

                        <!-- Actions -->
                        <div class="profile-actions">
                            <button type="button" class="btn btn-danger" onclick="UserProfileComponent.hideEditProfileModal()">
                                <i class="fas fa-times"></i> 取消
                            </button>
                            <button type="submit" class="btn btn-success" id="save-profile-btn">
                                <i class="fas fa-save"></i> 保存
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        `;

        // 点击模态框外部关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.hideEditProfileModal();
            }
        });

        return modal;
    },

    // 隐藏编辑资料模态框
    hideEditProfileModal() {
        const modal = document.getElementById('edit-profile-modal');
        if (modal) {
            modal.remove();
        }
        // 重置编辑数据
        this.profileEditData = {
            avatarFile: null,
            avatarCID: null,
            bannerFile: null,
            bannerCID: null
        };
    },

    // 绑定编辑资料事件
    bindProfileEditEvents() {
        // 字符计数事件
        const displayNameInput = document.getElementById('edit-display-name');
        const descriptionInput = document.getElementById('edit-description');

        if (displayNameInput) {
            displayNameInput.addEventListener('input', () =>
                this.updateProfileCharCount('display-name', APP_CONSTANTS.TEXT_LIMITS.DISPLAY_NAME_MAX_LENGTH)
            );
        }

        if (descriptionInput) {
            descriptionInput.addEventListener('input', () =>
                this.updateProfileCharCount('description', APP_CONSTANTS.TEXT_LIMITS.DESCRIPTION_MAX_LENGTH)
            );
        }

        // 文件选择事件
        const avatarInput = document.getElementById('avatar-input');
        const bannerInput = document.getElementById('banner-input');

        if (avatarInput) {
            avatarInput.addEventListener('change', (e) => this.handleAvatarSelect(e));
        }

        if (bannerInput) {
            bannerInput.addEventListener('change', (e) => this.handleBannerSelect(e));
        }

        // 表单提交事件
        const form = document.getElementById('edit-profile-form');
        if (form) {
            form.addEventListener('submit', (e) => this.handleProfileSubmit(e));
        }

        // 背景图点击上传
        const bannerPreview = document.getElementById('banner-preview');
        if (bannerPreview) {
            bannerPreview.addEventListener('click', () => this.selectBanner());
        }
    },

    // 更新字符计数
    updateProfileCharCount(type, maxLength) {
        const input = document.getElementById(`edit-${type.replace('-', '-')}`);
        const counter = document.getElementById(`${type}-count`);

        if (!input || !counter) return;

        const length = input.value.length;
        counter.textContent = length;

        const counterElement = counter.parentElement;
        if (length > maxLength * 0.9) {
            counterElement.className = 'char-counter warning';
        } else if (length >= maxLength) {
            counterElement.className = 'char-counter error';
        } else {
            counterElement.className = 'char-counter';
        }
    },

    // 选择头像
    selectAvatar() {
        document.getElementById('avatar-input').click();
    },

    // 选择背景图
    selectBanner() {
        document.getElementById('banner-input').click();
    },

    // 移除头像
    removeAvatar() {
        this.profileEditData.avatarFile = null;
        this.profileEditData.avatarCID = null;

        const avatarPreview = document.getElementById('avatar-preview');
        if (avatarPreview && window.App.currentUser) {
            avatarPreview.innerHTML = '<i class="fas fa-user"></i>';
            avatarPreview.style.background = Utils.getAvatarColor(window.App.currentUser.handle);
        }
    },

    // 移除背景图
    removeBanner() {
        this.profileEditData.bannerFile = null;
        this.profileEditData.bannerCID = null;

        const bannerPreview = document.getElementById('banner-preview');
        if (bannerPreview) {
            bannerPreview.innerHTML = '<i class="fas fa-mountain"></i><span>点击或拖拽上传背景图</span>';
        }
    },

    // 处理头像选择
    async handleAvatarSelect(event) {
        const file = event.target.files[0];
        if (!file) return;

        if (!file.type.startsWith('image/')) {
            Utils.showMessage('请选择图片文件', 'error');
            return;
        }

        if (file.size > APP_CONSTANTS.FILE_LIMITS.AVATAR_MAX_SIZE) {
            Utils.showMessage('头像文件大小不能超过5MB', 'error');
            return;
        }

        try {
            // 预览图片
            const reader = new FileReader();
            reader.onload = (e) => {
                const avatarPreview = document.getElementById('avatar-preview');
                if (avatarPreview) {
                    avatarPreview.innerHTML = `<img src="${e.target.result}" alt="头像预览">`;
                }
            };
            reader.readAsDataURL(file);

            // 上传文件
            const uploadResult = await ApiService.uploadFile(file);
            this.profileEditData.avatarFile = file;
            this.profileEditData.avatarCID = uploadResult.cid;

            Utils.showMessage('头像上传成功', 'success');
        } catch (error) {
            Utils.showMessage('头像上传失败: ' + error.message, 'error');
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 处理背景图选择
    async handleBannerSelect(event) {
        const file = event.target.files[0];
        if (!file) return;

        if (!file.type.startsWith('image/')) {
            Utils.showMessage('请选择图片文件', 'error');
            return;
        }

        if (file.size > APP_CONSTANTS.FILE_LIMITS.BANNER_MAX_SIZE) {
            Utils.showMessage('背景图文件大小不能超过10MB', 'error');
            return;
        }

        try {
            // 预览图片
            const reader = new FileReader();
            reader.onload = (e) => {
                const bannerPreview = document.getElementById('banner-preview');
                if (bannerPreview) {
                    bannerPreview.innerHTML = `<img src="${e.target.result}" alt="背景图预览">`;
                }
            };
            reader.readAsDataURL(file);

            // 上传文件
            const uploadResult = await ApiService.uploadFile(file);
            this.profileEditData.bannerFile = file;
            this.profileEditData.bannerCID = uploadResult.cid;

            Utils.showMessage('背景图上传成功', 'success');
        } catch (error) {
            Utils.showMessage('背景图上传失败: ' + error.message, 'error');
        }

        // 清空文件输入
        event.target.value = '';
    },

    // 处理资料提交
    async handleProfileSubmit(event) {
        event.preventDefault();

        const displayName = document.getElementById('edit-display-name').value.trim();
        const description = document.getElementById('edit-description').value.trim();

        if (displayName.length > APP_CONSTANTS.TEXT_LIMITS.DISPLAY_NAME_MAX_LENGTH) {
            Utils.showMessage('显示名称不能超过50个字符', 'error');
            return;
        }

        if (description.length > APP_CONSTANTS.TEXT_LIMITS.DESCRIPTION_MAX_LENGTH) {
            Utils.showMessage('个人简介不能超过300个字符', 'error');
            return;
        }

        const saveBtn = document.getElementById('save-profile-btn');
        if (saveBtn) {
            saveBtn.disabled = true;
            saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 保存中...';
        }

        try {
            // 构建更新数据
            const updateData = {};

            if (displayName) {
                updateData.displayName = displayName;
            }

            if (description) {
                updateData.description = description;
            }

            if (this.profileEditData.avatarCID) {
                updateData.avatarCID = this.profileEditData.avatarCID;
            }

            if (this.profileEditData.bannerCID) {
                updateData.bannerCID = this.profileEditData.bannerCID;
            }

            // 发送更新请求
            const result = await ApiService.updateProfile(updateData);
            console.log('资料更新成功:', result);

            // 更新当前用户信息
            if (displayName) window.App.currentUser.displayName = displayName;
            if (description) window.App.currentUser.description = description;
            if (this.profileEditData.avatarCID) {
                window.App.currentUser.avatar = `/api/blobs?id=${this.profileEditData.avatarCID}`;
            }
            if (this.profileEditData.bannerCID) {
                window.App.currentUser.banner = `/api/blobs?id=${this.profileEditData.bannerCID}`;
            }

            // 保存到localStorage
            Storage.storeUser(window.App.currentUser);

            // 更新显示
            this.updateUserDisplay(window.App.currentUser);

            // 关闭模态框
            this.hideEditProfileModal();

            Utils.showMessage('个人资料更新成功！', 'success');

        } catch (error) {
            console.error('更新个人资料失败:', error);
            Utils.showMessage('更新失败: ' + error.message, 'error');
        } finally {
            // 恢复按钮
            if (saveBtn) {
                saveBtn.disabled = false;
                saveBtn.innerHTML = '<i class="fas fa-save"></i> 保存';
            }
        }
    }
};

// 导出用户资料组件
window.UserProfileComponent = UserProfileComponent;