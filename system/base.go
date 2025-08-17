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

// HitEvent 文件点击事件
type HitEvent struct {
	Id uint64
}

const (
	red   = "red"
	black = "black"
)

// RedBlackNode 红黑树节点
type RedBlackNode struct {
	File   *File
	Left   *RedBlackNode
	Right  *RedBlackNode
	parent *RedBlackNode
	Color  string
}

// RedBlackTree 红黑树结构
type RedBlackTree struct {
	root *RedBlackNode
}

func (rbt *RedBlackTree) String() string {
	nodes := rbt.GetAllNodesDesc()
	var result []string
	for _, V := range nodes {
		result = append(result, V.String())
	}
	return strings.Join(result, "\n")
}

func (rbt *RedBlackTree) insert(file *File) {
	// 新节点初始化为红色
	node := &RedBlackNode{File: file, Color: red}
	if rbt.root == nil {
		rbt.root = node
	} else {
		// 从根节点开始插入
		current := rbt.root
		for {
			if file.Count > current.File.Count {
				if current.Left == nil {
					current.Left = node
					break
				} else {
					current = current.Left
				}
			} else if file.Count < current.File.Count {
				if current.Right == nil {
					current.Right = node
					break
				} else {
					current = current.Right
				}
			} else {
				// 点击次数相同，按文件ID排序
				if file.Id < current.File.Id {
					if current.Left == nil {
						current.Left = node
						break
					} else {
						current = current.Left
					}
				} else {
					if current.Right == nil {
						current.Right = node
						break
					} else {
						current = current.Right
					}
				}
			}
		}
		// 设置节点父亲
		node.parent = current
	}
	// 调整红黑树
	rbt.fixAfterInsert(node)
}

func (rbt *RedBlackTree) fixAfterInsert(node *RedBlackNode) {
	// 当前插入的节点父节点是红色时，才需要修复
	for node != rbt.root && node.parent.Color == red {
		if node.parent == node.parent.parent.Left {
			// 父节点是祖父节点的左子树
			uncle := node.parent.parent.Right
			if uncle != nil && uncle.Color == red {
				// Case 1: 叔叔节点是红色
				// 将父节点和叔叔节点变黑，祖父节点变红
				node.parent.Color = black
				uncle.Color = black
				node.parent.parent.Color = red
				node = node.parent.parent // 继续向上修复
			} else {
				// Case 2: 叔叔节点是黑色，且当前节点是右子树
				if node == node.parent.Right {
					node = node.parent
					rbt.leftRotate(node) // 左旋转
				}
				// Case 3: 叔叔节点是黑色，且当前节点是左子树
				node.parent.Color = black
				node.parent.parent.Color = red
				rbt.rightRotate(node.parent.parent) // 右旋转
			}
		} else {
			// 父节点是祖父节点的右子树
			uncle := node.parent.parent.Left
			if uncle != nil && uncle.Color == red {
				// Case 1: 叔叔节点是红色
				// 将父节点和叔叔节点变黑，祖父节点变红
				node.parent.Color = black
				uncle.Color = black
				node.parent.parent.Color = red
				node = node.parent.parent // 继续向上修复
			} else {
				// Case 2: 叔叔节点是黑色，且当前节点是左子树
				if node == node.parent.Left {
					node = node.parent
					rbt.rightRotate(node) // 右旋转
				}
				// Case 3: 叔叔节点是黑色，且当前节点是右子树
				node.parent.Color = black
				node.parent.parent.Color = red
				rbt.leftRotate(node.parent.parent) // 左旋转
			}
		}
	}
	// 确保根节点为黑色
	rbt.root.Color = black
}

func (rbt *RedBlackTree) update(file *File) {
	node := rbt.find(file)
	if node != nil {
		node.File.Count = file.Count
		rbt.fixAfterUpdate(node)
	}
}

