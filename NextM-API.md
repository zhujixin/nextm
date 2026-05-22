# NextM — OpenAPI 规范

**版本**: V1.0
**状态**: 初稿
**最后更新**: 2026-05-22

**OpenAPI**: 3.0.3
**基础路径**: `/api/v1`
**格式**: JSON
**认证**: Bearer JWT

---

## 1. 通用规范

### 1.1 认证方式

```
Authorization: Bearer <access_token>
```

### 1.2 通用请求头

| Header | 必填 | 说明 |
|--------|------|------|
| `Authorization` | 是 | Bearer JWT access token |
| `X-Space-ID` | 是 | 当前工作区 ID |
| `Content-Type` | 是 | `application/json` |
| `X-Request-ID` | 否 | 幂等键（创建操作建议提供） |

### 1.3 通用响应格式

**成功响应**：
```json
{
    "data": { ... },
    "meta": {
        "request_id": "uuid"
    }
}
```

**分页响应**：
```json
{
    "data": [ ... ],
    "meta": {
        "total": 100,
        "limit": 20,
        "offset": 0,
        "has_more": true
    }
}
```

**错误响应**：
```json
{
    "error": {
        "code": 4001,
        "message": "资源冲突",
        "details": {}
    }
}
```

### 1.4 通用错误码

| Code | HTTP Status | 说明 |
|------|-------------|------|
| 1000 | 500 | 未知错误 |
| 1001 | 500 | 内部错误 |
| 1002 | 503 | 服务不可用 |
| 1003 | 429 | 请求限流 |
| 2000 | 401 | 未认证 |
| 2001 | 401 | Token 过期 |
| 2002 | 401 | 凭据无效 |
| 3000 | 403 | 权限不足 |
| 4000 | 404 | 资源未找到 |
| 4001 | 409 | 资源冲突 |
| 5000 | 400 | 验证错误 |
| 5001 | 400 | 无效输入 |
| 5002 | 413 | 请求体过大 |

---

## 2. Schema 定义

### 2.1 知识对象 (Object)

```yaml
Object:
  type: object
  properties:
    id:
      type: string
      format: uuid
      description: 对象唯一标识
    title:
      type: string
      description: 标题 (AI 自动生成或手动)
    type_id:
      type: string
      format: uuid
      description: 对象类型 ID
    properties:
      type: object
      description: 动态属性 KV (根据类型定义)
      example: {"author": "张三", "status": "进行中"}
    tags:
      type: array
      items:
        type: string
      description: 标签 ID 数组
    source:
      type: string
      enum: [manual, video, web, camera, audio, import, clipboard, email, agent]
    created_at:
      type: integer
      description: Unix ms
    updated_at:
      type: integer
      description: Unix ms
  required: [title, type_id]
```

### 2.2 内容块 (Block)

```yaml
Block:
  type: object
  properties:
    id:
      type: string
      format: uuid
    object_id:
      type: string
      format: uuid
    parent_id:
      type: string
      format: uuid
      nullable: true
    type:
      type: string
      enum: [text, heading1, heading2, heading3, image, code, table, mermaid,
             todo, bullet_list, numbered_list, quote, divider, file, embed,
             callout, toggle, equation, canvas, block_reference]
    content:
      type: string
      description: Markdown / JSON / 纯文本
    properties:
      type: object
    position:
      type: number
      format: float
      description: 排序位置 (相邻平均法)
    depth:
      type: integer
      default: 0
  required: [type, content]
```

### 2.3 对象类型 (ObjectType)

```yaml
ObjectType:
  type: object
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
      example: "书籍"
    icon:
      type: string
      example: "📚"
    color:
      type: string
      example: "#6366f1"
    description:
      type: string
    fields:
      type: array
      items:
        $ref: '#/components/schemas/FieldDefinition'
    is_builtin:
      type: boolean
```

```yaml
FieldDefinition:
  type: object
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
      example: "作者"
    field_type:
      type: string
      enum: [text, number, date, select, multi_select, relation, rollup,
             formula, file, email, url, phone, progress, currency, rating]
    required:
      type: boolean
    options:
      type: array
      items:
        type: object
        properties:
          id: {type: string}
          name: {type: string}
          color: {type: string}
```

