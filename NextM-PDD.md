# NextM — 产品设计说明书

**版本**: V1.1
**状态**: 交叉验证修订版
**最后更新**: 2026-05-22

> **范围说明**: 本文档覆盖 PRD 中 P0（核心体验）和大部分 P1 功能的设计。以下 PRD 识别的 P1/P2 需求在本版本 PDD 中暂未展开详细设计，将在后续 Phase 中补充：会议录音、AI Link 摘要、微信/Telegram 转发、系统音频捕获、代码片段检索、热力图、名片识别、思维导图识别、CLI 工具、插件市场、Webhook。详见 PRD 路线图。

---

## 1. 系统总体架构

### 1.1 分层架构图

```
┌──────────────────────────────────────────────────────────────────┐
│                        表示层 (UI Layer)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────┐  │
│  │  Web App │  │ Desktop  │  │  Mobile  │  │  Br. Extension │  │
│  │  React+TS│  │ Tauri+TS │  │  Flutter │  │  Chrome/FF     │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └───────┬────────┘  │
│       │              │             │                │           │
│  ┌────┴──────────────┴─────────────┴────────────────┴───────┐  │
│  │             状态管理层 (State Management)                   │  │
│  │           TanStack Query + Zustand + Signals              │  │
│  └───────────────────────────┬───────────────────────────────┘  │
├──────────────────────────────┼──────────────────────────────────┤
│                   Go 业务逻辑层 (Business Logic)                 │
│  ┌──────────┐ ┌──────────┐ ┌───────┐ ┌────────┐ ┌──────────┐  │
│  │ Object   │ │ Collection│ │Search │ │ Vision │ │ Agent    │  │
│  │ Service  │ │ Service  │ │Service│ │Service │ │ Service  │  │
│  │ (Go)     │ │ (Go)     │ │(Go)   │ │(Go)    │ │ (Go)     │  │
│  └────┬─────┘ └────┬─────┘ └───┬───┘ └───┬────┘ └────┬─────┘  │
│       │             │           │         │           │        │
│  ┌────┴─────────────┴───────────┴─────────┴───────────┴─────┐  │
│  │              同步引擎 (Sync Engine — Go)                   │  │
│  │        CRDT (Go-Yrs FFI) + WebSocket + Offline Queue     │  │
│  └───────────────────────────┬───────────────────────────────┘  │
├──────────────────────────────┼──────────────────────────────────┤
│                   Go 持久化层 (Persistence)                     │
│  ┌────────────────┐ ┌───────────────┐ ┌───────────────────┐   │
│  │  Local DB      │ │  File Store   │ │  Vector Index     │   │
│  │  SQLite (Go)   │ │  File System  │ │  LanceDB (Go SDK) │   │
│  └───────┬────────┘ └───────┬───────┘ └────────┬──────────┘   │
│          │                  │                   │              │
│  ┌───────┴──────────────────┴───────────────────┴──────────┐  │
│  │              云存储层 (Cloud Storage)                     │  │
│  │    PostgreSQL+pgvector / MinIO(S3) / Redis / NATS       │  │
│  │       (Go pgx / go-redis / nats.go)                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │          AI 服务层 (AI Service Layer — Go + LiteLLM)      │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────┐  │  │
│  │  │ 本地 LLM │ │ 云端 LLM │ │ Embedding│ │ OCR Pipeline │  │  │
│  │  │ Ollama   │ │ Claude/GPT│ │ BGE/M3E │ │ PaddleOCR   │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └─────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

### 1.2 技术栈详细选型

| 层级 | 技术 | 版本 | 选型依据 |
|------|------|------|---------|
| **后端语言** | Go | 1.24+ | 单二进制、goroutine 并发模型适合智能体、工具链完善 |
| **API 框架** | Chi Router + net/http | latest | stdlib 兼容、中间件链式组合、无重型框架 |
| **内部通信** | gRPC + Protocol Buffers | latest | 服务间高效通信、强类型 schema |
| **SQL 工具链** | SQLc + database/sql | latest | SQL 即代码 → 类型安全 Go 结构体，无 ORM 开销 |
| **本地数据库驱动** | modernc.org/sqlite (pure Go) | latest | 纯 Go 实现 SQLite，无 CGo 依赖可选 |
| **云端数据库驱动** | pgx (PostgreSQL) | v5 | 高性能纯 Go PG 驱动 |
| **数据库迁移** | golang-migrate | latest | 多数据库支持，CLI + Go 集成 |
| **Web 框架** | React + TypeScript | React 19+ | 类型安全、生态成熟 |
| **Web 构建** | Vite + Tailwind CSS 4 | Vite 6+ | 极速 HMR、原子化 CSS |
| **桌面端** | Tauri 2 (Rust) | — | 体积 ~5MB，内存占用 < Electron 60% |
| **移动端** | Flutter 3 | Dart 3 | 统一 iOS/Android，原生性能 |
| **状态管理** | Zustand + TanStack Query | — | 轻量、Server Cache |
| **搜索引擎** | Meilisearch | latest | 开箱即用的模糊搜索，中文友好 |
| **向量数据库** | LanceDB (本地) / pgvector (云端) | — | 本地嵌入式、云端可扩展 |
| **对象存储** | MinIO (自建) / Backblaze B2 | — | S3 兼容、成本可控 |
| **实时同步** | CRDT (Go-Yrs FFI) + WebSocket | — | Go 协程模型处理大量 WS 连接 |
| **消息队列** | NATS (nats.go) | latest | 轻量、JetStream 持久化、智能体通信 |
| **任务队列** | Asynq (go-redis) | latest | 延迟任务、定时调度、Go 原生 |
| **AI 网关** | LiteLLM (sidecar) + Go HTTP 客户端 | — | 统一多模型路由、成本追踪 |
| **OCR** | PaddleOCR + LayoutLMv3 | — | 中文 SOTA，轻量可本地部署 |
| **图像增强** | Real-ESRGAN (ncnn) | — | 移动端可运行的超分辨率 |
| **音视频** | Whisper.cpp (本地) / Deepgram API | — | 本地语音转文字 |
| **数据可视化** | D3.js + vis-network + Three.js | — | 图谱、3D、自定义图表 |
| **MCP SDK** | Go MCP 实现 (自建) | — | 标准 AI 协议支持 |
| **配置管理** | Viper + envconfig | latest | 文件/环境变量/远程配置 |
| **日志** | zap / slog (stdlib) | latest | 结构化、高性能 |
| **可观测** | OpenTelemetry Go SDK | latest | 分布式追踪、指标 |
| **CI/CD** | GitHub Actions + Docker | — | 自动化构建、Go 原生交叉编译 |

---

## 2. 模块详细设计

### 2.1 核心数据模型

```
┌──────────────────┐       ┌──────────────────┐
│      Space       │       │      Object      │
│  (工作区)        │ 1:N   │  (知识对象)      │
│ ─────────────── │──────▶│ ───────────────  │
│ uuid: string    │       │ uuid: string     │
│ name: string    │       │ spaceId: string  │
│ type: personal  │       │ type: ObjectType │ ← 动态类型系统
│ /team           │       │ title: string    │
│ accountId       │       │ content: Block[] │
│ createdAt       │       │ properties: Map  │
│ updatedAt       │       │ tags: string[]   │
│ encrypted: bool │       │ source: Video/   │
└──────────────────┘       │  Web/Camera     │
       │                   │ sourceMeta: {}  │
       │1:N                │ embedding: vec  │
       ▼                   │ createdAt       │
┌──────────────────┐       │ updatedAt       │
│   ObjectType     │       └───────┬──────────┘
│  (类型定义)      │               │
│ ───────────────  │               │N:M
│ uuid: string     │               │
│ name: string     │               ▼
│ icon: string     │       ┌──────────────────┐
│ fields: Field[]  │       │   Relation       │
│ color: string    │       │  (关系)          │
│ createdBy        │       │ ───────────────  │
└──────────────────┘       │ sourceId: string │
                           │ targetId: string │
