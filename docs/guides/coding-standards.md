# 编码规范

本文档定义 NextM 项目的编码约定。所有代码须保持一致风格。

---

## 1. Go

### 1.1 通用规则

- 遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- 使用 `go fmt` 格式化代码（或 `gofumpt` 更严格的格式化）
- 使用 `go vet` 静态分析，无警告
- 使用 `golangci-lint`（配置见 `.golangci.yml`）做完整 lint

### 1.2 命名

| 类型 | 规则 | 示例 |
|------|------|------|
| 包名 | 全小写、单数形式、无下划线 | `service`, `repository` |
| 文件名 | 蛇形 snake_case | `object_service.go` |
| 接口名 | 方法名 + `er` 后缀 | `Storer`, `Parser` |
| 导出类型 | 驼峰 PascalCase | `KnowledgeObject` |
| 未导出类型 | 驼峰 camelCase | `objectService` |
| 导出函数 | 驼峰 PascalCase | `CreateObject` |
| 未导出函数 | 驼峰 camelCase | `validateObject` |
| 常量 | 驼峰 PascalCase 或全大写 | `MaxRetries`, `MAX_RETRIES` |
| 枚举值 | 类型 + PascalCase | `StatusActive`, `StatusArchived` |
| 接收者 | 1-2 字母缩写 | `s *Service`, `r *Repository` |

### 1.3 包组织结构

```
internal/
├── api/
│   ├── handler/       # HTTP/gRPC 处理器 — 每个资源一个文件
│   ├── middleware/     # 中间件 — 每个关注点一个文件
│   ├── router/         # 路由注册
│   └── dto/            # 请求/响应 DTO
├── service/            # 业务逻辑层 — 每个领域一个子包
├── repository/         # 数据访问层 — SQLc 生成 + 手工扩展
└── model/              # 领域模型
```

**依赖方向**: Handler → Service → Repository（单向依赖，禁止循环）

### 1.4 接口设计

- 在调用方侧定义接口（Consumer-side interface）
- 接口保持小粒度（建议 1-3 个方法）
- 使用值接收者（除非需要修改接收者）
- 返回 struct 值而非指针（除非为 nil 语义）

```go
// 好的做法
type ObjectService interface {
    Create(ctx context.Context, dto CreateObjectDTO) (*model.Object, error)
    Get(ctx context.Context, id string) (*model.Object, error)
    List(ctx context.Context, filter ObjectFilter) ([]model.Object, error)
}

// 避免大型接口
type ObjectService interface {
    Create(ctx context.Context, ...) ...
    Get(ctx context.Context, ...) ...
    Update(ctx context.Context, ...) ...
    Delete(ctx context.Context, ...) ...
    List(ctx context.Context, ...) ...
    BatchCreate(ctx context.Context, ...) ...
    BatchUpdate(ctx context.Context, ...) ...
    BatchDelete(ctx context.Context, ...) ...
    Search(ctx context.Context, ...) ...
    // ... 超过 5 个方法说明职责过重
}
```

### 1.5 错误处理

- 使用 `fmt.Errorf("context: %w", err)` 包装错误
- 定义领域错误类型（`var ErrNotFound = errors.New("object not found")`）
- API 层统一错误响应格式
- 避免在业务层直接记录错误，上抛给调用方处理

### 1.6 并发

- 使用 errgroup 管理 goroutine 生命周期
- 使用 channel 传递信号，sync.WaitGroup 等待完成
- 避免裸露的 goroutine（始终使用 errgroup 或 waitgroup 管理）
- 使用 `-race` 标志运行测试，确保无数据竞争

### 1.7 导入顺序

```
标准库
空行
第三方包
空行
内部包
```

```go
import (
    "context"
    "fmt"

    "github.com/chibisov/go-sqlite3"
    "github.com/nats-io/nats.go"

    "github.com/your-org/nextm/internal/model"
    "github.com/your-org/nextm/internal/pkg/logger"
)
```

---

## 2. TypeScript / React

### 2.1 通用规则

