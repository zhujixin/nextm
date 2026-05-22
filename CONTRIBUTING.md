# 贡献指南

感谢你对 NextM 的关注！欢迎参与开发。

---

## 行为准则

请保持专业、友善的沟通态度。欢迎建设性的讨论和代码审查，拒绝人身攻击和恶意行为。

---

## 如何参与

### 报告 Bug

1. 在 [Issues](https://github.com/your-org/nextm/issues) 搜索是否已有相同报告
2. 如无，创建新 Issue，使用 Bug Report 模板
3. 包含：环境信息、复现步骤、期望行为、实际行为、截图/日志

### 提出新功能

1. 先在 [Discussions](https://github.com/your-org/nextm/discussions) 发起讨论
2. 收集反馈后创建 Feature Request Issue
3. 说明：使用场景、目标用户、期望行为、备选方案

### 提交代码

1. 在 Issue 下留言认领任务，避免重复工作
2. 遵循 [Git 工作流规范](docs/guides/git-workflow.md)
3. 遵循 [编码规范](docs/guides/coding-standards.md)
4. 确保测试覆盖
5. 提交 Pull Request

---

## 开发流程

```
Issue → 讨论 → 认领 → 特性分支 → 开发 → 测试 → PR → Review → 合并
```

### 分支命名

| 类型 | 格式 |
|------|------|
| 特性 | `feat/<简短描述>` |
| 修复 | `fix/<简短描述>` |
| 文档 | `docs/<简短描述>` |
| 重构 | `refactor/<简短描述>` |
| 测试 | `test/<简短描述>` |

### Commit 格式

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<类型>(<范围>): <简短描述>

<正文>
```

类型：`feat` | `fix` | `docs` | `style` | `refactor` | `perf` | `test` | `chore`

---

## 代码要求

- **测试覆盖**：新功能/修复必须有对应测试
- **Lint 通过**：`make lint` 无报错
- **编译通过**：`go build ./...` 或对应前端的类型检查
- **无数据竞争**：`go test -race` 通过
- **文档同步**：API/DB 变更需同步更新对应设计文档

---

## Pull Request 流程

1. 在 PR 模板中填写完整信息
2. 关联对应的 Issue：`Closes #123`
3. 至少等待 1 人 Review
4. 回复 Review 意见并修改
5. Review 通过后由作者自行合并

### PR 大小建议

- **200 行以内**：最佳，Review 效率最高
- **200-500 行**：可接受，尽量拆分
- **500 行以上**：建议拆分为多个 PR

---

## 搭建开发环境

详见 [开发环境搭建指南](docs/guides/setup.md)。

快速开始：

```bash
git clone https://github.com/your-org/nextm
cd nextm
make deps
make dev
```

---

## 疑问？

- 在 [Discussions](https://github.com/your-org/nextm/discussions) 发起讨论
- 在 Issue 中 @ 维护者
