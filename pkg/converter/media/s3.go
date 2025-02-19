package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Handler struct {
	client     *s3.Client
	bucket     string
	pathPrefix string
	urlPrefix  string
}

func NewS3Handler(bucket, region, pathPrefix, urlPrefix string) (*S3Handler, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("无法加载 AWS 配置: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3Handler{
		client:     client,
		bucket:     bucket,
		pathPrefix: pathPrefix,
		urlPrefix:  urlPrefix,
	}, nil
}

func (h *S3Handler) SaveMedia(url string) (string, error) {
	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取内容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取内容失败: %w", err)
	}

	// 生成文件名
	filename := filepath.Base(url)
	if filename == "" || !strings.Contains(filename, ".") {
		ext := ".bin"
		switch resp.Header.Get("Content-Type") {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "video/mp4":
			ext = ".mp4"
		}
		filename = fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	}

	// 构建 S3 路径
	s3Path := filepath.Join(h.pathPrefix, filename)
	if !strings.HasPrefix(s3Path, "/") {
		s3Path = "/" + s3Path
	}

	// 上传到 S3
	_, err = h.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(h.bucket),
		Key:         aws.String(s3Path),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(resp.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("上传到 S3 失败: %w", err)
	}

	// 返回可访问的 URL
	return h.urlPrefix + "/" + filename, nil
}

func (h *S3Handler) SupportedTypes() []string {
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
