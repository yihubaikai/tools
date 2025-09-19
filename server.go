package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const uploadDir = "files"
const tmpPrefix = "tmp_"
const tmpCleanupAge = 24 * time.Hour // 清理 tmp 文件阈值，可以改

// 全局状态
var (
	host       string
	httpPort   string
	md5Map     = map[string]string{} // md5 -> final filename
	md5MapLock sync.RWMutex

	inProgress     = map[string]bool{} // md5 -> uploading?
	inProgressLock sync.Mutex
)

type CheckResp struct {
	Status string `json:"status"` // exists | notfound | uploading | error
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

type UploadResp struct {
	Status string `json:"status"` // ok | exists | uploading | error
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	// load .env
	_ = godotenv.Load(".env")
	host = os.Getenv("HOST")
	httpPort = os.Getenv("HTTP_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if httpPort == "" {
		httpPort = "8080"
	}

	// ensure upload dir
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("[FATAL] 无法创建上传目录: %v", err)
	}

	// 初始化已有文件映射（按文件名前缀解析 md5）
	loadExistingFiles()

	// 启动后台清理 tmp 文件
	go cleanupTmpLoop()

	// 注册 HTTP 路由
	http.HandleFunc("/check", checkHandler)   // ?md5=...
	http.HandleFunc("/upload", uploadHandler) // ?md5=...&filename=...
	// files 静态服务
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(uploadDir))))

	addr := "0.0.0.0:" + httpPort
	log.Printf("[INFO] HTTP 服务启动: %s (外网访问使用 HOST=%s)\n", addr, host)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("[FATAL] ListenAndServe 失败: %v", err)
	}
}

// loadExistingFiles: 从 uploadDir 读取文件，将文件名的 md5 前缀注册到 md5Map
func loadExistingFiles() {
	list, err := ioutil.ReadDir(uploadDir)
	if err != nil {
		log.Printf("[WARN] 读取 %s 失败: %v\n", uploadDir, err)
		return
	}
	cnt := 0
	for _, fi := range list {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		if strings.HasPrefix(name, tmpPrefix) {
			continue
		}
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		// 简单判断 base 是否可能是 md5
		if len(base) >= 32 {
			md := base[:32]
			// 尝试 decode hex
			if _, err := hex.DecodeString(md); err == nil {
				md5MapLock.Lock()
				md5Map[md] = name
				md5MapLock.Unlock()
				cnt++
			}
		}
	}
	log.Printf("[INFO] 初始化已知文件映射，记录 %d 个文件\n", cnt)
}

// /check?md5=...
func checkHandler(w http.ResponseWriter, r *http.Request) {
	md5q := r.URL.Query().Get("md5")
	if md5q == "" {
		writeJSON(w, http.StatusBadRequest, CheckResp{Status: "error", Error: "missing md5 param"})
		return
	}

	// 先检查是否存在
	md5MapLock.RLock()
	if fname, ok := md5Map[md5q]; ok {
		md5MapLock.RUnlock()
		url := fmt.Sprintf("http://%s:%s/files/%s", host, httpPort, fname)
		writeJSON(w, http.StatusOK, CheckResp{Status: "exists", URL: url})
		return
	}
	md5MapLock.RUnlock()

	// 检查是否正在上传
	inProgressLock.Lock()
	if inProgress[md5q] {
		inProgressLock.Unlock()
		writeJSON(w, http.StatusOK, CheckResp{Status: "uploading"})
		return
	}
	inProgressLock.Unlock()

	// not found
	writeJSON(w, http.StatusOK, CheckResp{Status: "notfound"})
}

