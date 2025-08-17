package system

import (
	"fmt"
	"strings"
)

// Res 响应体
type Res struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ResSuccess 响应成功
func ResSuccess(data interface{}) *Res {
	return &Res{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// ResFailed 响应失败
func ResFailed(message string) *Res {
	return &Res{
		Code:    -1,
		Message: message,
		Data:    "",
	}
}

// File 文件结构体
type File struct {
	Id       uint64 `json:"id"`
	FileName string `json:"fileName"`
	Count    uint64 `json:"count"`
}

// String 返回文件信息的格式化字符串
func (f *File) String() string {
	return fmt.Sprintf("文件ID: %d, 文件名: %s, 点击次数: %d", f.Id, f.FileName, f.Count)
}

type EventType int

const (
	HitEvent EventType = iota
	DeleteEvent
)

// FileEvent 文件点击事件
type FileEvent struct {
	Id   uint64
	Type EventType
}

// LinkedNode 双向链表节点
type LinkedNode struct {
	File *File
	Prev *LinkedNode
	Next *LinkedNode
}

// LRUList 基于LRU思想的排行榜实现
type LRUList struct {
	head    *LinkedNode
	tail    *LinkedNode
	fileMap map[uint64]*LinkedNode // 用于快速查找节点
}

// NewLRURanking 创建新的LRU排行榜实例
func NewLRURanking() *LRUList {
	return &LRUList{
		fileMap: make(map[uint64]*LinkedNode),
	}
}

// Hit 处理文件点击事件，如果文件不存在则创建新节点插入，如果存在则更新计数
func (lru *LRUList) hit(fileId uint64) {
	// 通过map快速查找节点
	node, exists := lru.fileMap[fileId]
	if !exists {
		fileInfo, _ := GetFileByID(fileId)
		// 节点不存在，创建新文件并插入
		newFile := &File{
			Id:       fileId,
			FileName: fileInfo.Name,
			Count:    1, // 初始点击次数为1
		}
		lru.insert(newFile)
	} else {
		// 节点存在，增加点击次数并更新位置
		node.File.Count++
		// 重新排序以确保链表按点击次数正确排列
		lru.update(node.File)
	}
}

// insert 插入文件节点到链表中的正确位置（按count降序排序）
func (lru *LRUList) insert(file *File) {
	// 检查节点是否已存在
	if _, exists := lru.fileMap[file.Id]; exists {
		return // 节点已存在，不重复插入
	}

	newNode := &LinkedNode{
		File: file,
	}

	// 添加到map中
	lru.fileMap[file.Id] = newNode

	// 如果链表为空
	if lru.head == nil {
		lru.head = newNode
		lru.tail = newNode
		return
	}

	// 寻找正确的插入位置
	var curr *LinkedNode
	for curr = lru.head; curr != nil; curr = curr.Next {
		// 按count降序排序，count相同时按ID升序排序
		if curr.File.Count < file.Count ||
			(curr.File.Count == file.Count && curr.File.Id > file.Id) {
			break
		}
	}

	// 插入节点到正确位置
	if curr == nil {
		// 插入到末尾
		lru.tail.Next = newNode
		newNode.Prev = lru.tail
		lru.tail = newNode
	} else if curr.Prev == nil {
		// 插入到头部
		newNode.Next = lru.head
		lru.head.Prev = newNode
		lru.head = newNode
	} else {
		// 插入到中间
		newNode.Next = curr
		newNode.Prev = curr.Prev
		curr.Prev.Next = newNode
		curr.Prev = newNode
	}
}

// update 更新文件节点在链表中的位置
func (lru *LRUList) update(file *File) {
	// 通过map快速查找节点
	node, _ := lru.fileMap[file.Id]

	// 更新点击数
	node.File.Count = file.Count
	if node.Prev != nil && node.Prev.File.Count < node.File.Count {
		lru.swapWithPrev(node)
	}
}

// swapWithPrev 将节点与其前一个节点交换位置
func (lru *LRUList) swapWithPrev(node *LinkedNode) {
	prev := node.Prev
	if prev == nil {
		return
	}

	// 调整prev的prev节点的Next指针
	if prev.Prev != nil {
		prev.Prev.Next = node
	} else {
		lru.head = node
	}

	// 调整node的Next节点的Prev指针
	if node.Next != nil {
		node.Next.Prev = prev
	} else {
		lru.tail = prev
	}

	// 交换prev和node的连接
	node.Prev = prev.Prev
	prev.Next = node.Next
	node.Next = prev
	prev.Prev = node
}

// TopN 获取点击次数前N的文件（按count降序排列）
func (lru *LRUList) TopN(n int) []*File {
	var result []*File
	curr := lru.head
	count := 0

	for curr != nil && count < n {
		result = append(result, curr.File)
		curr = curr.Next
		count++
	}

	return result
}

// TopAll 获取所有文件，按点击次数降序排列
func (lru *LRUList) TopAll() []*File {
	var result []*File
	curr := lru.head

	for curr != nil {
		result = append(result, curr.File)
		curr = curr.Next
	}

	return result
}

// String 返回链表的字符串表示（按count降序排列）
func (lru *LRUList) String() string {
	nodes := lru.TopAll()
	var result []string
	for _, v := range nodes {
		result = append(result, v.String())
	}
	return strings.Join(result, "\n")
}

// Remove 从排行榜中移除指定ID的文件
func (lru *LRUList) Remove(id uint64) {
	node, exists := lru.fileMap[id]
	if !exists {
		return
	}

	// 从链表中移除节点
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		lru.head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		lru.tail = node.Prev
	}

	// 从map中移除
	delete(lru.fileMap, id)
}
