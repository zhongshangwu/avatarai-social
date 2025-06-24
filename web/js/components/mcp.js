// MCP 管理组件
const MCPComponent = {
    // 数据状态
    servers: [],
    selectedServer: null,

    // 初始化
    init() {
        this.renderMCPContainer();
        this.bindEvents();
        this.loadMCPServers();
    },

    // 渲染 MCP 容器
    renderMCPContainer() {
        const container = document.getElementById('mcp-container');
        if (!container) return;

        container.innerHTML = `
            <div class="mcp-header">
                <div class="mcp-title">
                    <h4><i class="fas fa-plug"></i> MCP 服务器管理</h4>
                    <p>管理您的 Model Context Protocol 服务器连接</p>
                </div>
                <div class="mcp-actions">
                    <button id="refresh-servers-btn" class="btn btn-sm">
                        <i class="fas fa-sync-alt"></i> 刷新
                    </button>
                    <button id="install-server-btn" class="btn btn-success btn-sm">
                        <i class="fas fa-plus"></i> 安装服务器
                    </button>
                </div>
            </div>

            <div id="mcp-servers-list" class="mcp-servers-list">
                <div class="loading-state">
                    <div class="spinner"></div>
                    <p>正在加载 MCP 服务器...</p>
                </div>
            </div>
        `;
    },

    // 绑定事件
    bindEvents() {
        // 刷新服务器列表
        const refreshBtn = document.getElementById('refresh-servers-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.loadMCPServers());
        }

        // 安装服务器
        const installBtn = document.getElementById('install-server-btn');
        if (installBtn) {
            installBtn.addEventListener('click', () => this.showInstallModal());
        }
    },

    // 加载 MCP 服务器列表
    async loadMCPServers() {
        console.log('MCP: 开始加载 MCP 服务器列表...');
        try {
            console.log('MCP: 发起 API 请求到 /api/mcp/servers');
            const response = await ApiService.get('/api/mcp/servers');
            console.log('MCP: 收到响应', response);

            if (response.ok) {
                const data = await response.json();
                console.log('MCP: 解析到的数据', data);
                this.servers = data.servers || [];
                console.log('MCP: 服务器列表', this.servers);
                this.renderServersList();
            } else {
                console.error('MCP: 响应状态不正常', response.status, response.statusText);
                Utils.showMessage('加载 MCP 服务器失败', 'error');
            }
        } catch (error) {
            console.error('MCP: 加载 MCP 服务器失败:', error);
            Utils.showMessage('加载 MCP 服务器失败', 'error');
        }
    },

    // 渲染服务器列表
    renderServersList() {
        const container = document.getElementById('mcp-servers-list');
        if (!container) return;

        if (this.servers.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">
                        <i class="fas fa-server"></i>
                    </div>
                    <div class="empty-text">
                        <h4>暂无 MCP 服务器</h4>
                        <p>点击"安装服务器"按钮添加您的第一个 MCP 服务器</p>
                    </div>
                </div>
            `;
            return;
        }

        const serversHTML = this.servers.map(server => this.renderServerCard(server)).join('');
        container.innerHTML = `<div class="servers-grid">${serversHTML}</div>`;
    },

    // 渲染单个服务器卡片
    renderServerCard(server) {
        const statusClass = server.status === 'connected' ? 'status-connected' : 'status-disconnected';
        const statusIcon = server.status === 'connected' ? 'fas fa-check-circle' : 'fas fa-times-circle';
        const authStatusClass = server.authorization.status === 'active' ? 'auth-active' : 'auth-inactive';
        const authIcon = server.authorization.status === 'active' ? 'fas fa-shield-alt' : 'fas fa-shield';

        return `
            <div class="server-card" data-server-id="${server.mcpId}">
                <div class="server-header">
                    <div class="server-info">
                        <div class="server-icon">
                            ${server.icon ? `<img src="${server.icon}" alt="${server.name}">` : '<i class="fas fa-server"></i>'}
                        </div>
                        <div class="server-details">
                            <h5 class="server-name">${server.name}</h5>
                            <p class="server-description">${server.description || '暂无描述'}</p>
                            <div class="server-meta">
                                <span class="server-author">
                                    <i class="fas fa-user"></i> ${server.author || '未知作者'}
                                </span>
                                <span class="server-version">
                                    <i class="fas fa-tag"></i> ${server.version || 'v1.0.0'}
                                </span>
                                ${server.isBuiltin ? '<span class="builtin-badge">内置</span>' : ''}
                            </div>
                        </div>
                    </div>
                    <div class="server-status">
                        <div class="status-indicator ${statusClass}">
                            <i class="${statusIcon}"></i>
                            <span>${server.status === 'connected' ? '已连接' : '未连接'}</span>
                        </div>
                        <div class="auth-indicator ${authStatusClass}">
                            <i class="${authIcon}"></i>
                            <span>${server.authorization.status === 'active' ? '已授权' : '未授权'}</span>
                        </div>
                    </div>
                </div>

                <div class="server-body">
                    ${server.about ? `<p class="server-about">${server.about}</p>` : ''}

                    <div class="server-config">
                        <div class="config-item">
                            <label class="switch">
                                <input type="checkbox" ${server.enabled ? 'checked' : ''}
                                       onchange="MCPComponent.toggleEnabled('${server.mcpId}', this.checked)">
                                <span class="slider"></span>
                            </label>
                            <span class="config-label">启用服务器</span>
                        </div>

                        <div class="config-item">
                            <label class="switch">
                                <input type="checkbox" ${server.syncResources ? 'checked' : ''}
                                       onchange="MCPComponent.toggleSyncResources('${server.mcpId}', this.checked)">
                                <span class="slider"></span>
                            </label>
                            <span class="config-label">同步资源</span>
                        </div>
                    </div>
                </div>

                <div class="server-actions">
                    <button class="btn btn-sm" onclick="MCPComponent.showServerDetail('${server.mcpId}')">
                        <i class="fas fa-info-circle"></i> 详情
                    </button>

                    ${server.authorization.method === 'oauth2' ? `
                        <button class="btn btn-primary btn-sm" onclick="MCPComponent.handleOAuth('${server.mcpId}')">
                            <i class="fas fa-key"></i> ${server.authorization.status === 'active' ? '重新授权' : '授权'}
                        </button>
                    ` : ''}

                    ${!server.isBuiltin ? `
                        <button class="btn btn-danger btn-sm" onclick="MCPComponent.uninstallServer('${server.mcpId}')">
                            <i class="fas fa-trash"></i> 卸载
                        </button>
                    ` : ''}
                </div>
            </div>
        `;
    },

    // 显示服务器详情
    async showServerDetail(mcpId) {
        try {
            const response = await ApiService.get(`/api/mcp/servers/${mcpId}`);
            if (response.ok) {
                const server = await response.json();
                this.showServerDetailModal(server);
            } else {
                Utils.showMessage('获取服务器详情失败', 'error');
            }
        } catch (error) {
            console.error('获取服务器详情失败:', error);
            Utils.showMessage('获取服务器详情失败', 'error');
        }
    },

    // 显示服务器详情模态框
    showServerDetailModal(server) {
        const modalContent = `
            <div class="server-detail-modal">
                <div class="server-detail-header">
                    <div class="server-icon-large">
                        ${server.icon ? `<img src="${server.icon}" alt="${server.name}">` : '<i class="fas fa-server"></i>'}
                    </div>
                    <div class="server-info">
                        <h3>${server.name}</h3>
                        <p class="server-description">${server.description || '暂无描述'}</p>
                    </div>
                </div>

                <div class="server-detail-body">
                    <div class="detail-section">
                        <h4><i class="fas fa-info-circle"></i> 基本信息</h4>
                        <div class="detail-grid">
                            <div class="detail-item">
                                <label>作者:</label>
                                <span>${server.author || '未知'}</span>
                            </div>
                            <div class="detail-item">
                                <label>版本:</label>
                                <span>${server.version || 'v1.0.0'}</span>
                            </div>
                            <div class="detail-item">
                                <label>协议版本:</label>
                                <span>${server.protocolVersion || '1.0.0'}</span>
                            </div>
                            <div class="detail-item">
                                <label>类型:</label>
                                <span>${server.isBuiltin ? '内置服务器' : '用户安装'}</span>
                            </div>
                        </div>
                    </div>

                    ${server.about ? `
                        <div class="detail-section">
                            <h4><i class="fas fa-file-alt"></i> 详细说明</h4>
                            <p class="about-text">${server.about}</p>
                        </div>
                    ` : ''}

                    <div class="detail-section">
                        <h4><i class="fas fa-cog"></i> 连接配置</h4>
                        <div class="config-info">
                            <div class="config-item">
                                <label>端点类型:</label>
                                <span>${server.endpoint?.type || '未配置'}</span>
                            </div>
                            ${server.endpoint?.url ? `
                                <div class="config-item">
                                    <label>URL:</label>
                                    <span class="url-text">${server.endpoint.url}</span>
                                </div>
                            ` : ''}
                            ${server.endpoint?.command ? `
                                <div class="config-item">
                                    <label>命令:</label>
                                    <span class="command-text">${server.endpoint.command}</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>

                    <div class="detail-section">
                        <h4><i class="fas fa-shield-alt"></i> 授权信息</h4>
                        <div class="auth-info">
                            <div class="auth-item">
                                <label>授权方法:</label>
                                <span class="auth-method">${this.getAuthMethodText(server.authorization.method)}</span>
                            </div>
                            <div class="auth-item">
                                <label>授权状态:</label>
                                <span class="auth-status ${server.authorization.status}">${this.getAuthStatusText(server.authorization.status)}</span>
                            </div>
                            ${server.authorization.scopes ? `
                                <div class="auth-item">
                                    <label>权限范围:</label>
                                    <span class="auth-scopes">${server.authorization.scopes}</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>

                    ${server.capabilities && Object.keys(server.capabilities).length > 0 ? `
                        <div class="detail-section">
                            <h4><i class="fas fa-tools"></i> 服务器能力</h4>
                            <div class="capabilities-list">
                                ${Object.entries(server.capabilities).map(([key, value]) => `
                                    <div class="capability-item">
                                        <i class="fas fa-check"></i>
                                        <span>${key}: ${JSON.stringify(value)}</span>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;

        ModalsComponent.showModal('服务器详情', modalContent, {
            size: 'large',
            showFooter: false
        });
    },

    // 处理 OAuth 授权
    async handleOAuth(mcpId) {
        try {
            const currentDomain = window.location.origin;
            const returnUri = `${currentDomain}/oauth-callback.html`;

            const response = await ApiService.get(`/api/mcp/oauth/authorize?mcpId=${mcpId}&returnUri=${encodeURIComponent(returnUri)}`);

            if (response.ok) {
                const data = await response.json();

                if (data.already_authorized) {
                    Utils.showMessage('服务器已经授权，无需重复授权', 'info');
                    return;
                }

                // 打开授权窗口
                const authWindow = window.open(
                    data.authorization_url,
                    'mcp-oauth',
                    'width=600,height=700,scrollbars=yes,resizable=yes'
                );

                // 监听授权完成
                this.waitForOAuthCallback(authWindow, mcpId);

            } else {
                const errorData = await response.json();
                Utils.showMessage(errorData.error || '启动授权流程失败', 'error');
            }
        } catch (error) {
            console.error('OAuth 授权失败:', error);
            Utils.showMessage('OAuth 授权失败', 'error');
        }
    },

    // 等待 OAuth 回调
    waitForOAuthCallback(authWindow, mcpId) {
        const checkClosed = setInterval(() => {
            if (authWindow.closed) {
                clearInterval(checkClosed);
                // 授权窗口关闭后刷新服务器状态
                setTimeout(() => {
                    this.loadMCPServers();
                    Utils.showMessage('OAuth 授权完成', 'success');
                }, 1000);
            }
        }, 1000);

        // 监听来自授权窗口的消息
        const messageHandler = (event) => {
            if (event.data && event.data.type === 'mcp-oauth-success') {
                authWindow.close();
                window.removeEventListener('message', messageHandler);
                clearInterval(checkClosed);

                // 刷新服务器列表
                this.loadMCPServers();
                Utils.showMessage('OAuth 授权成功', 'success');
            }
        };

        window.addEventListener('message', messageHandler);
    },

    // 切换启用状态
    async toggleEnabled(mcpId, enabled) {
        try {
            const response = await ApiService.post(`/api/mcp/servers/${mcpId}/toggle-enabled`, {
                enabled: enabled
            });

            if (response.ok) {
                Utils.showMessage(`服务器已${enabled ? '启用' : '禁用'}`, 'success');
                // 更新本地状态
                const server = this.servers.find(s => s.mcpId === mcpId);
                if (server) {
                    server.enabled = enabled;
                    server.status = enabled ? 'connected' : 'disconnected';
                }
                this.renderServersList();
            } else {
                Utils.showMessage('更新服务器状态失败', 'error');
                // 恢复开关状态
                const checkbox = document.querySelector(`input[onchange*="${mcpId}"][onchange*="toggleEnabled"]`);
                if (checkbox) {
                    checkbox.checked = !enabled;
                }
            }
        } catch (error) {
            console.error('更新服务器状态失败:', error);
            Utils.showMessage('更新服务器状态失败', 'error');
        }
    },

    // 切换同步资源状态
    async toggleSyncResources(mcpId, syncResources) {
        try {
            const response = await ApiService.post(`/api/mcp/servers/${mcpId}/toggle-sync-resources`, {
                syncResources: syncResources
            });

            if (response.ok) {
                Utils.showMessage(`资源同步已${syncResources ? '启用' : '禁用'}`, 'success');
                // 更新本地状态
                const server = this.servers.find(s => s.mcpId === mcpId);
                if (server) {
                    server.syncResources = syncResources;
                }
            } else {
                Utils.showMessage('更新同步状态失败', 'error');
                // 恢复开关状态
                const checkbox = document.querySelector(`input[onchange*="${mcpId}"][onchange*="toggleSyncResources"]`);
                if (checkbox) {
                    checkbox.checked = !syncResources;
                }
            }
        } catch (error) {
            console.error('更新同步状态失败:', error);
            Utils.showMessage('更新同步状态失败', 'error');
        }
    },

    // 卸载服务器
    async uninstallServer(mcpId) {
        const server = this.servers.find(s => s.mcpId === mcpId);
        if (!server) return;

        const confirmed = await ModalsComponent.showConfirm(
            '确认卸载',
            `确定要卸载 MCP 服务器 "${server.name}" 吗？此操作不可撤销。`
        );

        if (!confirmed) return;

        try {
            const response = await ApiService.delete(`/api/mcp/servers/uninstall?mcpId=${mcpId}`);

            if (response.ok) {
                Utils.showMessage('MCP 服务器卸载成功', 'success');
                this.loadMCPServers();
            } else {
                Utils.showMessage('卸载 MCP 服务器失败', 'error');
            }
        } catch (error) {
            console.error('卸载 MCP 服务器失败:', error);
            Utils.showMessage('卸载 MCP 服务器失败', 'error');
        }
    },

    // 显示安装模态框
    showInstallModal() {
        const modalContent = `
            <form id="install-server-form" class="install-server-form">
                <div class="form-group">
                    <label for="server-name">
                        <i class="fas fa-tag"></i> 服务器名称
                    </label>
                    <input
                        type="text"
                        id="server-name"
                        name="name"
                        placeholder="输入服务器名称"
                        required
                    >
                </div>

                <div class="form-group">
                    <label for="endpoint-type">
                        <i class="fas fa-plug"></i> 端点类型
                    </label>
                    <select id="endpoint-type" name="endpointType" required>
                        <option value="">选择端点类型</option>
                        <option value="stdio">标准输入输出 (stdio)</option>
                        <option value="sse">服务器发送事件 (SSE)</option>
                        <option value="websocket">WebSocket</option>
                        <option value="streamableHttp">可流式 HTTP</option>
                    </select>
                </div>

                <div id="stdio-config" class="endpoint-config hidden">
                    <div class="form-group">
                        <label for="command">
                            <i class="fas fa-terminal"></i> 命令
                        </label>
                        <input
                            type="text"
                            id="command"
                            name="command"
                            placeholder="例如: node server.js"
                        >
                    </div>
                    <div class="form-group">
                        <label for="args">
                            <i class="fas fa-list"></i> 参数 (每行一个)
                        </label>
                        <textarea
                            id="args"
                            name="args"
                            placeholder="--port=3000&#10;--verbose"
                            rows="3"
                        ></textarea>
                    </div>
                </div>

                <div id="http-config" class="endpoint-config hidden">
                    <div class="form-group">
                        <label for="url">
                            <i class="fas fa-link"></i> URL
                        </label>
                        <input
                            type="url"
                            id="url"
                            name="url"
                            placeholder="https://api.example.com/mcp"
                        >
                    </div>
                    <div class="form-group">
                        <label for="headers">
                            <i class="fas fa-code"></i> 请求头 (JSON 格式)
                        </label>
                        <textarea
                            id="headers"
                            name="headers"
                            placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
                            rows="3"
                        ></textarea>
                    </div>
                </div>

                <div class="form-group">
                    <label for="env-vars">
                        <i class="fas fa-cog"></i> 环境变量 (JSON 格式)
                    </label>
                    <textarea
                        id="env-vars"
                        name="env"
                        placeholder='{"API_KEY": "your-api-key", "DEBUG": "true"}'
                        rows="3"
                    ></textarea>
                </div>
            </form>
        `;

        ModalsComponent.showModal('安装 MCP 服务器', modalContent, {
            confirmText: '安装',
            onConfirm: () => this.handleInstallServer()
        });

        // 绑定端点类型变化事件
        const endpointTypeSelect = document.getElementById('endpoint-type');
        if (endpointTypeSelect) {
            endpointTypeSelect.addEventListener('change', (e) => {
                this.handleEndpointTypeChange(e.target.value);
            });
        }
    },

    // 处理端点类型变化
    handleEndpointTypeChange(type) {
        const stdioConfig = document.getElementById('stdio-config');
        const httpConfig = document.getElementById('http-config');

        // 隐藏所有配置
        stdioConfig.classList.add('hidden');
        httpConfig.classList.add('hidden');

        // 显示对应配置
        if (type === 'stdio') {
            stdioConfig.classList.remove('hidden');
        } else if (['sse', 'websocket', 'streamableHttp'].includes(type)) {
            httpConfig.classList.remove('hidden');
        }
    },

    // 处理安装服务器
    async handleInstallServer() {
        const form = document.getElementById('install-server-form');
        if (!form) return false;

        const formData = new FormData(form);
        const data = {
            name: formData.get('name'),
            endpoint: {
                type: formData.get('endpointType')
            }
        };

        // 根据端点类型添加配置
        const endpointType = formData.get('endpointType');
        if (endpointType === 'stdio') {
            data.endpoint.command = formData.get('command');
            const argsText = formData.get('args');
            if (argsText) {
                data.endpoint.args = argsText.trim().split('\n').filter(arg => arg.trim());
            }
        } else if (['sse', 'websocket', 'streamableHttp'].includes(endpointType)) {
            data.endpoint.url = formData.get('url');
            const headersText = formData.get('headers');
            if (headersText) {
                try {
                    data.endpoint.headers = JSON.parse(headersText);
                } catch (e) {
                    Utils.showMessage('请求头格式不正确，请使用有效的 JSON 格式', 'error');
                    return false;
                }
            }
        }

        // 环境变量
        const envText = formData.get('env');
        if (envText) {
            try {
                data.endpoint.env = JSON.parse(envText);
            } catch (e) {
                Utils.showMessage('环境变量格式不正确，请使用有效的 JSON 格式', 'error');
                return false;
            }
        }

        try {
            const response = await ApiService.post('/api/mcp/servers/install', data);

            if (response.ok) {
                const result = await response.json();
                Utils.showMessage('MCP 服务器安装成功', 'success');
                ModalsComponent.hideModal();
                this.loadMCPServers();
                return true;
            } else {
                const errorData = await response.json();
                Utils.showMessage(errorData.message || '安装 MCP 服务器失败', 'error');
                return false;
            }
        } catch (error) {
            console.error('安装 MCP 服务器失败:', error);
            Utils.showMessage('安装 MCP 服务器失败', 'error');
            return false;
        }
    },

    // 获取授权方法文本
    getAuthMethodText(method) {
        const methods = {
            'none': '无需授权',
            'oauth2': 'OAuth 2.0',
            'api_key': 'API 密钥',
            'bearer_token': 'Bearer 令牌'
        };
        return methods[method] || method;
    },

    // 获取授权状态文本
    getAuthStatusText(status) {
        const statuses = {
            'active': '已激活',
            'inactive': '未激活',
            'disabled': '已禁用',
            'expired': '已过期'
        };
        return statuses[status] || status;
    }
};

// 导出 MCP 组件
window.MCPComponent = MCPComponent;