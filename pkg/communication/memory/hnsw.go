package memory

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// Vector 表示一个向量
type Vector []float64

// Node 表示图中的一个节点
type Node struct {
	ID        int
	Vector    Vector
	Level     int
	Neighbors []map[int]*Node // 每层的邻居节点
}

// HNSW 主结构
type HNSW struct {
	Nodes          []*Node
	EntryPoint     *Node
	MaxLevel       int
	M              int     // 每层最大连接数
	Ml             float64 // level 生成参数
	Ef             int     // 搜索时的候选集大小
	EfConstruction int     // 构建时的候选集大小
}

// NewHNSW 创建新的 HNSW 索引
func NewHNSW(m int, efConstruction int, ml float64) *HNSW {
	return &HNSW{
		Nodes:          make([]*Node, 0),
		M:              m,
		Ml:             ml,
		Ef:             efConstruction,
		EfConstruction: efConstruction,
		MaxLevel:       0,
	}
}

// 计算两个向量的欧几里得距离
func EuclideanDistance(a, b Vector) float64 {
	if len(a) != len(b) {
		panic("向量维度不匹配")
	}

	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// 候选节点结构，用于搜索过程
type Candidate struct {
	Node     *Node
	Distance float64
}

// 候选节点优先队列（最小堆）
type CandidateQueue []*Candidate

func (pq CandidateQueue) Len() int { return len(pq) }
func (pq CandidateQueue) Less(i, j int) bool {
	return pq[i].Distance < pq[j].Distance
}
func (pq CandidateQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

// 随机选择层级
func (h *HNSW) selectLevel() int {
	level := int(math.Floor(-math.Log(rand.Float64()) * h.Ml))
	return level
}

// 搜索最近邻居
func (h *HNSW) searchLayer(query Vector, entryPoints []*Node, numClosest int, level int) []*Candidate {
	visited := make(map[int]bool)
	candidates := make([]*Candidate, 0)
	dynamic := make([]*Candidate, 0)

	// 初始化候选集
	for _, ep := range entryPoints {
		if !visited[ep.ID] {
			dist := EuclideanDistance(query, ep.Vector)
			candidate := &Candidate{Node: ep, Distance: dist}
			candidates = append(candidates, candidate)
			dynamic = append(dynamic, candidate)
			visited[ep.ID] = true
		}
	}

	for len(dynamic) > 0 {
		// 获取距离最近的候选点
		sort.Slice(dynamic, func(i, j int) bool {
			return dynamic[i].Distance < dynamic[j].Distance
		})

		current := dynamic[0]
		dynamic = dynamic[1:]

		// 检查是否需要继续搜索
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Distance < candidates[j].Distance
		})

		if len(candidates) >= numClosest && current.Distance > candidates[numClosest-1].Distance {
			break
		}

		// 检查当前节点的邻居
		if level < len(current.Node.Neighbors) {
			for _, neighbor := range current.Node.Neighbors[level] {
				if !visited[neighbor.ID] {
					visited[neighbor.ID] = true
					dist := EuclideanDistance(query, neighbor.Vector)
					candidate := &Candidate{Node: neighbor, Distance: dist}

					sort.Slice(candidates, func(i, j int) bool {
						return candidates[i].Distance < candidates[j].Distance
					})

					if len(candidates) < numClosest || dist < candidates[len(candidates)-1].Distance {
						candidates = append(candidates, candidate)
						dynamic = append(dynamic, candidate)

						// 保持候选集大小
						sort.Slice(candidates, func(i, j int) bool {
							return candidates[i].Distance < candidates[j].Distance
						})
						if len(candidates) > numClosest {
							candidates = candidates[:numClosest]
						}
					}
				}
			}
		}
	}

	// 返回最近的 numClosest 个候选点
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	if len(candidates) > numClosest {
		candidates = candidates[:numClosest]
	}

	return candidates
}

// 选择邻居节点（启发式算法）
func (h *HNSW) selectNeighbors(candidates []*Candidate, m int) []*Node {
	if len(candidates) <= m {
		result := make([]*Node, len(candidates))
		for i, c := range candidates {
			result[i] = c.Node
		}
		return result
	}

	// 简单策略：选择距离最近的 m 个节点
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	result := make([]*Node, m)
	for i := 0; i < m; i++ {
		result[i] = candidates[i].Node
	}

	return result
}

// 插入新节点
func (h *HNSW) Insert(vector Vector) {
	level := h.selectLevel()
	node := &Node{
		ID:        len(h.Nodes),
		Vector:    vector,
		Level:     level,
		Neighbors: make([]map[int]*Node, level+1),
	}

	// 初始化每层的邻居映射
	for i := 0; i <= level; i++ {
		node.Neighbors[i] = make(map[int]*Node)
	}

	h.Nodes = append(h.Nodes, node)

	if h.EntryPoint == nil {
		h.EntryPoint = node
		h.MaxLevel = level
		return
	}

	currentMaxLevel := h.MaxLevel
	entryPoints := []*Node{h.EntryPoint}

	// 从顶层搜索到目标层+1
	for lev := currentMaxLevel; lev > level; lev-- {
		entryPoints = h.getNodesFromCandidates(h.searchLayer(vector, entryPoints, 1, lev))
	}

	// 在每一层进行搜索和连接
	for lev := min(level, currentMaxLevel); lev >= 0; lev-- {
		candidates := h.searchLayer(vector, entryPoints, h.EfConstruction, lev)
		neighbors := h.selectNeighbors(candidates, h.M)

		// 建立双向连接
		for _, neighbor := range neighbors {
			node.Neighbors[lev][neighbor.ID] = neighbor
			neighbor.Neighbors[lev][node.ID] = node

			// 如果邻居节点的连接数超过限制，需要修剪
			if len(neighbor.Neighbors[lev]) > h.M {
				h.pruneConnections(neighbor, lev)
			}
		}

		entryPoints = neighbors
	}

	// 更新入口点
	if level > h.MaxLevel {
		h.MaxLevel = level
		h.EntryPoint = node
	}
}