### 2.4 集合视图 (CollectionView)

```yaml
CollectionView:
  type: object
  properties:
    id:
      type: string
    name:
      type: string
    type:
      type: string
      enum: [table, kanban, gallery, calendar, timeline, list]
    filters:
      type: array
      items:
        $ref: '#/components/schemas/FilterGroup'
    sorts:
      type: array
      items:
        type: object
        properties:
          field_id: {type: string}
          direction: {type: string, enum: [asc, desc]}
    visible_fields:
      type: array
      items:
        type: string
    group_by:
      type: string
```

```yaml
FilterGroup:
  type: object
  properties:
    id: {type: string}
    operator: {type: string, enum: [and, or]}
    conditions:
      type: array
      items:
        $ref: '#/components/schemas/FilterCondition'

FilterCondition:
  type: object
  properties:
    field_id: {type: string}
    operator:
      type: string
      enum: [eq, neq, gt, gte, lt, lte, contains, startsWith,
             isEmpty, isNotEmpty, in, notIn]
    value: {}
```

### 2.5 搜索 (Search)

```yaml
SearchResult:
  type: object
  properties:
    items:
      type: array
      items:
        type: object
        properties:
          id: {type: string}
          title: {type: string}
          type_id: {type: string}
          snippet:
            type: string
            description: 匹配上下文片段
          score:
            type: number
            description: 相关度分数 [0,1]
          updated_at: {type: integer}
    total: {type: integer}
    suggestion: {type: string}
    related_queries:
      type: array
      items: {type: string}
```

### 2.6 标签/关系

```yaml
Tag:
  type: object
  properties:
    id: {type: string}
    name: {type: string}
    color: {type: string}
    parent_id: {type: string, nullable: true}
    ai_generated: {type: boolean}
    object_count: {type: integer}

Relation:
  type: object
  properties:
    id: {type: string}
    source_id: {type: string}
    target_id: {type: string}
    type:
      type: string
      enum: [link, reference, citation, parent, child, related, custom]
    weight: {type: number, format: float}
    ai_generated: {type: boolean}
```

### 2.7 智能体 (Agent)

```yaml
Agent:
  type: object
  properties:
    id: {type: string}
    name: {type: string}
    agent_type:
      type: string
      enum: [summarizer, tagger, linker, review, writer, collector,
             refactor, orchestrator, scheduler, custom]
    triggers:
      type: array
      items:
        type: object
        properties:
          type:
            type: string
            enum: [onCreate, onUpdate, onDelete, onSchedule, onEvent, onManual, onWebhook]
          config: {type: object}
    enabled: {type: boolean}
    last_run_at: {type: integer}
  required: [name, agent_type]

AgentRun:
  type: object
  properties:
    id: {type: string}
    agent_id: {type: string}
    status:
      type: string
      enum: [pending, running, completed, failed, cancelled, timeout]
    input: {type: object}
    output: {type: object}
    error: {type: string}
    tokens_used: {type: integer}
    processing_time: {type: integer}
    started_at: {type: integer}
    completed_at: {type: integer}
```

### 2.8 认证 (Auth)

```yaml
LoginRequest:
  type: object
  required: [email, password]
  properties:
    email: {type: string, format: email}
    password: {type: string, minLength: 8}

LoginResponse:
  type: object
  properties:
    access_token: {type: string}
    refresh_token: {type: string}
    expires_in: {type: integer}
    user:
      type: object
      properties:
        id: {type: string}
        name: {type: string}
        email: {type: string}
        avatar_url: {type: string}

RefreshRequest:
  type: object
  properties:
    refresh_token: {type: string}
```

### 2.9 采集 (Capture)

```yaml
ScreenshotSubmit:
  type: object
  required: [image_data]
  properties:
    image_data:
      type: string
      format: byte
      description: Base64 编码的图片
    source:
      type: string
      enum: [video, camera, screenshot, gallery, clipboard]
    source_app:
      type: string
      description: 来源 App 名称
    source_url:
      type: string
      description: 视频/文章 URL
    playback_position:
      type: integer
      description: 视频播放位置 (秒)

ScreenshotResult:
  type: object
  properties:
    screenshot_id: {type: string}
    status:
      type: string
      enum: [pending, processing, completed, duplicate, low_quality]
    object_id: {type: string}
    ocr_text: {type: string}
    tags: {type: array, items: {type: string}}
```

