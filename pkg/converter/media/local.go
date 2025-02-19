package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type LocalHandler struct {
	savePath  string
	urlPrefix string
	category  string
	article   string
}

func NewLocalHandler(savePath, urlPrefix string) *LocalHandler {
	return &LocalHandler{
		savePath:  savePath,
		urlPrefix: urlPrefix,
	}
}

// 设置当前处理的文章信息
func (h *LocalHandler) SetContext(category, article string) {
	h.category = category
	h.article = article
}

func (h *LocalHandler) SaveMedia(url string) (string, error) {
	// 清理 URL 中的查询参数
	cleanURL := strings.Split(url, "?")[0]

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	// 生成文件名
	filename := filepath.Base(cleanURL)
	if filename == "" || filepath.Ext(filename) == "" {
		contentType := resp.Header.Get("Content-Type")
		ext := ".bin"
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "video/mp4":
			ext = ".mp4"
		}
		filename = fmt.Sprintf("image-%d%s", len(filename), ext)
	}

	// 构建保存路径
	relativePath := filename
	if h.category != "" && h.article != "" {
		relativePath = filepath.Join(h.category, h.article, filename)
	}

	fullPath := filepath.Join(h.savePath, relativePath)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建目标文件
	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	// 保存文件
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	// 返回相对 URL
	return h.urlPrefix + "/" + strings.ReplaceAll(relativePath, "\\", "/"), nil
}

func (h *LocalHandler) SupportedTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"video/mp4",
		"video/webm",
		"audio/mpeg",
		"audio/wav",
		"application/pdf",
	}
}
