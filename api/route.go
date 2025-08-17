package api

import (
	"encoding/json"
	"fileClick/service"
	"fileClick/system"
	"net/http"
)

func InitRouter() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/click", methodGuard(http.MethodPut, service.Click))
	mux.HandleFunc("/topN", methodGuard(http.MethodGet, service.GetTopN))
	mux.HandleFunc("/topAll", methodGuard(http.MethodGet, service.GetTopAll))

	mux.HandleFunc("/upload", methodGuard(http.MethodPost, service.UploadFile))
	mux.HandleFunc("/download", methodGuard(http.MethodGet, service.DownloadFile))
	mux.HandleFunc("/delete", methodGuard(http.MethodDelete, service.DeleteFile))
	mux.HandleFunc("/all", methodGuard(http.MethodGet, service.GetAllFile))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return srv
}

// methodGuard 是一个中间件，用于限制允许的HTTP方法
func methodGuard(allowedMethod string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == allowedMethod {
			handler(w, r)
			return
		}

		// 如果方法不被允许，返回405状态码
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(
			system.ResSuccess("Method not allowed. Allowed methods " + allowedMethod))
	}
}