### 2.10 图片处理 (Vision)

```yaml
VisionRequest:
  type: object
  required: [image_data]
  properties:
    image_data:
      type: string
      format: byte
    format:
      type: string
      enum: [png, jpg, webp, heic]
      default: png
    options:
      type: object
      properties:
        enhance: {type: boolean}
        extract_tables: {type: boolean}
        extract_formula: {type: boolean}
        languages:
          type: array
          items: {type: string}
          example: ["zh", "en"]

ExtractionResult:
  type: object
  properties:
    raw_text: {type: string}
    layout:
      type: array
      items:
        type: object
        properties:
          type: {type: string, enum: [heading, paragraph, table, list, figure]}
          text: {type: string}
          bbox: {type: array, items: {type: number}}
    tables:
      type: array
      items:
        type: object
        properties:
          headers: {type: array, items: {type: string}}
          rows:
            type: array
            items: {type: array, items: {type: string}}
    entities:
      type: array
      items:
        type: object
        properties:
          type: {type: string, enum: [person, date, location, organization, term]}
          text: {type: string}
    quality:
      type: object
      properties:
        blur_score: {type: number}
        text_coverage: {type: number}
        overall_score: {type: number}
```

---

## 3. API 路径

### 3.1 认证

```
#### 登录
POST /api/v1/auth/login
  Content-Type: application/json
  Body: LoginRequest
  Response 200: LoginResponse
  Response 401: {error: {code: 2002, message: "邮箱或密码错误"}}

#### 注册
POST /api/v1/auth/register
  Content-Type: application/json
  Body:
    email: string
    password: string
    name: string
  Response 201: LoginResponse
  Response 409: {error: {code: 4003, message: "邮箱已被注册"}}

#### 刷新 Token
POST /api/v1/auth/refresh
  Content-Type: application/json
  Body: RefreshRequest
  Response 200: LoginResponse

#### 获取已登录账号列表
GET /api/v1/auth/accounts
  Authorization: Bearer <token>
  Response 200:
    data:
      - id: string
        name: string
        email: string
        avatar_url: string
        is_current: boolean

#### 切换账号
POST /api/v1/auth/switch
  Authorization: Bearer <token>
  Content-Type: application/json
  Body:
    account_id: string
  Response 200:
    access_token: string
    refresh_token: string

#### 登出
DELETE /api/v1/auth/accounts/:id
  Authorization: Bearer <token>
  Response 204: No Content
```

### 3.2 知识对象 CRUD

```
#### 获取对象列表
GET /api/v1/objects
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Query Parameters:
    type_id: string          # 按类型筛选
    tag_ids: string          # 按标签筛选 (逗号分隔)
    source: string           # 按来源筛选
    q: string                # 关键词搜索
    sort: string             # updated_at | created_at | title (默认 updated_at)
    order: string            # asc | desc (默认 desc)
    limit: integer           # 默认 20, 最大 100
    offset: integer          # 默认 0
  Response 200:
    data: [Object]
    meta: { total, limit, offset, has_more }

#### 创建对象
POST /api/v1/objects
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  X-Request-ID: <idempotency_key>
  Content-Type: application/json
  Body:
    title: string
    type_id: string (uuid)
    properties: object (optional)
    tags: string[] (optional)
    source: string (optional)
    blocks: Block[] (optional)
  Response 201: Object
  Response 422: {error: {code: 5000, message: "验证失败"}}

#### 获取对象详情
GET /api/v1/objects/:id
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Response 200:
    data: Object + blocks + relations
  Response 404: {error: {code: 4000, message: "对象不存在"}}

#### 更新对象
PUT /api/v1/objects/:id
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Content-Type: application/json
  Body:
    title: string (optional)
    properties: object (optional)
    tags: string[] (optional)
  Response 200: Object
  Response 409: {error: {code: 4001, message: "版本冲突"}}

#### 删除对象 (软删除)
DELETE /api/v1/objects/:id
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Response 204: No Content

#### 归档/取消归档
PATCH /api/v1/objects/:id/archive
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Body:
    archived: boolean
  Response 200: Object
```