┌──────────────────┐       │ type: link/      │
│      Block       │       │  reference/      │
│  (内容块)        │       │  citation        │
│ ───────────────  │       │ weight: float    │
│ id: string       │       │ createdAt        │
│ type: text/      │       └──────────────────┘
│  image/code/     │
│  table/          │       ┌──────────────────┐
│  mermaid/        │       │   Tag            │
│ children: []     │       │ ───────────────  │
│ properties: {}   │       │ id: string       │
│ position (canvas)│       │ name: string     │
└──────────────────┘       │ color: string    │
                           │ aiGenerated: bool│
┌──────────────────┐       └──────────────────┘
│    Collection    │
│  (数据库)        │       ┌──────────────────┐
│ ───────────────  │       │   Template       │
│ id: string       │       │ ───────────────  │
│ name: string     │       │ id: string       │
│ views: View[]    │       │ name: string     │
│ filters: Filter[]│       │ blocks: Block[]  │
│ sort: Sort[]     │       │ fields: Field[]  │
└──────────────────┘       └──────────────────┘
```

### 2.2 信息采集模块

#### 2.2.1 视频截屏采集

```
┌─────────────────────────────────────────────────────────────────┐
│                    截图采集服务 (Capture Service)                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │ System      │  │ Screenshot   │  │ Floating Button     │   │
│  │ Screenshot  │──│ Watcher      │  │ Overlay Service     │   │
│  │ Listener    │  │ (polling)    │  │ (悬浮按钮)           │   │
│  └──────┬──────┘  └──────┬───────┘  └─────────┬────────────┘   │
│         │                │                     │                │
│         └────────────────┼─────────────────────┘                │
│                          ▼                                      │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                  Screenshot Queue                        │   │
│  │        内存缓冲队列 → 去重 → 定时 Flush                    │   │
│  └──────────────────────────┬──────────────────────────────┘   │
│                             ▼                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Metadata Collector                          │   │
│  │  ┌─────────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐  │   │
│  │  │ Foreground  │ │ Video    │ │ Timestamp│ │Subtitle│  │   │
│  │  │ App Detector│ │ Title    │ │          │ │Capture │  │   │
│  │  └─────────────┘ └──────────┘ └──────────┘ └────────┘  │   │
│  └──────────────────────────┬──────────────────────────────┘   │
│                             ▼                                   │
│                   → Vision Service (OCR)                       │
│                   → AI Service (理解/标签/关联)                  │
│                   → Persistence Service (入库)                  │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

**关键接口定义**：

```go
// CaptureService 截图采集服务接口
type CaptureService interface {
    // StartFloatingButton 启动悬浮按钮 (Android Overlay / iOS Picture in Picture)
    StartFloatingButton(ctx context.Context, config FloatingButtonConfig) error
    // StopFloatingButton 停止悬浮按钮
    StopFloatingButton() error
    // WatchSystemScreenshot 监听系统截图 (返回 channel 接收截图路径)
    WatchSystemScreenshot(ctx context.Context) (<-chan string, error)
    // StartBurstMode 批量模式
    StartBurstMode(ctx context.Context) (*ScreenshotBatch, error)
    // ImportFromGallery 从相册导入
    ImportFromGallery(ctx context.Context, uris []string) ([]ProcessResult, error)
    // GetQueueStatus 获取截图处理队列状态
    GetQueueStatus(ctx context.Context) (*QueueStatus, error)
}

// ScreenshotMetadata 截图元数据
type ScreenshotMetadata struct {
    AppPackageName   string `json:"appPackageName,omitempty"`   // 前台 App 包名
    AppName          string `json:"appName,omitempty"`          // 前台 App 名称
    VideoTitle       string `json:"videoTitle,omitempty"`       // 视频标题
    VideoURL         string `json:"videoUrl,omitempty"`         // 视频链接
    Timestamp        int64  `json:"timestamp"`                  // 截图时间戳 (Unix ms)
    PlaybackPosition int64  `json:"playbackPosition,omitempty"` // 视频播放位置(秒)
    Subtitle         string `json:"subtitle,omitempty"`         // 当前字幕/弹幕
    ScreenshotPath   string `json:"screenshotPath"`             // 截图本地路径
    ImageHash        string `json:"imageHash"`                  // pHash 用于去重
}

// ProcessResult 处理结果
type ProcessResult struct {
    ScreenshotID string   `json:"screenshotId"`
    Status       string   `json:"status"`       // pending | processing | completed | duplicate | low_quality
    ObjectID     string   `json:"objectId,omitempty"`     // 生成的知识对象 ID
    OcrText      string   `json:"ocrText,omitempty"`
    Tags         []string `json:"tags,omitempty"`
    Error        string   `json:"error,omitempty"`
}
```

#### 2.2.2 图片信息提取 Pipeline

```go
// VisionService 图片信息提取 Pipeline 接口
type VisionService interface {
    // ProcessImage OCR + 版面分析完整流程
    ProcessImage(ctx context.Context, input ImageInput) (*ExtractionResult, error)

    // ProcessBatch 批量处理
    ProcessBatch(ctx context.Context, inputs []ImageInput, opts *BatchOptions) ([]ExtractionResult, error)

    // RecognizeFormula 公式识别
    RecognizeFormula(ctx context.Context, image ImageData) (*LatexFormula, error)

    // ExtractChartData 图表数据提取
    ExtractChartData(ctx context.Context, image ImageData) (*ChartData, error)
}

// ImageInput 图片输入
type ImageInput struct {
    Data   []byte `json:"data"`              // 文件 buffer
    Format string `json:"format"`            // png | jpg | webp | heic
    Source string `json:"source"`            // camera | screenshot | gallery | clipboard | url
    Options *ImageOptions `json:"options,omitempty"`
}

// ImageOptions 处理选项
type ImageOptions struct {
    Enhance        bool     `json:"enhance,omitempty"`          // 图像增强
    Language       []string `json:"language,omitempty"`         // OCR 语言 ["zh", "en"]
    ExtractTables  bool     `json:"extractTables,omitempty"`    // 提取表格
    ExtractFormula bool     `json:"extractFormula,omitempty"`   // 提取公式
}

// ExtractionResult 提取结果
type ExtractionResult struct {
    RawText  string        `json:"rawText"`            // 完整 OCR 文本
    Layout   []LayoutBlock `json:"layout"`             // 版面结构
    Tables   []TableData   `json:"tables,omitempty"`   // 表格数据
    Formulas []FormulaData `json:"formulas,omitempty"` // 公式
    Entities []Entity      `json:"entities,omitempty"` // 实体
    Quality  QualityInfo   `json:"quality"`            // 质量信息
    EnhancedImage []byte   `json:"enhancedImage,omitempty"` // 增强后的图片
}

// QualityInfo 质量信息
type QualityInfo struct {
    BlurScore     float64 `json:"blurScore"`     // 模糊度 (0-1, 1=最清晰)
    TextCoverage  float64 `json:"textCoverage"`  // 文字覆盖率
    OverallScore  float64 `json:"overallScore"`  // 综合质量分
}
```

### 2.3 信息组织模块

#### 2.3.1 对象类型系统

