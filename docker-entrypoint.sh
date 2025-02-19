#!/bin/sh
set -e

# 如果没有配置文件，使用默认配置
if [ ! -f "notion.config.json" ]; then
    echo "未找到 notion.config.json，使用默认配置"
    exit 1
fi

# 运行转换器
exec notion2md 