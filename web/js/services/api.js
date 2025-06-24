// API 服务
const ApiService = {
    // 获取访问令牌
    getAccessToken() {
        return window.App?.accessToken || Storage.getStoredToken();
    },

    // 通用请求方法
    async request(url, options = {}) {
        console.log('ApiService: 发起请求', url, options);
        const token = this.getAccessToken();
        console.log('ApiService: 使用访问令牌', token ? '存在' : '不存在');

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

        console.log('ApiService: 最终请求选项', finalOptions);
        const response = await fetch(url, finalOptions);
        console.log('ApiService: 收到响应', response.status, response.statusText);

        if (!response.ok) {
            const error = await response.text();
            console.error('ApiService: 请求失败', response.status, error);
            throw new Error(error || `请求失败: ${response.status}`);
        }

        return response;
    },

    // HTTP 方法便捷函数
    async get(url, options = {}) {
        return this.request(url, { ...options, method: 'GET' });
    },

    async post(url, data, options = {}) {
        return this.request(url, {
            ...options,
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async put(url, data, options = {}) {
        return this.request(url, {
            ...options,
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    async delete(url, options = {}) {
        return this.request(url, { ...options, method: 'DELETE' });
    },

    // 兼容旧版本的 request 方法，返回 JSON 数据
    async requestJson(url, options = {}) {
        const response = await this.request(url, options);
        return response.json();
    },

    // 用户相关API
    async getCurrentUser() {
        return this.requestJson('/api/avatar/profile');
    },

    async updateProfile(profileData) {
        return this.requestJson('/api/avatar/profile', {
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
        return this.requestJson('/api/moments', {
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

        return this.requestJson(`/api/feeds?${params}`);
    },

    async getThread(momentUri, depth = 10) {
        const params = new URLSearchParams();
        params.set('uri', momentUri);
        params.set('depth', depth.toString());

        return this.requestJson(`/api/moments/thread?${params}`);
    },

    // 聊天历史API
    async getChatHistory(roomId = 'default', threadId = 'default', limit = 20) {
        const params = new URLSearchParams();
        params.set('roomId', roomId);
        params.set('threadId', threadId);
        params.set('limit', limit.toString());

        return this.requestJson(`/api/messages/history?${params}`);
    },

    // OAuth相关API
    async refreshToken(refreshToken) {
        return this.requestJson('/api/oauth/refresh', {
            method: 'POST',
            body: JSON.stringify({
                refresh_token: refreshToken
            })
        });
    },

    async logout() {
        return this.requestJson('/api/oauth/logout', {
            method: 'GET'
        });
    }
};

// 导出API服务
window.ApiService = ApiService;