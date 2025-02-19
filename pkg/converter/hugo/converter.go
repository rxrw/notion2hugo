package hugo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"notion2md/pkg/converter"
	"notion2md/pkg/converter/media"
	"notion2md/pkg/converter/notion"

	"github.com/jomei/notionapi"
	"github.com/mozillazg/go-pinyin"
)

type HugoConverter struct {
	outputPath     string
	templatePath   string
	blockProcessor converter.BlockProcessor
	metaProcessor  converter.MetadataProcessor
}

func New(blockProcessor converter.BlockProcessor, metaProcessor converter.MetadataProcessor) *HugoConverter {
	return &HugoConverter{
		blockProcessor: blockProcessor,
		metaProcessor:  metaProcessor,
	}
}

func (h *HugoConverter) Convert(page notionapi.Page, blocks []notionapi.Block) error {
	// 处理元数据
	metadata, err := h.metaProcessor.ProcessMetadata(page)
	if err != nil {
		return fmt.Errorf("处理元数据失败: %w", err)
	}

	// 获取第一个分类作为目录
	category := "uncategorized"
	if categoryDir, ok := metadata["category_dir"].(string); ok {
		category = categoryDir
	} else if categories, ok := metadata["categories"].([]string); ok && len(categories) > 0 {
		category = categories[0]
	}

	articleDir := h.generateFilename(page, metadata)
	articleDir = strings.TrimSuffix(articleDir, ".md")

	if handler, ok := h.blockProcessor.(*notion.BlockProcessor); ok {
		if mediaHandler, ok := handler.GetMediaHandler().(*media.LocalHandler); ok {
			mediaHandler.SetContext(category, articleDir)
		}
	}

	// 处理内容
	var content bytes.Buffer
	for _, block := range blocks {
		if err := h.blockProcessor.ProcessBlock(block, &content); err != nil {
			return fmt.Errorf("处理块失败: %w", err)
		}
	}

	// 渲染模板
	tmpl, err := template.ParseFiles(h.templatePath)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	// 处理元数据值
	data := map[string]interface{}{
		"Title":       getOrDefault(metadata, "title", ""),
		"MetaTitle":   getOrDefault(metadata, "meta_title", ""),
		"Description": getOrDefault(metadata, "description", ""),
		"Date":        getOrDefault(metadata, "date", ""),
		"Image":       getOrDefault(metadata, "cover", ""),
		"Author":      getOrDefault(metadata, "author", ""),
		"Draft":       getOrDefault(metadata, "draft", false),
		"Weight":      getOrDefault(metadata, "weight", 0),
		"Content":     content.String(),
		"Toc":         getOrDefault(metadata, "toc", false),
		"Comments":    getOrDefault(metadata, "comments", false),
		"Slug":        getOrDefault(metadata, "slug", ""),
		"Lastmod":     getOrDefault(metadata, "lastmod", ""),
	}

	// 特殊处理标签
	if tags, ok := metadata["tags"].([]string); ok && len(tags) > 0 {
		quotedTags := make([]string, len(tags))
		for i, tag := range tags {
			quotedTags[i] = fmt.Sprintf(`"%s"`, tag)
		}
		data["Tags"] = "[" + strings.Join(quotedTags, ", ") + "]"
	} else {
		data["Tags"] = "[]"
	}

	// 特殊处理分类
	if categories, ok := metadata["categories"].([]string); ok && len(categories) > 0 {
		quotedCats := make([]string, len(categories))
		for i, cat := range categories {
			quotedCats[i] = fmt.Sprintf(`"%s"`, cat)
		}
		data["Categories"] = "[" + strings.Join(quotedCats, ", ") + "]"
	} else {
		data["Categories"] = "[]"
	}

	// 创建输出文件
	filename := h.generateFilename(page, metadata)
	outputFile := filepath.Join(h.outputPath, category, filename)

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("渲染模板失败: %w", err)
	}

	return nil
}

func (h *HugoConverter) SetOutput(path string) error {
	h.outputPath = path
	return nil
}

func (h *HugoConverter) SetTemplate(template string) error {
	h.templatePath = template
	return nil
}

func (h *HugoConverter) generateFilename(page notionapi.Page, metadata map[string]interface{}) string {
	// 获取标题，如果为空则使用 ID
	title, ok := metadata["title"].(string)
	if !ok || title == "" {
		return string(page.ID) + ".md"
	}

	// 转换为拼音
	args := pinyin.NewArgs()
	args.Separator = "-" // 分隔符
	pys := pinyin.LazyPinyin(title, args)
	filename := strings.Join(pys, "-")

	// 清理文件名
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	filename = reg.ReplaceAllString(filename, "-")
	filename = strings.Trim(filename, "-")

	// 如果清理后文件名为空，使用 ID
	if filename == "" {
		filename = string(page.ID)
	}

	return filename + ".md"
}

func getOrDefault(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if v, ok := m[key]; ok && v != nil {
		return v
	}
	return defaultValue
}
