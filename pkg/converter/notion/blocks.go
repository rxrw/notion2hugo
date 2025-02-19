package notion

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"notion2md/pkg/converter"

	"github.com/jomei/notionapi"
)

type BlockProcessor struct {
	mediaHandler converter.MediaHandler
	codeStyle    string
	config       struct {
		UseShortcodes bool
		Image         struct {
			MaxWidth int
			Quality  int
			Formats  []string
		}
	}
}

func NewBlockProcessor(mediaHandler converter.MediaHandler, config *converter.Config) *BlockProcessor {
	return &BlockProcessor{
		mediaHandler: mediaHandler,
		codeStyle:    "github",
		config: struct {
			UseShortcodes bool
			Image         struct {
				MaxWidth int
				Quality  int
				Formats  []string
			}
		}{
			UseShortcodes: true,
			Image: struct {
				MaxWidth int
				Quality  int
				Formats  []string
			}{
				MaxWidth: config.Image.MaxWidth,
				Quality:  config.Image.Quality,
				Formats:  config.Image.Formats,
			},
		},
	}
}

func (p *BlockProcessor) ProcessBlock(block notionapi.Block, w io.Writer) error {
	switch b := block.(type) {
	case *notionapi.Heading1Block:
		return p.processHeading(w, b.Heading1.RichText, 1)
	case *notionapi.Heading2Block:
		return p.processHeading(w, b.Heading2.RichText, 2)
	case *notionapi.Heading3Block:
		return p.processHeading(w, b.Heading3.RichText, 3)
	case *notionapi.ParagraphBlock:
		return p.processParagraph(w, b)
	case *notionapi.BulletedListItemBlock:
		return p.processBulletList(w, b)
	case *notionapi.NumberedListItemBlock:
		return p.processNumberedList(w, b)
	case *notionapi.ToDoBlock:
		return p.processTodo(w, b)
	case *notionapi.ToggleBlock:
		return p.processToggle(w, b)
	case *notionapi.QuoteBlock:
		return p.processQuote(w, b)
	case *notionapi.CodeBlock:
		return p.processCode(w, b)
	case *notionapi.CalloutBlock:
		return p.processCallout(w, b)
	case *notionapi.ImageBlock:
		return p.processImage(w, b)
	case *notionapi.VideoBlock:
		return p.processVideo(w, b)
	case *notionapi.FileBlock:
		return p.processFile(w, b)
	case *notionapi.BookmarkBlock:
		return p.processBookmark(w, b)
	case *notionapi.EquationBlock:
		return p.processEquation(w, b)
	case *notionapi.DividerBlock:
		_, err := fmt.Fprintln(w, "---")
		return err
	case *notionapi.TableBlock:
		return p.processTable(w, b)
	case *notionapi.ColumnListBlock:
		return p.processColumns(w, b)
	}
	return nil
}

func (p *BlockProcessor) processHeading(w io.Writer, text []notionapi.RichText, level int) error {
	prefix := strings.Repeat("#", level)
	_, err := fmt.Fprintf(w, "%s %s\n\n", prefix, p.processRichText(text))
	return err
}

func (p *BlockProcessor) processParagraph(w io.Writer, block *notionapi.ParagraphBlock) error {
	text := p.processRichText(block.Paragraph.RichText)
	if text == "" {
		_, err := fmt.Fprintln(w)
		return err
	}
	_, err := fmt.Fprintf(w, "%s\n\n", text)
	return err
}

func (p *BlockProcessor) processRichText(text []notionapi.RichText) string {
	var buf bytes.Buffer
	for _, t := range text {
		switch t.Type {
		case notionapi.ObjectTypeText:
			content := t.Text.Content
			if t.Annotations.Bold {
				content = fmt.Sprintf("**%s**", content)
			}
			if t.Annotations.Italic {
				content = fmt.Sprintf("*%s*", content)
			}
			if t.Annotations.Strikethrough {
				content = fmt.Sprintf("~~%s~~", content)
			}
			if t.Annotations.Code {
				content = fmt.Sprintf("`%s`", content)
			}
			if t.Text.Link != nil {
				content = fmt.Sprintf("[%s](%s)", content, t.Text.Link.Url)
			}
			buf.WriteString(content)
		}
	}
	return buf.String()
}

