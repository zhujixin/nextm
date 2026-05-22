# 安全设计说明书

本文档覆盖 NextM 的安全架构、数据保护策略和隐私合规要求。

---

## 1. 安全原则

- **最小权限**：每个组件只获得完成任务所需的最小权限
- **纵深防御**：多层安全控制，单层失效不导致全系统暴露
- **零信任架构**：所有请求需验证身份，无论来源
- **本地优先隐私**：用户数据默认存储在本地，云端仅作为同步副本
- **加密默认开启**：传输加密和静态加密默认启用

---

## 2. 数据分类

| 分类 | 定义 | 示例 | 存储要求 |
|------|------|------|---------|
| 公开 | 用户主动公开分享 | 分享链接 | 标准加密 |
| 私有 | 用户个人数据 | 笔记、图片、录音 | AES-256-GCM 加密 |
| 凭证 | 认证与密钥 | JWT Token、API Key | 加密 + 硬件隔离 |
| 元数据 | 使用数据 | 操作日志、统计 | 标准加密 |

---

## 3. 传输安全

### 3.1 TLS

- 所有外部 API 通信强制 TLS 1.3
- 内部服务间通信（gRPC）启用 mTLS
- 证书自动轮转（Cert-Manager / Let's Encrypt）
- HSTS 头：`Strict-Transport-Security: max-age=31536000`

### 3.2 WebSocket

- 连接使用 WSS（WebSocket over TLS）
- 连接建立时验证 JWT Token
- 消息级别签名校验（HMAC-SHA256）

---

## 4. 存储加密

### 4.1 本地存储

```
┌─────────────────────┐
│    应用沙箱目录       │
│  ┌────────────────┐  │
│  │ SQLite 数据库   │←── AES-256-GCM 加密
│  │ (加密文件)       │  │
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ 附件文件         │←── AES-256-GCM 加密
│  │ (图片/音频等)    │  │
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ LanceDB 向量索引 │←── 索引数据加密
│  └────────────────┘  │
└─────────────────────┘
```

- 加密密钥派生自主密码（Argon2id）
- 主密码永不上传到服务器
- 密钥存储在操作系统密钥链（Keychain / Credential Manager）

### 4.2 云端存储

- PostgreSQL 存储层加密（TDE 或列级加密）
- 附件存储使用 S3 服务端加密（SSE-S3 / SSE-KMS）
- 数据库连接使用 TLS

### 4.3 密钥管理

```
用户主密码
    ↓ Argon2id
派生密钥 (KEK)
    ↓ HKDF
数据加密密钥 (DEK) ──→ AES-256-GCM 加密数据
    ↓ HKDF
HMAC 密钥 ──→ 消息签名
```

---

## 5. 认证与授权

### 5.1 认证方案

| 方法 | 适用场景 | 安全等级 |
|------|---------|---------|
| 邮箱 + 密码 | Web / Desktop 登录 | 标准 |
| OAuth 2.0 (Google/GitHub) | 第三方登录 | 高 |
| 生物认证 | 移动端解锁 | 高 |
| API Key | 开发者 API 调用 | 高 |

### 5.2 JWT 策略

```yaml
access_token:
  type: JWT (HS256/RS256)
  expire: 15 分钟
  storage: 内存 / HTTP Only Cookie

refresh_token:
  type: Opaque Token
  expire: 7 天
  rotation: 每次使用轮转
  storage: localStorage / Secure Cookie

recovery_token:
  type: Opaque Token
  expire: 1 小时
  usage: 密码重置
```

### 5.3 密码策略

- 最小长度：8 字符
- 建议使用密码管理器生成的随机密码
- 密码哈希：bcrypt (cost=12) / Argon2id
- 登录限流：5 次失败后锁定 15 分钟

### 5.4 权限模型

```
Account ── 1:N ── Space ── N:N ── Member
  │                    │
  │                    ├── owner（完全控制）
  │                    ├── admin（管理成员、内容）
  │                    ├── editor（读写内容）
  │                    └── viewer（只读）
  │
  └── Personal Space（独有）
```

---

## 6. API 安全

### 6.1 防护措施

