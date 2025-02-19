package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"notion2md/pkg/converter"
	"notion2md/pkg/converter/hugo"
	"notion2md/pkg/converter/media"
	"notion2md/pkg/converter/notion"

	"github.com/briandowns/spinner"
	"github.com/jomei/notionapi"
	"github.com/schollz/progressbar/v3"
)

var (
	configFile string
	envFile    string
)

func init() {
	// 优先查找工作目录下的 notion.config.json
	if _, err := os.Stat("notion.config.json"); err == nil {
		configFile = "notion.config.json"
	}

	flag.StringVar(&configFile, "config", configFile, "配置文件路径")
	flag.StringVar(&envFile, "env", ".env", "环境变量文件路径")
}

func main() {
	flag.Parse()

	// 加载配置
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 检查必要的配置
	if config.DatabaseID == "" {
		log.Fatal("未设置 Notion 数据库 ID")
	}

	// 初始化 Notion 客户端
	token := os.Getenv("NOTION_SECRET")
	if token == "" {
		log.Fatal("未设置 NOTION_SECRET 环境变量")
	}
	client := notionapi.NewClient(notionapi.Token(token))

	// 初始化媒体处理器
	var mediaHandler converter.MediaHandler
	switch config.Storage.Type {
	case "local":
		mediaHandler = media.NewLocalHandler(
			config.Storage.Local.Path,
			config.Storage.Local.URLPrefix,
		)
	case "s3":
		var err error
		mediaHandler, err = media.NewS3Handler(
			config.Storage.S3.Bucket,
			config.Storage.S3.Region,
			config.Storage.S3.PathPrefix,
			config.Storage.S3.URLPrefix,
		)
		if err != nil {
			log.Fatalf("初始化 S3 处理器失败: %v", err)
		}
	default:
		log.Fatalf("不支持的存储类型: %s", config.Storage.Type)
	}

	// 初始化块处理器
	blockProcessor := notion.NewBlockProcessor(mediaHandler, config)

	// 初始化元数据处理器
	metaProcessor := notion.NewMetadataProcessor(config)

	// 初始化转换器
	conv := hugo.New(blockProcessor, metaProcessor)
	if err := conv.SetTemplate(config.Content.Archetype); err != nil {
		log.Fatalf("设置模板失败: %v", err)
	}
	if err := conv.SetOutput(config.Content.Folder); err != nil {
		log.Fatalf("设置输出目录失败: %v", err)
	}

	// 查询数据库
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "查询 Notion 数据库 "
	s.Start()
	pages, err := queryDatabase(client, config)
	s.Stop()
	if err != nil {
		log.Fatalf("查询数据库失败: %v", err)
	}
	fmt.Printf("✓ 找到 %d 篇文章\n", len(pages))

	// 处理每个页面
	bar := progressbar.Default(int64(len(pages)), "转换进度")
	for _, page := range pages {
		title := getPageTitle(page)
		bar.Describe(fmt.Sprintf("处理: %s", title))

		// 处理页面，如果返回 nil 说明跳过了这篇文章
		if err := processPage(client, conv, page, config); err != nil {
			if err == ErrSkipPage {
				log.Printf("⚠️ 跳过文章 [%s]: 未配置分类映射", title)
				continue
			}
			log.Printf("❌ 处理页面失败 [%s]: %v", title, err)
			continue
		}

		// 只有成功处理的文章才更新状态
		if err := updateStatus(client, page, config.Notion.Status.Published); err != nil {
			log.Printf("⚠️ 更新状态失败 [%s]: %v", title, err)
		} else {
			log.Printf("✓ 已完成: %s", title)
		}

		bar.Add(1)
	}
}

// 定义一个特殊的错误类型表示跳过文章
var ErrSkipPage = fmt.Errorf("跳过文章")

func queryDatabase(client *notionapi.Client, config *converter.Config) ([]notionapi.Page, error) {
	// 查询待发布的文章
	readyQuery := &notionapi.DatabaseQueryRequest{
		Filter: &notionapi.PropertyFilter{
			Property: "Status",
			Status: &notionapi.StatusFilterCondition{
				Equals: config.Notion.Status.Ready,
			},
		},
		PageSize: 100,
	}

	readyResp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(config.DatabaseID), readyQuery)
	if err != nil {
		return nil, fmt.Errorf("查询待发布文章失败: %w", err)
	}

	// 查询待删除的文章
	deleteQuery := &notionapi.DatabaseQueryRequest{
		Filter: &notionapi.PropertyFilter{
			Property: "Status",
			Status: &notionapi.StatusFilterCondition{
				Equals: config.Notion.Status.ToDelete,
			},
		},
		PageSize: 100,
	}

	deleteResp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(config.DatabaseID), deleteQuery)
	if err != nil {
		return nil, fmt.Errorf("查询待删除文章失败: %w", err)
	}

	// 合并结果
	return append(readyResp.Results, deleteResp.Results...), nil
}

func processPage(client *notionapi.Client, conv converter.Converter, page notionapi.Page, config *converter.Config) error {
	// 检查状态
	if status, ok := page.Properties["Status"].(*notionapi.StatusProperty); ok {
		if status.Status.Name == config.Notion.Status.ToDelete {
			log.Printf("🗑 删除文章: %s", getPageTitle(page))
			if err := updateStatus(client, page, config.Notion.Status.Deleted); err != nil {
				return fmt.Errorf("更新状态失败: %w", err)
			}
			return nil
		}
	}

	// 处理正常文章
	blocks, err := getPageBlocks(client, page.ID)
	if err != nil {
		return fmt.Errorf("获取页面内容失败: %w", err)
	}

	if err := conv.Convert(page, blocks); err != nil {
		if err == ErrSkipPage {
			return ErrSkipPage
		}
		return fmt.Errorf("转换内容失败: %w", err)
	}

	return nil
}

func getPageBlocks(client *notionapi.Client, pageID notionapi.ObjectID) ([]notionapi.Block, error) {
	resp, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(pageID), &notionapi.Pagination{
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	var blocks []notionapi.Block
	for _, block := range resp.Results {
		// 递归获取子块
		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.Paragraph.Children = children
		case *notionapi.BulletedListItemBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.BulletedListItem.Children = children
		case *notionapi.NumberedListItemBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.NumberedListItem.Children = children
		case *notionapi.ToDoBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.ToDo.Children = children
		case *notionapi.ToggleBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.Toggle.Children = children
		case *notionapi.QuoteBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.Quote.Children = children
		case *notionapi.CalloutBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.Callout.Children = children
		case *notionapi.ColumnListBlock:
			children, err := getPageBlocks(client, notionapi.ObjectID(b.ID))
			if err != nil {
				return nil, err
			}
			b.ColumnList.Children = children
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func updateStatus(client *notionapi.Client, page notionapi.Page, newStatus string) error {
	props := notionapi.Properties{
		"Status": notionapi.StatusProperty{
			Status: notionapi.Status{
				Name: newStatus,
			},
		},
	}

	_, err := client.Page.Update(context.Background(), notionapi.PageID(page.ID), &notionapi.PageUpdateRequest{
		Properties: props,
	})
	return err
}

func loadConfig(path string) (*converter.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config converter.Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Println("config", string(data))
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

func getPageTitle(page notionapi.Page) string {
	if title, ok := page.Properties["Name"].(*notionapi.TitleProperty); ok {
		if len(title.Title) > 0 {
			return title.Title[0].PlainText
		}
	}
	return string(page.ID)
}
