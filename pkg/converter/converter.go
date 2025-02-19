package converter

import (
	"io"

	"github.com/jomei/notionapi"
)

// Converter 定义了内容转换器的接口
type Converter interface {
	// Convert 将 Notion 页面转换为目标格式
	Convert(page notionapi.Page, blocks []notionapi.Block) error

	// SetOutput 设置输出位置
	SetOutput(path string) error

	// SetTemplate 设置模板
	SetTemplate(template string) error
}

// BlockProcessor 定义了块处理器的接口
type BlockProcessor interface {
	// ProcessBlock 处理单个块
	ProcessBlock(block notionapi.Block, w io.Writer) error

	// SupportedBlocks 返回支持的块类型
	SupportedBlocks() []string
}

// MediaHandler 定义了媒体处理器的接口
type MediaHandler interface {
	// SaveMedia 保存媒体文件并返回可访问的 URL
	SaveMedia(url string) (string, error)

	// SupportedTypes 返回支持的媒体类型
	SupportedTypes() []string
}

// MetadataProcessor 定义了元数据处理器的接口
type MetadataProcessor interface {
	// ProcessMetadata 处理页面元数据
	ProcessMetadata(page notionapi.Page) (map[string]interface{}, error)
}
