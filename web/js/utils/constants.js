// 应用常量
const APP_CONSTANTS = {
    // WebSocket 事件类型
    WS_EVENTS: {
        MESSAGE_SEND: 'message.send',
        MESSAGE_RECEIVED: 'message_received',
        AGENT_MESSAGE_CREATED: 'agent_message.created',
        AGENT_MESSAGE_IN_PROGRESS: 'agent_message.in_progress',
        AGENT_MESSAGE_COMPLETED: 'agent_message.completed',
        AGENT_MESSAGE_FAILED: 'agent_message.failed',
        AGENT_MESSAGE_INCOMPLETE: 'agent_message.incomplete',
        OUTPUT_ITEM_ADDED: 'agent_message.output_item.added',
        CONTENT_PART_ADDED: 'agent_message.content_part.added',
        TEXT_DELTA: 'agent_message.output_text.delta',
        TEXT_DONE: 'agent_message.output_text.done',
        CONTENT_PART_DONE: 'agent_message.content_part.done',
        OUTPUT_ITEM_DONE: 'agent_message.output_item.done',
        AGENT_MESSAGE_INTERRUPT: 'agent_message.interrupt',
        ERROR: 'error'
    },

    // 消息类型
    MESSAGE_TYPES: {
        TEXT: 1,
        IMAGE: 3,
        AGENT_MESSAGE: 9
    },

    // 文件大小限制
    FILE_LIMITS: {
        IMAGE_MAX_SIZE: 10 * 1024 * 1024, // 10MB
        VIDEO_MAX_SIZE: 100 * 1024 * 1024, // 100MB
        AVATAR_MAX_SIZE: 5 * 1024 * 1024, // 5MB
        BANNER_MAX_SIZE: 10 * 1024 * 1024, // 10MB
        MAX_IMAGES: 4
    },

    // 文本长度限制
    TEXT_LIMITS: {
        POST_MAX_LENGTH: 3000,
        DISPLAY_NAME_MAX_LENGTH: 50,
        DESCRIPTION_MAX_LENGTH: 300
    },

    // 令牌刷新时间
    TOKEN_REFRESH_TIME: 23 * 60 * 60 * 1000, // 23小时

    // 默认值
    DEFAULTS: {
        ROOM_ID: 'default',
        THREAD_ID: 'default',
        FEED_LIMIT: 20,
        HISTORY_LIMIT: 20
    }
};

// 导出常量
window.APP_CONSTANTS = APP_CONSTANTS;