#!/bin/bash
# ============================================================
# build.sh - 编译 alert-webhook（交叉编译为 Linux amd64 二进制）
# 目标平台: CentOS 7 / Linux amd64
# ============================================================

set -e

APP_NAME="alert-webhook"
BUILD_DIR="./build"

echo "==> 清理旧构建..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

echo "==> 整理依赖..."
go mod tidy

echo "==> 交叉编译为 Linux amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o "${BUILD_DIR}/${APP_NAME}" \
    .

echo "==> 复制配置和资源文件..."
cp -r templates "${BUILD_DIR}/"
cp -r static    "${BUILD_DIR}/"
cp .env.example "${BUILD_DIR}/" 2>/dev/null || true

echo ""
echo "✅ 编译完成！输出目录: ${BUILD_DIR}/"
echo ""
echo "   部署步骤（在 CentOS 7 服务器上执行）:"
echo "   1. 上传整个 build/ 目录到服务器"
echo "   2. 复制 .env.example 为 .env 并修改数据库配置"
echo "   3. chmod +x ${APP_NAME}"
echo "   4. 使用 systemd 或 nohup 运行:"
echo "      nohup ./${APP_NAME} > /dev/null 2>&1 &"
echo "      或配置 systemd service（见 README.md）"
