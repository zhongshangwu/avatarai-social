// 主应用
const App = {
    // 应用状态
    accessToken: null,
    currentUser: null,

    // 初始化应用
    init() {
        console.log('AvatarAI Social 应用启动');

        // 初始化各个组件
        this.initComponents();

        // 启动认证流程
        AuthComponent.init();
    },

    // 初始化组件
    initComponents() {
        // 初始化用户资料组件
        UserProfileComponent.init();

        // 初始化聊天组件
        ChatComponent.init();

        // 初始化帖子组件
        PostComponent.init();

        // 初始化Feed组件
        FeedComponent.init();

        // 初始化MCP组件
        MCPComponent.init();

        // 初始化模态框组件
        ModalsComponent.init();
    },

    // 设置用户信息
    setUser(user) {
        this.currentUser = user;
        Storage.storeUser(user);
    },

    // 设置访问令牌
    setAccessToken(token) {
        this.accessToken = token;
        Storage.storeToken(token);
    },

    // 清除用户会话
    clearSession() {
        this.accessToken = null;
        this.currentUser = null;
        Storage.clearStoredAuth();
        WebSocketService.disconnect();
    }
};

// 将App挂载到全局
window.App = App;

// 页面加载完成后启动应用
function initializeApp() {
    // 检查所有必要的依赖是否已加载
    const requiredComponents = [
        { name: 'Utils', obj: Utils },
        { name: 'Storage', obj: Storage },
        { name: 'ApiService', obj: ApiService },
        { name: 'WebSocketService', obj: WebSocketService },
        { name: 'AuthComponent', obj: AuthComponent },
        { name: 'MCPComponent', obj: MCPComponent }
    ];

    const missingComponents = requiredComponents.filter(comp => typeof comp.obj === 'undefined');

    if (missingComponents.length > 0) {
        console.log('等待组件加载:', missingComponents.map(c => c.name).join(', '));
        setTimeout(initializeApp, 100);
        return;
    }

    console.log('所有组件已加载，启动应用');
    App.init();
}

// 检查DOM是否已经加载完成
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeApp);
} else {
    // DOM已经加载完成，直接初始化
    initializeApp();
}