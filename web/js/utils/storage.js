// 存储管理
const Storage = {
    // Token 相关
    getStoredToken() {
        // 优先从localStorage获取，然后从cookie
        let token = localStorage.getItem('avatarai_access_token');
        if (!token) {
            token = this.getCookie('access_token') || this.getCookie('avatarai_token');
        }
        return token;
    },

    storeToken(token) {
        localStorage.setItem('avatarai_access_token', token);
        document.cookie = `access_token=${token}; path=/; max-age=86400`;
        document.cookie = `avatarai_token=${token}; path=/; max-age=86400`;
    },

    getStoredRefreshToken() {
        return localStorage.getItem('avatarai_refresh_token');
    },

    storeRefreshToken(token) {
        if (token) {
            localStorage.setItem('avatarai_refresh_token', token);
        }
    },

    // 用户信息相关
    getStoredUser() {
        const userStr = localStorage.getItem('avatarai_user');
        if (userStr) {
            try {
                return JSON.parse(userStr);
            } catch (e) {
                console.error('解析存储的用户信息失败:', e);
                localStorage.removeItem('avatarai_user');
            }
        }
        return null;
    },

    storeUser(user) {
        if (user) {
            localStorage.setItem('avatarai_user', JSON.stringify(user));
        }
    },

    // 清除所有认证信息
    clearStoredAuth() {
        // 清除localStorage
        localStorage.removeItem('avatarai_access_token');
        localStorage.removeItem('avatarai_refresh_token');
        localStorage.removeItem('avatarai_user');

        // 清除cookies
        this.deleteCookie('access_token');
        this.deleteCookie('avatarai_token');
    },

    // Cookie 操作
    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
    },

    deleteCookie(name) {
        document.cookie = name + '=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    }
};

// 导出存储管理
window.Storage = Storage;