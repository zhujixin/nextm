# Git 工作流规范

本文档定义 NextM 项目的 Git 协作规范，包括分支策略、Commit 规范、PR 流程和版本管理。

---

## 1. 分支策略

采用 **Trunk-Based Development** 为主、短生命周期特性分支为辅的策略。

### 1.1 分支类型

| 分支 | 命名格式 | 来源 | 合并到 |
|------|---------|------|--------|
| 主分支 | `main` | — | — |
| 开发分支 | `develop` | `main` | `main` |
| 特性分支 | `feat/<简短描述>` | `develop` | `develop` |
| 修复分支 | `fix/<简短描述>` | `develop` | `develop` |
| 紧急修复 | `hotfix/<简短描述>` | `main` | `main`, `develop` |
| 发布分支 | `release/v<版本号>` | `develop` | `main`, `develop` |
| 实验分支 | `experiment/<描述>` | `develop` | 不合并 |

### 1.2 分支生命周期

```
main  ─────●────────────●────────────●────
           \          / \          /
develop ────●──●──●──●───●──●──●──●───────
              \    /       \    /
feat/login     ●──●         ●──●
                           /
fix/header               ●
```

1. 从 `develop` 创建特性/修复分支
2. 在特性分支上开发
3. 提交 PR 到 `develop`
4. Code Review 通过后合并
5. `develop` 稳定后创建 `release/v*` 分支
6. 发布分支验证后合并到 `main` 并打 Tag

### 1.3 分支命名约定

```
feat/video-capture        # 新功能
feat/object-type-system   # 新功能
fix/ocr-crash             # 修复 bug
fix/login-redirect-loop   # 修复 bug
hotfix/security-auth-bypass  # 紧急安全修复
release/v1.0.0            # 发布分支
chore/update-deps         # 杂项（依赖更新等）
docs/api-usage            # 文档
refactor/service-layer    # 重构
test/sync-integration     # 测试
```

---

## 2. Commit 规范

采用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 2.1 格式

```
<类型>(<范围>): <简短描述>

<正文（可选）>

<脚注（可选）>
```

### 2.2 类型

| 类型 | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复 bug |
| `docs` | 文档变更 |
| `style` | 代码格式（不影响功能） |
| `refactor` | 重构（非新功能、非修复） |
| `perf` | 性能优化 |
| `test` | 添加/修改测试 |
| `chore` | 构建过程、依赖、工具 |

### 2.3 示例

```
feat(object): 添加视频截屏自动 OCR 提取

实现视频帧捕获后的 OCR 流水线，支持中文和英文文本提取。
使用 PaddleOCR 引擎，通过 LiteLLM 网关调用。

Closes #123
```

```
fix(search): 修复中文拼音模糊搜索的边界条件

当输入单个字符时，FTS5 解析器会产生空结果集。
添加对单字符输入的 fallback 处理。

Fixes #456
```

```
docs: 更新 API 文档中的认证流程

修正 JWT Refresh Token 的过期时间描述。
```

```
chore(deps): 升级 golangci-lint 至 v1.60

修复新 lint 规则导致的警告。
```

### 2.4 原则

- 每个 commit 保持原子性（一个 commit 一个逻辑变更）
- 简短描述使用祈使句、不超过 72 字符
- 正文说明 WHY 而非 WHAT
- 关联 Issue：`Closes #123`、`Fixes #456`、`Refs #789`

---

## 3. Pull Request 流程

### 3.1 PR 创建

1. 确保分支名符合规范，commit 已推送
2. 创建 PR 到目标分支（通常为 `develop`）
3. 填写 PR 模板

### 3.2 PR 模板

```markdown
## 摘要

<!-- 1-3 句描述变更内容 -->

## 相关 Issue

Closes #<!-- issue 编号 -->

## 变更类型

- [ ] 新功能 (feat)
- [ ] 修复 (fix)
- [ ] 重构 (refactor)
- [ ] 文档 (docs)
- [ ] 测试 (test)
- [ ] 杂项 (chore)

## 测试

- [ ] 单元测试
- [ ] 集成测试
- [ ] 手动测试（描述测试步骤）

## 截图（UI 变更时）

<!-- 可选 -->

## 检查清单

- [ ] 代码遵循项目编码规范
- [ ] 新增代码有对应的测试
- [ ] 所有测试通过
- [ ] 无新增 lint 警告
- [ ] API 变更已更新 OpenAPI 文档
- [ ] 数据库变更已更新 DB 文档
```

### 3.3 合并条件

- 至少 1 人 Code Review 通过
- CI 全部通过（lint + test + build）
- 无未解决的 Discussion
- PR 作者自行合并（保持所有权）

### 3.4 合并策略

| 情况 | 策略 |
|------|------|
| 特性分支合并到 develop | **Squash merge** — 保持 develop 历史整洁 |
| release 合并到 main | **Merge commit** — 保留发布分支历史 |
| hotfix 合并到 main | **Squash merge** — 快速合并 |

---

## 4. 版本管理

### 4.1 版本号

遵循 [Semantic Versioning 2.0](https://semver.org/)：

```
主版本.次版本.补丁 (例如 v1.0.0)
```

- **主版本**：不兼容的 API 变更
- **次版本**：向后兼容的新功能
- **补丁**：向后兼容的修复

### 4.2 Tag 规范

```bash
# 发布版本
git tag -a v1.0.0 -m "Release v1.0.0"
git tag -a v1.0.1 -m "Release v1.0.1"
git tag -a v2.0.0 -m "Release v2.0.0"

# 候选版本
git tag -a v1.0.0-rc.1 -m "Release candidate v1.0.0-rc.1"
```

### 4.3 Changelog

每个版本更新 `CHANGELOG.md`，格式参考 [Keep a Changelog](https://keepachangelog.com/)：

```markdown
## [v1.0.0] - 2026-06-15

### Added
- 视频截屏自动导入知识库功能 (#123)
- 全文搜索支持中文拼音模糊匹配 (#456)

### Fixed
- 修复 OCR 引擎在低光照条件下的识别率 (#789)

### Changed
- 升级 Go 版本至 1.24 (#101)
```

---

## 5. 工作流示例

### 5.1 日常开发

```bash
# 同步最新代码
git checkout develop
git pull

# 创建特性分支
git checkout -b feat/video-capture

# 开发 ... 多次 commit
git add .
git commit -m "feat(capture): 添加视频帧捕获接口"

# 推送并创建 PR
git push -u origin feat/video-capture
# → 在 GitHub 创建 PR → 等待 Review → 合并
```

### 5.2 处理 Review 反馈

```bash
# 在特性分支上修改
git add .
git commit -m "fix(capture): 修复 Review 指出的帧率计算错误"

# 推送更新 PR
git push
```

---

## 6. 禁止行为

- ❌ 直接向 `main` 推送（仅通过 PR 或 release 分支合并）
- ❌ 强制推送（force push）到共享分支（如 `develop`、`main`）
- ❌ 在 Review 未通过时自行合并 PR
- ❌ 提交包含密钥、Token、密码等敏感信息
- ❌ 提交大量无关文件的混合变更