### 3.3 内容块管理

```
#### 获取块列表
GET /api/v1/objects/:id/blocks
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Response 200:
    data: [Block]  # 按 position 排序

#### 添加块
POST /api/v1/objects/:id/blocks
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Body: Block (without id)
  Response 201: Block

#### 更新块
PUT /api/v1/blocks/:id
  Authorization: Bearer <token>
  Body:
    content: string
    type: string (optional)
    properties: object (optional)
  Response 200: Block

#### 删除块
DELETE /api/v1/blocks/:id
  Authorization: Bearer <token>
  Response 204

#### 批量重排
PATCH /api/v1/blocks/reorder
  Authorization: Bearer <token>
  Body:
    object_id: string
    blocks:
      - id: string
        position: number
        parent_id: string (optional)
  Response 200: {data: [Block]}
```

### 3.4 对象类型

```
#### 获取类型列表
GET /api/v1/types
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Response 200:
    data: [ObjectType]

#### 创建自定义类型
POST /api/v1/types
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Body:
    name: string
    icon: string (optional)
    color: string (optional)
    fields: [FieldDefinition]
  Response 201: ObjectType

#### 更新类型
PUT /api/v1/types/:id
  Authorization: Bearer <token>
  Response 200: ObjectType

#### 删除类型
DELETE /api/v1/types/:id
  Authorization: Bearer <token>
  Response 204
  Note: 内置类型 (is_builtin=true) 不可删除
```

### 3.5 数据库视图 (Collection)

```
#### 获取集合列表
GET /api/v1/collections
  Response 200: [CollectionView]

#### 创建集合
POST /api/v1/collections
  Body:
    name: string
    type: string (默认 table)
    source_type: type | tag | mixed | manual
    source_config: object
  Response 201: CollectionView

#### 更新集合
PUT /api/v1/collections/:id
  Response 200: CollectionView

#### 删除集合
DELETE /api/v1/collections/:id
  Response 204
```

### 3.6 搜索

```
#### 全文搜索
GET /api/v1/search
  Authorization: Bearer <token>
  X-Space-ID: <space_id>
  Query Parameters:
    q: string (required)
    type_ids: string (逗号分隔)
    tag_ids: string (逗号分隔)
    limit: integer (默认 20)
    offset: integer (默认 0)
  Response 200: SearchResult

#### 语义搜索
POST /api/v1/search/semantic
  Authorization: Bearer <token>
  Body:
    query: string
    limit: integer
    filters: [SearchFilter] (optional)
  Response 200: SearchResult

#### 混合搜索 (RRF)
POST /api/v1/search/hybrid
  Authorization: Bearer <token>
  Body:
    query: string
    fulltext_weight: number (default 0.5)
    semantic_weight: number (default 0.5)
    limit: integer
    filters: [SearchFilter] (optional)
  Response 200: SearchResult

#### 深度搜索 (AI)
POST /api/v1/search/deep
  Authorization: Bearer <token>
  Body:
    query: string
    model: string (optional, 默认用户偏好模型)
  Response 200: DeepSearchResult

#### 关联推荐
GET /api/v1/objects/:id/related
  Authorization: Bearer <token>
  Query Parameters:
    limit: integer (默认 5)
  Response 200:
    data:
      - object: Object
        score: number
        reason: string
```

### 3.7 图片处理 (Vision)

```
#### OCR 识别
POST /api/v1/vision/ocr
  Body: VisionRequest (未启用 options.extract_tables/formula)
  Response 200: ExtractionResult

#### 完整提取
POST /api/v1/vision/extract
  Body: VisionRequest
  Response 200: ExtractionResult (含 layout, tables, entities)

#### 批量处理
POST /api/v1/vision/batch
  Body:
    images: [VisionRequest]
  Response 200:
    data: [ExtractionResult]
    meta: { total, completed, failed }

#### 图像增强
POST /api/v1/vision/enhance
  Body:
    image_data: string (base64)
    scale: integer (default 2, enum [2, 4])
  Response 200:
    enhanced_image: string (base64)
    original_size: { width, height }
    enhanced_size: { width, height }
```

