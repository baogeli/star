#!/bin/bash

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}开始编译项目...${NC}"

# 进入项目目录
cd "$(dirname "$0")"

# 设置目标操作系统
export GOOS=linux
export GOARCH=amd64

# 创建 build目录（如果不存在）
mkdir -p ./build

# 编译项目
echo -e "${YELLOW}编译目标：Linux AMD64${NC}"
go build -o ./build/star main.go

# 检查编译结果
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 编译成功！${NC}"
    echo -e "${YELLOW}输出文件：${NC}./build/star"
    
    # 显示文件信息
    if command -v file &> /dev/null; then
        echo -e "${YELLOW}文件信息：${NC}"
        file ./build/star
    fi
    
    # 显示文件大小
    echo -e "${YELLOW}文件大小：${NC}"
    ls -lh ./build/star | awk '{print $5}'
else
    echo -e "${RED}✗ 编译失败！${NC}"
    exit 1
fi