func (rbt *RedBlackTree) fixAfterUpdate(node *RedBlackNode) {
	// 判断节点的父节点是否为红色，如果是红色的话需要修复红黑树
	for node != rbt.root && node.parent.Color == red {
		if node.parent == node.parent.parent.Left {
			// 父节点是祖父节点的左子树
			uncle := node.parent.parent.Right
			if uncle != nil && uncle.Color == red {
				// Case 1: 叔叔节点是红色
				// 将父节点和叔叔节点变黑，祖父节点变红
				node.parent.Color = black
				uncle.Color = black
				node.parent.parent.Color = red
				node = node.parent.parent // 继续向上修复
			} else {
				// Case 2: 叔叔节点是黑色，且当前节点是右子树
				if node == node.parent.Right {
					node = node.parent
					rbt.leftRotate(node) // 左旋转
				}
				// Case 3: 叔叔节点是黑色，且当前节点是左子树
				node.parent.Color = black
				node.parent.parent.Color = red
				rbt.rightRotate(node.parent.parent) // 右旋转
			}
		} else {
			// 父节点是祖父节点的右子树
			uncle := node.parent.parent.Left
			if uncle != nil && uncle.Color == red {
				// Case 1: 叔叔节点是红色
				// 将父节点和叔叔节点变黑，祖父节点变红
				node.parent.Color = black
				uncle.Color = black
				node.parent.parent.Color = red
				node = node.parent.parent // 继续向上修复
			} else {
				// Case 2: 叔叔节点是黑色，且当前节点是左子树
				if node == node.parent.Left {
					node = node.parent
					rbt.rightRotate(node) // 右旋转
				}
				// Case 3: 叔叔节点是黑色，且当前节点是右子树
				node.parent.Color = black
				node.parent.parent.Color = red
				rbt.leftRotate(node.parent.parent) // 左旋转
			}
		}
	}

	// 确保根节点为黑色
	rbt.root.Color = black
}

func (rbt *RedBlackTree) TopN(n int) (result []*File) {
	rbt.topNFromRoot(rbt.root, &result, n)
	return result
}

// 从树的根节点开始，遍历TopN个文件
func (rbt *RedBlackTree) topNFromRoot(node *RedBlackNode, result *[]*File, n int) {
	if node == nil || len(*result) >= n {
		return
	}

	// 遍历左子树
	rbt.topNFromRoot(node.Left, result, n)

	// 添加当前节点到结果中
	if len(*result) < n {
		*result = append(*result, node.File)
	}

	// 遍历右子树，再遍历左子树
	rbt.topNFromRoot(node.Right, result, n)
}

// GetAllNodesDesc 获取所有节点并按倒序排列
func (rbt *RedBlackTree) GetAllNodesDesc() (result []*File) {
	rbt.getAllNodesDescFromRoot(rbt.root, &result)
	return result
}

// 从树的根节点开始，遍历所有节点并按倒序排列
func (rbt *RedBlackTree) getAllNodesDescFromRoot(node *RedBlackNode, result *[]*File) {
	if node == nil {
		return
	}
	// 遍历左子树
	rbt.getAllNodesDescFromRoot(node.Left, result)

	// 添加当前节点到结果中
	*result = append(*result, node.File)

	// 遍历右子树
	rbt.getAllNodesDescFromRoot(node.Right, result)
}

func (rbt *RedBlackTree) find(file *File) *RedBlackNode {
	current := rbt.root
	for current != nil {
		if file.Count > current.File.Count {
			current = current.Left
		} else if file.Count < current.File.Count {
			current = current.Right
		} else {
			// 点击次数相同，按文件ID排序
			if file.Id < current.File.Id {
				current = current.Left
			} else if file.Id > current.File.Id {
				current = current.Right
			} else {
				return current
			}
		}
	}
	return nil
}

// 左旋转操作
func (rbt *RedBlackTree) leftRotate(x *RedBlackNode) {
	y := x.Right
	x.Right = y.Left
	if y.Left != nil {
		y.Left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		rbt.root = y
	} else if x == x.parent.Left {
		x.parent.Left = y
	} else {
		x.parent.Right = y
	}
	y.Left = x
	x.parent = y
}

// 右旋转操作
func (rbt *RedBlackTree) rightRotate(x *RedBlackNode) {
	y := x.Left
	x.Left = y.Right
	if y.Right != nil {
		y.Right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		rbt.root = y
	} else if x == x.parent.Right {
		x.parent.Right = y
	} else {
		x.parent.Left = y
	}
	y.Right = x
	x.parent = y
}