```go
// ObjectTypeDefinition 类型定义
type ObjectTypeDefinition struct {
    ID          string               `json:"id"`
    SpaceID     string               `json:"spaceId"`
    Name        string               `json:"name"`            // "书籍"、"论文"、"项目"、"人脉"
    Icon        string               `json:"icon"`            // 📚 📄 🚀 👤
    Description string               `json:"description,omitempty"`
    Color       string               `json:"color"`           // 图谱显示颜色
    Fields      []FieldDefinition    `json:"fields"`
    Relations   []RelationDefinition `json:"relations"`
    Templates   []string             `json:"templates"`       // 关联模板 ID
    IsBuiltin   bool                 `json:"isBuiltin"`       // 是否系统内置
    CreatedAt   int64                `json:"createdAt"`
    UpdatedAt   int64                `json:"updatedAt"`
}

// FieldType 字段类型枚举
type FieldType string

const (
    FieldTypeText        FieldType = "text"
    FieldTypeNumber      FieldType = "number"
    FieldTypeDate        FieldType = "date"
    FieldTypeSelect      FieldType = "select"
    FieldTypeMultiSelect FieldType = "multi_select"
    FieldTypeRelation    FieldType = "relation"
    FieldTypeRollup      FieldType = "rollup"
    FieldTypeFormula     FieldType = "formula"
    FieldTypeFile        FieldType = "file"
    FieldTypeEmail       FieldType = "email"
    FieldTypeURL         FieldType = "url"
)

// FieldDefinition 字段定义
type FieldDefinition struct {
    ID           string        `json:"id"`
    Name         string        `json:"name"`                   // "作者"、"出版社"、"状态"
    Type         FieldType     `json:"type"`
    Required     bool          `json:"required"`
    DefaultValue any           `json:"defaultValue,omitempty"`
    Options      []SelectOption `json:"options,omitempty"`     // 针对 select 类型
    Config       map[string]any `json:"config,omitempty"`
}

// SelectOption 下拉选项
type SelectOption struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Color string `json:"color,omitempty"`
}

// RelationDefinition 关系定义
type RelationDefinition struct {
    ID           string `json:"id"`
    Name         string `json:"name"`                   // "引用"、"隶属于"
    TargetTypeID string `json:"targetTypeId"`           // 关联的目标类型
    Type         string `json:"type"`                   // oneToOne | oneToMany | manyToMany
    ReverseName  string `json:"reverseName,omitempty"`  // 反向关系名称
}

// KnowledgeObject 知识对象实例
type KnowledgeObject struct {
    ID         string            `json:"id"`
    SpaceID    string            `json:"spaceId"`
    TypeID     string            `json:"typeId"`         // 关联 ObjectTypeDefinition
    Title      string            `json:"title"`
    Content    []Block           `json:"content"`
    Properties map[string]any    `json:"properties"`     // 根据类型定义动态生成
    Tags       []string          `json:"tags"`
    CoverImage string            `json:"coverImage,omitempty"`
    Source     ObjectSource      `json:"source"`
    Embedding  []float32         `json:"embedding"`      // 向量嵌入
    CreatedAt  int64             `json:"createdAt"`
    UpdatedAt  int64             `json:"updatedAt"`
    LastReadAt int64             `json:"lastReadAt,omitempty"`
}
```

#### 2.3.2 数据库视图引擎

```go
// ViewType 视图类型枚举
type ViewType string

const (
    ViewTypeTable    ViewType = "table"
    ViewTypeKanban   ViewType = "kanban"
    ViewTypeGallery  ViewType = "gallery"
    ViewTypeCalendar ViewType = "calendar"
    ViewTypeTimeline ViewType = "timeline"
    ViewTypeList     ViewType = "list"
)

// CollectionView 数据库视图
type CollectionView struct {
    ID            string                `json:"id"`
    Name          string                `json:"name"`
    Type          ViewType              `json:"type"`
    Filters       []FilterGroup         `json:"filters"`
    Sorts         []SortRule            `json:"sorts"`
    GroupBy       string                `json:"groupBy,omitempty"`     // 分组的字段 ID
    VisibleFields []string              `json:"visibleFields"`
    ColumnWidths  map[string]float64    `json:"columnWidths,omitempty"`
    CalendarField string                `json:"calendarField,omitempty"` // 日历视图使用的日期字段
    KanbanField   string                `json:"kanbanField,omitempty"`  // 看板视图使用的状态字段
}

// FilterGroup 筛选组
type FilterGroup struct {
    ID         string             `json:"id"`
    Operator   string             `json:"operator"`   // and | or
    Conditions []FilterCondition  `json:"conditions"`
}

// FilterCondition 筛选条件
type FilterCondition struct {
    FieldID  string `json:"fieldId"`
    Operator string `json:"operator"`   // eq | neq | gt | gte | lt | lte | contains | startsWith | isEmpty | isNotEmpty | in | notIn
    Value    any    `json:"value"`
}
```

### 2.4 搜索模块

```go
// SearchService 搜索服务接口
type SearchService interface {
    // FullTextSearch 全文搜索 (Meilisearch/SQLite FTS5)
    FullTextSearch(ctx context.Context, query SearchQuery) (*SearchResult, error)
    // SemanticSearch 语义搜索 (向量)
    SemanticSearch(ctx context.Context, query SearchQuery) (*SearchResult, error)
    // HybridSearch 混合搜索 (RRF 融合)
    HybridSearch(ctx context.Context, query SearchQuery) (*SearchResult, error)
    // DeepSearch 深度搜索 (AI 多步推理)
    DeepSearch(ctx context.Context, q string) (*DeepSearchResult, error)
    // Reindex 索引管理
    Reindex(ctx context.Context, spaceID ...string) error
    // GetRelated 推荐相关内容
    GetRelated(ctx context.Context, objectID string, limit int) ([]RelatedItem, error)
}

// SortOrder 排序枚举
type SortOrder string

const (
    SortRelevance   SortOrder = "relevance"
    SortCreatedAt   SortOrder = "createdAt"
    SortUpdatedAt   SortOrder = "updatedAt"
    SortTitle       SortOrder = "title"
)

// SearchQuery 搜索查询
type SearchQuery struct {
    Query     string          `json:"query"`
    SpaceID   string          `json:"spaceId,omitempty"`
    Filters   []SearchFilter  `json:"filters,omitempty"`
    Sort      SortOrder       `json:"sort,omitempty"`
    Limit     int             `json:"limit"`
    Offset    int             `json:"offset"`
    ObjectTypeIDs []string    `json:"objectTypeIds,omitempty"` // 限定类型
    TagIDs    []string        `json:"tagIds,omitempty"`
}

// SearchResult 搜索结果
type SearchResult struct {
    Items          []SearchItem `json:"items"`
    Total          int          `json:"total"`
    Suggestion     string       `json:"suggestion,omitempty"`     // AI 搜索建议
    RelatedQueries []string     `json:"relatedQueries,omitempty"` // 相关搜索词
}

// DeepSearchResult 深度搜索结果
type DeepSearchResult struct {
    Answer            string       `json:"answer"`              // AI 综合答案
    Sources           []SearchItem `json:"sources"`             // 引用来源
    Reasoning         string       `json:"reasoning"`           // 推理过程
    FollowUpQuestions []string     `json:"followUpQuestions"`   // 追问建议
}
```

### 2.5 智能体模块

#### 2.5.1 智能体定义

```go
// AgentType 智能体类型
type AgentType string

const (
    AgentSummarizer   AgentType = "summarizer"   // 总结
    AgentTagger       AgentType = "tagger"       // 标签
    AgentLinker       AgentType = "linker"       // 关联
    AgentReview       AgentType = "review"       // 复习
    AgentWriter       AgentType = "writer"       // 写作
    AgentCollector    AgentType = "collector"    // 采集
    AgentRefactor     AgentType = "refactor"     // 整理
    AgentOrchestrator AgentType = "orchestrator" // 编排
    AgentScheduler    AgentType = "scheduler"    // 定时
    AgentCustom       AgentType = "custom"       // 用户自定义
)

// AgentDefinition 智能体定义
type AgentDefinition struct {
    ID             string          `json:"id"`
    Name           string          `json:"name"`
    Description    string          `json:"description,omitempty"`
    Type           AgentType       `json:"type"`
    Triggers       []AgentTrigger  `json:"triggers"`       // 触发条件
    Actions        []AgentAction   `json:"actions"`        // 执行动作
    Schedule       string          `json:"schedule,omitempty"`       // cron 表达式
    EventPatterns  []EventPattern  `json:"eventPatterns,omitempty"`  // 事件监听
    Dependencies   []string        `json:"dependencies,omitempty"`   // 依赖的智能体
    TimeoutMs      int64           `json:"timeoutMs"`                // 超时 ms
    MaxRetries     int             `json:"maxRetries"`
    MaxTokensPerRun int            `json:"maxTokensPerRun"`
    MaxMemoryItems  int            `json:"maxMemoryItems"`
    Enabled        bool            `json:"enabled"`
    CreatedAt      int64           `json:"createdAt"`
}

// TriggerType 触发类型
type TriggerType string

const (
    TriggerOnCreate   TriggerType = "onCreate"
    TriggerOnUpdate   TriggerType = "onUpdate"
    TriggerOnDelete   TriggerType = "onDelete"
    TriggerOnSchedule TriggerType = "onSchedule"
    TriggerOnEvent    TriggerType = "onEvent"
    TriggerOnManual   TriggerType = "onManual"
    TriggerOnWebhook  TriggerType = "onWebhook"
)

// AgentTrigger 智能体触发器
type AgentTrigger struct {
    Type   TriggerType      `json:"type"`
    Config map[string]any   `json:"config,omitempty"`
}

// ActionType 动作类型
type ActionType string

const (
    ActionLLMCall      ActionType = "llm_call"
    ActionAPICall      ActionType = "api_call"
    ActionSearch       ActionType = "search"
    ActionCreateObject ActionType = "create_object"
    ActionUpdateObject ActionType = "update_object"
    ActionNotification ActionType = "send_notification"
    ActionRunWorkflow  ActionType = "run_workflow"
)

// AgentAction 智能体动作
type AgentAction struct {
    Type   ActionType     `json:"type"`
    Params map[string]any `json:"params"`
}
```