- 使用 TypeScript `strict` 模式
- 遵循 [React 官方文档](https://react.dev/) 推荐模式
- 使用 ESLint + Prettier（配置见 `frontend/web/.eslintrc.cjs`）

### 2.2 命名

| 类型 | 规则 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `ObjectCard.tsx` |
| Hook 文件 | camelCase + `use` 前缀 | `useObject.ts` |
| 工具文件 | camelCase | `formatDate.ts` |
| 组件名 | PascalCase | `ObjectCard` |
| 函数名 | camelCase | `formatDate` |
| 类型/接口 | PascalCase + `Type`/`Props` 后缀 | `ObjectType`, `ObjectCardProps` |
| 枚举 | PascalCase | `ObjectStatus` |

### 2.3 组件结构

```tsx
// 推荐的组件组织
interface ObjectCardProps {
  object: KnowledgeObject
  onSelect?: (id: string) => void
}

export function ObjectCard({ object, onSelect }: ObjectCardProps) {
  // 逻辑在顶部
  const handleClick = useCallback(() => {
    onSelect?.(object.id)
  }, [object.id, onSelect])

  // JSX 在底部
  return (
    <div onClick={handleClick}>
      <h3>{object.title}</h3>
    </div>
  )
}
```

### 2.4 状态管理

- 优先使用 React 内置状态（`useState` / `useReducer`）
- 跨组件共享使用 React Context
- 仅对复杂跨页面状态使用 Zustand（按模块拆分 store）
- 避免 prop drilling 超过 3 层

### 2.5 样式

- 使用 Tailwind CSS 4 utility classes
- 提取复用的 class 组合到组件内部常量
- 复杂组件样式使用 CSS modules（`.module.css`）

---

## 3. Dart / Flutter

### 3.1 通用规则

- 遵循 [Effective Dart](https://dart.dev/guides/language/effective-dart)
- 使用 `dart format` 格式化
- 使用 `flutter analyze` 静态分析

### 3.2 命名

| 类型 | 规则 | 示例 |
|------|------|------|
| 文件 | 蛇形 snake_case | `object_card.dart` |
| 类名 | PascalCase | `ObjectCard` |
| 变量/函数 | camelCase | `objectTitle`, `formatDate` |
| 常量 | lowerCamelCase | `maxRetries` |
| 枚举 | PascalCase | `ObjectStatus` |
| 私有成员 | 下划线前缀 | `_internalState` |

### 3.3 Widget 组织

```
frontend/mobile/
├── lib/
│   ├── main.dart
│   ├── app.dart              # 应用入口
│   ├── core/                 # 基础设施
│   │   ├── theme/
│   │   ├── router/
│   │   └── network/
│   ├── features/             # 按功能模块分组
│   │   ├── object/
│   │   │   ├── models/
│   │   │   ├── providers/
│   │   │   ├── screens/
│   │   │   ├── widgets/
│   │   │   └── services/
│   │   └── collection/
│   └── shared/
│       ├── widgets/
│       └── utils/
```

---

## 4. SQL（SQLc）

### 4.1 命名

- SQL 文件名蛇形命名：`objects.sql`
- 查询命名使用 `camelCase`：`-- name: GetObject :one`
- 表和字段使用蛇形：`knowledge_objects`, `created_at`

### 4.2 约定

- 每个查询使用命名参数：`sqlc.arg(name)` 或 `@name`
- 手写 SQL 优先可读性，避免过度优化
- 复杂查询添加注释说明意图
- Big-endian UUID 存储

---

## 5. 通用约定

### 5.1 提交前检查清单

- [ ] `make lint` 通过
- [ ] `make test` 通过
- [ ] 新增代码有对应的测试
- [ ] 无硬编码的敏感信息（密钥、Token 等）
- [ ] API 变更同步更新了 OpenAPI 规范

### 5.2 注释原则

- 写 **Why** 而非 **What**（好的代码本身说明 What）
- 公共导出项必须写 Go doc 注释
- TODO 注释标明作者和 Issue 编号：`// TODO(username): issue #123 修复边界情况`
- 避免无意义注释：`// Increment counter` → 不要写，代码本身已体现

### 5.3 安全生产

- 密钥/Token/密码绝不可提交到代码仓库
- 本地配置使用 `.env.local` 或 `config.local.yaml`
- 添加新依赖前评估许可证兼容性和安全风险
- 提交前检查是否包含调试代码（`fmt.Println`, `console.log`, `print()`）
