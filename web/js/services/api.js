// API 服务
const ApiService = {
    // 获取访问令牌
    getAccessToken() {
        return window.App?.accessToken || Storage.getStoredToken();
    },

    // 通用请求方法
    async request(url, options = {}) {
        const token = this.getAccessToken();
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                ...(token && { 'Authorization': `Bearer ${token}` })
            }
        };

        const finalOptions = {
            ...defaultOptions,
            ...options,
            headers: {
                ...defaultOptions.headers,
                ...options.headers
            }
        };

        const response = await fetch(url, finalOptions);

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || `请求失败: ${response.status}`);
        }

        return response.json();
    },

    // 用户相关API
    async getCurrentUser() {
        return this.request('/api/avatar/profile');
    },

    async updateProfile(profileData) {
        return this.request('/api/avatar/profile', {
            method: 'POST',
            body: JSON.stringify(profileData)
        });
    },

    // 文件上传
    async uploadFile(file) {
        const token = this.getAccessToken();
        const formData = new FormData();
        formData.append('file', file);

        const response = await fetch('/api/blobs', {
            method: 'POST',
            headers: {
                ...(token && { 'Authorization': `Bearer ${token}` })
            },
            body: formData
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || '上传失败');
        }

        return response.json();
    },

    // 帖子相关API
    async createMoment(momentData) {
        return this.request('/api/moments', {
            method: 'POST',
            body: JSON.stringify(momentData)
        });
    },

    async getFeed(cursor = null, limit = 20) {
        const params = new URLSearchParams();
        params.set('limit', limit.toString());
        if (cursor) {
            params.set('cursor', cursor);
        }

        return this.request(`/api/feeds?${params}`);
    },

    async getThread(momentUri, depth = 10) {
        const params = new URLSearchParams();
        params.set('uri', momentUri);
        params.set('depth', depth.toString());

        return this.request(`/api/moments/thread?${params}`);
    },

    // 聊天历史API
    async getChatHistory(roomId = 'default', threadId = 'default', limit = 20) {
        const params = new URLSearchParams();
        params.set('roomId', roomId);
        params.set('threadId', threadId);
        params.set('limit', limit.toString());

        return this.request(`/api/messages/history?${params}`);
    },

    // OAuth相关API
    async refreshToken(refreshToken) {
        return this.request('/api/oauth/refresh', {
            method: 'POST',
            body: JSON.stringify({
                refresh_token: refreshToken
            })
        });
    },

    async logout() {
        return this.request('/api/oauth/logout', {
            method: 'GET'
        });
    }
};

// 导出API服务
window.ApiService = ApiService;