#### 2.5.2 长周期自主智能体

```go
// LHAState 长周期智能体状态
type LHAState string

const (
    LHAStateRunning   LHAState = "running"
    LHAStatePaused    LHAState = "paused"
    LHAStateCompleted LHAState = "completed"
    LHAStateFailed    LHAState = "failed"
)

// LongHorizonAgent 长周期自主智能体
type LongHorizonAgent struct {
    ID               string         `json:"id"`
    SpaceID          string         `json:"spaceId"`
    Name             string         `json:"name"`
    Goal             string         `json:"goal"`            // "追踪 AI 芯片行业动态"
    Plan             []AgentStep    `json:"plan"`            // 分解后的步骤
    State            LHAState       `json:"state"`
    ShortTermMemory  []MemoryItem   `json:"shortTermMemory,omitempty"`  // 当前任务上下文
    Progress         float64        `json:"progress"`        // 0-100
    StartedAt        int64          `json:"startedAt"`
    LastActiveAt     int64          `json:"lastActiveAt,omitempty"`
    Deadline         int64          `json:"deadline,omitempty"`
    Deliverables     []Deliverable  `json:"deliverables,omitempty"`     // 阶段性产出
    NextCheckIn      int64          `json:"nextCheckIn,omitempty"`      // 下次向用户汇报时间
}

// StepType 步骤类型
type StepType string

const (
    StepCollect   StepType = "collect"
    StepAnalyze   StepType = "analyze"
    StepSummarize StepType = "summarize"
    StepRelate    StepType = "relate"
    StepReport    StepType = "report"
)

// StepStatus 步骤状态
type StepStatus string

const (
    StepPending   StepStatus = "pending"
    StepRunning   StepStatus = "running"
    StepCompleted StepStatus = "completed"
    StepSkipped   StepStatus = "skipped"
    StepFailed    StepStatus = "failed"
)

// AgentStep 智能体步骤
type AgentStep struct {
    ID            string      `json:"id"`
    Type          StepType    `json:"type"`
    Status        StepStatus  `json:"status"`
    AssignedAgent string      `json:"assignedAgent"`       // 执行智能体 ID
    DependsOn     []string    `json:"dependsOn"`           // 依赖的上一步
    Result        any         `json:"result,omitempty"`
    Error         string      `json:"error,omitempty"`
    StartedAt     int64       `json:"startedAt,omitempty"`
    CompletedAt   int64       `json:"completedAt,omitempty"`
}
```

#### 2.5.3 智能体消息总线

```go
// MsgType 消息类型
type MsgType string

const (
    MsgTaskRequest MsgType = "task_request"
    MsgTaskResult  MsgType = "task_result"
    MsgQuery       MsgType = "query"
    MsgEvent       MsgType = "event"
    MsgHeartbeat   MsgType = "heartbeat"
)

// Priority 优先级
type Priority string

const (
    PriorityLow      Priority = "low"
    PriorityNormal   Priority = "normal"
    PriorityHigh     Priority = "high"
    PriorityCritical Priority = "critical"
)

// AgentMessage 智能体消息
type AgentMessage struct {
    ID             string         `json:"id"`
    Type           MsgType        `json:"type"`
    Source         string         `json:"source"`            // 发送方 agent ID
    Target         string         `json:"target,omitempty"`  // 接收方 agent ID (空 = 广播)
    ConversationID string         `json:"conversationId"`    // 会话 ID
    Payload        any            `json:"payload"`
    Context        MessageContext `json:"context"`
    Priority       Priority       `json:"priority"`
    Timestamp      int64          `json:"timestamp"`
    TTL            int64          `json:"ttl"`               // 消息存活时间 ms
    RetryCount     int            `json:"retryCount"`
}

// MessageContext 消息上下文
type MessageContext struct {
    UserID    string `json:"userId"`
    SpaceID   string `json:"spaceId"`
    ObjectID  string `json:"objectId,omitempty"`
    SessionID string `json:"sessionId,omitempty"`
}

// MessageBus 智能体消息总线接口 (基于 NATS JetStream)
type MessageBus interface {
    Publish(ctx context.Context, msg AgentMessage) (string, error)
    Subscribe(ctx context.Context, pattern MessagePattern, handler MessageHandler) (Unsubscribe, error)
    Request(ctx context.Context, msg AgentMessage, timeout time.Duration) (*AgentMessage, error)
    GetStatus(ctx context.Context, agentID string) (*AgentStatus, error)
    GetQueueLength(ctx context.Context) (map[string]int, error)
}

// MessageHandler 消息处理函数
type MessageHandler func(ctx context.Context, msg AgentMessage) error

// Unsubscribe 取消订阅函数
type Unsubscribe func() error

// Orchestrator 编排引擎接口
type Orchestrator interface {
    ExecuteWorkflow(ctx context.Context, workflow WorkflowDefinition, contextData any) (*WorkflowResult, error)
    ExecuteDAG(ctx context.Context, dag DagDefinition, contextData any) (*DagResult, error)
    ExecuteLongTerm(ctx context.Context, goal string, opts *LHAOptions) (*LongHorizonAgent, error)
}
```

### 2.6 同步引擎

```go
// SyncEngine CRDT 同步引擎接口
type SyncEngine interface {
    // Connect 连接同步服务
    Connect(ctx context.Context, spaceID, userID string) error
    // Disconnect 断开连接
    Disconnect() error

    // PushChange 本地变更 → CRDT → 推送
    PushChange(ctx context.Context, objectID string, change Change) error

    // OnRemoteChange 接收远程变更 channel
    OnRemoteChange(ctx context.Context) (<-chan Change, error)

    // GetStatus 获取同步状态
    GetStatus(ctx context.Context) (*SyncStatus, error)

    // FlushOfflineQueue 离线日志回放
    FlushOfflineQueue(ctx context.Context) error

    // ResolveConflict CRDT 冲突解决
    ResolveConflict(ctx context.Context, objectID string, local, remote ObjectState) (*ResolvedState, error)

    // SetSyncScope 选择性同步
    SetSyncScope(ctx context.Context, scope SyncScope) error
}

// ConnState 连接状态
type ConnState string

const (
    ConnConnected    ConnState = "connected"
    ConnConnecting   ConnState = "connecting"
    ConnDisconnected ConnState = "disconnected"
    ConnSyncing      ConnState = "syncing"
    ConnOffline      ConnState = "offline"
)

// SyncStatus 同步状态
type SyncStatus struct {
    State          ConnState `json:"state"`
    LastSyncAt     int64     `json:"lastSyncAt,omitempty"`
    PendingChanges int       `json:"pendingChanges"`
    Conflicts      int       `json:"conflicts"`
    PeerConnections int      `json:"peerConnections"`
}

// SyncMsgType 同步消息类型
type SyncMsgType string

const (
    SyncMsgDelta    SyncMsgType = "delta"
    SyncMsgSnapshot SyncMsgType = "snapshot"
    SyncMsgAck      SyncMsgType = "ack"
    SyncMsgPing     SyncMsgType = "ping"
    SyncMsgPong     SyncMsgType = "pong"
    SyncMsgConflict SyncMsgType = "conflict"
)

// SyncMessage WebSocket 同步消息 (goroutine-per-conn 模型)
type SyncMessage struct {
    Type       SyncMsgType `json:"type"`
    SpaceID    string      `json:"spaceId"`
    UserID     string      `json:"userId"`
    ObjectType string      `json:"objectType"`   // object | collection | type | tag
    ObjectID   string      `json:"objectId"`
    Data       []byte      `json:"data"`         // CRDT 二进制变更
    Timestamp  int64       `json:"timestamp"`
    Version    int64       `json:"version"`
}
```