### 3.8 截图采集

```
#### 提交截图
POST /api/v1/capture/screenshot
  Body: ScreenshotSubmit
  Response 201: ScreenshotResult

#### 获取队列状态
GET /api/v1/capture/queue
  Response 200:
    total: integer
    pending: integer
    processing: integer
    completed: integer
    failed: integer
    estimated_wait_ms: integer

#### 批量提交
POST /api/v1/capture/batch
  Body:
    screenshots: [ScreenshotSubmit]
  Response 201:
    data: [ScreenshotResult]

#### 采集历史
GET /api/v1/capture/sessions
  Query Parameters:
    limit, offset, source_app
  Response 200:
    data:
      - session_id: string
        source_app: string
        source_url: string
        image_count: integer
        created_at: integer
        status: string
```

### 3.9 智能体

```
#### 获取智能体列表
GET /api/v1/agents
  Response 200: [Agent]

#### 创建智能体
POST /api/v1/agents
  Body:
    name: string
    agent_type: string
    triggers: [AgentTrigger]
    actions: [AgentAction]
    schedule: string (cron, optional)
    model_preference: string (optional)
  Response 201: Agent

#### 手动触发智能体
POST /api/v1/agents/:id/trigger
  Body:
    params: object (optional)
    object_id: string (可选，指定处理对象)
  Response 202: {run_id: string, status: "pending"}

#### 运行历史
GET /api/v1/agents/:id/runs
  Query Parameters:
    limit, offset, status
  Response 200: [AgentRun]

#### 启动长周期任务
POST /api/v1/agents/:id/long-term
  Body:
    goal: string
    deadline: integer (optional)
  Response 201: LongHorizonAgent

#### 执行工作流
POST /api/v1/workflow
  Body:
    steps: [WorkflowStep]
    context: object
  Response 202: {workflow_id: string, status: "running"}
```

### 3.10 关系与图谱

```
#### 查询关系
GET /api/v1/relations
  Query Parameters:
    object_id: string (查询对象的所有关系)
    type: string (关系类型筛选)
    direction: source | target | both (默认 both)
  Response 200: [Relation]

#### 创建关系
POST /api/v1/relations
  Body:
    source_id: string
    target_id: string
    type: string
    weight: number (optional, default 1.0)
  Response 201: Relation

#### 删除关系
DELETE /api/v1/relations/:id
  Response 204

#### 获取知识图谱子图
GET /api/v1/graph/subgraph
  Query Parameters:
    center_id: string (中心节点)
    depth: integer (默认 2, 最大 5)
    type_ids: string (节点类型筛选)
  Response 200:
    nodes:
      - id, title, type_id, type_name, color, size
    edges:
      - source_id, target_id, type, weight
```

### 3.11 MCP 协议

```
#### 获取 MCP 工具列表
GET /api/v1/mcp/tools
  Response 200:
    tools:
      - name: string
        description: string
        input_schema: object

#### 调用 MCP 工具
POST /api/v1/mcp/call
  Body:
    tool_name: string
    arguments: object
  Response 200: {result: any}

#### 获取 MCP 资源
GET /api/v1/mcp/resources
  Response 200: [MCPResourceDefinition]
```

### 3.12 同步

```
#### WebSocket 同步连接
WS /api/v1/sync
  Query Parameters:
    token: string (JWT)
    space_id: string
    device_id: string
  Messages (text JSON):
    客户端 → 服务端:
      - {type: "ping"}
      - {type: "delta", object_id, object_type, data: "base64"}
    服务端 → 客户端:
      - {type: "pong"}
      - {type: "delta", object_id, object_type, data: "base64", version}
      - {type: "conflict", object_id, local_version, remote_version}

#### 推送同步变更
POST /api/v1/sync/push
  Body:
    object_id: string
    object_type: string
    data: string (base64 CRDT binary)
    version: integer
  Response 200: {ack: true, version: integer}

#### 拉取变更
GET /api/v1/sync/pull
  Query Parameters:
    object_type: string
    since_version: integer
    limit: integer (默认 100)
  Response 200:
    changes: [{object_id, object_type, data, version, created_at}]

#### 同步状态
GET /api/v1/sync/status
  Response 200:
    state: connected | syncing | offline
    last_sync_at: integer
    pending_changes: integer
    conflicts: integer
```