// 修剪连接（保持每个节点的连接数不超过 M）
func (h *HNSW) pruneConnections(node *Node, level int) {
	if len(node.Neighbors[level]) <= h.M {
		return
	}

	// 计算所有邻居的距离
	candidates := make([]*Candidate, 0, len(node.Neighbors[level]))
	for _, neighbor := range node.Neighbors[level] {
		dist := EuclideanDistance(node.Vector, neighbor.Vector)
		candidates = append(candidates, &Candidate{Node: neighbor, Distance: dist})
	}

	// 选择保留的邻居
	selected := h.selectNeighbors(candidates, h.M)

	// 重建连接
	newNeighbors := make(map[int]*Node)
	for _, neighbor := range selected {
		newNeighbors[neighbor.ID] = neighbor
	}

	// 移除不再连接的邻居
	for id, neighbor := range node.Neighbors[level] {
		if _, exists := newNeighbors[id]; !exists {
			delete(neighbor.Neighbors[level], node.ID)
		}
	}

	node.Neighbors[level] = newNeighbors
}

// 辅助函数：从候选节点中提取节点
func (h *HNSW) getNodesFromCandidates(candidates []*Candidate) []*Node {
	nodes := make([]*Node, len(candidates))
	for i, c := range candidates {
		nodes[i] = c.Node
	}
	return nodes
}

// 搜索 K 个最近邻
func (h *HNSW) Search(query Vector, k int) []*Candidate {
	if h.EntryPoint == nil {
		return []*Candidate{}
	}

	entryPoints := []*Node{h.EntryPoint}

	// 从顶层搜索到第1层
	for level := h.MaxLevel; level > 0; level-- {
		entryPoints = h.getNodesFromCandidates(h.searchLayer(query, entryPoints, 1, level))
	}

	// 在第0层进行详细搜索
	candidates := h.searchLayer(query, entryPoints, max(h.Ef, k), 0)

	// 返回前 k 个结果
	if len(candidates) > k {
		candidates = candidates[:k]
	}

	return candidates
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 示例使用
func HNSWExample() {
	rand.Seed(time.Now().UnixNano())

	// 创建 HNSW 索引
	hnsw := NewHNSW(16, 200, 1.0/math.Log(2.0))

	fmt.Println("=== HNSW 索引算法演示 ===")

	// 插入一些示例向量
	vectors := []Vector{
		{1.0, 2.0, 3.0},
		{2.0, 3.0, 4.0},
		{3.0, 4.0, 5.0},
		{10.0, 11.0, 12.0},
		{11.0, 12.0, 13.0},
		{1.5, 2.5, 3.5},
		{9.0, 10.0, 11.0},
		{2.1, 3.1, 4.1},
	}

	fmt.Printf("插入 %d 个向量到索引中...\n", len(vectors))
	for i, vec := range vectors {
		hnsw.Insert(vec)
		fmt.Printf("插入向量 %d: %v (层级: %d)\n", i, vec, hnsw.Nodes[i].Level)
	}

	fmt.Printf("\n索引构建完成！\n")
	fmt.Printf("- 总节点数: %d\n", len(hnsw.Nodes))
	fmt.Printf("- 最大层级: %d\n", hnsw.MaxLevel)
	fmt.Printf("- 入口点: 节点 %d (层级 %d)\n", hnsw.EntryPoint.ID, hnsw.EntryPoint.Level)

	// 搜索示例
	query := Vector{2.0, 3.0, 4.0}
	k := 3

	fmt.Printf("\n=== 搜索演示 ===\n")
	fmt.Printf("查询向量: %v\n", query)
	fmt.Printf("搜索最近的 %d 个邻居...\n", k)

	results := hnsw.Search(query, k)

	fmt.Printf("\n搜索结果:\n")
	for i, result := range results {
		fmt.Printf("%d. 节点 %d: %v (距离: %.4f)\n",
			i+1, result.Node.ID, result.Node.Vector, result.Distance)
	}

	// 显示图的连接信息
	fmt.Printf("\n=== 图连接信息 ===\n")
	for _, node := range hnsw.Nodes {
		fmt.Printf("节点 %d (层级 %d):\n", node.ID, node.Level)
		for level := 0; level <= node.Level; level++ {
			neighbors := make([]int, 0, len(node.Neighbors[level]))
			for id := range node.Neighbors[level] {
				neighbors = append(neighbors, id)
			}
			sort.Ints(neighbors)
			fmt.Printf("  层级 %d: 连接到 %v\n", level, neighbors)
		}
	}
}
