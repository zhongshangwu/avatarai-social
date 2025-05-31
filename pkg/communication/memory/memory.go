package memory

import "github.com/zhongshangwu/avatarai-social/pkg/types/messages"

type Chunk interface {
	ID() string
	Type() string
	Metadata() map[string]interface{}
}

type ChunkType string

const (
	ChunkTypeText ChunkType = "text"
)

type Memory interface {
	Write(chunk Chunk) error
	Retrieve(query Chunk) ([]Chunk, error)
	Close() error
}

type SimpleTextChunk struct {
	id        string
	chunkType ChunkType
	metadata  map[string]interface{}
	content   string
}

func (c *SimpleTextChunk) ID() string {
	return c.id
}

func (c *SimpleTextChunk) Type() ChunkType {
	return ChunkTypeText
}

func (c *SimpleTextChunk) Metadata() map[string]interface{} {
	return c.metadata
}

type MessageChunk struct {
	id        string
	chunkType ChunkType
	metadata  map[string]interface{}
	content   *messages.Message
}

// - 连续上下文
// - 隔离上下文
// - 自动上下文

// class SimpleCognitiveMemoryNetwork(CognitiveMemoryNetwork):
//     def __init__(self):
//         self.working_memory = TransformerEncoder()
//         self.episodic_memory = MemformerMemorySlots()
//         self.semantic_memory = HNSWIndex()
//         self.procedural_memory = RLPolicyLibrary()

//     def encode(self, input_data):
//         # 使用Transformer编码输入
//         return self.working_memory.encode(input_data)

//     def retrieve(self, query, context=None):
//         # 1. 工作记忆直接注意力检索
//         wm_result = self.working_memory.attend(query)
//         # 2. 情景记忆用Memformer记忆槽检索
//         em_result = self.episodic_memory.cross_attention(query)
//         # 3. 语义记忆用HNSW近似最近邻检索
//         sm_result = self.semantic_memory.search(query)
//         # 4. 程序记忆用行为库检索
//         pm_result = self.procedural_memory.match(query)
//         # 5. 融合多层检索结果
//         return self._integrate_results(wm_result, em_result, sm_result, pm_result)

//     def write(self, information, importance=0.5, tag=None):
//         # 重要信息写入情景记忆和/或语义记忆
//         if importance > 0.7:
//             self.episodic_memory.write(information)
//             if importance > 0.9:
//                 self.semantic_memory.add(information)
//         # 行为模式写入程序记忆
//         if tag == "action":
//             self.procedural_memory.store(information)

//     def consolidate(self):
//         # 定期将频繁访问的情景记忆转移到语义记忆
//         candidates = self.episodic_memory.get_frequent()
//         for info in candidates:
//             self.semantic_memory.add(info)

//     def forget(self):
//         # 清理低强度记忆
//         self.episodic_memory.prune()
//         self.semantic_memory.prune()

//     def update(self, feedback):
//         # 根据反馈调整记忆权重
//         self.episodic_memory.update_strength(feedback)
//         self.semantic_memory.update_strength(feedback)
