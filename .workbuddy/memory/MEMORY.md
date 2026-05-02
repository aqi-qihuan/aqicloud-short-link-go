# MEMORY.md - 长期记忆

## 项目概况：aqicloud-short-link-go
- **类型**: Go 短链接系统，6 个微服务架构
- **框架**: Gin v1.10.0 + GORM + MySQL + Redis + RabbitMQ + Kafka
- **认证**: JWT HS256, 密钥前缀 `dcloud-link`, 有效期 7 天, 密钥通过 `JWT_SECRET` 环境变量配置
- **端口**: Gateway:8888, Account:8001, Data:8002, Link:8003, Shop:8005, AI:8006
- **特殊架构**: 应用层分库分表（3 个数据库）、MQ 异步处理写操作
- **测试现状**: 2026-05-02 分析时测试覆盖率为 0，已制定完整测试方案
- **Git 工作流**: 2026-05-02 完成 Git 工作流建设
  - 分支策略: Git Flow 简化版 (main + develop + feature/fix/refactor/hotfix)
  - 提交规范: Conventional Commits（pre-commit + commit-msg hooks 自动校验）
  - CI: GitHub Actions (lint → test → build)
  - 代码质量: golangci-lint (errcheck/gosimple/govet/staticcheck 等)
  - 密钥目录 `密钥*/` 已被 .gitignore 排除，切勿提交
