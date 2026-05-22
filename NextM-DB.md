# NextM — 数据库设计说明书

**版本**: V1.0
**状态**: 初稿
**最后更新**: 2026-05-22

---

## 1. 设计原则

### 1.1 通用约定

| 约定 | 说明 |
|------|------|
| 命名风格 | 所有表名、字段名使用 `snake_case` |
| 主键 | UUID v4 文本存储 (`TEXT`)，非自增整数，避免同步冲突 |
| 时间戳 | Unix 毫秒时间戳 (`INTEGER`)，统一 UTC，前端负责本地化显示 |
| 软删除 | 核心数据表保留 `is_deleted` 标记，数据移入 `trash` 表 |
| 版本号 | `version` 字段用于乐观锁，+1 自增 |
| JSON 存储 | 不常查询的结构化元数据使用 JSON TEXT 存储；需查询的字段拆为独立列 |
| 空间隔离 | 所有多租户数据通过 `space_id` 字段隔离，索引必须包含 `space_id` |
| 设备标识 | 所有变更记录 `device_id`，用于 CRDT 向量时钟 |

### 1.2 类型映射

| SQLite | PostgreSQL | 用途 |
|--------|-----------|------|
| TEXT | UUID | 主键/外键 (UUID v4) |
| TEXT | VARCHAR(255) | 短文本 |
| TEXT | TEXT | 长文本/JSON |
| INTEGER | BIGINT | 时间戳/计数 |
| REAL | DOUBLE PRECISION | 浮点数/权重 |
| BLOB | BYTEA | 二进制/CRDT 变更 |
| — | vector(768) | pgvector 嵌入向量 (仅云端) |
| INTEGER | BOOLEAN | 布尔值 (0/1) |

### 1.3 本地 vs 云端职责划分

| 维度 | 本地 SQLite | 云端 PostgreSQL |
|------|-----------|----------------|
| 角色 | 离线主存储、CRDT 变更日志 | 持久化权威副本、多设备同步枢纽 |
| 用户数据 | 全量数据（除账号凭据） | 全量数据 |
| 账号 | 仅缓存当前账号 ID | 账号注册、认证、MFA |
| 向量搜索 | LanceDB (独立引擎) | pgvector 索引 |
| 全文搜索 | SQLite FTS5 | Meilisearch (独立引擎) |
| 文件存储 | 本地文件系统路径 | MinIO/S3 对象存储 |
| 消息队列 | — | NATS (智能体通信) |
| 配置缓存 | — | Redis |

---

## 2. 实体关系总览

```
┌───────────┐       ┌────────────────┐       ┌───────────────┐
│  Account   │ 1:N   │    Space       │ 1:N   │  ObjectType   │
│  (云端)    │──────▶│  (工作区)      │──────▶│  (类型定义)   │
└───────────┘       │                │       └───────────────┘
                    │  id            │       ┌───────────────┐
┌───────────┐       │  name          │ 1:N   │  Collection   │
│Subscription│ 1:1  │  type          │──────▶│  (数据库视图) │
│  (云端)   │◀─────│  account_id    │       └───────────────┘
└───────────┘       │  encrypted     │       ┌───────────────┐
                    └───────┬───────┘  1:N  │  Agent        │
                            │──────────────▶│  (智能体)     │
                            │               └───────────────┘
                            │ 1:N   ┌──────────────────┐
                            ├──────▶│     Object       │
                            │       │  (知识对象)      │
                            │       │──────────────────│
                            │       │ title, properties│
                            │       │ content_version  │
                            │       │ type_id (FK)     │
                            │       └───────┬──────────┘
                            │               │ 1:N
                            │               ▼
                            │       ┌──────────────────┐
                            │       │     Block        │
                            │       │  (内容块)        │
                            │       │──────────────────│
                            │       │ type, content    │
                            │       │ parent_id (自引) │
                            │       │ position         │
                            │       └──────────────────┘
                            │ 1:N
                            ├──────▶┌─────────────┐    ┌──────────────┐
                            │       │   Tag       │◀───│ Object_Tag   │
                            │       │─────────────│    │──────────────│
                            │       │ name, color │    │ object_id    │
                            │       │ ai_generated│    │ tag_id       │
                            │       └─────────────┘    └──────────────┘
                            │ 1:N
                            ├──────▶┌─────────────┐    ┌──────────────┐
                            │       │  Relation   │◀───│    Object    │
                            │       │─────────────│    │ (source/target)
                            │       │ source_id   │    └──────────────┘
                            │       │ target_id   │
                            │       │ type, weight│
                            │       └─────────────┘
                            │
                            ├──────▶┌─────────────────────┐
                            │       │    sync_log         │
                            │       │ (CRDT 变更日志)     │
                            │       └─────────────────────┘
```

---

## 3. 表定义

### 3.1 空间与账号体系

#### 3.1.1 `accounts` — 账号表 (云端专用)

```sql
CREATE TABLE accounts (
    id              TEXT PRIMARY KEY,                          -- UUID v4
    email           TEXT NOT NULL UNIQUE,                      -- 登录邮箱
    name            TEXT NOT NULL,                             -- 显示名称
    avatar_url      TEXT,                                      -- 头像 URL
    auth_provider   TEXT NOT NULL DEFAULT 'email',             -- email | google | wechat | apple | sso
    password_hash   TEXT,                                      -- bcrypt (provider=email 时非空)
    mfa_enabled     INTEGER DEFAULT 0,                         -- 是否开启 MFA
    mfa_secret      TEXT,                                      -- TOTP Secret (加密存储)
    locale          TEXT DEFAULT 'zh-CN',                      -- 语言偏好
    timezone        TEXT DEFAULT 'Asia/Shanghai',              -- 时区
    is_active       INTEGER DEFAULT 1,                         -- 账号是否激活
    last_login_at   INTEGER,                                   -- 最后登录时间
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_accounts_email ON accounts(email);
```

#### 3.1.2 `oauth_accounts` — OAuth 第三方账号绑定 (云端专用)

```sql
CREATE TABLE oauth_accounts (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,                             -- google | wechat | apple
    provider_id     TEXT NOT NULL,                             -- 第三方用户 ID
    provider_email  TEXT,                                      -- 第三方邮箱
    access_token    TEXT,                                      -- 加密存储
    refresh_token   TEXT,                                      -- 加密存储
    expires_at      INTEGER,                                   -- Token 过期时间
    created_at      INTEGER NOT NULL,
    UNIQUE(provider, provider_id)
);
CREATE INDEX idx_oauth_account ON oauth_accounts(account_id);
```

#### 3.1.3 `refresh_tokens` — 刷新令牌 (云端专用)

```sql
CREATE TABLE refresh_tokens (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,                             -- SHA-256(token)
    device_id       TEXT NOT NULL,                             -- 设备标识
    device_name     TEXT,
    device_type     TEXT,                                      -- web | desktop | mobile
    ip_address      TEXT,
    user_agent      TEXT,
    expires_at      INTEGER NOT NULL,
    revoked         INTEGER DEFAULT 0,
    rotated_by      TEXT,                                      -- 被哪个新 token 轮换
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_account ON refresh_tokens(account_id);
```

#### 3.1.4 `mfa_devices` — MFA 设备 (云端专用)

```sql
CREATE TABLE mfa_devices (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK(type IN ('totp', 'sms', 'email', 'backup_code')),
    name            TEXT,
    secret          TEXT,                                      -- 加密存储
    backup_codes    TEXT,                                      -- JSON: ["xxxxx", ...]
    last_used_at    INTEGER,
    is_primary      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_mfa_devices_account ON mfa_devices(account_id);
```

#### 3.1.5 `spaces` — 工作区

```sql
CREATE TABLE spaces (
    id              TEXT PRIMARY KEY,                          -- UUID v4
    name            TEXT NOT NULL,
    type            TEXT NOT NULL CHECK(type IN ('personal', 'team')),
    account_id      TEXT NOT NULL,                             -- 拥有者账号 ID (云端 FK -> accounts.id)
    icon            TEXT DEFAULT '📁',
    description     TEXT,
    encrypted       INTEGER DEFAULT 0,                         -- 是否 E2EE
    encryption_key  TEXT,                                      -- 加密密钥提示
    settings        TEXT DEFAULT '{}',                          -- JSON: 空间级别配置
    object_count    INTEGER DEFAULT 0,                         -- 对象计数 (缓存)
    sync_status     TEXT DEFAULT 'ready',                       -- ready | syncing | conflict
    is_deleted      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_spaces_account ON spaces(account_id);
CREATE INDEX idx_spaces_type ON spaces(type);
```

#### 3.1.6 `space_members` — 空间成员 (团队空间, 云端专用)

