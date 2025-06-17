// 工具函数集合
const Utils = {
    // HTML转义
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    // 生成UUID
    generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    },

    // 根据handle生成头像颜色
    getAvatarColor(handle) {
        let hash = 0;
        for (let i = 0; i < handle.length; i++) {
            hash = handle.charCodeAt(i) + ((hash << 5) - hash);
        }
        const colors = [
            '#6366f1', '#8b5cf6', '#06b6d4', '#10b981',
            '#f59e0b', '#ef4444', '#ec4899', '#84cc16'
        ];
        return colors[Math.abs(hash) % colors.length];
    },

    // 自动调整文本框高度
    autoResizeTextarea(textarea) {
        textarea.style.height = 'auto';
        textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px';
    },

    // 显示消息提示
    showMessage(message, type = 'info') {
        const container = document.getElementById('message-container');
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;

        let icon = 'fas fa-info-circle';
        if (type === 'error') icon = 'fas fa-exclamation-triangle';
        else if (type === 'success') icon = 'fas fa-check-circle';
        else if (type === 'warning') icon = 'fas fa-exclamation-circle';

        messageDiv.innerHTML = `<i class="${icon}"></i> ${this.escapeHtml(message)}`;
        container.appendChild(messageDiv);

        setTimeout(() => {
            messageDiv.remove();
        }, 5000);
    },

    // 显示/隐藏元素
    showElement(id) {
        document.getElementById(id).classList.remove('hidden');
    },

    hideElement(id) {
        document.getElementById(id).classList.add('hidden');
    },

    // 显示/隐藏加载状态
    showLoading(id) {
        document.getElementById(id).classList.add('show');
    },

    hideLoading(id) {
        document.getElementById(id).classList.remove('show');
    },

    // 解析文本中的facets（标签、链接、提及）
    parseFacets(text) {
        const facets = [];

        // 解析标签 #tag
        const tagRegex = /#(\w+)/g;
        let tagMatch;
        while ((tagMatch = tagRegex.exec(text)) !== null) {
            facets.push({
                index: {
                    byteStart: tagMatch.index,
                    byteEnd: tagMatch.index + tagMatch[0].length
                },
                features: [{
                    $type: 'app.bsky.richtext.facet#tag',
                    tag: tagMatch[1]
                }]
            });
        }

        // 解析链接 http(s)://
        const linkRegex = /https?:\/\/[^\s]+/g;
        let linkMatch;
        while ((linkMatch = linkRegex.exec(text)) !== null) {
            facets.push({
                index: {
                    byteStart: linkMatch.index,
                    byteEnd: linkMatch.index + linkMatch[0].length
                },
                features: [{
                    $type: 'app.bsky.richtext.facet#link',
                    uri: linkMatch[0]
                }]
            });
        }

        // 解析提及 @handle
        const mentionRegex = /@(\w+(?:\.\w+)*)/g;
        let mentionMatch;
        while ((mentionMatch = mentionRegex.exec(text)) !== null) {
            facets.push({
                index: {
                    byteStart: mentionMatch.index,
                    byteEnd: mentionMatch.index + mentionMatch[0].length
                },
                features: [{
                    $type: 'app.bsky.richtext.facet#mention',
                    did: 'did:plc:unknown' // 实际应该解析handle得到DID
                }]
            });
        }

        return facets.length > 0 ? facets : undefined;
    },

    // 提取标签
    extractTags(text) {
        const tags = [];
        const tagRegex = /#(\w+)/g;
        let match;

        while ((match = tagRegex.exec(text)) !== null) {
            const tag = match[1];
            if (!tags.includes(tag)) {
                tags.push(tag);
            }
        }

        return tags.length > 0 ? tags : undefined;
    },

    // 构建moment URI
    buildMomentURI(momentId, userDid) {
        if (!userDid) return '';
        return `at://${userDid}/app.vtri.activity.moment/${momentId}`;
    },

    // 格式化时间
    formatTime(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN');
    },

    // 格式化相对时间
    formatRelativeTime(timestamp) {
        const now = Date.now();
        const diff = now - timestamp;

        if (diff < 60000) return '刚刚';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
        if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`;

        return new Date(timestamp).toLocaleDateString('zh-CN');
    }
};

// 导出工具函数
window.Utils = Utils;