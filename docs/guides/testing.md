# 测试策略

本文档定义 NextM 项目的测试规范和最佳实践。

---

## 1. 测试金字塔

```
        ╱╲
       ╱ E2E ╲
      ╱────────╲
     ╱ 集成测试  ╲
    ╱──────────────╲
   ╱   单元测试      ╲     ← 核心，覆盖率 > 90%
  ╱────────────────────╲
```

| 层 | 工具 | 覆盖目标 | 运行频率 |
|----|------|---------|---------|
| 单元测试 | `go test` + testify | 核心逻辑 > 90% | 每次 commit |
| 集成测试 | testcontainers-go | Repository + Service | 每次 PR (CI) |
| API 测试 | httptest | Handler | 每次 PR (CI) |
| E2E 测试 | Playwright | 关键用户路径 | 每次发布 |

---

## 2. Go 后端测试

### 2.1 测试文件组织

测试文件与源码同目录，以 `_test.go` 结尾：

```
internal/
├── service/
│   ├── object/
│   │   ├── service.go
│   │   ├── service_test.go       # 单元测试
│   │   └── service_integration_test.go  # 集成测试（build tag）
│   └── search/
│       ├── search.go
│       └── search_test.go
└── repository/
    └── db/
        ├── queries/
        │   ├── objects.sql
        │   ├── objects_test.go    # SQLc 查询测试
        └── migrations/
```

### 2.2 单元测试

```go
// service/object/service_test.go
package object_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateObject(t *testing.T) {
    // Arrange
    mockRepo := new(mockObjectRepository)
    svc := NewService(mockRepo)

    // Act
    obj, err := svc.Create(ctx, CreateObjectDTO{...})

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, obj)
    assert.Equal(t, "test title", obj.Title)
}
```

### 2.3 集成测试

使用 build tag `//go:build integration` 标记：

```go
// service/object/service_integration_test.go
//go:build integration

package object_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
)

func TestCreateObjectWithPostgres(t *testing.T) {
    // 使用 testcontainers 启动 PostgreSQL
    pg, err := testcontainers.GenericContainer(ctx, ...)
    // ...
}
```

运行方式：

```bash
# 仅单元测试（快）
make test-unit  # go test ./internal/... -short

# 包含集成测试（需要 Docker）
make test-int   # go test ./internal/... -tags=integration

# 全部测试
make test       # go test ./...
```

### 2.4 Mock 策略

- 使用 `mockgen` 从接口生成 mock
- mock 生成文件存放于 `internal/pkg/mock/` 目录
- 优先 mock 接口（repository、service），不 mock 外部 API
- 外部 API 使用 test server 替代 mock

```bash
# 生成 mock
mockgen -source=internal/service/object/service.go \
        -destination=internal/mock/service/object.go \
        -package=mock_service
```

### 2.5 Race Detection

所有测试默认开启 race detection：

```bash
go test -race ./...
# 在 CI 中强制执行
```

### 2.6 表驱动测试

```go
func TestValidateObject(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateObjectDTO
        wantErr bool
    }{
        {"empty title", CreateObjectDTO{Title: ""}, true},
        {"long title", CreateObjectDTO{Title: strings.Repeat("a", 1000)}, true},
        {"valid object", CreateObjectDTO{Title: "test", Type: "note"}, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateObject(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2.7 Fuzz 测试

```go
func FuzzParseObjectID(f *testing.F) {
    f.Add("550e8400-e29b-41d4-a716-446655440000")
    f.Fuzz(func(t *testing.T, id string) {
        result, err := ParseObjectID(id)
        if err == nil {
            assert.NotEmpty(t, result)
        }
    })
}
```

---

## 3. API 测试

使用 `net/http/httptest`：

```go
// internal/api/handler/object_test.go
func TestCreateObjectHandler(t *testing.T) {
    // 设置 handler
    handler := NewObjectHandler(mockService)
    router := chi.NewRouter()
    router.Post("/api/v1/objects", handler.Create)

    // 构造请求
    body := `{"title":"test","type":"note"}`
    req := httptest.NewRequest("POST", "/api/v1/objects", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    // 执行
    router.ServeHTTP(rec, req)

    // 断言
    assert.Equal(t, http.StatusCreated, rec.Code)
    assert.JSONEq(t, `{"id":"...","title":"test"}`, rec.Body.String())
}
```

---

## 4. 前端测试

### 4.1 Web (Vitest + Testing Library)

```tsx
// frontend/web/src/features/object/ObjectCard.test.tsx
import { render, screen } from '@testing-library/react'
import { ObjectCard } from './ObjectCard'

describe('ObjectCard', () => {
    it('renders title', () => {
        render(<ObjectCard title="测试标题" />)
        expect(screen.getByText('测试标题')).toBeDefined()
    })

    it('fires onSelect when clicked', async () => {
        const onSelect = vi.fn()
        render(<ObjectCard title="test" onSelect={onSelect} />)
        await userEvent.click(screen.getByRole('button'))
        expect(onSelect).toHaveBeenCalledOnce()
    })
})
```

### 4.2 Flutter (flutter_test)

```dart
// frontend/mobile/test/features/object/object_card_test.dart
void main() {
    testWidgets('ObjectCard displays title', (tester) async {
        await tester.pumpWidget(ObjectCard(title: '测试标题'));
        expect(find.text('测试标题'), findsOneWidget);
    });
}
```

---

## 5. E2E 测试

使用 Playwright 覆盖关键用户路径：

```bash
# 运行 E2E 测试
cd frontend/web
pnpm exec playwright test
```

### E2E 场景清单（优先级由高到低）

| 场景 | 覆盖范围 | 优先级 |
|------|---------|--------|
| 用户注册/登录 | 完整的认证流程 | P0 |
| 创建笔记 | 编辑器、保存、展示 | P0 |
| 搜索 | 全文搜索、搜索结果 | P0 |
| 文件导入 | 拖拽、OCR 结果 | P1 |
| 同步 | 多设备同步状态 | P1 |

---

## 6. 性能测试

使用 k6 进行 API 性能测试：

```bash
# 执行性能测试
k6 run scripts/load-test.js
```

性能目标：

| 指标 | 目标 |
|------|------|
| API P95 延迟 | < 200ms |
| 搜索 P95 延迟 | < 500ms |
| 同步 P95 延迟 | < 1s |
| 并发用户（单节点） | 1000 |

---

## 7. 测试覆盖率要求

```bash
# 查看覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

| 包 | 目标覆盖率 |
|----|-----------|
| `internal/model/` | > 90% |
| `internal/service/` | > 85% |
| `internal/api/handler/` | > 80% |
| `internal/repository/` | > 70%（集成测试覆盖） |
| `internal/pkg/` | > 90% |

---

## 8. CI 集成

GitHub Actions 自动运行：

```yaml
# .github/workflows/test.yml（示例）
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go test -race ./...
      - run: make lint
```

---

## 9. 最佳实践

- **FIRST 原则**：Fast（快速）、Isolated（隔离）、Repeatable（可重复）、Self-validating（自验证）、Timely（及时）
- 测试不依赖外部网络（离线可运行）
- 测试数据使用工厂函数生成，避免硬编码
- 集成测试使用 testcontainers，不用共享数据库
- 避免 flaky 测试（如依赖时间的测试用 mock clock）
- 失败时提供清晰的消息
