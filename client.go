package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type CheckResp struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

type UploadResp struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

func loadEnvDefaults() (string, string) {
	_ = godotenv.Load(".env")
	h := os.Getenv("HOST")
	p := os.Getenv("HTTP_PORT")
	if h == "" {
		h = "127.0.0.1"
	}
	if p == "" {
		p = "80"
	}
	return h, p
}

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

// ProgressReader 用于显示上传进度
type ProgressReader struct {
	io.Reader
	Total     int64
	ReadSoFar int64
	LastPrint time.Time
	Fname     string
}

func (p *ProgressReader) Read(b []byte) (int, error) {
	n, err := p.Reader.Read(b)
	if n > 0 {
		p.ReadSoFar += int64(n)
		now := time.Now()
		if now.Sub(p.LastPrint) >= 1*time.Second || p.ReadSoFar == p.Total {
			percent := float64(p.ReadSoFar) / float64(p.Total) * 100
			fmt.Printf("[PROGRESS] %s %.2f%% (%d/%d)\n", p.Fname, percent, p.ReadSoFar, p.Total)
			p.LastPrint = now
		}
	}
	return n, err
}

func main() {
	fileFlag := flag.String("file", "", "上传的文件路径（必填）")
	serverFlag := flag.String("server", "", "server 地址（host:port），优先级高于 .env")
	flag.Parse()

	if *fileFlag == "" {
		fmt.Println("用法: client -file=<path> [-server=host:port]")
		return
	}

	// 读取 .env 默认
	envHost, envPort := loadEnvDefaults()
	baseURL := fmt.Sprintf("http://%s:%s", envHost, envPort)
	if *serverFlag != "" {
		// 支持带 http:// 的写法或不带
		if strings.HasPrefix(*serverFlag, "http://") || strings.HasPrefix(*serverFlag, "https://") {
			baseURL = (*serverFlag)
		} else {
			baseURL = "http://" + *serverFlag
		}
	}
	fmt.Printf("[INFO] 使用服务器: %s\n", baseURL)

	// 计算本地 MD5
	fmt.Printf("[INFO] 计算本地文件 MD5: %s\n", *fileFlag)
	md5sum, err := calcMD5(*fileFlag)
	if err != nil {
		fmt.Printf("[ERROR] 计算 MD5 失败: %v\n", err)
		return
	}
	fmt.Printf("[INFO] 文件 MD5: %s\n", md5sum)

	// 1) 先 pre-check
	checkURL := fmt.Sprintf("%s/check?md5=%s", strings.TrimRight(baseURL, "/"), url.QueryEscape(md5sum))
	fmt.Printf("[INFO] 预检 (GET) => %s\n", checkURL)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(checkURL)
	if err != nil {
		fmt.Printf("[ERROR] 预检请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	var cr CheckResp
	_ = json.NewDecoder(resp.Body).Decode(&cr)
	switch cr.Status {
	case "exists":
		fmt.Printf("[SUCCESS] 服务端已存在相同文件，URL: %s\n", cr.URL)
		return
	case "uploading":
		fmt.Printf("[INFO] 服务端正在有客户端上传同文件 (uploading)，将轮询等待 60s ...\n")
		// 轮询等待
		waited := 0
		for waited < 60 {
			time.Sleep(3 * time.Second)
			waited += 3
			resp2, err2 := client.Get(checkURL)
			if err2 != nil {
				fmt.Printf("[WARN] 轮询请求失败: %v\n", err2)
				continue
			}
			var cr2 CheckResp
			_ = json.NewDecoder(resp2.Body).Decode(&cr2)
			resp2.Body.Close()
			if cr2.Status == "exists" {
				fmt.Printf("[SUCCESS] 上传完成（他人），URL: %s\n", cr2.URL)
				return
			} else if cr2.Status == "notfound" {
				// 对方可能失败或还未完成，继续上传
				break
			}
		}
		fmt.Println("[INFO] 轮询结束，继续尝试上传...")
	case "notfound":
		// 继续上传
	default:
		fmt.Printf("[WARN] 预检返回异常: %+v\n", cr)
	}

	// 2) 上传文件（multipart streaming，带进度）
	file, err := os.Open(*fileFlag)
	if err != nil {
		fmt.Printf("[ERROR] 打开文件失败: %v\n", err)
		return
	}
	defer file.Close()
	stat, _ := file.Stat()
	total := stat.Size()
	fmt.Printf("[INFO] 开始上传: %s (size=%d)\n", *fileFlag, total)

	uploadURL := fmt.Sprintf("%s/upload?md5=%s&filename=%s", strings.TrimRight(baseURL, "/"), url.QueryEscape(md5sum), url.QueryEscape(filepath.Base(*fileFlag)))
	fmt.Printf("[INFO] 上传地址: %s\n", uploadURL)

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// 在 goroutine 中写 multipart body（stream）
	go func() {
		defer pw.Close()
		part, err := writer.CreateFormFile("file", filepath.Base(*fileFlag))
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		// wrap file with progress reader
		prd := &ProgressReader{Reader: file, Total: total, Fname: filepath.Base(*fileFlag)}
		if _, err := io.Copy(part, prd); err != nil {
			pw.CloseWithError(err)
			return
		}
		writer.Close()
	}()

	req, err := http.NewRequest("POST", uploadURL, pr)
	if err != nil {
		fmt.Printf("[ERROR] 创建上传请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// 适当超时
	client2 := &http.Client{Timeout: 0} // 上传大文件不设置超时
	resp2, err := client2.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] 上传请求失败: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	var ur UploadResp
	_ = json.NewDecoder(resp2.Body).Decode(&ur)
	if ur.Status == "ok" || ur.Status == "exists" {
		fmt.Printf("[SUCCESS] 上传结果: %s\n", ur.URL)
		return
	} else if ur.Status == "uploading" {
		fmt.Printf("[INFO] 服务器提示上传正在进行，请稍后轮询/check\n")
		return
	} else {
		fmt.Printf("[ERROR] 上传失败: %v\n", ur.Error)
		return
	}
}
