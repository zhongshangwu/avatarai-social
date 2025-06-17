// 认证组件
const AuthComponent = {
    tokenRefreshTimer: null,

    // 初始化认证
    init() {
        this.bindEvents();
        this.handleOAuthCallback();

        // 检查是否有保存的访问令牌
        const accessToken = Storage.getStoredToken();
        if (accessToken) {
            const storedUser = Storage.getStoredUser();
            if (storedUser) {
                window.App.accessToken = accessToken;
                window.App.currentUser = storedUser;
                this.setupTokenRefresh();
                this.checkTokenExpiry();
                this.showUserPanel();
                WebSocketService.connect();
            } else {
                this.fetchCurrentUser();
            }
        } else {
            this.showWelcome();
        }
    },

    // 绑定事件
    bindEvents() {
        // OAuth登录表单
        const oauthLoginForm = document.getElementById('oauth-login-form');
        if (oauthLoginForm) {
            oauthLoginForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleOAuthLogin();
            });
        }

        // 登出按钮
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => this.handleLogout());
        }

        // 刷新令牌按钮
        const refreshTokenBtn = document.getElementById('refresh-token-btn');
        if (refreshTokenBtn) {
            refreshTokenBtn.addEventListener('click', () => this.refreshAccessToken());
        }

        // 欢迎页面登录按钮
        const welcomeLoginBtn = document.getElementById('welcome-login-btn');
        if (welcomeLoginBtn) {
            welcomeLoginBtn.addEventListener('click', () => this.showLoginForm());
        }
    },

    // 处理OAuth回调
    handleOAuthCallback() {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');
        const error = urlParams.get('error');

        if (error) {
            Utils.showMessage('OAuth 认证失败: ' + decodeURIComponent(error), 'error');
            this.showWelcome();
            window.history.replaceState({}, document.title, window.location.pathname);
            return;
        }

        if (code) {
            Utils.showElement('oauth-callback-handler');

            // 使用code换取token
            fetch('/api/oauth/token', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ code: code })
            })
            .then(response => response.json())
            .then(data => {
                Utils.hideElement('oauth-callback-handler');
                if (data.access_token) {
                    window.App.accessToken = data.access_token;
                    Storage.storeToken(data.access_token);
                    Storage.storeRefreshToken(data.refresh_token);

                    // 创建用户对象并存储
                    const user = {
                        did: data.did,
                        handle: data.handle
                    };
                    window.App.currentUser = user;
                    Storage.storeUser(user);

                    this.setupTokenRefresh();
                    this.showUserPanel();
                    WebSocketService.connect();
                    window.history.replaceState({}, document.title, window.location.pathname);
                    Utils.showMessage('登录成功！', 'success');
                } else {
                    Utils.showMessage('OAuth 认证失败: ' + (data.error || '未知错误'), 'error');
                    this.showWelcome();
                }
            })
            .catch(error => {
                Utils.hideElement('oauth-callback-handler');
                console.error('OAuth callback error:', error);
                Utils.showMessage('OAuth 认证失败: ' + error.message, 'error');
                this.showWelcome();
            });
        }
    },

    // 处理OAuth登录
    handleOAuthLogin() {
        const username = document.getElementById('username').value.trim();
        if (!username) {
            Utils.showMessage('请输入用户名', 'error');
            return;
        }

        Utils.showLoading('login-loading');

        // 直接提交表单到后端
        const form = document.createElement('form');
        form.method = 'POST';
        form.action = '/api/oauth/signin?platform=web';

        const usernameInput = document.createElement('input');
        usernameInput.type = 'hidden';
        usernameInput.name = 'username';
        usernameInput.value = username;

        form.appendChild(usernameInput);
        document.body.appendChild(form);
        form.submit();
    },

    // 处理登出
    async handleLogout() {
        if (window.App.accessToken) {
            try {
                await ApiService.logout();
            } catch (error) {
                console.error('Logout error:', error);
            }
        }

        this.clearAuth();
        WebSocketService.disconnect();
        this.showWelcome();
        Utils.showMessage('已成功登出', 'success');
    },

    // 获取当前用户信息
    async fetchCurrentUser() {
        if (!window.App.accessToken) {
            this.showWelcome();
            return;
        }

        try {
            const data = await ApiService.getCurrentUser();

            // 处理头像和背景图URL
            if (data.avatar && !data.avatar.startsWith('http')) {
                data.avatar = `/api/blobs?id=${data.avatar}`;
            }
            if (data.banner && !data.banner.startsWith('http')) {
                data.banner = `/api/blobs?id=${data.banner}`;
            }

            window.App.currentUser = data;
            Storage.storeUser(data);
            this.setupTokenRefresh();
            this.showUserPanel();
            WebSocketService.connect();
        } catch (error) {
            console.error('获取用户信息失败:', error);
            this.clearAuth();
            this.showWelcome();
        }
    },

    // 刷新访问令牌
    async refreshAccessToken() {
        const refreshToken = Storage.getStoredRefreshToken();
        if (!refreshToken) {
            Utils.showMessage('没有刷新令牌，请重新登录', 'warning');
            this.clearAuth();
            this.showWelcome();
            return;
        }

        try {
            const data = await ApiService.refreshToken(refreshToken);

            if (data.access_token) {
                window.App.accessToken = data.access_token;
                Storage.storeToken(data.access_token);
                if (data.refresh_token) {
                    Storage.storeRefreshToken(data.refresh_token);
                }

                // 更新用户信息
                if (data.did && data.handle) {
                    const user = {
                        did: data.did,
                        handle: data.handle
                    };
                    window.App.currentUser = user;
                    Storage.storeUser(user);
                }

                this.setupTokenRefresh();
                Utils.showMessage('访问令牌已刷新', 'success');
            } else {
                throw new Error('响应中没有新的访问令牌');
            }
        } catch (error) {
            console.error('刷新令牌失败:', error);
            Utils.showMessage('刷新令牌失败，请重新登录: ' + error.message, 'error');
            this.clearAuth();
            this.showWelcome();
        }
    },

    // 设置自动刷新令牌
    setupTokenRefresh() {
        if (this.tokenRefreshTimer) {
            clearTimeout(this.tokenRefreshTimer);
        }

        const refreshTime = (typeof APP_CONSTANTS !== 'undefined' && APP_CONSTANTS.TOKEN_REFRESH_TIME)
            ? APP_CONSTANTS.TOKEN_REFRESH_TIME
            : 23 * 60 * 60 * 1000; // 默认23小时

        this.tokenRefreshTimer = setTimeout(() => {
            console.log('自动刷新访问令牌');
            this.refreshAccessToken();
        }, refreshTime);
    },

    // 检查令牌是否需要刷新
    checkTokenExpiry() {
        const refreshToken = Storage.getStoredRefreshToken();
        if (!refreshToken || !window.App.accessToken) {
            return;
        }

        try {
            const tokenParts = window.App.accessToken.split('.');
            if (tokenParts.length === 3) {
                const payload = JSON.parse(atob(tokenParts[1]));
                const currentTime = Math.floor(Date.now() / 1000);
                const expiryTime = payload.exp;

                // 如果令牌在5分钟内过期，自动刷新
                if (expiryTime && (expiryTime - currentTime) < 300) {
                    console.log('令牌即将过期，自动刷新');
                    this.refreshAccessToken();
                }
            }
        } catch (e) {
            console.error('解析令牌失败:', e);
        }
    },

    // 清除认证信息
    clearAuth() {
        Storage.clearStoredAuth();

        if (this.tokenRefreshTimer) {
            clearTimeout(this.tokenRefreshTimer);
            this.tokenRefreshTimer = null;
        }

        window.App.accessToken = null;
        window.App.currentUser = null;
    },

    // 显示欢迎页面
    showWelcome() {
        Utils.hideElement('user-panel');
        Utils.hideElement('login-form');
        Utils.hideElement('oauth-callback-handler');
        Utils.hideElement('user-greeting');
        Utils.hideElement('nav-menu');
        Utils.showElement('welcome-message');
    },

    // 显示登录表单
    showLoginForm() {
        Utils.hideElement('user-panel');
        Utils.hideElement('welcome-message');
        Utils.hideElement('oauth-callback-handler');
        Utils.hideElement('user-greeting');
        Utils.hideElement('nav-menu');
        Utils.showElement('login-form');

        const usernameInput = document.getElementById('username');
        if (usernameInput) {
            usernameInput.focus();
        }
    },

    // 显示用户面板
    showUserPanel() {
        Utils.hideElement('login-form');
        Utils.hideElement('welcome-message');
        Utils.hideElement('oauth-callback-handler');
        Utils.showElement('user-panel');
        Utils.showElement('user-greeting');
        Utils.showElement('nav-menu');

        if (window.App.currentUser) {
            UserProfileComponent.updateUserDisplay(window.App.currentUser);

            // 延迟加载其他组件
            setTimeout(() => {
                ChatComponent.loadHistory();
                FeedComponent.loadFeedData(true);
            }, 1000);
        }
    }
};

// 导出认证组件
window.AuthComponent = AuthComponent;