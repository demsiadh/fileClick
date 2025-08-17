package system

import (
	"encoding/json"
	"fileClick/config"
	"fmt"
	"os"
	"strconv"
)

type FileInfo struct {
	Name string `json:"fileName"`
	Path string `json:"path"`
}

// AddFileToJSON 将文件信息添加到JSON文件
func AddFileToJSON(id uint64, file *FileInfo) error {

	// 读取现有数据
	files, err := GetAllFiles()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// 添加新文件信息
	files[id] = *file

	// 写入JSON文件
	data, err := json.Marshal(files)
	if err != nil {
		return err
	}

	return os.WriteFile(config.FileInfoPath, data, 0644)
}

// RemoveFileFromJSON 从JSON文件中移除指定ID的文件记录
func RemoveFileFromJSON(id uint64) error {
	files, err := GetAllFiles()
	if err != nil {
		return err
	}

	// 删除指定ID的文件
	delete(files, id)

	// 写入更新后的数据
	data, err := json.Marshal(files)
	if err != nil {
		return err
	}

	return os.WriteFile(config.FileInfoPath, data, 0644)
}

// GetAllFiles 获取所有文件信息
func GetAllFiles() (map[uint64]FileInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(config.FileInfoPath); os.IsNotExist(err) {
		return make(map[uint64]FileInfo), nil
	}

	// 读取文件内容
	data, err := os.ReadFile(config.FileInfoPath)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var files map[uint64]FileInfo
	err = json.Unmarshal(data, &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// GetFileByID 根据ID获取文件信息
func GetFileByID(id uint64) (*FileInfo, error) {
	// 获取map格式的数据
	filesMap, err := GetAllFiles()
	if err != nil {
		return nil, err
	}

	// 直接通过key查找
	if file, exists := filesMap[id]; exists {
		return &file, nil
	}

	return nil, fmt.Errorf("文件不存在, Id: " + strconv.FormatUint(id, 10))
}