### 2.7 MCP 连接器

```go
// MCPConnector MCP 协议实现接口
type MCPConnector interface {
    // GetTools 获取工具定义 (暴露给外部 AI 的工具)
    GetTools(ctx context.Context) ([]MCPToolDefinition, error)
    // HandleToolCall 处理 MCP 请求
    HandleToolCall(ctx context.Context, toolName string, args any) (any, error)
    // GetResources 获取资源定义
    GetResources(ctx context.Context) ([]MCPResourceDefinition, error)
    // ReadResource 处理资源读取
    ReadResource(ctx context.Context, uri string) (*ResourceContent, error)
}

// MCPToolDefinition MCP 工具定义
type MCPToolDefinition struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema JSONSchema  `json:"inputSchema"`
    Handler     func(ctx context.Context, args any) (any, error) `json:"-"`
}

// MCPResourceDefinition MCP 资源定义
type MCPResourceDefinition struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent 资源内容
type ResourceContent struct {
    URI      string `json:"uri"`
    MimeType string `json:"mimeType"`
    Text     string `json:"text,omitempty"`
    Blob     []byte `json:"blob,omitempty"`
}
```

---

## 3. 数据库设计

### 3.1 设计概览

NextM 采用**本地优先 + 云端同步**的双库架构：

| 维度 | 本地 SQLite | 云端 PostgreSQL |
|------|-----------|----------------|
| 角色 | 离线主存储、CRDT 变更日志 | 持久化权威副本、多设备同步枢纽 |
| 账号 | 仅缓存当前账号 ID | 账号注册、认证、MFA、订阅管理 |
| 向量搜索 | LanceDB (独立引擎) | pgvector 索引 |
| 全文搜索 | SQLite FTS5 | Meilisearch (独立引擎) |
| 文件存储 | 本地文件系统 | MinIO/S3 对象存储 |

### 3.2 核心表结构

| 模块 | 表名 | 说明 |
|------|------|------|
| 空间与账号 | `spaces`, `space_members`, `account_devices` | 工作区与多账号隔离 |
| 对象系统 | `object_types`, `type_fields`, `type_relations` | 动态类型定义（归一化存储） |
| 知识对象 | `objects`, `blocks`, `object_properties`, `object_versions` | 核心内容 + EAV 属性 + 版本历史 |
| 关系标签 | `relations`, `tags`, `object_tags` | 知识图谱基础 + AI 自动关联 |
| 集合视图 | `collections`, `collection_views`, `collection_items` | Notion 式多视图引擎 |
| 模板 | `templates` | 预设模板 + 模板市场 |
| 采集视觉 | `capture_sessions`, `image_queue`, `vision_jobs` | 视频截屏 + OCR Pipeline |
| 智能体 | `agents`, `agent_runs`, `agent_messages`, `long_horizon_agents`, `agent_steps`, `agent_memory` | 多智能体协作 + 消息总线 + 长周期任务 |
| 同步 | `sync_log`, `sync_checkpoint`, `sync_conflicts` | CRDT 变更日志 + 冲突管理 |
| 文件 | `files`, `file_chunks` | 附件存储 + 分片上传 |
| AI | `ai_cache`, `ai_usage_logs`, `search_history` | 响应缓存 + 用量计费 |
| MCP | `mcp_connectors`, `mcp_audit_log` | 开放协议 + 审计 |
| 通知分享 | `notifications`, `share_links`, `export_jobs` | 推送 + 分享 + 导出 |
| 回收站 | `trash` | 30 天自动清理 |

> 完整建表 SQL、索引策略、FTS5 配置、向量搜索（pgvector + LanceDB）、分区策略、迁移方案详见独立的 **[NextM-DB.md](./NextM-DB.md)**。

### 3.3 关键设计决策

- **主键**: UUID v4 文本存储，避免 CRDT 同步冲突
- **时间戳**: Unix 毫秒整数 (INTEGER)，统一 UTC
- **软删除**: 核心表 `is_deleted` 标记，数据移入 `trash`
- **JSON 字段**: 不常查询的元数据使用 JSON；需查询的字段拆为独立列（如 `type_fields` 归一化）
- **Block Position**: 浮点数相邻平均法，支持中间插入无需重新编号
- **空间隔离**: 所有多租户数据通过 `space_id` 字段隔离，索引必含 `space_id`
- **47 张表**（本地 37 张 + 云端独有 10 张）

---

## 4. API 设计

### 4.1 API 规范

- **基础路径**: `/api/v1`
- **格式**: JSON (request/response)
- **认证**: Bearer JWT + Refresh Token 轮换
- **分页**: `?limit=20&offset=0` 或 `?cursor=xxx`
- **错误格式**: `{ "error": { "code": "xxx", "message": "xxx", "details": {} } }`

### 4.2 核心 API

```
# ─── 对象 CRUD ─────────────────────────────────────────────

GET    /api/v1/objects                    # 列表(支持筛选/排序/分页)
POST   /api/v1/objects                    # 创建对象
GET    /api/v1/objects/:id               # 获取详情(含块内容)
PUT    /api/v1/objects/:id               # 更新对象
DELETE /api/v1/objects/:id               # 软删除
PATCH  /api/v1/objects/:id/archive       # 归档/取消归档

# ─── 内容块 ────────────────────────────────────────────────

GET    /api/v1/objects/:id/blocks        # 获取块列表
POST   /api/v1/objects/:id/blocks        # 添加块
PUT    /api/v1/blocks/:id                # 更新块
DELETE /api/v1/blocks/:id                # 删除块
PATCH  /api/v1/blocks/reorder            # 批量重排

# ─── 对象类型 ───────────────────────────────────────────────

GET    /api/v1/types                     # 获取所有类型定义
POST   /api/v1/types                     # 创建自定义类型
PUT    /api/v1/types/:id                 # 更新类型
DELETE /api/v1/types/:id                 # 删除类型

# ─── 数据库视图 ─────────────────────────────────────────────

GET    /api/v1/collections               # 列表
POST   /api/v1/collections               # 创建视图
PUT    /api/v1/collections/:id           # 更新
DELETE /api/v1/collections/:id           # 删除

# ─── 搜索 ──────────────────────────────────────────────────

GET    /api/v1/search?q=&type=&tags=&limit=&offset=
POST   /api/v1/search/semantic           # 语义搜索
POST   /api/v1/search/hybrid             # 混合搜索
POST   /api/v1/search/deep               # 深度搜索(AI)
GET    /api/v1/objects/:id/related       # 关联推荐

# ─── 图片处理 ──────────────────────────────────────────────

POST   /api/v1/vision/ocr                # OCR 识别
POST   /api/v1/vision/extract            # 完整提取(OCR+版面)
POST   /api/v1/vision/batch              # 批量处理
POST   /api/v1/vision/enhance            # 图像增强

# ─── 视频截屏 ──────────────────────────────────────────────

POST   /api/v1/capture/screenshot        # 提交截图
GET    /api/v1/capture/queue             # 获取队列状态
POST   /api/v1/capture/batch             # 批量提交
GET    /api/v1/capture/sessions          # 历史采集记录

# ─── 智能体 ────────────────────────────────────────────────

GET    /api/v1/agents                    # 智能体列表
POST   /api/v1/agents                    # 创建智能体
POST   /api/v1/agents/:id/trigger        # 手动触发
GET    /api/v1/agents/:id/runs           # 运行历史
POST   /api/v1/agents/:id/long-term      # 启动长周期任务
POST   /api/v1/workflow                  # 执行工作流

# ─── 关系 ──────────────────────────────────────────────────

GET    /api/v1/relations                 # 查询关系
POST   /api/v1/relations                 # 创建关系
DELETE /api/v1/relations/:id             # 删除关系
GET    /api/v1/graph/subgraph            # 获取子图(图谱渲染)

# ─── 同步 ──────────────────────────────────────────────────

WS     /api/v1/sync                      # WebSocket 同步连接
POST   /api/v1/sync/push                 # 推送变更
GET    /api/v1/sync/pull                 # 拉取变更
GET    /api/v1/sync/status               # 同步状态

# ─── MCP ───────────────────────────────────────────────────

GET    /api/v1/mcp/tools                 # 获取 MCP 工具列表
POST   /api/v1/mcp/call                  # 调用 MCP 工具
GET    /api/v1/mcp/resources             # 获取 MCP 资源

# ─── 账号 ──────────────────────────────────────────────────

POST   /api/v1/auth/login                # 登录
POST   /api/v1/auth/register             # 注册
POST   /api/v1/auth/refresh              # 刷新 Token
GET    /api/v1/auth/accounts             # 已登录账号列表
POST   /api/v1/auth/switch              # 切换账号
DELETE /api/v1/auth/accounts/:id         # 登出账号
```