```sql
CREATE TABLE space_members (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK(role IN ('owner', 'editor', 'commenter', 'reader')),
    invite_email    TEXT,
    invite_status   TEXT DEFAULT 'accepted' CHECK(invite_status IN ('pending', 'accepted', 'declined', 'revoked')),
    joined_at       INTEGER,
    created_at      INTEGER NOT NULL,
    UNIQUE(space_id, account_id)
);
CREATE INDEX idx_space_members_account ON space_members(account_id);
CREATE INDEX idx_space_members_space ON space_members(space_id);
```

#### 3.1.7 `subscriptions` — 订阅信息 (云端专用)

```sql
CREATE TABLE subscriptions (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL UNIQUE REFERENCES accounts(id),
    plan            TEXT NOT NULL CHECK(plan IN ('free', 'pro', 'team', 'self_hosted')),
    ai_tokens_used  INTEGER DEFAULT 0,
    ai_tokens_limit INTEGER DEFAULT 10000,
    storage_used    BIGINT DEFAULT 0,
    storage_limit   BIGINT DEFAULT 104857600,                   -- 默认 100MB (bytes)
    seats_total     INTEGER DEFAULT 1,                          -- 团队席位总数
    seats_used      INTEGER DEFAULT 1,
    starts_at       INTEGER NOT NULL,
    expires_at      INTEGER,
    auto_renew      INTEGER DEFAULT 1,
    payment_method  TEXT,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
```

#### 3.1.8 `account_devices` — 设备管理

```sql
CREATE TABLE account_devices (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL,                              -- 云端 FK -> accounts.id
    device_name     TEXT NOT NULL,
    device_type     TEXT NOT NULL CHECK(device_type IN ('web', 'desktop_windows', 'desktop_macos', 'desktop_linux', 'mobile_ios', 'mobile_android')),
    device_fingerprint TEXT,                                    -- 设备指纹 SHA-256
    public_key      TEXT,                                       -- E2E 密钥对公钥
    push_token      TEXT,                                       -- 推送通知 Token
    last_sync_at    INTEGER,
    last_ip         TEXT,
    is_current      INTEGER DEFAULT 0,
    is_active       INTEGER DEFAULT 1,
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_devices_account ON account_devices(account_id);
CREATE INDEX idx_devices_fingerprint ON account_devices(device_fingerprint);
```

#### 3.1.9 `user_preferences` — 用户偏好

```sql
CREATE TABLE user_preferences (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL,                              -- 云端 FK -> accounts.id
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    theme           TEXT DEFAULT 'light',                        -- light | dark | system
    font_size       INTEGER DEFAULT 16,
    editor_mode     TEXT DEFAULT 'wysiwyg',                      -- wysiwyg | markdown | source
    sidebar_width   INTEGER DEFAULT 280,
    sidebar_collapsed INTEGER DEFAULT 0,
    language        TEXT DEFAULT 'zh-CN',
    ai_model        TEXT DEFAULT 'default',                      -- 用户偏好的 AI 模型
    custom_settings TEXT DEFAULT '{}',                           -- JSON: 自定义设置
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(account_id, space_id)
);
```

---

### 3.2 对象类型系统

#### 3.2.1 `object_types` — 类型定义

```sql
CREATE TABLE object_types (
    id              TEXT PRIMARY KEY,                          -- UUID v4
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                             -- "书籍"、"论文"、"项目"、"人脉"
    icon            TEXT DEFAULT '📄',
    color           TEXT DEFAULT '#6366f1',                    -- 图谱显示颜色
    description     TEXT,
    is_builtin      INTEGER DEFAULT 0,                         -- 是否系统内置 (不可删除)
    is_archived     INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_object_types_space ON object_types(space_id);
```

#### 3.2.2 `type_fields` — 类型的字段定义 (归一化存储)

```sql
CREATE TABLE type_fields (
    id              TEXT PRIMARY KEY,
    type_id         TEXT NOT NULL REFERENCES object_types(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                             -- "作者"、"出版社"、"状态"
    field_type      TEXT NOT NULL CHECK(field_type IN (
                        'text', 'number', 'date', 'select', 'multi_select',
                        'relation', 'rollup', 'formula', 'file', 'email',
                        'url', 'phone', 'progress', 'currency', 'rating'
                    )),
    position        INTEGER NOT NULL DEFAULT 0,                -- 排序
    required        INTEGER DEFAULT 0,
    default_value   TEXT,                                      -- 默认值 (JSON 编码)
    options         TEXT,                                      -- JSON: SelectOption[] (select 类型专用)
    config          TEXT DEFAULT '{}',                          -- JSON: 字段级配置
    is_builtin      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_type_fields_type ON type_fields(type_id);
```

#### 3.2.3 `type_relations` — 类型的关系定义

```sql
CREATE TABLE type_relations (
    id              TEXT PRIMARY KEY,
    type_id         TEXT NOT NULL REFERENCES object_types(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                             -- "引用"、"隶属于"
    target_type_id  TEXT NOT NULL REFERENCES object_types(id),
    relation_type   TEXT NOT NULL CHECK(relation_type IN ('one_to_one', 'one_to_many', 'many_to_many')),
    reverse_name    TEXT,                                      -- 反向关系名称
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_type_relations_type ON type_relations(type_id);
CREATE INDEX idx_type_relations_target ON type_relations(target_type_id);
```

---

### 3.3 知识对象与内容

#### 3.3.1 `objects` — 知识对象 (核心表)

```sql
CREATE TABLE objects (
    id              TEXT PRIMARY KEY,                          -- UUID v4
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    type_id         TEXT NOT NULL REFERENCES object_types(id),
    title           TEXT NOT NULL DEFAULT '',                    -- 标题 (可为空, AI 自动生成)
    properties      TEXT DEFAULT '{}',                          -- JSON: 动态属性 KV
    tags            TEXT DEFAULT '[]',                          -- JSON: tag_id 数组
    cover_image     TEXT,                                       -- 封面图 URL/路径
    source          TEXT DEFAULT 'manual' CHECK(source IN (
                        'manual', 'video', 'web', 'camera', 'audio',
                        'import', 'clipboard', 'email', 'agent'
                    )),
    source_meta     TEXT DEFAULT '{}',                          -- JSON: 来源元数据
    embedding_id    TEXT,                                       -- 向量索引 ID (LanceDB)
    word_count      INTEGER DEFAULT 0,
    version         INTEGER DEFAULT 1,                          -- 乐观锁/内容版本
    is_archived     INTEGER DEFAULT 0,
    is_deleted      INTEGER DEFAULT 0,
    last_read_at    INTEGER,
    sync_status     TEXT DEFAULT 'synced' CHECK(sync_status IN (
                        'synced', 'pending', 'conflict', 'error'
                    )),
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_objects_space_type ON objects(space_id, type_id);
CREATE INDEX idx_objects_type ON objects(type_id);
CREATE INDEX idx_objects_source ON objects(source);
CREATE INDEX idx_objects_updated ON objects(updated_at DESC);
CREATE INDEX idx_objects_deleted ON objects(is_deleted) WHERE is_deleted = 0;
```

#### 3.3.2 `blocks` — 内容块

```sql
CREATE TABLE blocks (
    id              TEXT PRIMARY KEY,                          -- UUID v4
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    parent_id       TEXT REFERENCES blocks(id) ON DELETE CASCADE, -- 父块 (块级嵌套)
    type            TEXT NOT NULL CHECK(type IN (
                        'text', 'heading1', 'heading2', 'heading3', 'heading4',
                        'image', 'code', 'table', 'table_row', 'table_cell',
                        'mermaid', 'todo', 'bullet_list', 'numbered_list',
                        'list_item', 'quote', 'divider', 'file', 'embed',
                        'callout', 'toggle', 'bookmark', 'equation',
                        'canvas', 'link_preview', 'block_reference'
                    )),
    content         TEXT NOT NULL DEFAULT '',                   -- Markdown / JSON / 文本
    properties      TEXT DEFAULT '{}',                          -- JSON: 额外属性
    position        REAL NOT NULL DEFAULT 0,                    -- 使用浮点数支持中间插入 (相邻平均法)
    depth           INTEGER DEFAULT 0,                          -- 缩进层级
    collapsed       INTEGER DEFAULT 0,                          -- toggle 折叠状态
    color           TEXT,                                       -- 块级颜色标记
    version         INTEGER DEFAULT 1,
    sync_status     TEXT DEFAULT 'synced',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_blocks_object ON blocks(object_id, position);
CREATE INDEX idx_blocks_parent ON blocks(parent_id);
CREATE INDEX idx_blocks_type ON blocks(object_id, type);
```

#### 3.3.3 `object_properties` — 对象属性 EAV (可选扩展)

当 `objects.properties` JSON 不满足复杂查询时使用此表。

