package service

import (
	"encoding/json"
	"fileClick/system"
	"net/http"
	"strconv"
)

func Click(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if id < 1 || err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("非法id"))
		return
	}

	// 记录点击事件
	system.RankEngine.Click(id)

	_ = json.NewEncoder(w).Encode(system.ResSuccess(id))
}

func GetTopN(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	topNStr := r.URL.Query().Get("topN")
	topN, err := strconv.Atoi(topNStr)
	if topN < 1 || err != nil {
		_ = json.NewEncoder(w).Encode(system.ResFailed("topN必须为正整数"))
		return
	}
	files := system.RankEngine.TopN(topN)

	_ = json.NewEncoder(w).Encode(system.ResSuccess(files))
}

func GetTopAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	files := system.RankEngine.TopAll()

	_ = json.NewEncoder(w).Encode(system.ResSuccess(files))
}