---

## 5. UI/UX 设计

### 5.1 页面结构树

```
/                     → 着陆页 (最近笔记 / 今日概览)
├── /notes            → 笔记列表 (数据库视图)
│   └── /notes/:id    → 笔记详情 (编辑器)
├── /graph            → 知识图谱
├── /search           → 搜索中心
├── /collections      → 收藏/数据库视图管理
├── /capture          → 采集记录 (视频/图片)
├── /agents           → 智能体管理
│   └── /agents/:id   → 智能体详情/配置
├── /types            → 对象类型管理
├── /templates        → 模板市场
├── /settings         → 设置
│   ├── /accounts     → 多账号管理
│   ├── /sync         → 同步配置
│   ├── /ai           → AI 设置 (模型选择/限额)
│   └── /plugins      → 插件管理
└── /trash            → 回收站
```

### 5.2 核心组件设计

#### 5.2.1 编辑器 (Editor)

```
┌─────────────────────────────────────────────────────┐
│  /笔记标题 (可编辑 / 自动 AI 生成)                    │
├─────────────────────────────────────────────────────┤
│  / 工具栏                                            │
│  [Text] [H1] [H2] [B] [I] [Code] [Table] [Image]   │
│  [Mermaid] [Todo] [Quote] [Canvas] [/AI]            │
├─────────────────────────────────────────────────────┤
│                                                      │
│  正文内容 (ProseMirror + TipTap)                     │
│                                                      │
│  # 一级标题                                          │
│                                                      │
│  这是一个段落正文，支持 **加粗**、*斜体*、`代码`      │
│                                                      │
│  [[双向链接]] 自动补全 → 弹出关联笔记列表              │
│                                                      │
│  > 引用块                                            │
│                                                      │
│  - [ ] 待办事项                                      │
│                                                      │
│  ![](图片) → 点击进入图片编辑 (裁剪/OCR/增强)         │
│                                                      │
│  ```mermaid                                         │
│  graph LR                                            │
│  A-->B                                              │
│  ```                                                │
│                                                      │
│  / 输入 / → 弹出命令菜单:                             │
│  /AI 总结  /AI 改写  /AI 翻译  /AI 提取待办          │
│  /日期  /模板  /文件  /表情                           │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**编辑器技术选型**：

| 组件 | 选型 | 理由 |
|------|------|------|
| 编辑器框架 | ProseMirror + TipTap | 可扩展性强、Block 级操作 |
| 渲染 | React TipTap wrapper | 与 React 生态无缝集成 |
| 拖拽 | dnd-kit | 轻量、可访问性好 |
| Markdown 解析 | remark + rehype | 生态成熟 |
| 代码高亮 | Shiki | 支持 200+ 语言 |
| 数学公式 | KaTeX | 快速、自渲染 |
| 画布 | tldraw / Excalidraw | 白板/画布模式 |
| 思维导图 | Markmap | 从 Markdown 自动生成 |

#### 5.2.2 知识图谱

```
┌─────────────────────────────────────────────────────┐
│  知识图谱 · 共 342 节点  筛选: [标签] [类型] [时间]  │
├─────────────────────────────────────────────────────┤
│                                                      │
│               🔵 AI Agent                            │
│              /        \                              │
│             /          \                             │
│   🟢 RAG  ──────── 🟣 Embedding                    │
│      |                  |                            │
│      |                  |                            │
│   🟡 VectorDB ───── 🟠 LlamaIndex                   │
│        \              /                              │
│         \            /                               │
│          🟣 LLM  🟢 知识管理-PRD                    │
│             \      /                                 │
│              \    /                                   │
│            🔵 个人知识库                              │
│                                                      │
│  [缩放: 75%]  [布局: 力导向/径向/分层] [3D 切换]    │
│                                                      │
│  选中节点: 🟣 Embedding                              │
│  ┌──────────────────────────────────────────────┐   │
│  │ 关联节点: AI Agent(强)  RAG(强)  VectorDB(中)│   │
│  │ 路径: Embedding → VectorDB → 个人知识库      │   │
│  │ 盲点检测: 缺少"语义缓存"相关节点                │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

**图谱技术选型**：

| 组件 | 选型 | 理由 |
|------|------|------|
| 图谱引擎 | vis-network (2D) | 性能好、交互丰富 |
| 3D 引擎 | Three.js + ForceGraph3D | 大规模节点浏览 |
| 布局算法 | d3-force | 力导向布局 |
| 聚类 | Leiden 算法 | 社区检测 |
| 盲点检测 | LLM + 图分析 | 识别结构空洞 |
| 关系路径 | Dijkstra / BFS | 最短路径分析 |

#### 5.2.3 智能体管理面板