```sql
CREATE TABLE object_properties (
    id              TEXT PRIMARY KEY,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    field_id        TEXT NOT NULL REFERENCES type_fields(id),   -- 对应类型字段
    value_text      TEXT,                                      -- 文本值
    value_number    REAL,                                      -- 数值
    value_date      INTEGER,                                   -- 日期值 (Unix ms)
    value_ref       TEXT,                                      -- 关联对象 ID
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(object_id, field_id)
);
CREATE INDEX idx_obj_props_object ON object_properties(object_id);
CREATE INDEX idx_obj_props_field ON object_properties(field_id);
CREATE INDEX idx_obj_props_text ON object_properties(value_text);
CREATE INDEX idx_obj_props_number ON object_properties(value_number);
CREATE INDEX idx_obj_props_date ON object_properties(value_date);
```

#### 3.3.4 `object_versions` — 对象版本历史

```sql
CREATE TABLE object_versions (
    id              TEXT PRIMARY KEY,                          -- UUID v4
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    version         INTEGER NOT NULL,                          -- 版本号 (递增)
    title           TEXT NOT NULL,
    content_snapshot TEXT NOT NULL,                             -- JSON: blocks[] 快照
    properties      TEXT DEFAULT '{}',
    tags            TEXT DEFAULT '[]',
    word_count      INTEGER DEFAULT 0,
    change_summary  TEXT,                                      -- AI 生成的变更摘要
    device_id       TEXT,                                      -- 变更来源设备
    account_id      TEXT,                                      -- 变更者
    diff            TEXT,                                       -- JSON: 与上一版本的差异
    created_at      INTEGER NOT NULL                            -- 版本创建时间
);
CREATE INDEX idx_obj_versions_object ON object_versions(object_id, version DESC);
CREATE INDEX idx_obj_versions_created ON object_versions(created_at DESC);
```

#### 3.3.5 `trash` — 回收站

```sql
CREATE TABLE trash (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL,                              -- 原对象 ID
    object_type     TEXT NOT NULL,                              -- 'object' | 'type' | 'collection' | 'tag' | 'agent'
    data            TEXT NOT NULL,                               -- JSON: 完整数据快照 (用于恢复)
    deleted_by      TEXT,                                       -- 账号 ID
    auto_delete_at  INTEGER,                                    -- 自动清理时间 (30 天后)
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_trash_space ON trash(space_id);
CREATE INDEX idx_trash_autodelete ON trash(auto_delete_at) WHERE auto_delete_at IS NOT NULL;
```

---

### 3.4 关系与标签

#### 3.4.1 `relations` — 对象间关系

```sql
CREATE TABLE relations (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    source_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    target_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK(type IN (
                        'link', 'reference', 'citation', 'parent', 'child',
                        'related', 'custom', 'duplicate', 'sequence'
                    )),
    custom_type_id  TEXT,                                      -- 用户自定义关系类型 (type=custom 时)
    weight          REAL DEFAULT 1.0 CHECK(weight >= 0 AND weight <= 1),  -- 关系强度 [0,1]
    metadata        TEXT DEFAULT '{}',                          -- JSON: 附加信息 (如引用页码)
    ai_generated    INTEGER DEFAULT 0,                          -- AI 自动发现
    source_object_type TEXT,                                    -- 来源对象类型 (冗余，用于快速筛选)
    target_object_type TEXT,                                    -- 目标对象类型 (冗余)
    created_at      INTEGER NOT NULL,
    UNIQUE(source_id, target_id, type)
);
CREATE INDEX idx_relations_source ON relations(source_id);
CREATE INDEX idx_relations_target ON relations(target_id);
CREATE INDEX idx_relations_space ON relations(space_id);
CREATE INDEX idx_relations_type ON relations(type);
CREATE INDEX idx_relations_weight ON relations(weight DESC);
```

#### 3.4.2 `tags` — 标签

```sql
CREATE TABLE tags (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    color           TEXT DEFAULT '#94a3b8',
    parent_id       TEXT REFERENCES tags(id) ON DELETE SET NULL,
    description     TEXT,
    ai_generated    INTEGER DEFAULT 0,
    object_count    INTEGER DEFAULT 0,                          -- 缓存: 关联对象数
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(space_id, name)
);
CREATE INDEX idx_tags_space ON tags(space_id);
CREATE INDEX idx_tags_parent ON tags(parent_id);
```

#### 3.4.3 `object_tags` — 对象与标签关联

```sql
CREATE TABLE object_tags (
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    tag_id          TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at      INTEGER NOT NULL,
    PRIMARY KEY (object_id, tag_id)
);
CREATE INDEX idx_object_tags_tag ON object_tags(tag_id);
```

---

### 3.5 集合与视图

#### 3.5.1 `collections` — 数据库集合

```sql
CREATE TABLE collections (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    icon            TEXT DEFAULT '📁',
    description     TEXT,
    source_type     TEXT NOT NULL CHECK(source_type IN ('type', 'tag', 'mixed', 'manual', 'saved_search')),
    source_config   TEXT DEFAULT '{}',                          -- JSON: 来源配置
    layout          TEXT DEFAULT 'table',                        -- table | kanban | gallery | calendar | timeline
    is_archived     INTEGER DEFAULT 0,
    position        INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_collections_space ON collections(space_id);
```

#### 3.5.2 `collection_views` — 集合的多种视图

```sql
CREATE TABLE collection_views (
    id              TEXT PRIMARY KEY,
    collection_id   TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                             -- "表格视图"、"看板视图"
    view_type       TEXT NOT NULL CHECK(view_type IN (
                        'table', 'kanban', 'gallery', 'calendar', 'timeline', 'list'
                    )),
    filters         TEXT DEFAULT '[]',                          -- JSON: FilterGroup[]
    sorts           TEXT DEFAULT '[]',                          -- JSON: SortRule[]
    group_by        TEXT,                                       -- 分组字段 ID
    visible_fields  TEXT DEFAULT '[]',                          -- JSON: field_id[]
    column_widths   TEXT DEFAULT '{}',                          -- JSON: { field_id: width }
    calendar_field  TEXT,                                       -- 日历视图专用: 日期字段 ID
    kanban_field    TEXT,                                       -- 看板视图专用: 状态字段 ID
    is_default      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_collection_views_parent ON collection_views(collection_id);
```

#### 3.5.3 `collection_items` — 手动集合的条目

```sql
CREATE TABLE collection_items (
    id              TEXT PRIMARY KEY,
    collection_id   TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    position        INTEGER DEFAULT 0,
    note            TEXT,                                       -- 集合内备注
    created_at      INTEGER NOT NULL,
    UNIQUE(collection_id, object_id)
);
CREATE INDEX idx_collection_items_collection ON collection_items(collection_id);
CREATE INDEX idx_collection_items_object ON collection_items(object_id);
```

---

### 3.6 模板

#### 3.6.1 `templates` — 模板定义

```sql
CREATE TABLE templates (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    type_id         TEXT REFERENCES object_types(id) ON DELETE SET NULL,
    name            TEXT NOT NULL,
    icon            TEXT DEFAULT '📋',
    description     TEXT,
    category        TEXT DEFAULT 'general' CHECK(category IN (
                        'general', 'meeting', 'book', 'project', 'daily',
                        'weekly', 'person', 'vocabulary', 'custom'
                    )),
    blocks          TEXT NOT NULL DEFAULT '[]',                  -- JSON: Block[] 模板
    properties_def  TEXT DEFAULT '{}',                           -- JSON: 属性默认值
    is_builtin      INTEGER DEFAULT 0,
    is_public       INTEGER DEFAULT 0,                           -- 是否公开到模板市场
    use_count       INTEGER DEFAULT 0,                           -- 使用计数
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_templates_space ON templates(space_id);
CREATE INDEX idx_templates_type ON templates(type_id);
CREATE INDEX idx_templates_category ON templates(category);
```

---

### 3.7 采集与视觉

#### 3.7.1 `capture_sessions` — 采集会话

```sql
CREATE TABLE capture_sessions (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    source          TEXT NOT NULL CHECK(source IN (
                        'video', 'camera', 'screen_recording', 'burst', 'gallery'
                    )),
    source_app      TEXT,                                       -- 来源 App 名称 (如 "哔哩哔哩")
    source_url      TEXT,                                       -- 视频/文章 URL
    source_title    TEXT,                                       -- 视频/课程标题
    playback_position INTEGER,                                  -- 视频播放位置 (秒)
    subtitle        TEXT,                                       -- 捕获时的字幕/弹幕
    image_count     INTEGER DEFAULT 0,                          -- 截取图片数
    status          TEXT DEFAULT 'active' CHECK(status IN (
                        'active', 'paused', 'completed', 'cancelled'
                    )),
    device_id       TEXT,
    started_at      INTEGER NOT NULL,
    ended_at        INTEGER
);
CREATE INDEX idx_capture_sessions_space ON capture_sessions(space_id);
CREATE INDEX idx_capture_sessions_source ON capture_sessions(source_app);
```

#### 3.7.2 `image_queue` — 图片处理队列

