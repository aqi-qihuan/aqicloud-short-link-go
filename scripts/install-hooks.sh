#!/bin/bash
# 安装 Git Hooks
# 用法: bash scripts/install-hooks.sh

set -e

echo "🔗 安装 Git Hooks..."

# 设置 hooks 路径为项目目录
git config core.hooksPath scripts/githooks

echo "✅ Git Hooks 已安装！"
echo ""
echo "已启用的 hooks:"
echo "  - pre-commit: gofmt / go vet / golangci-lint / 敏感信息检查"
echo "  - commit-msg: Conventional Commits 格式校验"
echo ""
echo "跳过检查: git commit --no-verify"