```
┌─────────────────────────────────────────────────────┐
│  智能体                              [+ 创建智能体] │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ 🔄 关联智能体                    ● 运行中     │   │
│  │ 上次运行: 2 分钟前  触发: 新笔记创建          │   │
│  │ 今日处理: 23 篇  发现关联: 47 条             │   │
│  ├─────────────────────────────────────────────┤   │
│  │ 最近活动                                      │   │
│  │ 12:30  发现 "RAG" ↔ "向量数据库" 关联 ✓     │   │
│  │ 12:28  处理 "知识管理PRD.md" ✓               │   │
│  │ 12:25  发现知识盲区: "语义缓存" 无对应笔记 ⚠  │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ 📅 定时报告智能体                ● 运行中     │   │
│  │ 下次执行: 明日 09:00  频率: 每日             │   │
│  │ 上次报告: 今日 09:00 · 12 页知识更新         │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ 🧠 长周期追踪: AI 芯片行业        ● 运行中   │   │
│  │ 进度: ████████░░ 80%  · 已运行 12 天        │   │
│  │ 已收集: 34 篇  · 已生成 3 篇报告            │   │
│  │ 下次汇报: 后天 · 阶段 4/5 分析              │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

### 5.3 关键交互规范

#### 5.3.1 视频截屏 (移动端)

```
┌──────────────────────────────────────────────────┐
│   触发方式                                        │
│                                                   │
│  悬浮按钮 ──── 点击 ──── 振动反馈 + 动画          │
│    ✚                    │                         │
│  (半透明, 可拖拽)        │                         │
│                          ▼                         │
│  ┌──────────────────────────────────────────┐     │
│  │ ✓ 已保存到知识库                          │     │
│  │ 来源: 哔哩哔哩 · 12:34                   │     │
│  │ [查看] [继续观看] [忽略]                  │     │
│  └──────────────────────────────────────────┘     │
│                          │                         │
│                          ▼                         │
│  后台处理中... (OCR + AI 标签 + 关联)              │
│                          │                         │
│                          ▼                         │
│  通知: "处理完成 · 已生成 3 个标签"                │
│                                                   │
│   交互原则:                                       │
│   - 截图后 1 秒内弹出确认浮窗                      │
│   - 不打断当前视频播放                             │
│   - 支持批量模式下不弹确认直接收集                  │
└──────────────────────────────────────────────────┘
```

#### 5.3.2 多账号切换

```
┌──────────────────────────────────────────────────┐
│  侧边栏底部 ──── 点击头像 ──── 弹出账号面板        │
│                                                   │
│  ┌──────────────────────────────────────┐         │
│  │  当前: 👤 工作账号                    │         │
│  │  ─────────────────────               │         │
│  │  👤 个人账号                          │         │
│  │     (点击切换)                        │         │
│  │  👤 团队账号 · 3 条未读               │         │
│  │     (点击切换)                        │         │
│  │  ─────────────────────               │         │
│  │  [+ 添加新账号]                       │         │
│  │  [⚙ 管理账号]                        │         │
│  └──────────────────────────────────────┘         │
│                                                   │
│  切换后:                                         │
│  1. 清空当前工作区状态                            │
│  2. 加载目标账号的 Space 列表                     │
│  3. 重建侧边栏导航树                              │
│  4. 显示目标账号的最近笔记                        │
│  5. 切换过程 < 500ms (本地优先)                   │
│                                                   │
│  桌面端: 多开窗口 (每账号独立窗口)                 │
│  移动端: 应用内切换                               │
└──────────────────────────────────────────────────┘
```

#### 5.3.3 AI 自动整理

```
┌──────────────────────────────────────────────────┐
│  用户: 在草稿笔记中点击 [AI 整理]                  │
│                                                   │
│  ┌──────────────────────────────────────────┐     │
│  │  AI 整理                                  │     │
│  │                                           │     │
│  │  ☑ 清理格式 → 统一 Markdown 风格          │     │
│  │  ☑ 提取标题 → "关于 RAG 系统的笔记"        │     │
│  │  ☑ 推荐标签 → AI, RAG, 向量数据库          │     │
│  │  ☑ 检测关联 → 找到 3 篇相关笔记            │     │
│  │  ☑ 提取待办 → 2 个待办已识别               │     │
│  │  ☑ 建议分类 → 移动到"技术/AI"笔记本        │     │
│  │                                           │     │
│  │  [一键执行]  [逐项确认]  [取消]             │     │
│  └──────────────────────────────────────────┘     │
│                                                   │
│  执行后:                                         │
│  1. 笔记内容被清洗重构                            │
│  2. 自动生成标题和标签                            │
│  3. 建立到关联笔记的双向链接                      │
│  4. 待办自动提取到今日待办列表                    │
│  5. Toast 提示 "已整理完成"                      │
└──────────────────────────────────────────────────┘
```

---

## 6. 安全设计

### 6.1 数据加密

| 数据类型 | 传输加密 | 存储加密 | 说明 |
|---------|---------|---------|------|
| 笔记内容 | TLS 1.3 | AES-256-GCM / 明文可选 | 本地存储可明文(性能)，云端必加密 |
| 图片/附件 | TLS 1.3 | AES-256 (S3 SSE) | 对象存储级加密 |
| 向量嵌入 | TLS 1.3 | AES-256-GCM | 敏感信息可能嵌入向量中 |
| 账号凭据 | TLS 1.3 | bcrypt + salt | 密码哈希存储 |
| 同步数据 | TLS 1.3 | CRDT 数据可选 E2EE | E2EE 模式下服务端无法读取 |
| 本地数据库 | — | SQLite Encryption Extension (可选) | 用户可选择加密 |

### 6.2 MCP 连接安全

```go
// MCPAuthConfig MCP 连接认证配置
type MCPAuthConfig struct {
    AuthType string   `json:"authType"`            // oauth | api_key | none
    Scopes   []string `json:"scopes"`              // search:read | object:read | object:write | agent:trigger | admin
    RateLimit struct {
        RequestsPerMinute int `json:"requestsPerMinute"`
        TokensPerMinute   int `json:"tokensPerMinute"`
    } `json:"rateLimit"`
    AuditLog  bool     `json:"auditLog"`
    AllowedIPs []string `json:"allowedIPs,omitempty"`   // IP 白名单 (自托管)
}
```

---

## 7. 部署架构

### 7.1 云服务部署

```
                      ┌─────────────┐
                      │  Cloudflare │
                      │   CDN + DNS │
                      └──────┬──────┘
                             │
                      ┌──────▼──────┐
                      │  Nginx /    │
                      │  Traefik    │
                      └──────┬──────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
       ┌──────▼──────┐ ┌────▼────┐ ┌──────▼──────┐
       │  Web App    │ │  WS      │ │  API        │
       │  (Vercel/   │ │  Gateway │ │  (K8s)      │
       │   Docker)   │ │         │ │              │
       └─────────────┘ └─────────┘ └──────┬───────┘
                                          │
              ┌──────────────┬────────────┼──────────────┐
              │              │            │              │
       ┌──────▼──────┐ ┌────▼────┐ ┌─────▼──────┐ ┌─────▼──────┐
       │ PostgreSQL  │ │  Redis  │ │  Qdrant    │ │  NATS      │
       │ + pgvector  │ │ Cache   │ │  VectorDB  │ │  Message   │
       └─────────────┘ └─────────┘ └────────────┘ │  Queue     │
                                                   └────────────┘
       ┌─────────────┐ ┌──────────┐ ┌───────────────────────┐
       │  MinIO S3   │ │  AI      │ │  OCR Worker           │
       │  Attachments│ │  Gateway │ │  (GPU Pod)            │
       └─────────────┘ └──────────┘ └───────────────────────┘
```

### 7.2 本地优先架构

```
┌─────────────────────────────────────────────────────────┐
│                    客户端设备                             │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │                主进程 (UI + Logic)                 │   │
│  ├──────────────────────────────────────────────────┤   │
│  │  SQLite DB    │  Markdown 文件  │ 向量索引(LanceDB)│   │
│  ├──────────────────────────────────────────────────┤   │
│  │              Sync Engine (Go-Yrs CRDT)             │   │
│  │     ┌─────────────┐    ┌────────────────────┐    │   │
│  │     │ Online Flow │    │ Offline Queue      │    │   │
│  │     │ WebSocket   │    │ IndexedDB / File   │    │   │
│  │     └─────────────┘    └────────────────────┘    │   │
│  ├──────────────────────────────────────────────────┤   │
│  │  Local AI: Ollama / Whisper.cpp / PaddleOCR      │   │
│  │  (可选, 按需下载模型)                              │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  离线时 → 全部本地读写                                    │
│  在线时 → 本地 + 云端双向同步                             │
└──────────────────────────────────────────────────────────┘
```

---

## 8. 性能指标

| 指标 | 目标 | 测量方式 |
|------|------|---------|
| 编辑器输入延迟 | < 50ms | RAIL 模型 |
| 全文搜索响应 | < 200ms (本地) / < 500ms (云端) | P95 分位 |
| 语义搜索响应 | < 500ms (本地) / < 1s (云端) | P95 分位 |
| 知识图谱渲染 | < 1s (1000 节点) | 首次渲染 |
| 图谱交互帧率 | > 30fps (拖拽/缩放) | Chrome DevTools |
| 图片 OCR 处理 | < 3s (单张) | 端到端 |
| 同步延迟 | < 2s (同设备同网络) | 变更→确认 |
| 首次加载 | < 2s (Web) / < 1s (桌面端) | Lighthouse |
| 本地存储上限 | > 10 万篇笔记 | SQLite 基准测试 |
| 并发编辑 | 10 人同时编辑一篇笔记无冲突 | CRDT 测试 |
| 电量消耗 (移动端) | 后台监听 < 1%/h | Android Battery Historian |

### 8.1 AI 服务成本控制策略

| 策略 | 说明 | 实现方式 |
|------|------|---------|
| **本地模型优先** | OCR、Embedding 优先本地推理，减少 API 调用 | PaddleOCR（本地）、BGE-M3（本地 ONNX）|
| **分级模型** | 简单任务用小模型，复杂任务用大模型 | LiteLLM 网关路由：标签/摘要用 Haiku/Flash，深度分析用 Opus/Pro |
| **缓存命中** | 相似图片/文档命中缓存跳过 AI 处理 | ai_cache 表 + 向量相似度去重 |
| **Token 预算** | 每用户每日限额，不同付费等级差异化 | 读取 subscriptions.ai_tokens_limit |
| **批量处理** | 延迟非实时任务到低峰期批量执行 | vision_jobs 队列 + NATS 延迟消费 |
| **去重过滤** | 低质量/重复内容不入库，节省处理成本 | image_queue 的 pHash 去重 + quality_score 过滤 |

---

## 9. 错误处理规范

### 9.1 错误码体系

```go
// ErrorCode 错误码
type ErrorCode int

