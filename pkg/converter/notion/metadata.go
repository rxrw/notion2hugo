package notion

import (
	"fmt"
	"strings"
	"time"

	"notion2md/pkg/converter"

	"github.com/jomei/notionapi"
)

type MetadataProcessor struct {
	config struct {
		CategoryMap map[string]string
		Status      struct {
			Draft string
		}
		Properties struct {
			Title       string `json:"title"`
			Categories  string `json:"categories"`
			Tags        string `json:"tags"`
			Status      string `json:"status"`
			Description string `json:"description"`
			Author      string `json:"author"`
			MetaTitle   string `json:"metaTitle"`
			Slug        string `json:"slug"`
			Toc         string `json:"toc"`
			Comments    string `json:"comments"`
			Weight      string `json:"weight"`
		}
	}
}

func NewMetadataProcessor(config *converter.Config) *MetadataProcessor {
	p := &MetadataProcessor{}
	p.config.CategoryMap = config.Notion.CategoryMap
	p.config.Status.Draft = config.Notion.Status.Draft
	p.config.Properties = config.Notion.Properties
	return p
}

func (p *MetadataProcessor) ProcessMetadata(page notionapi.Page) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// 调试输出所有属性
	for name, prop := range page.Properties {
		fmt.Printf("Property %s: %T\n", name, prop)
	}

	// 基本字段
	if title, ok := page.Properties["Name"].(*notionapi.TitleProperty); ok {
		metadata["title"] = processRichText(title.Title)
	}
	metadata["date"] = page.CreatedTime.Format(time.RFC3339)
	metadata["lastmod"] = page.LastEditedTime.Format(time.RFC3339)

	// 处理封面图片
	if page.Cover != nil {
		metadata["cover"] = page.Cover.GetURL()
	}

	// 处理分类
	var originalCategories []string
	var mappedCategories []string

	if cats, ok := page.Properties["Category"].(*notionapi.SelectProperty); ok {
		// 单选
		if cats.Select.Name != "" {
			if mapped, exists := p.config.CategoryMap[cats.Select.Name]; exists {
				originalCategories = []string{cats.Select.Name}
				mappedCategories = []string{mapped}
			} else {
				// 如果没有映射，跳过这篇文章
				fmt.Printf("警告: 分类 '%s' 未配置映射，跳过文章\n", cats.Select.Name)
				return nil, nil
			}
		}
	} else if cats, ok := page.Properties["Categories"].(*notionapi.MultiSelectProperty); ok {
		// 多选
		originalCategories = make([]string, 0, len(cats.MultiSelect))
		mappedCategories = make([]string, 0, len(cats.MultiSelect))

		for _, cat := range cats.MultiSelect {
			if mapped, exists := p.config.CategoryMap[cat.Name]; exists {
				originalCategories = append(originalCategories, cat.Name)
				mappedCategories = append(mappedCategories, mapped)
			} else {
				// 如果没有映射，跳过这篇文章
				fmt.Printf("警告: 分类 '%s' 未配置映射，跳过文章\n", cat.Name)
				return nil, nil
			}
		}
	}

	if len(originalCategories) > 0 {
		metadata["categories"] = originalCategories    // front matter 使用原始值
		metadata["category_dir"] = mappedCategories[0] // 目录使用映射值
	}

	// 处理标签
	if tags, ok := page.Properties["Tags"].(*notionapi.MultiSelectProperty); ok {
		tagList := make([]string, 0, len(tags.MultiSelect))
		for _, tag := range tags.MultiSelect {
			tagList = append(tagList, tag.Name)
		}
		if len(tagList) > 0 {
			metadata["tags"] = tagList
		}
	}

	metadata["author"] = page.CreatedBy.Name

	// 处理状态（草稿）
	if status, ok := page.Properties["Status"].(*notionapi.StatusProperty); ok {
		metadata["draft"] = status.Status.Name == p.config.Status.Draft
	}

	// 处理描述
	if desc, ok := page.Properties["Description"].(*notionapi.RichTextProperty); ok {
		description := processRichText(desc.RichText)
		if description != "" {
			metadata["description"] = description
		}
	}

	// 处理元标题
	if metaTitle, ok := page.Properties["Meta Title"].(*notionapi.RichTextProperty); ok {
		if len(metaTitle.RichText) > 0 {
			metadata["meta_title"] = processRichText(metaTitle.RichText)
		}
	}

	// 处理可选字段
	p.processOptionalFields(page, metadata)

	return metadata, nil
}

func (p *MetadataProcessor) processOptionalFields(page notionapi.Page, metadata map[string]interface{}) {
	// TOC
	if toc, ok := page.Properties["Toc"].(*notionapi.CheckboxProperty); ok {
		metadata["toc"] = toc.Checkbox
	}

	// Comments
	if comments, ok := page.Properties["Comments"].(*notionapi.CheckboxProperty); ok {
		metadata["comments"] = comments.Checkbox
	}

	// Slug
	if slug, ok := page.Properties["Slug"].(*notionapi.RichTextProperty); ok {
		if len(slug.RichText) > 0 {
			metadata["slug"] = processRichText(slug.RichText)
		}
	}
}

// 辅助函数
func processRichText(text []notionapi.RichText) string {
	var parts []string
	for _, t := range text {
		if t.Type == notionapi.ObjectTypeText {
			parts = append(parts, t.Text.Content)
		}
	}
	return strings.Join(parts, "")
}
