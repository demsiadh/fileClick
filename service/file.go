package service

import (
	"encoding/json"
	"fileClick/config"
	"fileClick/system"
	"fileClick/util"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// UploadFile 上传文件
func UploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1.解析表单数据，包括文件
	err := r.ParseMultipartForm(config.FileMaxSize)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("解析表单数据失败: " + err.Error()))
		return
	}

	// 2.获取上传的文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(system.ResFailed("获取文件失败: " + err.Error()))
		return
	}
	defer file.Close()

	// 3.创建目标文件
	id := util.GetIdGenerator().GenerateID()

	sourceFileName := handler.Filename
	systemFilename := strconv.FormatUint(id, 10) + getFileExtension(sourceFileName)
	filepath := config.FilePath + systemFilename

	dst, err := os.Create(filepath)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("创建文件失败: " + err.Error()))
		return
	}
	defer dst.Close()
	// 4.复制文件内容
	_, err = io.Copy(dst, file)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("保存文件失败: " + err.Error()))
		return
	}

	// 5.维护数据json
	err = system.AddFileToJSON(id, &system.FileInfo{
		Name: sourceFileName, Path: filepath,
	})
	if err != nil {
		panic("保存文件信息json失败！")
	}

	// 6.返回成功响应
	_ = json.NewEncoder(w).Encode(system.ResSuccess(id))
}

// DownloadFile 下载文件
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	// 从URL参数获取文件ID
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		_ = json.NewEncoder(w).Encode(system.ResFailed("文件ID不能为空"))
		return
	}

	// 将 fileID 从 string 转换为 uint64
	id, err := strconv.ParseUint(fileID, 10, 64)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("无效的文件ID"))
		return
	}

	// 读取文件信息
	fileInfo, err := system.GetFileByID(id)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("文件不存在: " + err.Error()))
		return
	}

	// 设置下载响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.Name)

	// 返回文件内容
	http.ServeFile(w, r, fileInfo.Path)
}

// DeleteFile 删除文件
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 从表单数据获取文件ID
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		_ = json.NewEncoder(w).Encode(system.ResFailed("文件ID不能为空"))
		return
	}

	// 将 fileID 从 string 转换为 uint64
	id, err := strconv.ParseUint(fileID, 10, 64)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("无效的文件ID"))
		return
	}

	// 读取文件信息
	fileInfo, err := system.GetFileByID(id)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("文件不存在: " + err.Error()))
		return
	}

	// 删除物理文件
	err = os.Remove(fileInfo.Path)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("删除文件失败: " + err.Error()))
		return
	}

	// 从JSON文件中移除记录
	err = system.RemoveFileFromJSON(id)
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("更新文件记录失败: " + err.Error()))
		return
	}

	_ = json.NewEncoder(w).Encode(system.ResSuccess(fileID))
}

// GetAllFile 获取所有文件
func GetAllFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	files, err := system.GetAllFiles()
	if err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("获取文件列表失败: " + err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(system.ResSuccess(files))
}

// getFileExtension 获取文件扩展名
func getFileExtension(filename string) string {
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return "" // 没有扩展名
	}
	return filename[lastDot:]
}