const (
    // 通用 (1xxx)
    ErrUnknown           ErrorCode = 1000
    ErrInternal          ErrorCode = 1001
    ErrServiceUnavail    ErrorCode = 1002
    ErrRateLimited       ErrorCode = 1003

    // 认证 (2xxx)
    ErrUnauthorized      ErrorCode = 2000
    ErrTokenExpired      ErrorCode = 2001
    ErrInvalidCreds      ErrorCode = 2002
    ErrMFARequired       ErrorCode = 2003
    ErrAccountLocked     ErrorCode = 2004

    // 权限 (3xxx)
    ErrForbidden         ErrorCode = 3000
    ErrSpaceFull         ErrorCode = 3001
    ErrExportLimit       ErrorCode = 3002

    // 资源 (4xxx)
    ErrNotFound          ErrorCode = 4000
    ErrConflict          ErrorCode = 4001
    ErrVersionConflict   ErrorCode = 4002 // CRDT 冲突
    ErrDuplicateEntry    ErrorCode = 4003

    // 验证 (5xxx)
    ErrValidation        ErrorCode = 5000
    ErrInvalidInput      ErrorCode = 5001
    ErrPayloadTooLarge   ErrorCode = 5002
    ErrUnsupportedFormat ErrorCode = 5003

    // AI (6xxx)
    ErrAIQuotaExceeded   ErrorCode = 6000
    ErrAITimeout         ErrorCode = 6001
    ErrAIContentFiltered ErrorCode = 6002
    ErrAIModelUnavail    ErrorCode = 6003

    // 同步 (7xxx)
    ErrSyncConflict      ErrorCode = 7000
    ErrSyncStaleData     ErrorCode = 7001
    ErrSyncOffline       ErrorCode = 7002

    // 图片处理 (8xxx)
    ErrOCRFailed         ErrorCode = 8000
    ErrImageTooLarge     ErrorCode = 8001
    ErrLowImageQuality   ErrorCode = 8002
    ErrUnsupportedImage  ErrorCode = 8003
)

// APIError API 错误响应
type APIError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details any       `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return e.Message
}
```

### 9.2 降级策略

| 场景 | 降级行为 | 用户体验 |
|------|---------|---------|
| 云端 AI 不可用 | 降级到本地小模型 | 功能可用，精度略降 |
| 网络断开 | 切换到完全离线模式 | 除同步外所有功能正常 |
| OCR 服务过载 | 加入队列，回调通知 | 提示"处理中，稍后通知" |
| 同步冲突 | CRDT 自动合并 | 无感 |
| 搜索超时 | 降级到基础关键词搜索 | 搜索结果略少，有提示 |
| 图片处理超时 | 简化处理流程 | 文字仍提取，版面分析跳过 |

---

## 10. 测试策略

| 层级 | 工具 | 覆盖目标 |
|------|------|---------|
| 单元测试 | `go test` (stdlib) + testify | 核心逻辑 90%+ |
| 组件测试 | Testing Library + Playwright | 关键 UI 组件 |
| 集成测试 | testcontainers-go (PG/SQLite) + Playwright | Repository/Service 层 + UI 流程 |
| CRDT 测试 | Go-Yrs 测试套件 | 冲突合并正确性 |
| 同步测试 | 自定义 Go 多客户端模拟 (goroutine) | 离线/在线/多设备 |
| 性能测试 | k6 (API) / Go pprof (CPU/内存) | API < 200ms P95 |
| AI 评测 | 自定义 Go 评估框架 | OCR 准确率 > 95% |
| 安全测试 | golangci-lint (gosec) + OWASP ZAP | 无高危漏洞 |
| Fuzz 测试 | `go test -fuzz` | 输入边缘用例 |

---

## 11. 项目结构

```
nextm/
├── cmd/                       # Go 应用入口
│   ├── server/                # 主 API 服务器 (Chi Router)
│   │   └── main.go
│   ├── sync/                  # 同步服务 (WebSocket)
│   │   └── main.go
│   ├── worker/                # 后台任务 Worker (Asynq)
│   │   └── main.go
│   ├── migrator/              # 数据库迁移工具
│   │   └── main.go
│   └── cli/                   # CLI 工具 (Cobra)
│       └── main.go
├── internal/                  # Go 私有包
│   ├── api/                   # HTTP/gRPC 层
│   │   ├── handler/           # Request handlers
│   │   ├── middleware/        # Auth, CORS, 日志, 限流
│   │   ├── router/            # Chi 路由定义
│   │   └── dto/               # 数据传输对象
│   ├── service/               # 业务逻辑层
│   │   ├── object/            # 知识对象 CRUD
│   │   ├── collection/        # 数据库视图引擎
│   │   ├── search/            # 全文/语义搜索
│   │   ├── vision/            # OCR Pipeline 编排
│   │   ├── capture/           # 截图采集服务
│   │   ├── agent/             # 多智能体编排
│   │   │   ├── orchestrator/  # DAG 编排引擎
│   │   │   ├── messagebus/    # NATS 消息总线
│   │   │   └── longhorizon/   # 长周期智能体
│   │   ├── relation/          # 知识图谱关系
│   │   ├── sync/              # CRDT 同步引擎
│   │   ├── mcp/               # MCP 协议实现
│   │   ├── auth/              # 认证 & 多账号
│   │   ├── ai/                # AI 网关 (LiteLLM 集成)
│   │   └── export/            # 数据导出
│   ├── repository/            # 数据访问层 (SQLc 生成)
│   │   ├── db/                # SQL 查询 + 迁移
│   │   │   ├── migrations/    # golang-migrate SQL
│   │   │   ├── queries/       # SQLc 查询文件
│   │   │   ├── sqlc.yaml      # SQLc 配置
│   │   │   ├── models.go      # SQLc 生成模型
│   │   │   └── sqlite/        # SQLite 方言查询
│   │   │   └── postgres/      # PG 方言查询
│   │   └── cache/             # Redis 缓存
│   ├── model/                 # 领域模型
│   ├── eventbus/              # NATS 事件总线
│   ├── crdt/                  # CRDT 合并引擎 (Go-Yrs FFI)
│   ├── vector/                # LanceDB Go SDK 向量搜索
│   ├── config/                # Viper 配置管理
│   ├── telemetry/             # OpenTelemetry 追踪/指标
│   └── pkg/                   # 共享工具
│       ├── logger/            # zap/slog 结构化日志
│       ├── idgen/             # UUID v4 生成
│       ├── crypto/            # AES-256-GCM 加密
│       ├── validator/         # 输入校验
│       └── httputil/          # HTTP 辅助函数
├── frontend/                  # 客户端应用
│   ├── web/                   # Web 应用 (React + Vite + Tailwind)
│   │   ├── src/
│   │   │   ├── components/   # 共享组件
│   │   │   ├── pages/        # 页面
│   │   │   ├── hooks/        # 自定义 Hooks
│   │   │   ├── stores/       # Zustand stores
│   │   │   └── services/     # API 客户端
│   │   └── ...
│   ├── desktop/               # 桌面端 (Tauri)
│   │   ├── src-tauri/         # Rust 壳
│   │   │   └── src/
│   │   │       └── main.rs    # Tauri 入口
│   │   └── src/               # 前端 (React)
│   ├── mobile/                # 移动端 (Flutter)
│   │   ├── lib/
│   │   │   ├── screens/
│   │   │   ├── widgets/
│   │   │   └── services/
│   │   └── ...
│   └── extension/             # 浏览器插件
│       └── ...
├── infra/                     # 部署基础设施
│   ├── k8s/                   # Kubernetes 清单
│   ├── docker/                # Docker Compose
│   │   ├── Dockerfile         # Go 多阶段构建
│   │   └── docker-compose.yml
│   └── terraform/             # IaC
├── docs/                      # 文档
│   ├── api/                   # OpenAPI 规范
│   ├── architecture/          # 架构文档
│   └── guides/                # 开发指南
├── scripts/                   # 构建 & 开发脚本
│   ├── build.sh               # Go 交叉编译
│   └── dev.sh                 # 本地开发环境
├── go.mod                     # Go 模块定义
├── go.sum
├── Taskfile.yml               # Task 任务运行器 (替代 Makefile)
└── golangci.yml               # golangci-lint 配置
```
