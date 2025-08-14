package system

import (
	"fmt"
	"testing"
)

// 测试插入后红黑树的平衡性
func TestRedBlackTreeInsert(t *testing.T) {
	rbt := &RedBlackTree{}
	files := []*File{
		{Id: 1, FileName: "file1", Count: 10},
		{Id: 2, FileName: "file2", Count: 20},
		{Id: 3, FileName: "file3", Count: 5},
	}

	for _, file := range files {
		rbt.insert(file)
	}

	// 验证根节点为黑色
	if rbt.root.Color != black {
		t.Errorf("Root node color should be black, got %s", rbt.root.Color)
	}

	// 验证查找功能
	found := rbt.find(&File{Id: 2, Count: 20})
	if found == nil || found.File.Id != 2 {
		t.Errorf("Failed to find node with Id=2")
	}

	// 验证未命中场景
	notFound := rbt.find(&File{Id: 99, Count: 99})
	if notFound != nil {
		t.Errorf("Unexpectedly found non-existent node")
	}
	fmt.Println(rbt)
}

// 测试更新后红黑树的平衡性
func TestRedBlackTreeUpdate(t *testing.T) {
	rbt := &RedBlackTree{}
	file := &File{Id: 1, FileName: "file1", Count: 10}
	rbt.insert(file)

	// 更新点击次数
	file.Count++
	rbt.update(file)

	// 验证更新后的值
	updatedNode := rbt.find(file)
	if updatedNode == nil || updatedNode.File.Count != 11 {
		t.Logf("Current tree state: %+v", rbt)
		t.Logf("Expected node: %+v", file)
		t.Logf("Actual node: %+v", updatedNode)
		t.Errorf("Update failed, expected Count=30, got %v", updatedNode)
	}

	// 验证红黑树规则
	if rbt.root.Color != black {
		t.Errorf("Root node color should be black after update")
	}
}

// 模拟多文件点击场景
func TestRedBlackTreeMultipleFiles(t *testing.T) {
	rbt := &RedBlackTree{}
	files := []*File{
		{Id: 1, FileName: "file1", Count: 10},
		{Id: 2, FileName: "file2", Count: 20},
		{Id: 3, FileName: "file3", Count: 30},
	}

	// 插入文件
	for _, file := range files {
		rbt.insert(file)
	}

	// 模拟点击操作
	for i := 0; i < 5; i++ {
		for _, file := range files {
			file.Count++
			rbt.update(file)
		}
	}

	// 验证排序
	current := rbt.root
	prevCount := current.File.Count
	for current != nil {
		if current.File.Count > prevCount {
			t.Errorf("RedBlackTree is not sorted in descending order: %v > %v", current.File.Count, prevCount)
		}
		prevCount = current.File.Count
		current = current.Right
	}

	// 打印最终树状态
	t.Logf("Final tree state: %+v", rbt)
}