---

## 4. 使用示例

### 4.1 完整请求流程

```
#### 1. 注册账号
POST /api/v1/auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "SecurePass123",
    "name": "张三"
}

Response 201:
{
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
        "expires_in": 900,
        "user": {
            "id": "a1b2c3d4-...",
            "name": "张三",
            "email": "user@example.com"
        }
    }
}

#### 2. 创建知识对象
POST /api/v1/objects
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
X-Space-ID: sp-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Content-Type: application/json

{
    "title": "关于 RAG 系统的学习笔记",
    "type_id": "type-book-xxxx",
    "tags": ["tag-ai", "tag-rag"],
    "blocks": [
        {
            "type": "heading1",
            "content": "RAG 系统概述",
            "position": 0
        },
        {
            "type": "text",
            "content": "RAG (Retrieval-Augmented Generation) 是一种...",
            "position": 1
        },
        {
            "type": "todo",
            "content": "阅读 LlamaIndex 官方文档",
            "properties": {"checked": false},
            "position": 2
        }
    ]
}

Response 201:
{
    "data": {
        "id": "obj-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
        "title": "关于 RAG 系统的学习笔记",
        "type_id": "type-book-xxxx",
        "tags": ["tag-ai", "tag-rag"],
        "source": "manual",
        "created_at": 1747929600000,
        "updated_at": 1747929600000
    }
}

#### 3. 搜索
GET /api/v1/search/hybrid
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
X-Space-ID: sp-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Content-Type: application/json

{
    "query": "RAG 向量数据库",
    "limit": 10
}

Response 200:
{
    "data": {
        "items": [
            {
                "id": "obj-xxxxxxxx-...",
                "title": "关于 RAG 系统的学习笔记",
                "snippet": "...RAG (Retrieval-Augmented Generation) 是一种结合<mark>向量数据库</mark>...",
                "score": 0.92,
                "updated_at": 1747929600000
            }
        ],
        "total": 1
    }
}

#### 4. WebSocket 同步
WS /api/v1/sync?token=eyJhbGciOiJIUzI1NiIs...&space_id=sp-xxxx&device_id=dev-xxxx

→ {"type": "ping"}
← {"type": "pong"}
→ {"type": "delta", "object_id": "obj-xxxx", "object_type": "object", "data": "<base64_crdt>"}
← {"type": "delta", "object_id": "obj-xxxx", "object_type": "object", "data": "<base64_crdt>", "version": 5}
```

---

## 5. OpenAPI 3.0 YAML 参考

以下是将上述规范转换为标准 OpenAPI 3.0 YAML 的参考映射：

```yaml
openapi: 3.0.3
info:
  title: NextM API
  version: 1.0.0
  description: NextM 个人知识管理工具 REST API

servers:
  - url: https://api.nextm.app/api/v1
    description: 生产环境
  - url: http://localhost:8080/api/v1
    description: 本地开发

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  parameters:
    SpaceID:
      name: X-Space-ID
      in: header
      required: true
      schema:
        type: string
        format: uuid

  schemas:
    Object:
      $ref: '#/definitions/Object'
    Block:
      $ref: '#/definitions/Block'
    # ... 其余 schema 定义

security:
  - BearerAuth: []

paths:
  /auth/login:
    post:
      tags: [Auth]
      summary: 用户登录
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: 登录成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'

  /objects:
    get:
      tags: [Objects]
      summary: 获取对象列表
      parameters:
        - $ref: '#/components/parameters/SpaceID'
        - name: type_id
          in: query
          schema: {type: string}
        # ... 其余参数
      responses:
        '200':
          description: 对象列表
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Object'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'
```