func (p *BlockProcessor) processBulletList(w io.Writer, block *notionapi.BulletedListItemBlock) error {
	text := p.processRichText(block.BulletedListItem.RichText)
	_, err := fmt.Fprintf(w, "- %s\n", text)
	if err != nil {
		return err
	}

	// 处理子项
	if len(block.BulletedListItem.Children) > 0 {
		for _, child := range block.BulletedListItem.Children {
			_, err := fmt.Fprint(w, "  ") // 缩进
			if err != nil {
				return err
			}
			if err := p.ProcessBlock(child, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *BlockProcessor) processNumberedList(w io.Writer, block *notionapi.NumberedListItemBlock) error {
	text := p.processRichText(block.NumberedListItem.RichText)
	_, err := fmt.Fprintf(w, "1. %s\n", text)
	if err != nil {
		return err
	}

	// 处理子项
	if len(block.NumberedListItem.Children) > 0 {
		for _, child := range block.NumberedListItem.Children {
			_, err := fmt.Fprint(w, "   ") // 缩进
			if err != nil {
				return err
			}
			if err := p.ProcessBlock(child, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *BlockProcessor) processTodo(w io.Writer, block *notionapi.ToDoBlock) error {
	text := p.processRichText(block.ToDo.RichText)
	checkbox := "[ ]"
	if block.ToDo.Checked {
		checkbox = "[x]"
	}
	_, err := fmt.Fprintf(w, "- %s %s\n", checkbox, text)
	return err
}

func (p *BlockProcessor) processToggle(w io.Writer, block *notionapi.ToggleBlock) error {
	summary := p.processRichText(block.Toggle.RichText)
	_, err := fmt.Fprintf(w, "<details>\n<summary>%s</summary>\n\n", summary)
	if err != nil {
		return err
	}

	// 处理子内容
	for _, child := range block.Toggle.Children {
		if err := p.ProcessBlock(child, w); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(w, "</details>")
	return err
}

func (p *BlockProcessor) processQuote(w io.Writer, block *notionapi.QuoteBlock) error {
	text := p.processRichText(block.Quote.RichText)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		_, err := fmt.Fprintf(w, "> %s\n", line)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *BlockProcessor) processCode(w io.Writer, block *notionapi.CodeBlock) error {
	code := p.processRichText(block.Code.RichText)
	language := block.Code.Language
	if language == "plain text" {
		language = ""
	}
	_, err := fmt.Fprintf(w, "```%s\n%s\n```\n\n", language, code)
	return err
}

func (p *BlockProcessor) processCallout(w io.Writer, block *notionapi.CalloutBlock) error {
	text := p.processRichText(block.Callout.RichText)
	icon := "💡" // 默认图标
	if block.Callout.Icon != nil {
		switch block.Callout.Icon.Type {
		case "emoji":
			if block.Callout.Icon.Emoji != nil {
				icon = string(*block.Callout.Icon.Emoji) // 将 Emoji 类型转换为 string
			}
		case "external":
			icon = "🔗"
		case "file":
			icon = "📎"
		}
	}
	_, err := fmt.Fprintf(w, "> %s %s\n\n", icon, text)
	return err
}

func (p *BlockProcessor) processImage(w io.Writer, block *notionapi.ImageBlock) error {
	caption := p.processRichText(block.Image.Caption)
	if caption == "" {
		caption = "image"
	}

	url := block.Image.File.URL
	if block.Image.Type == "external" {
		url = block.Image.External.URL
	}

	// 如果配置了媒体处理器，使用它处理图片
	if p.mediaHandler != nil {
		newURL, err := p.mediaHandler.SaveMedia(url)
		if err != nil {
			return fmt.Errorf("处理图片失败: %w", err)
		}
		url = newURL
	}

	_, err := fmt.Fprintf(w, "![%s](%s)\n\n", caption, url)
	return err
}

func (p *BlockProcessor) processVideo(w io.Writer, block *notionapi.VideoBlock) error {
	url := block.Video.File.URL
	if block.Video.Type == "external" {
		url = block.Video.External.URL
	}

	// 处理 YouTube 视频
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		videoID := extractYouTubeID(url)
		_, err := fmt.Fprintf(w, "{{< youtube %s >}}\n\n", videoID)
		return err
	}

	// 其他视频使用 HTML5 video 标签
	_, err := fmt.Fprintf(w, "<video controls src=\"%s\"></video>\n\n", url)
	return err
}

func (p *BlockProcessor) processFile(w io.Writer, block *notionapi.FileBlock) error {
	var url, filename string

	switch block.File.Type {
	case "external":
		url = block.File.External.URL
		filename = filepath.Base(url)
	case "file":
		url = block.File.File.URL
		filename = filepath.Base(url)
	}

	if url == "" {
		return nil
	}

	// 如果是 PDF，使用特殊处理
	if strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		if p.config.UseShortcodes {
			_, err := fmt.Fprintf(w, "{{< pdf src=\"%s\" >}}\n\n", url)
			return err
		}
		_, err := fmt.Fprintf(w, "<embed src=\"%s\" type=\"application/pdf\" width=\"100%%\" height=\"600px\">\n\n", url)
		return err
	}

	// 普通文件生成下载链接
	_, err := fmt.Fprintf(w, "[%s](%s)\n\n", filename, url)
	return err
}

func (p *BlockProcessor) processBookmark(w io.Writer, block *notionapi.BookmarkBlock) error {
	title := block.Bookmark.URL
	if len(block.Bookmark.Caption) > 0 {
		title = p.processRichText(block.Bookmark.Caption)
	}
	_, err := fmt.Fprintf(w, "[%s](%s)\n\n", title, block.Bookmark.URL)
	return err
}

func (p *BlockProcessor) processEquation(w io.Writer, block *notionapi.EquationBlock) error {
	_, err := fmt.Fprintf(w, "$$\n%s\n$$\n\n", block.Equation.Expression)
	return err
}

func (p *BlockProcessor) processTable(w io.Writer, block *notionapi.TableBlock) error {
	if len(block.Table.Children) == 0 {
		return nil
	}

	// 处理表头
	firstRow := block.Table.Children[0].(*notionapi.TableRowBlock)
	for i, cell := range firstRow.TableRow.Cells {
		if i > 0 {
			fmt.Fprint(w, " | ")
		}
		fmt.Fprint(w, p.processRichText(cell))
	}
	fmt.Fprintln(w)

	// 分隔线
	for i := 0; i < len(firstRow.TableRow.Cells); i++ {
		if i > 0 {
			fmt.Fprint(w, " | ")
		}
		fmt.Fprint(w, "---")
	}
	fmt.Fprintln(w)

	// 处理数据行
	for i := 1; i < len(block.Table.Children); i++ {
		row := block.Table.Children[i].(*notionapi.TableRowBlock)
		for j, cell := range row.TableRow.Cells {
			if j > 0 {
				fmt.Fprint(w, " | ")
			}
			fmt.Fprint(w, p.processRichText(cell))
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
	return nil
}

func (p *BlockProcessor) processColumns(w io.Writer, block *notionapi.ColumnListBlock) error {
	_, err := fmt.Fprintln(w, "<div class=\"row\">")
	if err != nil {
		return err
	}

	for _, column := range block.ColumnList.Children {
		col, ok := column.(*notionapi.ColumnBlock)
		if !ok {
			continue
		}

		_, err = fmt.Fprintln(w, "<div class=\"col\">")
		if err != nil {
			return err
		}

		for _, block := range col.Column.Children {
			if err := p.ProcessBlock(block, w); err != nil {
				return err
			}
		}

		_, err = fmt.Fprintln(w, "</div>")
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(w, "</div>")
	return err
}

// 辅助函数
func extractYouTubeID(url string) string {
	if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) == 2 {
			return strings.Split(parts[1], "?")[0]
		}
	}
	if strings.Contains(url, "watch?v=") {
		parts := strings.Split(url, "watch?v=")
		if len(parts) == 2 {
			return strings.Split(parts[1], "&")[0]
		}
	}
	return url
}

func (p *BlockProcessor) SupportedBlocks() []string {
	return []string{
		"paragraph",
		"heading_1",
		"heading_2",
		"heading_3",
		"bulleted_list_item",
		"numbered_list_item",
		"to_do",
		"toggle",
		"quote",
		"code",
		"callout",
		"image",
		"video",
		"file",
		"bookmark",
		"equation",
		"divider",
		"table",
		"column_list",
	}
}

// 添加 getter 方法
func (p *BlockProcessor) GetMediaHandler() converter.MediaHandler {
	return p.mediaHandler
}