```sql
CREATE TABLE image_queue (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    session_id      TEXT REFERENCES capture_sessions(id) ON DELETE SET NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK(status IN (
                        'pending', 'processing', 'completed', 'failed', 'duplicate', 'low_quality', 'cancelled'
                    )),
    source          TEXT NOT NULL,
    file_path       TEXT NOT NULL,                              -- 本地路径 / 云端 URL
    file_size       INTEGER,                                    -- 文件大小 (bytes)
    image_hash      TEXT,                                       -- pHash 用于去重
    width           INTEGER,
    height          INTEGER,
    metadata        TEXT DEFAULT '{}',                          -- JSON: ScreenshotMetadata
    result_object_id TEXT REFERENCES objects(id) ON DELETE SET NULL,  -- 处理后生成的对象
    ocr_text        TEXT,                                       -- OCR 结果缓存
    quality_score   REAL,                                       -- 质量评分 [0,1]
    retry_count     INTEGER DEFAULT 0,
    priority        INTEGER DEFAULT 0,                          -- 优先级 (越大越优先)
    error           TEXT,
    processing_time INTEGER,                                    -- 处理耗时 (ms)
    enhanced_path   TEXT,                                       -- 增强后图片路径
    created_at      INTEGER NOT NULL,
    processed_at    INTEGER
);
CREATE INDEX idx_image_queue_status ON image_queue(status, priority);
CREATE INDEX idx_image_queue_session ON image_queue(session_id);
CREATE INDEX idx_image_queue_hash ON image_queue(image_hash);
CREATE INDEX idx_image_queue_space ON image_queue(space_id);
```

#### 3.7.3 `vision_jobs` — 视觉处理任务

```sql
CREATE TABLE vision_jobs (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    image_id        TEXT REFERENCES image_queue(id) ON DELETE CASCADE,
    job_type        TEXT NOT NULL CHECK(job_type IN (
                        'ocr', 'enhance', 'layout_analysis', 'formula_recognition',
                        'chart_extraction', 'entity_recognition', 'full_pipeline'
                    )),
    status          TEXT NOT NULL DEFAULT 'pending' CHECK(status IN (
                        'pending', 'queued', 'processing', 'completed', 'failed'
                    )),
    model_used      TEXT,                                       -- 使用的模型 (如 PaddleOCR-v4)
    input_path      TEXT NOT NULL,
    output_path     TEXT,                                       -- 处理结果路径
    result          TEXT,                                       -- JSON: ExtractionResult
    confidence      REAL,                                       -- 置信度 [0,1]
    processing_time INTEGER,                                    -- 耗时 (ms)
    error           TEXT,
    retry_count     INTEGER DEFAULT 0,
    worker_id       TEXT,                                       -- 处理节点 ID
    created_at      INTEGER NOT NULL,
    started_at      INTEGER,
    completed_at    INTEGER
);
CREATE INDEX idx_vision_jobs_status ON vision_jobs(status);
CREATE INDEX idx_vision_jobs_image ON vision_jobs(image_id);
CREATE INDEX idx_vision_jobs_space ON vision_jobs(space_id);
```

---

### 3.8 智能体

#### 3.8.1 `agents` — 智能体定义

```sql
CREATE TABLE agents (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    agent_type      TEXT NOT NULL CHECK(agent_type IN (
                        'summarizer', 'tagger', 'linker', 'review', 'writer',
                        'collector', 'refactor', 'orchestrator', 'scheduler', 'custom'
                    )),
    icon            TEXT DEFAULT '🤖',
    color           TEXT DEFAULT '#8b5cf6',
    triggers        TEXT DEFAULT '[]',                           -- JSON: AgentTrigger[]
    actions         TEXT DEFAULT '[]',                           -- JSON: AgentAction[]
    schedule        TEXT,                                       -- cron 表达式
    config          TEXT DEFAULT '{}',                           -- JSON: 额外配置
    model_preference TEXT,                                       -- 偏好的 AI 模型
    timeout_ms      INTEGER DEFAULT 300000,                     -- 超时 (ms), 默认 5min
    max_retries     INTEGER DEFAULT 3,
    max_tokens_per_run INTEGER DEFAULT 32000,
    max_memory_items INTEGER DEFAULT 100,
    enabled         INTEGER DEFAULT 1,
    last_run_at     INTEGER,
    last_error      TEXT,
    run_count       INTEGER DEFAULT 0,
    success_count   INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_agents_space ON agents(space_id);
CREATE INDEX idx_agents_type ON agents(agent_type);
CREATE INDEX idx_agents_enabled ON agents(enabled) WHERE enabled = 1;
```

#### 3.8.2 `agent_runs` — 智能体运行记录

```sql
CREATE TABLE agent_runs (
    id              TEXT PRIMARY KEY,
    agent_id        TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    trigger_type    TEXT NOT NULL,                               -- manual | schedule | event | workflow
    status          TEXT NOT NULL CHECK(status IN (
                        'pending', 'running', 'completed', 'failed', 'cancelled', 'timeout'
                    )),
    input           TEXT,                                       -- JSON: 输入参数
    output          TEXT,                                       -- JSON: 执行结果
    error           TEXT,                                       -- 错误信息
    tokens_used     INTEGER DEFAULT 0,
    processing_time INTEGER,                                    -- 耗时 (ms)
    steps_completed INTEGER DEFAULT 0,
    steps_total     INTEGER DEFAULT 0,
    started_at      INTEGER NOT NULL,
    completed_at    INTEGER
);
CREATE INDEX idx_agent_runs_agent ON agent_runs(agent_id, started_at DESC);
CREATE INDEX idx_agent_runs_space ON agent_runs(space_id);
CREATE INDEX idx_agent_runs_status ON agent_runs(status);
```

#### 3.8.3 `agent_messages` — 智能体消息总线持久化

```sql
CREATE TABLE agent_messages (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    conversation_id TEXT NOT NULL,                               -- 会话 ID (关联请求/响应)
    msg_type        TEXT NOT NULL CHECK(msg_type IN (
                        'task_request', 'task_result', 'query', 'event', 'heartbeat', 'error'
                    )),
    source          TEXT NOT NULL,                               -- 发送方 agent_id
    target          TEXT,                                       -- 接收方 agent_id (NULL=广播)
    priority        TEXT NOT NULL DEFAULT 'normal' CHECK(priority IN ('low', 'normal', 'high', 'critical')),
    payload         TEXT NOT NULL,                               -- JSON: 消息体
    context         TEXT DEFAULT '{}',                           -- JSON: 上下文 (userId, spaceId, objectId)
    status          TEXT DEFAULT 'pending' CHECK(status IN (
                        'pending', 'delivered', 'processed', 'failed', 'expired'
                    )),
    ttl             INTEGER DEFAULT 300000,                      -- 存活时间 (ms)
    retry_count     INTEGER DEFAULT 0,
    max_retries     INTEGER DEFAULT 3,
    timestamp       INTEGER NOT NULL,
    processed_at    INTEGER
);
CREATE INDEX idx_agent_messages_conv ON agent_messages(conversation_id);
CREATE INDEX idx_agent_messages_target ON agent_messages(target, status);
CREATE INDEX idx_agent_messages_space ON agent_messages(space_id);
CREATE INDEX idx_agent_messages_priority ON agent_messages(priority, timestamp DESC);
```

#### 3.8.4 `long_horizon_agents` — 长周期自主智能体

```sql
CREATE TABLE long_horizon_agents (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    goal            TEXT NOT NULL,                               -- "追踪 AI 芯片行业动态"
    plan            TEXT DEFAULT '[]',                           -- JSON: AgentStep[]
    state           TEXT NOT NULL DEFAULT 'running' CHECK(state IN (
                        'running', 'paused', 'completed', 'failed', 'cancelled'
                    )),
    progress        REAL DEFAULT 0 CHECK(progress >= 0 AND progress <= 100),
    short_term_memory TEXT DEFAULT '[]',                         -- JSON: MemoryItem[]
    deliverables    TEXT DEFAULT '[]',                           -- JSON: Deliverable[]
    deadline        INTEGER,
    started_at      INTEGER NOT NULL,
    last_active_at  INTEGER,
    next_check_in   INTEGER                                      -- 下次向用户汇报时间
);
CREATE INDEX idx_lha_space ON long_horizon_agents(space_id);
CREATE INDEX idx_lha_state ON long_horizon_agents(state);
CREATE INDEX idx_lha_checkin ON long_horizon_agents(next_check_in) WHERE state = 'running';
```

#### 3.8.5 `agent_steps` — 长周期智能体步骤

```sql
CREATE TABLE agent_steps (
    id              TEXT PRIMARY KEY,
    lha_id          TEXT NOT NULL REFERENCES long_horizon_agents(id) ON DELETE CASCADE,
    step_type       TEXT NOT NULL CHECK(step_type IN (
                        'collect', 'analyze', 'summarize', 'relate', 'report', 'custom'
                    )),
    title           TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK(status IN (
                        'pending', 'running', 'completed', 'skipped', 'failed'
                    )),
    assigned_agent  TEXT,                                       -- 执行智能体 ID
    depends_on      TEXT DEFAULT '[]',                           -- JSON: 依赖的 step_id[]
    input           TEXT,
    result          TEXT,                                       -- JSON 执行结果
    error           TEXT,
    position        INTEGER DEFAULT 0,
    started_at      INTEGER,
    completed_at    INTEGER
);
CREATE INDEX idx_agent_steps_lha ON agent_steps(lha_id, position);
```

