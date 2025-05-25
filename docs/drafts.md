
- Realtime
    - Websocket
    - WebRTC
    - HTTP
    - SSE

- Arogo
    - c2128690a0b44ea799d3fb099e08b5aa

- Dynamic Context
    - https://towardsdatascience.com/next-level-agents-unlocking-the-power-of-dynamic-context-68b8647eef89/
    - 应该同时支持两种机制？
- 不同的协议 or API
    - 会话级别:
        - Realtime API
        - Response API
    - 任务级别
        - Run API
        - A2A
    - 动态上下文

上下文：
1. 高效的资源利用
2. 信息熵理论

匹配用户的心智模型:
1. 控制感和可预测性

解决方案：
1. 单一连续上下文
2. 完全隔离上下文
3. 动态管理上下文，需要 模型 + 工程 + 产品设计

User: ........
Robot: xxxxxxxx

...

User: 我们继续昨天的 xxx 话题

TODO:

1. 消息协议设计:
    - 参考微信、飞书等
    - 通信协议的考虑
    - Websocket 作为信令和普通消息的通道
    - WebRTC 作为音视频通话的机制


1. Realtime API 直接使用 openai 的
    - conversation item
        - message
        - function call
        - function call output
2. Response API
    1. input items
        - text input
        - input message
        - output message
        - item
        - item reference
    2. response
        - output items
            - content part
                - output text

1. message 类型设计
    - File
2. message 和对话过程中的 part / artifacts 是一个东西么？？
3. message 还可能有其他的 card type

- room
    - room type: 单聊、自己、群组
    - message
    - topic  (理论上可以一直嵌套， 但是为保持用户体验，应该只有两层？)
        - message
            - parent_id
    - A2A
    - realtime api
        - audio call
        - video call
    -
-

- https://github.com/KiWi233333/JiwuChat/blob/main/composables/api/chat/message.ts
- https://github.com/lobehub/lobe-chat/tree/main/src/database

- session  不同的人会话
    - topic 不同的话题
        - thread 子话题 (是否包含上下文)
            - message 消息
- moments 里的 card 和 messages 之间的关系
- Messages 应该是一个通信协议，可能是
    - 一对一:
        - Human <-> Human
        - Agent <-> Agent
        - Human <-> Agent
    - 多对多
        - Coordinator 模式
        - P2P 模式
    - 如何实现
        1. human in loop
        2. 主动 chat
        3. multi agent
- UserProxyClient
- AgentProxyClient


- https://github.com/DevStack06/Whatsapp-Clone-Flutter/tree/master/lib
