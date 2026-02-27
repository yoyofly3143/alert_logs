#!/bin/bash
# ============================================================
# setup-git.sh - 初始化并推送到内网 GitLab
# 仓库: gitlab.ttpai.work:sre/alertlogs.git
# ============================================================

set -e

REMOTE_URL="git@gitlab.ttpai.work:sre/alertlogs.git"

echo "==> 检查 git 配置..."
# 如果还没有配置 git 用户信息，请先取消注释以下两行并填写
# git config user.name  "你的名字"
# git config user.email "你的邮箱@ttpai.work"

echo "==> 初始化本地仓库（如已初始化则跳过）..."
if [ ! -d ".git" ]; then
    git init
    echo "   已初始化 git 仓库"
else
    echo "   已有 git 仓库，跳过初始化"
fi

echo "==> 配置 .gitignore..."
cat > .gitignore << 'EOF'
# 编译产物
build/
*.exe
alert-webhook

# 日志
logs/
*.log

# 环境变量（含密码，不提交）
.env

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
EOF

echo "==> 确保远程仓库地址正确..."
if git remote get-url origin 2>/dev/null; then
    echo "   已存在 origin，更新为: ${REMOTE_URL}"
    git remote set-url origin "${REMOTE_URL}"
else
    echo "   添加 origin: ${REMOTE_URL}"
    git remote add origin "${REMOTE_URL}"
fi

echo "==> 暂存所有文件..."
git add -A

echo "==> 创建初始提交..."
git commit -m "feat: alert-webhook 初始版本

- 接收 AlertManager Webhook 并存储到 MySQL
- 支持字段: alertname/status/severity/instance/job/summary/description
- Web 界面: 深色主题告警监控Dashboard
- 支持过滤/分页/告警详情弹窗
- 日志写入 logs/ 目录，程序后台静默运行" 2>/dev/null || echo "   (无新变更可提交)"

echo "==> 推送到 GitLab (branch: main)..."
git branch -M main
git push -u origin main

echo ""
echo "✅ 推送完成！"
echo "   仓库地址: ${REMOTE_URL}"