| 措施 | 实现 |
|------|------|
| CSRF 防护 | SameSite=Strict Cookie + CSRF Token |
| XSS 防护 | Content-Security-Policy Header |
| Rate Limiting | 令牌桶算法，按用户/路由分级 |
| Request Body 限制 | 最大 10MB |
| CORS | 白名单域名配置 |
| SQL 注入防护 | SQLc 参数化查询（编译期保障） |

### 6.2 安全头

```
Content-Security-Policy: default-src 'self'; script-src 'self'; ...
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Referrer-Policy: strict-origin-when-cross-origin
```

---

## 7. 多智能体安全

- 智能体间通信经过 NATS 消息总线，受 JetStream 权限控制
- 智能体只能访问所属 Space 的数据
- 智能体执行的操作记录审计日志
- 外部 MCP 连接需要用户显式授权
- 智能体 token 使用量限流和预算控制

---

## 8. 隐私合规

### 8.1 合规框架适配

| 法规 | 适用范围 | 关键要求 |
|------|---------|---------|
| GDPR | 欧盟用户 | 数据可删除、可导出、处理同意 |
| CCPA | 加州用户 | 数据披露、删除选择权 |
| PIPL | 中国用户 | 数据本地化、处理同意、最小必要 |
| LGPD | 巴西用户 | GDPR 类似 |

### 8.2 数据主体权利

| 权利 | 实现方式 |
|------|---------|
| 知情权 | 隐私政策清晰说明数据处理范围 |
| 访问权 | 用户可导出所有数据（JSON/Markdown） |
| 删除权 | 账号删除彻底清除所有关联数据 |
| 可携带权 | 导出为标准格式（Markdown + YAML frontmatter） |
| 撤回同意 | 随时停止数据同步/AI 处理 |

### 8.3 数据生命周期

```
采集 ──→ 保存 ──→ 使用 ──→ 归档 ──→ 删除
 │         │        │        │        │
 用户控制  加密    合规使用   自动归档  彻底擦除
```

---

## 9. 审计与日志

### 9.1 审计事件

| 类别 | 事件示例 | 保留期 |
|------|---------|--------|
| 认证 | 登录成功/失败、密码修改 | 90 天 |
| 数据访问 | 导出、删除、批量操作 | 90 天 |
| 权限变更 | 邀请成员、角色修改 | 180 天 |
| 系统操作 | 智能体执行、API Key 创建 | 180 天 |

### 9.2 日志保护

- 审计日志只增不删（append-only）
- 日志传输加密
- 日志文件权限最小化
- 不记录敏感字段（密码、Token、内容全文）

---

## 10. 漏洞处理流程

```
1. 发现 ← 内部/外部安全研究员报告
2. 评估 ← 确认并定级（Critical/High/Medium/Low）
3. 修复 ← 开发修复补丁
4. 发布 ← 安全版本发布 + CVE 公开
5. 复盘 ← 根因分析 + 防护措施改进
```

### 响应时间

| 严重度 | 修复时间 | 通知方式 |
|--------|---------|---------|
| Critical | 48 小时内 | 邮件 + 应用内通知 |
| High | 7 天内 | 邮件通知 |
| Medium | 30 天内 | 发布说明 |
| Low | 下次发布 | — |

---

## 11. 安全测试

| 测试类型 | 频率 | 工具 |
|---------|------|------|
| SAST（静态分析） | 每次 commit | golangci-lint (gosec) |
| SCA（依赖扫描） | 每日 | Dependabot / Trivy |
| DAST（动态扫描） | 每次发布 | OWASP ZAP |
| 密钥扫描 | 每次 commit | GitLeaks / TruffleHog |
| 渗透测试 | 每季度 | 外部安全团队 |
| Fuzz 测试 | 持续 | go test -fuzz |

---

## 12. 安全 Checklist（发布前）

- [ ] SAST 扫描无 Critical/High 发现
- [ ] 依赖扫描无已知 CVE
- [ ] 密钥扫描未发现泄露
- [ ] 新 API 端点经过权限验证
- [ ] 新端点有 Rate Limiting 保护
- [ ] 所有组件使用 TLS 通信
- [ ] 日志不包含敏感数据
- [ ] 安全头已配置
- [ ] 数据库迁移回滚已验证