// /upload?md5=...&filename=...
// POST multipart/form-data field "file"
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	md5q := r.URL.Query().Get("md5")
	filenameParam := r.URL.Query().Get("filename")
	if md5q == "" {
		writeJSON(w, http.StatusBadRequest, UploadResp{Status: "error", Error: "missing md5 param"})
		return
	}

	// fast check: exists?
	md5MapLock.RLock()
	if fname, ok := md5Map[md5q]; ok {
		md5MapLock.RUnlock()
		url := fmt.Sprintf("http://%s:%s/files/%s", host, httpPort, fname)
		writeJSON(w, http.StatusOK, UploadResp{Status: "exists", URL: url})
		return
	}
	md5MapLock.RUnlock()

	// check in-progress
	inProgressLock.Lock()
	if inProgress[md5q] {
		inProgressLock.Unlock()
		writeJSON(w, http.StatusOK, UploadResp{Status: "uploading"})
		return
	}
	// mark in-progress
	inProgress[md5q] = true
	inProgressLock.Unlock()
	// ensure we clear inProgress on return
	defer func() {
		inProgressLock.Lock()
		delete(inProgress, md5q)
		inProgressLock.Unlock()
	}()

	log.Printf("[INFO] 开始接收上传请求 md5=%s filenameParam=%s from %s\n", md5q, filenameParam, r.RemoteAddr)

	// parse multipart form but don't set huge memory; FormFile will stream
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("[ERROR] 读取表单 file 失败: %v\n", err)
		writeJSON(w, http.StatusBadRequest, UploadResp{Status: "error", Error: "read form file failed: " + err.Error()})
		return
	}
	defer file.Close()

	// 临时文件名
	saveExt := filepath.Ext(header.Filename)
	if saveExt == "" && filenameParam != "" {
		saveExt = filepath.Ext(filenameParam)
	}
	tmpName := fmt.Sprintf("%s%d_%s", tmpPrefix, time.Now().UnixNano(), filepath.Base(header.Filename))
	tmpPath := filepath.Join(uploadDir, tmpName)

	out, err := os.Create(tmpPath)
	if err != nil {
		log.Printf("[ERROR] 创建临时文件失败: %v\n", err)
		writeJSON(w, http.StatusInternalServerError, UploadResp{Status: "error", Error: "create tmp file failed"})
		return
	}
	// copy stream
	written, err := io.Copy(out, file)
	out.Close()
	if err != nil {
		log.Printf("[ERROR] 写入临时文件失败: %v\n", err)
		os.Remove(tmpPath)
		writeJSON(w, http.StatusInternalServerError, UploadResp{Status: "error", Error: "write tmp failed"})
		return
	}
	log.Printf("[INFO] 临时文件已接收: %s (bytes=%d)\n", tmpPath, written)

	// 计算 md5
	calced, err := calcMD5(tmpPath)
	if err != nil {
		log.Printf("[ERROR] 计算 MD5 失败: %v\n", err)
		os.Remove(tmpPath)
		writeJSON(w, http.StatusInternalServerError, UploadResp{Status: "error", Error: "calc md5 failed"})
		return
	}
	if calced != md5q {
		log.Printf("[ERROR] MD5 校验失败，客户端声明=%s, 服务端计算=%s\n", md5q, calced)
		os.Remove(tmpPath)
		writeJSON(w, http.StatusBadRequest, UploadResp{Status: "error", Error: "md5 mismatch"})
		return
	}

	// final name md5 + ext
	finalName := calced + saveExt
	finalPath := filepath.Join(uploadDir, finalName)

	// someone may have created it meanwhile -> double check
	md5MapLock.Lock()
	if existing, ok := md5Map[md5q]; ok {
		// 已存在，删除 tmp
		md5MapLock.Unlock()
		os.Remove(tmpPath)
		url := fmt.Sprintf("http://%s:%s/files/%s", host, httpPort, existing)
		writeJSON(w, http.StatusOK, UploadResp{Status: "exists", URL: url})
		log.Printf("[INFO] 上传完成但已存在相同文件，返回已有 URL: %s\n", url)
		return
	}
	// register and move tmp -> final
	if err := os.Rename(tmpPath, finalPath); err != nil {
		// 尝试拷贝（跨文件系统时 rename 可能失败）
		log.Printf("[WARN] rename 失败，尝试拷贝: %v\n", err)
		if err := copyFile(tmpPath, finalPath); err != nil {
			log.Printf("[ERROR] 拷贝 tmp 到 final 失败: %v\n", err)
			os.Remove(tmpPath)
			md5MapLock.Unlock()
			writeJSON(w, http.StatusInternalServerError, UploadResp{Status: "error", Error: "save final failed"})
			return
		}
		os.Remove(tmpPath)
	}
	md5Map[md5q] = finalName
	md5MapLock.Unlock()

	url := fmt.Sprintf("http://%s:%s/files/%s", host, httpPort, finalName)
	log.Printf("[SUCCESS] 上传完成，md5=%s 保存为 %s, URL=%s\n", md5q, finalName, url)
	writeJSON(w, http.StatusOK, UploadResp{Status: "ok", URL: url})
}

// helper: write JSON with Content-Type
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// calcMD5 of a file path
func calcMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyFile src->dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// cleanupTmpLoop: 定期清理超过阈值的 tmp_ 文件
func cleanupTmpLoop() {
	for {
		files, err := ioutil.ReadDir(uploadDir)
		if err == nil {
			now := time.Now()
			for _, fi := range files {
				name := fi.Name()
				if strings.HasPrefix(name, tmpPrefix) {
					full := filepath.Join(uploadDir, name)
					if now.Sub(fi.ModTime()) > tmpCleanupAge {
						log.Printf("[CLEANUP] 删除旧临时文件: %s\n", full)
						_ = os.Remove(full)
					}
				}
			}
		}
		time.Sleep(1 * time.Hour)
	}
}