#### 3.8.6 `agent_memory` — 智能体长期记忆

```sql
CREATE TABLE agent_memory (
    id              TEXT PRIMARY KEY,
    agent_id        TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    lha_id          TEXT REFERENCES long_horizon_agents(id) ON DELETE CASCADE,
    memory_type     TEXT NOT NULL CHECK(memory_type IN (
                        'observation', 'insight', 'relation', 'decision', 'error', 'feedback'
                    )),
    content         TEXT NOT NULL,                               -- 记忆内容
    embedding       TEXT,                                       -- 向量嵌入 (base64)
    relevance_score REAL DEFAULT 1.0,                            -- 相关度 [0,1]
    source          TEXT,                                       -- 来源上下文
    expires_at      INTEGER,                                    -- 过期时间 (自动清理)
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_agent_memory_agent ON agent_memory(agent_id);
CREATE INDEX idx_agent_memory_type ON agent_memory(memory_type);
CREATE INDEX idx_agent_memory_relevance ON agent_memory(relevance_score DESC);
```

---

### 3.9 同步与版本

#### 3.9.1 `sync_log` — CRDT 同步日志

```sql
CREATE TABLE sync_log (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL,
    object_id       TEXT NOT NULL,
    object_type     TEXT NOT NULL CHECK(object_type IN (
                        'object', 'block', 'collection', 'type', 'tag', 'relation'
                    )),
    change_type     TEXT NOT NULL CHECK(change_type IN (
                        'create', 'update', 'delete', 'move', 'archive', 'restore'
                    )),
    data            BLOB,                                       -- CRDT 二进制变更 (Go-Yrs)
    version         INTEGER NOT NULL,
    device_id       TEXT NOT NULL,
    vector_clock    TEXT,                                       -- JSON: 版本向量 { device_id: version }
    parent_id       TEXT,                                       -- 父变更 ID (用于因果顺序)
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_sync_log_object ON sync_log(object_id, version);
CREATE INDEX idx_sync_log_space ON sync_log(space_id, created_at DESC);
CREATE INDEX idx_sync_log_device ON sync_log(device_id, created_at DESC);
```

#### 3.9.2 `sync_checkpoint` — 同步检查点

```sql
CREATE TABLE sync_checkpoint (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL,
    device_id       TEXT NOT NULL,
    object_type     TEXT NOT NULL,
    last_version    INTEGER NOT NULL,                           -- 已同步到的版本号
    last_log_id     TEXT,                                       -- 最后处理的日志 ID
    snapshot        BLOB,                                       -- CRDT 快照 (可选)
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(space_id, device_id, object_type)
);
CREATE INDEX idx_sync_checkpoint_space ON sync_checkpoint(space_id);
```

#### 3.9.3 `sync_conflicts` — 同步冲突记录

```sql
CREATE TABLE sync_conflicts (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL,
    object_type     TEXT NOT NULL,
    local_version   INTEGER NOT NULL,
    remote_version  INTEGER NOT NULL,
    local_data      BLOB,                                       -- 本地 CRDT 状态
    remote_data     BLOB,                                       -- 远程 CRDT 状态
    merged_data     BLOB,                                       -- 合并后状态
    resolution      TEXT DEFAULT 'auto' CHECK(resolution IN (
                        'auto_merged', 'manual_local', 'manual_remote', 'manual_custom', 'pending'
                    )),
    resolved_at     INTEGER,
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_sync_conflicts_object ON sync_conflicts(object_id);
CREATE INDEX idx_sync_conflicts_space ON sync_conflicts(space_id);
```

---

### 3.10 文件与附件

#### 3.10.1 `files` — 文件存储记录

```sql
CREATE TABLE files (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT REFERENCES objects(id) ON DELETE SET NULL, -- 关联对象 (可选)
    block_id        TEXT REFERENCES blocks(id) ON DELETE SET NULL,  -- 关联块 (可选)
    original_name   TEXT NOT NULL,                              -- 原始文件名
    mime_type       TEXT NOT NULL,                              -- image/png, application/pdf, ...
    file_size       INTEGER NOT NULL,                           -- 文件大小 (bytes)
    storage_type    TEXT NOT NULL CHECK(storage_type IN (
                        'local', 's3', 'minio', 'b2', 'temp'
                    )),
    storage_path    TEXT NOT NULL,                              -- 存储路径 (本地路径 / S3 Key)
    thumbnail_path  TEXT,                                       -- 缩略图路径
    blurhash        TEXT,                                       -- 图片占位符哈希
    width           INTEGER,                                    -- 图片/视频宽度
    height          INTEGER,                                    -- 图片/视频高度
    duration        INTEGER,                                    -- 音频/视频时长 (秒)
    md5_hash        TEXT,                                       -- 文件 MD5
    is_public       INTEGER DEFAULT 0,                          -- 是否公开可访问
    uploaded_at     INTEGER NOT NULL,
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_files_space ON files(space_id);
CREATE INDEX idx_files_object ON files(object_id);
CREATE INDEX idx_files_md5 ON files(md5_hash);
CREATE INDEX idx_files_mime ON files(mime_type);
```

#### 3.10.2 `file_chunks` — 大文件分片 (上传用, 云端专用)

```sql
CREATE TABLE file_chunks (
    id              TEXT PRIMARY KEY,
    file_id         TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    chunk_index     INTEGER NOT NULL,                           -- 分片序号
    chunk_size      INTEGER NOT NULL,
    chunk_hash      TEXT,                                       -- 分片 SHA-256
    uploaded        INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    UNIQUE(file_id, chunk_index)
);
```

---

### 3.11 AI 与搜索

#### 3.11.1 `ai_cache` — AI 响应缓存

```sql
CREATE TABLE ai_cache (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    cache_key       TEXT NOT NULL,                              -- SHA-256(input + model + params)
    model           TEXT NOT NULL,                               -- 使用的模型
    prompt_hash     TEXT NOT NULL,                               -- 输入 prompt 的哈希
    response        TEXT NOT NULL,                               -- JSON: AI 响应
    input_type      TEXT CHECK(input_type IN ('ocr', 'tagging', 'summary', 'embedding', 'qa', 'refactor')),
    tokens_in       INTEGER DEFAULT 0,
    tokens_out      INTEGER DEFAULT 0,
    confidence      REAL,
    hit_count       INTEGER DEFAULT 1,
    expires_at      INTEGER,                                    -- 缓存过期时间
    created_at      INTEGER NOT NULL,
    UNIQUE(space_id, cache_key)
);
CREATE INDEX idx_ai_cache_key ON ai_cache(cache_key);
CREATE INDEX idx_ai_cache_type ON ai_cache(input_type);
CREATE INDEX idx_ai_cache_expires ON ai_cache(expires_at) WHERE expires_at IS NOT NULL;
```

#### 3.11.2 `ai_usage_logs` — AI 用量日志 (云端专用)

```sql
CREATE TABLE ai_usage_logs (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id),
    space_id        TEXT NOT NULL,
    model           TEXT NOT NULL,                               -- claude-sonnet-4, gpt-4o, ...
    service         TEXT NOT NULL CHECK(service IN (
                        'llm', 'embedding', 'ocr', 'tts', 'stt', 'image_gen'
                    )),
    tokens_in       INTEGER DEFAULT 0,
    tokens_out      INTEGER DEFAULT 0,
    processing_time INTEGER,                                    -- 耗时 (ms)
    cost_usd        REAL DEFAULT 0,                              -- 估算成本
    source          TEXT,                                       -- 来源 (agent_id | 'user' | 'system')
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_ai_usage_account ON ai_usage_logs(account_id, created_at DESC);
CREATE INDEX idx_ai_usage_model ON ai_usage_logs(model);
CREATE INDEX idx_ai_usage_date ON ai_usage_logs(created_at);
```

#### 3.11.3 `search_history` — 搜索历史

```sql
CREATE TABLE search_history (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL,
    query           TEXT NOT NULL,
    search_type     TEXT NOT NULL CHECK(search_type IN (
                        'fulltext', 'semantic', 'hybrid', 'deep', 'related'
                    )),
    result_count    INTEGER DEFAULT 0,
    clicked_item    TEXT,                                       -- 点击的对象 ID
    clicked_position INTEGER,                                   -- 点击位置序号
    processing_time INTEGER,                                    -- 搜索耗时 (ms)
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_search_history_space ON search_history(space_id, created_at DESC);
CREATE INDEX idx_search_history_query ON search_history(query);
```

---

### 3.12 MCP 与开放生态

#### 3.12.1 `mcp_connectors` — MCP 连接器配置

