-- 000001_initial.up.sql
-- NextM 初始数据库迁移 — SQLite 方言
-- 按业务域分组，共 47 张表

-- ─── 空间与账号体系 ────────────────────────────────────

CREATE TABLE IF NOT EXISTS spaces (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    type            TEXT NOT NULL CHECK(type IN ('personal','team')),
    account_id      TEXT NOT NULL,
    icon            TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    encrypted       INTEGER DEFAULT 0,
    encryption_key  TEXT DEFAULT '',
    settings        TEXT DEFAULT '{}',
    object_count    INTEGER DEFAULT 0,
    sync_status     TEXT DEFAULT 'synced',
    is_deleted      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    id              TEXT PRIMARY KEY,
    email           TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT DEFAULT '',
    auth_provider   TEXT DEFAULT 'email',
    password_hash   TEXT DEFAULT '',
    mfa_enabled     INTEGER DEFAULT 0,
    mfa_secret      TEXT DEFAULT '',
    locale          TEXT DEFAULT 'zh-CN',
    timezone        TEXT DEFAULT 'Asia/Shanghai',
    is_active       INTEGER DEFAULT 1,
    last_login_at   INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS oauth_accounts (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,
    provider_id     TEXT NOT NULL,
    email           TEXT DEFAULT '',
    avatar_url      TEXT DEFAULT '',
    access_token    TEXT DEFAULT '',
    refresh_token   TEXT DEFAULT '',
    expires_at      INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(provider, provider_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,
    device_id       TEXT DEFAULT '',
    expires_at      INTEGER NOT NULL,
    revoked         INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS mfa_devices (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_type     TEXT NOT NULL CHECK(device_type IN ('totp','sms','email','recovery')),
    secret          TEXT DEFAULT '',
    verified        INTEGER DEFAULT 0,
    last_used_at    INTEGER,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS space_members (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK(role IN ('owner','editor','commenter','reader')),
    invite_email    TEXT DEFAULT '',
    invite_status   TEXT DEFAULT 'accepted' CHECK(invite_status IN ('pending','accepted','declined','revoked')),
    joined_at       INTEGER,
    created_at      INTEGER NOT NULL,
    UNIQUE(space_id, account_id)
);

CREATE TABLE IF NOT EXISTS account_devices (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_name     TEXT DEFAULT '',
    device_type     TEXT DEFAULT '',
    device_fingerprint TEXT NOT NULL,
    public_key      TEXT DEFAULT '',
    push_token      TEXT DEFAULT '',
    last_sync_at    INTEGER,
    last_ip         TEXT DEFAULT '',
    is_current      INTEGER DEFAULT 0,
    is_active       INTEGER DEFAULT 1,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    plan_type       TEXT NOT NULL CHECK(plan_type IN ('free','pro','team','self_hosted')),
    status          TEXT NOT NULL CHECK(status IN ('active','canceled','past_due','expired')),
    ai_tokens_limit INTEGER DEFAULT 0,
    ai_tokens_used  INTEGER DEFAULT 0,
    storage_limit   INTEGER DEFAULT 104857600,
    storage_used    INTEGER DEFAULT 0,
    expires_at      INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS user_preferences (
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    preferences     TEXT DEFAULT '{}',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(account_id, space_id)
);

-- ─── 对象类型系统 ──────────────────────────────────────

CREATE TABLE IF NOT EXISTS object_types (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    icon            TEXT DEFAULT '',
    color           TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    is_builtin      INTEGER DEFAULT 0,
    is_archived     INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS type_fields (
    id              TEXT PRIMARY KEY,
    type_id         TEXT NOT NULL REFERENCES object_types(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    field_type      TEXT NOT NULL CHECK(field_type IN (
                        'text','number','date','select','multi_select',
                        'relation','rollup','formula','file','email',
                        'url','phone','progress','currency','rating'
                    )),
    position        REAL NOT NULL,
    required        INTEGER DEFAULT 0,
    default_value   TEXT DEFAULT '',
    options         TEXT DEFAULT '[]',
    config          TEXT DEFAULT '{}',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS type_relations (
    id              TEXT PRIMARY KEY,
    type_id         TEXT NOT NULL REFERENCES object_types(id) ON DELETE CASCADE,
    target_type_id  TEXT NOT NULL,
    relation_type   TEXT NOT NULL CHECK(relation_type IN ('one_to_one','one_to_many','many_to_many')),
    reverse_name    TEXT DEFAULT '',
    created_at      INTEGER NOT NULL
);

-- ─── 知识对象与内容 ────────────────────────────────────

CREATE TABLE IF NOT EXISTS objects (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    type_id         TEXT NOT NULL REFERENCES object_types(id),
    title           TEXT NOT NULL DEFAULT '',
    properties      TEXT DEFAULT '{}',
    tags            TEXT DEFAULT '[]',
    cover_image     TEXT,
    source          TEXT NOT NULL CHECK(source IN (
                        'manual','video','web','camera','audio',
                        'import','clipboard','email','agent'
                    )),
    source_meta     TEXT DEFAULT '{}',
    embedding_id    TEXT,
    word_count      INTEGER DEFAULT 0,
    version         INTEGER DEFAULT 1,
    is_archived     INTEGER DEFAULT 0,
    is_deleted      INTEGER DEFAULT 0,
    last_read_at    INTEGER,
    sync_status     TEXT DEFAULT 'synced' CHECK(sync_status IN ('synced','pending','conflict','error')),
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS blocks (
    id              TEXT PRIMARY KEY,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    parent_id       TEXT REFERENCES blocks(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK(type IN (
                        'text','heading1','heading2','heading3','bullet_list',
                        'ordered_list','toggle','quote','code','callout',
                        'image','video','audio','file','bookmark','equation',
                        'mermaid','table','divider','breadcrumb','toc','canvas','embed'
                    )),
    content         TEXT DEFAULT '',
    properties      TEXT DEFAULT '{}',
    position        REAL NOT NULL,
    depth           INTEGER DEFAULT 0,
    collapsed       INTEGER DEFAULT 0,
    color           TEXT DEFAULT '',
    version         INTEGER DEFAULT 1,
    sync_status     TEXT DEFAULT 'synced' CHECK(sync_status IN ('synced','pending','conflict','error')),
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS object_properties (
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    field_id        TEXT NOT NULL REFERENCES type_fields(id) ON DELETE CASCADE,
    value_text      TEXT DEFAULT '',
    value_number    REAL,
    value_date      INTEGER,
    value_ref       TEXT,
    UNIQUE(object_id, field_id)
);

CREATE TABLE IF NOT EXISTS object_versions (
    id              TEXT PRIMARY KEY,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    version         INTEGER NOT NULL,
    title           TEXT DEFAULT '',
    content_snapshot TEXT DEFAULT '{}',
    change_summary  TEXT DEFAULT '',
    device_id       TEXT DEFAULT '',
    account_id      TEXT DEFAULT '',
    diff            TEXT DEFAULT '{}',
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS trash (
    id              TEXT PRIMARY KEY,
    object_id       TEXT NOT NULL,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_type     TEXT NOT NULL,
    data            TEXT NOT NULL,
    deleted_by      TEXT DEFAULT '',
    auto_delete_at  INTEGER,
    created_at      INTEGER NOT NULL
);

-- ─── 关系与标签 ────────────────────────────────────────

CREATE TABLE IF NOT EXISTS relations (
    id              TEXT PRIMARY KEY,
    source_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    target_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK(type IN (
                        'link','reference','citation','parent','child',
                        'related','custom','sequential'
                    )),
    custom_type_id  TEXT,
    weight          REAL DEFAULT 0.5 CHECK(weight >= 0 AND weight <= 1),
    metadata        TEXT DEFAULT '{}',
    ai_generated    INTEGER DEFAULT 0,
    source_object_type TEXT DEFAULT '',
    target_object_type TEXT DEFAULT '',
    created_at      INTEGER NOT NULL,
    UNIQUE(source_id, target_id, type)
);

CREATE TABLE IF NOT EXISTS tags (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    color           TEXT DEFAULT '',
    parent_id       TEXT REFERENCES tags(id) ON DELETE SET NULL,
    ai_generated    INTEGER DEFAULT 0,
    object_count    INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(space_id, name)
);

CREATE TABLE IF NOT EXISTS object_tags (
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    tag_id          TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (object_id, tag_id)
);

-- ─── 集合与视图 ────────────────────────────────────────

CREATE TABLE IF NOT EXISTS collections (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    source_type     TEXT NOT NULL CHECK(source_type IN ('type','tag','mixed','manual','saved_search')),
    source_config   TEXT DEFAULT '{}',
    layout          TEXT DEFAULT 'table',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS collection_views (
    id              TEXT PRIMARY KEY,
    collection_id   TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    view_type       TEXT NOT NULL CHECK(view_type IN (
                        'table','kanban','gallery','calendar','timeline','list'
                    )),
    filters         TEXT DEFAULT '[]',
    sorts           TEXT DEFAULT '[]',
    visible_fields  TEXT DEFAULT '[]',
    group_by        TEXT DEFAULT '',
    calendar_field  TEXT DEFAULT '',
    kanban_field    TEXT DEFAULT '',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS collection_items (
    id              TEXT PRIMARY KEY,
    collection_id   TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    position        REAL DEFAULT 0,
    note            TEXT DEFAULT '',
    UNIQUE(collection_id, object_id)
);

-- ─── 模板 ─────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS templates (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT DEFAULT '',
    icon            TEXT DEFAULT '',
    type_id         TEXT REFERENCES object_types(id),
    content         TEXT DEFAULT '{}',
    is_builtin      INTEGER DEFAULT 0,
    category        TEXT DEFAULT '',
    use_count       INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

-- ─── 采集与视觉 ────────────────────────────────────────

CREATE TABLE IF NOT EXISTS capture_sessions (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    source          TEXT NOT NULL CHECK(source IN (
                        'screenshot','camera','clipboard','system_audio','screen_record'
                    )),
    status          TEXT DEFAULT 'pending' CHECK(status IN ('pending','processing','completed','failed')),
    item_count      INTEGER DEFAULT 0,
    device_id       TEXT DEFAULT '',
    metadata        TEXT DEFAULT '{}',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS image_queue (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    session_id      TEXT REFERENCES capture_sessions(id),
    file_path       TEXT NOT NULL,
    file_hash       TEXT DEFAULT '',
    status          TEXT DEFAULT 'queued' CHECK(status IN (
                        'queued','processing','completed','failed','duplicate'
                    )),
    priority        INTEGER DEFAULT 0,
    quality_score   REAL,
    ocr_text        TEXT DEFAULT '',
    ocr_confidence  REAL,
    layout_data     TEXT DEFAULT '{}',
    error_message   TEXT DEFAULT '',
    retry_count     INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS vision_jobs (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    image_id        TEXT NOT NULL REFERENCES image_queue(id) ON DELETE CASCADE,
    job_type        TEXT NOT NULL CHECK(job_type IN (
                        'ocr','extract','enhance','batch','formula','chart'
                    )),
    status          TEXT DEFAULT 'queued' CHECK(status IN ('queued','processing','completed','failed')),
    result          TEXT DEFAULT '{}',
    model_used      TEXT DEFAULT '',
    processing_time INTEGER,
    tokens_used     INTEGER DEFAULT 0,
    error_message   TEXT DEFAULT '',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

-- ─── 智能体 ────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS agents (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    agent_type      TEXT NOT NULL CHECK(agent_type IN (
                        'summary','tagging','relation','writing','review',
                        'scheduler','research','workflow','long_horizon','custom'
                    )),
    description     TEXT DEFAULT '',
    triggers        TEXT DEFAULT '[]',
    actions         TEXT DEFAULT '[]',
    config          TEXT DEFAULT '{}',
    model_preference TEXT DEFAULT '',
    enabled         INTEGER DEFAULT 1,
    run_count       INTEGER DEFAULT 0,
    success_count   INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_runs (
    id              TEXT PRIMARY KEY,
    agent_id        TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    status          TEXT NOT NULL CHECK(status IN (
                        'queued','running','completed','failed','cancelled','pending'
                    )),
    input           TEXT DEFAULT '{}',
    output          TEXT DEFAULT '{}',
    triggered_by    TEXT DEFAULT 'manual',
    tokens_used     INTEGER DEFAULT 0,
    processing_time INTEGER DEFAULT 0,
    error_message   TEXT DEFAULT '',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_messages (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    conversation_id TEXT NOT NULL,
    msg_type        TEXT NOT NULL CHECK(msg_type IN (
                        'request','response','broadcast','command','event','error'
                    )),
    source          TEXT NOT NULL,
    target          TEXT NOT NULL,
    priority        TEXT DEFAULT 'normal' CHECK(priority IN ('low','normal','high','critical')),
    payload         TEXT DEFAULT '{}',
    status          TEXT DEFAULT 'pending' CHECK(status IN ('pending','delivered','processed','failed')),
    ttl             INTEGER DEFAULT 3600,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS long_horizon_agents (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    agent_id        TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    goal            TEXT NOT NULL,
    plan            TEXT DEFAULT '[]',
    state           TEXT NOT NULL CHECK(state IN (
                        'planning','running','paused','completed','failed'
                    )),
    progress        INTEGER DEFAULT 0 CHECK(progress >= 0 AND progress <= 100),
    short_term_memory TEXT DEFAULT '[]',
    long_term_memory  TEXT DEFAULT '[]',
    deadline        INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_steps (
    id              TEXT PRIMARY KEY,
    agent_id        TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    parent_step_id  TEXT REFERENCES agent_steps(id) ON DELETE CASCADE,
    step_order      INTEGER NOT NULL,
    step_type       TEXT NOT NULL CHECK(step_type IN (
                        'collect','analyze','summarize','relate','report','decide','act'
                    )),
    status          TEXT DEFAULT 'pending' CHECK(status IN ('pending','running','completed','failed','skipped')),
    input           TEXT DEFAULT '{}',
    output          TEXT DEFAULT '{}',
    error_message   TEXT DEFAULT '',
    started_at      INTEGER,
    completed_at    INTEGER,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_memory (
    id              TEXT PRIMARY KEY,
    agent_id        TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    memory_type     TEXT NOT NULL CHECK(memory_type IN (
                        'observation','conclusion','reference','feedback','context','state'
                    )),
    content         TEXT NOT NULL,
    embedding       BLOB,
    relevance_score REAL DEFAULT 0,
    expires_at      INTEGER,
    created_at      INTEGER NOT NULL
);

-- ─── 同步 ──────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sync_log (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL,
    object_type     TEXT NOT NULL CHECK(object_type IN ('object','block','collection','type','tag','relation')),
    change_type     TEXT NOT NULL CHECK(change_type IN ('create','update','delete','move','archive','restore')),
    data            BLOB,
    version         INTEGER NOT NULL,
    device_id       TEXT NOT NULL,
    vector_clock    TEXT DEFAULT '{}',
    parent_id       TEXT,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sync_checkpoint (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    device_id       TEXT NOT NULL,
    object_type     TEXT NOT NULL,
    last_version    INTEGER DEFAULT 0,
    last_log_id     TEXT,
    snapshot        BLOB,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    UNIQUE(space_id, device_id, object_type)
);

CREATE TABLE IF NOT EXISTS sync_conflicts (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL,
    object_type     TEXT NOT NULL,
    local_version   INTEGER NOT NULL,
    remote_version  INTEGER NOT NULL,
    local_data      BLOB,
    remote_data     BLOB,
    merged_data     BLOB,
    resolution      TEXT DEFAULT 'pending' CHECK(resolution IN (
                        'auto_merged','manual_local','manual_remote','manual_custom','pending'
                    )),
    resolved_at     INTEGER,
    created_at      INTEGER NOT NULL
);

-- ─── 文件存储 ──────────────────────────────────────────

CREATE TABLE IF NOT EXISTS files (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT REFERENCES objects(id) ON DELETE SET NULL,
    filename        TEXT NOT NULL,
    file_size       INTEGER NOT NULL,
    mime_type       TEXT DEFAULT 'application/octet-stream',
    storage_key     TEXT NOT NULL,
    storage_backend TEXT DEFAULT 'local',
    file_hash       TEXT DEFAULT '',
    width           INTEGER,
    height          INTEGER,
    duration        INTEGER,
    thumbnail_key   TEXT DEFAULT '',
    is_deleted      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS file_chunks (
    file_id         TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    chunk_index     INTEGER NOT NULL,
    chunk_size      INTEGER NOT NULL,
    storage_key     TEXT NOT NULL,
    file_hash       TEXT DEFAULT '',
    UNIQUE(file_id, chunk_index)
);

-- ─── AI 缓存与用量 ─────────────────────────────────────

CREATE TABLE IF NOT EXISTS ai_cache (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    input_hash      TEXT NOT NULL,
    model           TEXT NOT NULL,
    response        TEXT NOT NULL,
    tokens_saved    INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    expires_at      INTEGER
);

CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    model           TEXT NOT NULL,
    prompt_tokens   INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens    INTEGER DEFAULT 0,
    cost_usd        REAL DEFAULT 0,
    request_type    TEXT DEFAULT '',
    created_at      INTEGER NOT NULL
);

-- ─── MCP 协议 ──────────────────────────────────────────

CREATE TABLE IF NOT EXISTS mcp_connectors (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    connector_type  TEXT NOT NULL CHECK(connector_type IN ('stdio','sse','streamable_http')),
    command         TEXT DEFAULT '',
    url             TEXT DEFAULT '',
    api_key         TEXT DEFAULT '',
    tools           TEXT DEFAULT '[]',
    resources       TEXT DEFAULT '[]',
    enabled         INTEGER DEFAULT 1,
    last_heartbeat  INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS mcp_audit_log (
    id              TEXT PRIMARY KEY,
    connector_id    TEXT NOT NULL REFERENCES mcp_connectors(id) ON DELETE CASCADE,
    action          TEXT NOT NULL CHECK(action IN ('tool_call','resource_read','auth','error','config_change')),
    tool_name       TEXT DEFAULT '',
    arguments       TEXT DEFAULT '{}',
    result_summary  TEXT DEFAULT '',
    duration_ms     INTEGER DEFAULT 0,
    ip_address      TEXT DEFAULT '',
    created_at      INTEGER NOT NULL
);

-- ─── 通知与分享 ────────────────────────────────────────

CREATE TABLE IF NOT EXISTS notifications (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type            TEXT NOT NULL CHECK(type IN (
                        'mention','comment','share','sync','system','agent','invite'
                    )),
    title           TEXT NOT NULL,
    body            TEXT DEFAULT '',
    data            TEXT DEFAULT '{}',
    is_read         INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS share_links (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    object_id       TEXT NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    share_type      TEXT NOT NULL CHECK(share_type IN ('public','password','internal')),
    password_hash   TEXT DEFAULT '',
    expires_at      INTEGER,
    max_views       INTEGER DEFAULT 0,
    view_count      INTEGER DEFAULT 0,
    is_revoked      INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

-- ─── 导出 ──────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS export_jobs (
    id              TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    format          TEXT NOT NULL CHECK(format IN ('markdown','json','pdf')),
    scope           TEXT NOT NULL CHECK(scope IN ('space','collection','object')),
    scope_id        TEXT NOT NULL,
    status          TEXT DEFAULT 'queued' CHECK(status IN ('queued','processing','completed','failed')),
    progress        INTEGER DEFAULT 0,
    download_url    TEXT DEFAULT '',
    file_size       INTEGER DEFAULT 0,
    error_message   TEXT DEFAULT '',
    created_at      INTEGER NOT NULL,
    completed_at    INTEGER
);

-- ─── 搜索历史 ──────────────────────────────────────────

CREATE TABLE IF NOT EXISTS search_history (
    id              TEXT PRIMARY KEY,
    space_id        TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    query           TEXT NOT NULL,
    search_type     TEXT NOT NULL CHECK(search_type IN ('fulltext','semantic','hybrid','deep')),
    result_count    INTEGER DEFAULT 0,
    created_at      INTEGER NOT NULL
);

-- ─── 索引 ──────────────────────────────────────────────

-- 活跃对象索引
CREATE INDEX IF NOT EXISTS idx_objects_list ON objects(space_id, is_deleted, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_objects_type_tags ON objects(space_id, type_id, is_deleted);
CREATE INDEX IF NOT EXISTS idx_active_objects ON objects(space_id, type_id) WHERE is_deleted = 0;

-- 块索引
CREATE INDEX IF NOT EXISTS idx_blocks_object ON blocks(object_id, position);
CREATE INDEX IF NOT EXISTS idx_blocks_parent ON blocks(parent_id);

-- 关系索引
CREATE INDEX IF NOT EXISTS idx_relations_source ON relations(source_id);
CREATE INDEX IF NOT EXISTS idx_relations_target ON relations(target_id);

-- 标签索引
CREATE INDEX IF NOT EXISTS idx_tags_space ON tags(space_id);
CREATE INDEX IF NOT EXISTS idx_object_tags_tag ON object_tags(tag_id);

-- 集合索引
CREATE INDEX IF NOT EXISTS idx_collections_space ON collections(space_id);

-- 采集队列索引
CREATE INDEX IF NOT EXISTS idx_image_queue_status ON image_queue(status, priority);

-- 智能体索引
CREATE INDEX IF NOT EXISTS idx_enabled_agents ON agents(space_id) WHERE enabled = 1;
CREATE INDEX IF NOT EXISTS idx_agent_runs_agent ON agent_runs(agent_id, created_at DESC);

-- 长周期任务索引
CREATE INDEX IF NOT EXISTS idx_running_lha ON long_horizon_agents(space_id) WHERE state = 'running';

-- 同步日志索引
CREATE INDEX IF NOT EXISTS idx_sync_log_object ON sync_log(object_id, version);
CREATE INDEX IF NOT EXISTS idx_sync_log_space ON sync_log(space_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_log_device ON sync_log(device_id, created_at DESC);

-- 文件索引
CREATE INDEX IF NOT EXISTS idx_files_space ON files(space_id, is_deleted);

-- 通知索引
CREATE INDEX IF NOT EXISTS idx_notifications_account ON notifications(account_id, is_read, created_at DESC);

-- AI 缓存索引
CREATE INDEX IF NOT EXISTS idx_ai_cache_hash ON ai_cache(input_hash, model);

-- 导出任务索引
CREATE INDEX IF NOT EXISTS idx_export_jobs_account ON export_jobs(account_id, created_at DESC);

-- 搜索历史索引
CREATE INDEX IF NOT EXISTS idx_search_history_space ON search_history(space_id, created_at DESC);
