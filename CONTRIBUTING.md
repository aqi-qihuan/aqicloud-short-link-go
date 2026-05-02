# AqiCloud Short-Link Go — 团队协作规范

## 分支策略

采用 **Git Flow 简化版**，适合中小团队快速迭代：

```
main            ← 生产环境，始终可部署
  └── develop   ← 开发主线，集成分支
       ├── feature/xxx   ← 新功能
       ├── fix/xxx       ← Bug 修复
       └── refactor/xxx  ← 重构
```

### 分支规则

| 分支 | 来源 | 合入目标 | 命名规则 | 说明 |
|------|------|----------|----------|------|
| `main` | — | — | — | 生产代码，打 tag 发版 |
| `develop` | `main` | `main` | — | 日常开发集成分支 |
| `feature/*` | `develop` | `develop` | `feature/短链接-批量创建` | 新功能 |
| `fix/*` | `develop` 或 `main` | `develop` | `fix/redis连接泄漏` | Bug 修复 |
| `refactor/*` | `develop` | `develop` | `refactor/分库分表抽象` | 代码重构 |
| `hotfix/*` | `main` | `main` + `develop` | `hotfix/jwt过期判断` | 线上紧急修复 |

### 工作流程

```bash
# 1. 从 develop 拉取功能分支
git checkout develop && git pull
git checkout -b feature/批量创建短链接

# 2. 开发，小步提交
git add -p                    # 暂存区精细选择
git commit -m "feat(link): 添加批量创建短链接接口"

# 3. 保持分支与 develop 同步（避免大冲突）
git fetch origin
git rebase origin/develop     # 推荐 rebase 保持线性历史

# 4. 推送并创建 PR
git push origin feature/批量创建短链接
# → 在 GitHub/GitLab 创建 Pull Request → develop

# 5. Code Review + CI 通过后合并（Squash Merge）
# 6. 删除远程和本地功能分支
```

---

## 提交规范（Conventional Commits）

**格式**: `type(scope): 简短描述`

### Type 类型

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(link): 支持自定义短链接后缀` |
| `fix` | Bug 修复 | `fix(account): 修复 JWT 过期判断逻辑` |
| `refactor` | 重构（不改功能） | `refactor(common): 抽取 Redis 连接池配置` |
| `perf` | 性能优化 | `perf(link): 短链重定向热点路径优化` |
| `test` | 测试相关 | `test(link): 添加短链创建单元测试` |
| `docs` | 文档更新 | `docs: 更新部署文档` |
| `chore` | 构建/工具/依赖 | `chore: 升级 gin 到 v1.10.1` |
| `ci` | CI/CD 配置 | `ci: 添加 GitHub Actions 工作流` |
| `style` | 代码格式 | `style: gofmt 格式化` |
| `revert` | 回滚 | `revert: 回滚短链批量创建功能` |

### Scope 范围（对应服务/模块）

| Scope | 对应目录 |
|-------|----------|
| `gateway` | `cmd/gateway/`, `internal/gateway/` |
| `account` | `cmd/account/`, `internal/account/` |
| `link` | `cmd/link/`, `internal/link/` |
| `data` | `cmd/data/`, `internal/data/` |
| `shop` | `cmd/shop/`, `internal/shop/` |
| `ai` | `cmd/ai/`, `internal/ai/` |
| `common` | `internal/common/` |
| `docker` | `Dockerfile`, `docker-compose.yml` |
| `deploy` | `deploy/` |

### 提交示例

```bash
# 好的提交
feat(link): 支持自定义短链接后缀
fix(account): 修复登录时密码加密不一致的问题
refactor(common): 统一 Redis key 命名规范
perf(link): 短链重定向增加本地缓存

# 不好的提交（避免）
update code
fix bug
修改
WIP
提交
```

### 多行提交（复杂变更）

```
feat(link): 支持批量创建短链接

- 新增 POST /api/short-link/batch-create 接口
- 支持单次最多 100 条短链接批量创建
- 使用事务保证批量操作的原子性
- 添加速率限制防止滥用

Closes #42
```

---

## Pull Request 规范

### PR 标题
与提交规范保持一致：`feat(link): 支持批量创建短链接`

### PR 描述模板

```markdown
## 变更内容
- 新增/修改/删除了什么

## 影响范围
- [ ] gateway  [ ] account  [ ] link  [ ] data  [ ] shop  [ ] ai  [ ] common

## 测试
- [ ] 单元测试通过
- [ ] 本地集成测试通过
- [ ] 已验证不影响其他服务

## 截图/日志（如有 UI 变更或关键日志）

## 关联 Issue
Closes #
```

### Code Review 要点

- **必须**：至少 1 人 Approve 才能合并
- **必须**：CI 流水线通过（lint + test + build）
- **建议**：关注边界条件、错误处理、并发安全
- **建议**：关注是否涉及分库分表路由逻辑变更

---

## 合并策略

| 场景 | 合并方式 | 说明 |
|------|----------|------|
| 功能分支 → develop | **Squash Merge** | 压缩为一个整洁的提交 |
| hotfix → main | **Merge Commit** | 保留修复历史 |
| develop → main (发版) | **Merge Commit** | 保留版本历史 |

---

## Git 配置建议

```bash
# 设置 pull 默认使用 rebase
git config --global pull.rebase true

# 设置 push 默认使用当前分支
git config --global push.default current

# 自动处理换行符
git config --global core.autocrlf input    # macOS/Linux
git config --global core.autocrlf true     # Windows

# 设置默认分支名
git config --global init.defaultBranch main
```

---

## 版本发布流程

```bash
# 1. 从 develop 创建 release 分支
git checkout develop && git pull
git checkout -b release/v1.1.0

# 2. 最终测试和修复（只修 bug，不加功能）

# 3. 合并到 main 并打 tag
git checkout main && git merge release/v1.1.0
git tag -a v1.1.0 -m "Release v1.1.0: 批量创建短链接"

# 4. 同步回 develop
git checkout develop && git merge release/v1.1.0

# 5. 推送
git push origin main develop --tags

# 6. 删除 release 分支
git branch -d release/v1.1.0
```

---

## 常见问题

**Q: 我的分支落后 develop 很多怎么办？**
```bash
git fetch origin
git rebase origin/develop
# 如有冲突，逐个解决后 git rebase --continue
```

**Q: 我不小心提交到了 main 怎么办？**
```bash
git checkout -b feature/误提交的功能   # 先创建分支保存
git checkout main
git reset --hard HEAD~1               # 回退 main
git checkout feature/误提交的功能      # 继续在功能分支开发
```

**Q: 提交信息写错了怎么办？**
```bash
# 最后一次提交：修改提交信息
git commit --amend

# 已推送的提交：创建修正提交
git commit --amend
git push --force-with-lease           # 安全的 force push
```