```sql
CREATE TABLE mcp_connectors (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    auth_type       TEXT NOT NULL CHECK(auth_type IN ('oauth', 'api_key', 'none')),
    api_key         TEXT,                                       -- 加密存储
    scopes          TEXT DEFAULT '[]',                           -- JSON: 权限作用域
    rate_limit_rpm  INTEGER DEFAULT 60,                          -- 每分钟请求数
    rate_limit_tpm  INTEGER DEFAULT 100000,                      -- 每分钟 Token 数
    allowed_ips     TEXT DEFAULT '[]',                            -- JSON: IP 白名单
    audit_enabled   INTEGER DEFAULT 1,
    is_active       INTEGER DEFAULT 1,
    last_used_at    INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
CREATE INDEX idx_mcp_connectors_space ON mcp_connectors(space_id);
```

#### 3.12.2 `mcp_audit_log` — MCP 审计日志

```sql
CREATE TABLE mcp_audit_log (
    id              TEXT PRIMARY KEY,
    connector_id    TEXT NOT NULL REFERENCES mcp_connectors(id) ON DELETE CASCADE,
    tool_name       TEXT NOT NULL,                               -- 调用的工具
    input           TEXT,                                       -- 请求参数
    output          TEXT,                                       -- 响应摘要
    status          TEXT NOT NULL CHECK(status IN ('success', 'error', 'rate_limited', 'unauthorized')),
    ip_address      TEXT,
    user_agent      TEXT,
    processing_time INTEGER,
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_mcp_audit_connector ON mcp_audit_log(connector_id, created_at DESC);
CREATE INDEX idx_mcp_audit_tool ON mcp_audit_log(tool_name);
CREATE INDEX idx_mcp_audit_status ON mcp_audit_log(status);
```

---

### 3.13 通知与分享

#### 3.13.1 `notifications` — 通知

```sql
CREATE TABLE notifications (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL,                              -- 接收者
    type            TEXT NOT NULL CHECK(type IN (
                        'agent_complete', 'sync_conflict', 'share_received',
                        'mention', 'system', 'weekly_report', 'review_reminder',
                        'ai_finished', 'export_ready'
                    )),
    title           TEXT NOT NULL,
    body            TEXT,
    data            TEXT DEFAULT '{}',                           -- JSON: 附加数据
    is_read         INTEGER DEFAULT 0,
    read_at         INTEGER,
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_notifications_account ON notifications(account_id, is_read, created_at DESC);
CREATE INDEX idx_notifications_space ON notifications(space_id);
```

#### 3.13.2 `share_links` — 分享链接

```sql
CREATE TABLE share_links (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    code            TEXT NOT NULL UNIQUE,                        -- 短码 (可分享的 URL 标识)
    password_hash   TEXT,                                       -- 访问密码 (bcrypt)
    permission      TEXT DEFAULT 'read' CHECK(permission IN ('read', 'comment', 'edit')),
    expires_at      INTEGER,
    max_views       INTEGER,
    view_count      INTEGER DEFAULT 0,
    is_active       INTEGER DEFAULT 1,
    created_by      TEXT NOT NULL,                               -- 创建者账号 ID
    created_at      INTEGER NOT NULL
);
CREATE INDEX idx_share_links_code ON share_links(code);
CREATE INDEX idx_share_links_object ON share_links(object_id);
```

#### 3.13.3 `export_jobs` — 导出任务

```sql
CREATE TABLE export_jobs (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL,
    format          TEXT NOT NULL CHECK(format IN ('markdown', 'json', 'pdf', 'notion', 'obsidian')),
    scope           TEXT NOT NULL CHECK(scope IN ('space', 'collection', 'object')),
    scope_id        TEXT,                                       -- collection_id / object_id
    status          TEXT NOT NULL DEFAULT 'pending' CHECK(status IN (
                        'pending', 'processing', 'completed', 'failed'
                    )),
    file_path       TEXT,                                       -- 导出文件路径
    file_size       INTEGER,
    options         TEXT DEFAULT '{}',                           -- JSON: 导出选项
    error           TEXT,
    created_at      INTEGER NOT NULL,
    completed_at    INTEGER
);
CREATE INDEX idx_export_jobs_account ON export_jobs(account_id, created_at DESC);
```

---

## 4. 索引策略

### 4.1 核心查询模式与对应索引

| 查询场景 | SQL 模式 | 索引 |
|---------|----------|------|
| 按空间+类型列出对象 | `WHERE space_id=? AND type_id=? ORDER BY updated_at DESC` | `idx_objects_space_type` |
| 按空间+来源列出对象 | `WHERE space_id=? AND source=?` | — (在 `objects` 上已有 source 索引，可加组合) |
| 获取对象的块列表 | `SELECT * FROM blocks WHERE object_id=? ORDER BY position` | `idx_blocks_object` |
| 搜索未删除的对象 | `WHERE is_deleted=0` | `idx_objects_deleted` (部分索引) |
| 查找对象间关系 | `SELECT * FROM relations WHERE source_id=? OR target_id=?` | `idx_relations_source`, `idx_relations_target` |
| 待处理的图片队列 | `SELECT * FROM image_queue WHERE status='pending' ORDER BY priority DESC` | `idx_image_queue_status` |
| 运行中的长周期智能体 | `SELECT * FROM long_horizon_agents WHERE state='running'` | — (数据量小，全表扫描可接受) |
| 消息总线的待处理消息 | `SELECT * FROM agent_messages WHERE target=? AND status='pending'` | `idx_agent_messages_target` |
| 历史版本查询 | `SELECT * FROM object_versions WHERE object_id=? ORDER BY version DESC` | `idx_obj_versions_object` |
| 回收站自动清理 | `SELECT * FROM trash WHERE auto_delete_at < ?` | `idx_trash_autodelete` (部分索引) |

### 4.2 组合索引建议

```sql
-- 对象列表浏览 (按更新时间倒排)
CREATE INDEX idx_objects_list ON objects(space_id, is_deleted, updated_at DESC);

-- 按类型和标签的联合查询
CREATE INDEX idx_objects_type_tags ON objects(space_id, type_id, is_deleted);

-- 智能体消息优先处理
CREATE INDEX idx_agent_msgs_dispatch ON agent_messages(target, status, priority, timestamp);

-- MCP 审计按时间线查看
CREATE INDEX idx_mcp_audit_time ON mcp_audit_log(connector_id, created_at DESC);
```

### 4.3 部分索引 (SQLite 支持, PostgreSQL 也支持)

```sql
-- 只索引未删除的活跃对象 (减少索引大小)
CREATE INDEX idx_active_objects ON objects(space_id, type_id) WHERE is_deleted = 0;

-- 只索引启用的智能体
CREATE INDEX idx_enabled_agents ON agents(space_id) WHERE enabled = 1;

-- 只索引运行中的长周期任务
CREATE INDEX idx_running_lha ON long_horizon_agents(space_id) WHERE state = 'running';
```

---

## 5. 全文搜索

### 5.1 SQLite FTS5 (本地)

```sql
-- 创建 FTS5 虚拟表 (同步更新)
CREATE VIRTUAL TABLE objects_fts USING fts5(
    object_id UNINDEXED,
    title,
    content,        -- 从 blocks 聚合的文本
    tags_text,      -- 标签名字符串
    tokenize='unicode61 tokenchars "separators=。"'
);

-- 触发器：对象更新时同步 FTS
CREATE TRIGGER after_object_insert AFTER INSERT ON objects BEGIN
    INSERT OR REPLACE INTO objects_fts(object_id, title, content, tags_text)
    VALUES (new.id, new.title, '', '');
END;

-- 更新触发器需要关联 blocks 内容，实际由应用层维护
-- 应用层在 blocks 变更时，聚合对象所有块的文本后更新 FTS
```

### 5.2 全文搜索查询

```sql
-- FTS5 查询 (支持中文分词)
SELECT o.id, o.title, o.type_id, o.updated_at,
       rank
FROM objects_fts
JOIN objects o ON o.id = objects_fts.object_id
WHERE objects_fts MATCH ?                     -- 查询词
  AND o.space_id = ?
  AND o.is_deleted = 0
ORDER BY rank DESC
LIMIT ? OFFSET ?;
```

### 5.3 云端全文搜索 (Meilisearch)

云端全文搜索不通过 SQL 进行，由 Meilisearch 独立引擎处理，PostgreSQL 通过触发器或 CDC 将数据同步到 Meilisearch。

---

## 6. 向量搜索

### 6.1 云端 PostgreSQL pgvector

```sql
-- 创建向量扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 对象嵌入向量表
CREATE TABLE object_embeddings (
    object_id       TEXT PRIMARY KEY REFERENCES objects(id) ON DELETE CASCADE,
    embedding       vector(768),            -- BGE-M3 嵌入维度
    model           TEXT NOT NULL,           -- 使用的嵌入模型
    chunk_count     INTEGER DEFAULT 1,       -- 分块数
    updated_at      INTEGER NOT NULL
);

-- IVFFlat 索引 (100 聚类中心)
CREATE INDEX idx_object_embeddings 
    ON object_embeddings 
    USING ivfflat (embedding vector_cosine_ops) 
    WITH (lists = 100);

-- 向量相似度搜索
SELECT o.id, o.title, o.space_id,
       1 - (e.embedding <=> ?) AS similarity
FROM object_embeddings e
JOIN objects o ON o.id = e.object_id
WHERE o.space_id = ?
  AND o.is_deleted = 0
ORDER BY e.embedding <=> ?                  -- 余弦距离
LIMIT ?;
```

### 6.2 本地向量搜索 (LanceDB)

本地使用 LanceDB 作为嵌入式向量数据库，通过 LanceDB Go SDK 集成进 Go 后端。

```
objects/
├── id: string
├── embedding: fixed_size_list(float32, 768)
├── text: string
├── metadata: json
└── space_id: string
```

LanceDB 索引参数:
- IVF with PQ quantization
- metric: cosine
- num_partitions: 256
- num_sub_vectors: 96

---

## 7. 本地 vs 云端差异对照

| 组件 | 本地 (SQLite) | 云端 (PostgreSQL) |
|------|-------------|-------------------|
| 字符串类型 | TEXT | TEXT / VARCHAR(255) |
| JSON 操作 | `json_extract()` | `->>`, `@>`, `?` 等原生 JSONB |
| 布尔值 | INTEGER (0/1) | BOOLEAN |
| 全文搜索 | FTS5 插件 | Meilisearch (独立) |
| 向量搜索 | LanceDB (独立) | pgvector 插件 |
| UUID 生成 | Go `uuid.New()` (google/uuid) | `gen_random_uuid()` |
| 自增列 | `INTEGER PRIMARY KEY` | `SERIAL` / `IDENTITY` |
| UNIQUE 约束 | 支持 (含 NULL) | 支持 |
| 部分索引 | `WHERE is_deleted=0` | 同左 |
| 触发器和外键 | 创建时启用 `PRAGMA foreign_keys=ON` | 原生支持 |
| 账号表 | 不存在 | 存在 (`accounts`, `oauth_accounts` 等) |
| 订阅表 | 不存在 | 存在 (`subscriptions`) |
| MFA 表 | 不存在 | 存在 (`mfa_devices`) |
| 分片表 | 不存在 | 存在 (`file_chunks`) |
| AI 用量 | 可选缓存 | 详细计费日志 |

---

## 8. 分区策略 (云端 PostgreSQL)

对于云端大数据量表，按月或按空间 ID 哈希分区：

```sql
-- 按时间范围分区: sync_log
CREATE TABLE sync_log_partitioned (
    id TEXT NOT NULL,
    space_id TEXT NOT NULL,
    object_id TEXT NOT NULL,
    created_at INTEGER NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE sync_log_202605 PARTITION OF sync_log_partitioned
    FOR VALUES FROM (1748707200000) TO (1748822400000);  -- 2026-05

CREATE TABLE sync_log_202606 PARTITION OF sync_log_partitioned
    FOR VALUES FROM (1748822400000) TO (1751500800000);  -- 2026-06

-- 按空间哈希分区: object_versions
CREATE TABLE object_versions_partitioned (
    id TEXT NOT NULL,
    object_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    created_at INTEGER NOT NULL
) PARTITION BY HASH (space_id);

CREATE TABLE object_versions_p0 PARTITION OF object_versions_partitioned
    FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE object_versions_p1 PARTITION OF object_versions_partitioned
    FOR VALUES WITH (MODULUS 4, REMAINDER 1);
-- ...

-- 自动创建分区的函数 (每月运行)
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
    partition_date TEXT;
    partition_name TEXT;
    start_ts BIGINT;
    end_ts BIGINT;
BEGIN
    partition_date := to_char(now(), 'YYYYMM');
    partition_name := 'sync_log_' || partition_date;
    start_ts := extract(epoch FROM date_trunc('month', now())) * 1000;
    end_ts := extract(epoch FROM date_trunc('month', now()) + interval '1 month') * 1000;
    
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF sync_log_partitioned
         FOR VALUES FROM (%s) TO (%s)',
        partition_name, start_ts, end_ts
    );
END;
$$ LANGUAGE plpgsql;
```

---

## 9. 数据库初始化与迁移

### 9.1 本地 SQLite 初始化

```sql
-- 开启 WAL 模式 (性能)
PRAGMA journal_mode = WAL;
-- 开启外键约束
PRAGMA foreign_keys = ON;
-- 32MB 缓存
PRAGMA cache_size = -32768;
-- 同步模式 (NORMAL 平衡性能和安全)
PRAGMA synchronous = NORMAL;
-- 临时表内存存储
PRAGMA temp_store = MEMORY;
```

### 9.2 迁移策略

| 阶段 | 策略 | 说明 |
|------|------|------|
| MVP | `CREATE TABLE IF NOT EXISTS` + 版本号 | 简单直接，通过 app_version 检查 |
| Alpha 阶段 | 手动 DDL 脚本 | 不向后兼容，数据可重建 |
| Beta 阶段 | 增量 DDL 迁移 | 使用 migration 表跟踪，仅 ADD/ALTER |
| 正式版 | 谨慎迁移 | 引入版本号兼容，前向兼容 |

```sql
-- 迁移记录表 (本地 + 云端共用)
CREATE TABLE schema_migrations (
    version     INTEGER PRIMARY KEY,        -- 版本号 (递增)
    name        TEXT NOT NULL,               -- 迁移名称
    applied_at  INTEGER NOT NULL             -- 应用时间
);

-- 迁移示例: V002 添加字段
-- 检查迁移是否已执行，通过 app 层判断
```

---

## 10. 查询模式与优化建议

### 10.1 知识图谱查询

```sql
-- 获取某个对象的所有关联 (广度优先, 限制深度 3)
WITH RECURSIVE graph_traverse AS (
    -- 起点
    SELECT id, title, type_id, 0 AS depth
    FROM objects WHERE id = ? AND is_deleted = 0
    
    UNION ALL
    
    -- 递归展开关联
    SELECT DISTINCT o.id, o.title, o.type_id, gt.depth + 1
    FROM graph_traverse gt
    JOIN relations r ON r.source_id = gt.id OR r.target_id = gt.id
    JOIN objects o ON (o.id = CASE WHEN r.source_id = gt.id THEN r.target_id ELSE r.source_id END)
    WHERE gt.depth < 3
      AND o.is_deleted = 0
)
SELECT DISTINCT * FROM graph_traverse ORDER BY depth;
```

### 10.2 智能体消息轮询 (推荐带优先级)

```sql
-- 获取待处理的高优消息 (消息队列消费)
SELECT * FROM agent_messages
WHERE target = ?
  AND status = 'pending'
  AND ttl + timestamp > ?
ORDER BY priority DESC, timestamp ASC
LIMIT 10
FOR UPDATE SKIP LOCKED;          -- PostgreSQL 15+ 跳过锁冲突
```

### 10.3 同步差异检测

```sql
-- 获取本地缺失的远程变更
SELECT sl.* FROM sync_log sl
WHERE sl.space_id = ?
  AND sl.object_id = ?
  AND sl.version > (
      SELECT COALESCE(last_version, 0)
      FROM sync_checkpoint
      WHERE space_id = ? AND device_id = ? AND object_type = 'object'
  )
ORDER BY sl.version ASC;
```

### 10.4 对象属性筛选 (EAV 模式)

```sql
-- 查询 "状态=进行中" AND "优先级=高" 的所有项目对象
SELECT o.id, o.title
FROM objects o
WHERE o.space_id = ?
  AND o.type_id = 'project-type-id'
  AND o.is_deleted = 0
  AND EXISTS (
      SELECT 1 FROM object_properties op
      WHERE op.object_id = o.id
        AND op.field_id = 'status-field-id'
        AND op.value_text = '进行中'
  )
  AND EXISTS (
      SELECT 1 FROM object_properties op
      WHERE op.object_id = o.id
        AND op.field_id = 'priority-field-id'
        AND op.value_text = '高'
  )
ORDER BY o.updated_at DESC
LIMIT 50;
```

### 10.5 回收站自动清理

```sql
-- 每小时运行: 清理过期回收站条目
DELETE FROM trash WHERE auto_delete_at < strftime('%s','now') * 1000;
-- 对于关联的对象，执行物理删除
DELETE FROM objects WHERE id IN (
    SELECT object_id FROM trash WHERE auto_delete_at < strftime('%s','now') * 1000
);
```

---

## 11. 本地数据库文件布局

```
~/.nextm/
├── data/
│   ├── main.db              # 主数据库 (SQLite, 含 WAL/SHM)
│   ├── objects_fts.db        # FTS5 全文搜索 (单独文件)
│   ├── vectors/              # LanceDB 向量索引目录
│   │   ├── _metadata/
│   │   └── objects.lance/
│   ├── files/                # 附件文件存储
│   │   ├── images/           # 图片
│   │   ├── documents/        # 文档
│   │   └── audio/            # 音频
│   ├── cache/                # AI 响应缓存
│   │   └── ai_cache.db
│   └── temp/                 # 临时处理文件
│       └── vision/           # OCR 中间结果
├── config/
│   ├── preferences.json      # 用户偏好
│   └── accounts.json         # 账号配置缓存
└── logs/
    ├── sync.log              # 同步日志
    └── error.log             # 错误日志
```

---

## 12. SQL 兼容层设计

### 12.1 跨平台 SQL 函数抽象

由于 SQLite 和 PostgreSQL 在 SQL 语法上存在差异，定义统一访问层：

```go
// Dialect SQL 方言抽象接口 (Go 实现)
type Dialect interface {
    // JSONExtract JSON 字段提取:
    //   SQLite: json_extract(field, '$.key')
    //   PG:     field->>'key'
    JSONExtract(field, path string) string

    // CurrentTimestamp 当前时间戳 (Unix ms):
    //   SQLite: (strftime('%s','now') * 1000)
    //   PG:     (extract(epoch from now()) * 1000)::bigint
    CurrentTimestamp() string

    // Boolean 布尔值:
    //   SQLite: 0/1
    //   PG:     true/false
    Boolean(v bool) string

    // Paginate 分页: LIMIT ? OFFSET ?
    Paginate(limit, offset int) string
}

// SQLiteDialect SQLite 方言实现
type SQLiteDialect struct{}

func (d *SQLiteDialect) JSONExtract(field, path string) string {
    return fmt.Sprintf(`json_extract(%s, '$.%s')`, field, path)
}

func (d *SQLiteDialect) CurrentTimestamp() string {
    return `(strftime('%s','now') * 1000)`
}

func (d *SQLiteDialect) Boolean(v bool) string {
    if v {
        return "1"
    }
    return "0"
}

func (d *SQLiteDialect) Paginate(limit, offset int) string {
    return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

// PGDialect PostgreSQL 方言实现
type PGDialect struct{}

func (d *PGDialect) JSONExtract(field, path string) string {
    return fmt.Sprintf(`%s->>'%s'`, field, path)
}

func (d *PGDialect) CurrentTimestamp() string {
    return `(extract(epoch from now()) * 1000)::bigint`
}

func (d *PGDialect) Boolean(v bool) string {
    if v {
        return "true"
    }
    return "false"
}

func (d *PGDialect) Paginate(limit, offset int) string {
    return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
```

---

## 13. 数据库表总览

| # | 表名 | 类型 | 说明 | 核心关系 |
|---|------|------|------|---------|
| 1 | `accounts` | 云端 | 用户账号 | — |
| 2 | `oauth_accounts` | 云端 | OAuth 三方绑定 | N:1 accounts |
| 3 | `refresh_tokens` | 云端 | 刷新令牌 | N:1 accounts |
| 4 | `mfa_devices` | 云端 | MFA 设备 | N:1 accounts |
| 5 | `spaces` | 本地+云端 | 工作区 | N:1 accounts |
| 6 | `space_members` | 云端 | 空间成员 | N:1 spaces, N:1 accounts |
| 7 | `subscriptions` | 云端 | 订阅信息 | 1:1 accounts |
| 8 | `account_devices` | 云端 | 设备管理 | N:1 accounts |
| 9 | `user_preferences` | 云端 | 用户偏好 | N:1 accounts, N:1 spaces |
| 10 | `object_types` | 本地+云端 | 类型定义 | N:1 spaces |
| 11 | `type_fields` | 本地+云端 | 类型字段 | N:1 object_types |
| 12 | `type_relations` | 本地+云端 | 类型关系 | N:1 object_types |
| 13 | `objects` | 本地+云端 | **知识对象 (核心)** | N:1 spaces, N:1 object_types |
| 14 | `blocks` | 本地+云端 | 内容块 | N:1 objects, 自引用 |
| 15 | `object_properties` | 本地+云端 | 对象属性 EAV | N:1 objects, N:1 type_fields |
| 16 | `object_versions` | 本地+云端 | 对象版本历史 | N:1 objects |
| 17 | `trash` | 本地+云端 | 回收站 | N:1 spaces |
| 18 | `relations` | 本地+云端 | 对象间关系 | N:1 spaces, N:M objects |
| 19 | `tags` | 本地+云端 | 标签 | N:1 spaces, 自引用 |
| 20 | `object_tags` | 本地+云端 | 对象-标签关联 | N:M objects/tags |
| 21 | `collections` | 本地+云端 | 数据库集合 | N:1 spaces |
| 22 | `collection_views` | 本地+云端 | 集合视图 | N:1 collections |
| 23 | `collection_items` | 本地+云端 | 手动集合条目 | N:1 collections, N:1 objects |
| 24 | `templates` | 本地+云端 | 模板 | N:1 spaces, N:1 object_types |
| 25 | `capture_sessions` | 本地+云端 | 采集会话 | N:1 spaces |
| 26 | `image_queue` | 本地+云端 | 图片处理队列 | N:1 spaces, N:1 capture_sessions |
| 27 | `vision_jobs` | 云端 | 视觉处理任务 | N:1 spaces, N:1 image_queue |
| 28 | `agents` | 本地+云端 | 智能体定义 | N:1 spaces |
| 29 | `agent_runs` | 本地+云端 | 智能体运行记录 | N:1 agents, N:1 spaces |
| 30 | `agent_messages` | 云端 | 消息总线持久化 | N:1 spaces |
| 31 | `long_horizon_agents` | 本地+云端 | 长周期智能体 | N:1 spaces |
| 32 | `agent_steps` | 本地+云端 | 长周期智能体步骤 | N:1 long_horizon_agents |
| 33 | `agent_memory` | 本地+云端 | 智能体长期记忆 | N:1 agents |
| 34 | `sync_log` | 本地+云端 | CRDT 同步日志 | — |
| 35 | `sync_checkpoint` | 本地+云端 | 同步检查点 | — |
| 36 | `sync_conflicts` | 云端 | 同步冲突记录 | N:1 spaces |
| 37 | `files` | 本地+云端 | 文件存储记录 | N:1 spaces, N:1 objects |
| 38 | `file_chunks` | 云端 | 大文件分片 | N:1 files |
| 39 | `ai_cache` | 本地+云端 | AI 响应缓存 | N:1 spaces |
| 40 | `ai_usage_logs` | 云端 | AI 用量日志 | N:1 accounts |
| 41 | `search_history` | 本地+云端 | 搜索历史 | N:1 spaces |
| 42 | `mcp_connectors` | 云端 | MCP 连接器 | N:1 spaces |
| 43 | `mcp_audit_log` | 云端 | MCP 审计日志 | N:1 mcp_connectors |
| 44 | `notifications` | 云端 | 通知 | N:1 spaces, N:1 accounts |
| 45 | `share_links` | 云端 | 分享链接 | N:1 spaces, N:1 objects |
| 46 | `export_jobs` | 云端 | 导出任务 | N:1 spaces, N:1 accounts |
| 47 | `schema_migrations` | 本地+云端 | 数据库迁移记录 | — |

---

## 14. ER 关系一览

### 层级关系链

```
Account ──1:N──> Space ──1:N──> ObjectType ──1:N──> TypeField
                            │                      └──1:N──> TypeRelation
                            ├──1:N──> Object ──1:N──> Block
                            │         │         └──N:M──> Tag (via object_tags)
                            │         │         └──N:M──> Object (via relations)
                            │         │         └──N:1──> ObjectVersion
                            │         │         └──1:N──> ObjectProperty
                            │         │
                            ├──1:N──> Collection ──1:N──> CollectionView
                            │         └──N:M──> Object (via collection_items)
                            ├──1:N──> Tag (自引用 parent_id)
                            ├──1:N──> Agent ──1:N──> AgentRun
                            │         └──1:N──> AgentMemory
                            ├──1:N──> Template
                            ├──1:N──> LongHorizonAgent ──1:N──> AgentStep
                            ├──1:N──> CaptureSession ──1:N──> ImageQueue
                            └──1:N──> File ──1:N──> FileChunk
```

### 多对多关系

```
Object ────<object_tags>──── Tag
Object ────<relations>─────── Object (自关联, N:M)
Collection ──<collection_items>── Object
```

---

## 附录: 快速部署 SQL 脚本

完整的建表 SQL 脚本可拆分为:

- `schema_local.sql` — SQLite 本地库 (表 5,10-37,39-47, 不含云端专用表)
- `schema_cloud.sql` — PostgreSQL 云端库 (所有表)
- `migration_scripts/` — 增量迁移脚本目录

每个脚本包含:
1. 表创建语句 (含 IF NOT EXISTS)
2. 索引创建
3. 触发器创建 (FTS5 同步)
4. 初始种子数据 (内置类型、模板)
