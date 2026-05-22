# 前端架构设计

本文档覆盖 NextM 四个客户端的前端架构：Web 应用、桌面端（Tauri）、移动端（Flutter）、浏览器扩展。

---

## 1. 整体架构

### 1.1 四端概览

```
后端 API (REST + gRPC + WebSocket)
        ↕
┌──────────────────────────────────┐
│         前端客户端层              │
│                                  │
│  ┌──────┐ ┌──────┐ ┌────────┐  │
│  │ Web  │ │桌面端 │ │ 移动端  │  │
│  │React │ │Tauri │ │Flutter │  │
│  └──────┘ └──────┘ └────────┘  │
│  ┌──────────────────────────┐   │
│  │     浏览器扩展             │   │
│  └──────────────────────────┘   │
└──────────────────────────────────┘
```

### 1.2 跨端能力矩阵

| 功能 | Web | Desktop | Mobile | Extension |
|------|-----|---------|--------|-----------|
| 笔记编辑 | ✅ | ✅ | ✅ | ❌ |
| 快速剪藏 | ❌ | ✅ | ✅ | ✅ |
| 离线使用 | 有限 | ✅ | ✅ | ❌ |
| 视频截屏 | ❌ | ✅ | ✅ | ❌ |
| 语音速记 | ❌ | ✅ | ✅ | ❌ |
| 文件拖拽 | ✅ | ✅ | ✅ | ❌ |
| 系统托盘 | ❌ | ✅ | ❌ | ❌ |
| 全局快捷键 | ❌ | ✅ | ❌ | ✅ |

### 1.3 通用原则

- **本地优先**：所有客户端优先读写本地 SQLite，同步引擎后台处理云端同步
- **渐进式 Web**：Web 端作为快速入口，复杂功能下沉到桌面端/移动端
- **API 驱动**：四端共用同一套 REST + WebSocket API
- **组件复用**：Web 与 Desktop 共享 React 组件代码

---

## 2. Web 应用 (React 19 + Vite + Tailwind CSS 4)

### 2.1 技术栈

| 领域 | 选择 | 理由 |
|------|------|------|
| 框架 | React 19 | 成熟的生态、Server Component 支持 |
| 构建 | Vite 6 | 极速 HMR、ESM 原生 |
| 样式 | Tailwind CSS 4 | Utility-first、零运行时 |
| 路由 | TanStack Router | 类型安全的路由、懒加载 |
| 状态管理 | Zustand + React Query | 轻量、服务端状态缓存 |
| 编辑器 | TipTap (ProseMirror) | 扩展性强、块级编辑 |
| 国际化 | react-i18next | 成熟的 i18n 方案 |
| 测试 | Vitest + Testing Library | 与 Vite 深度集成 |

### 2.2 目录结构

```
frontend/web/
├── src/
│   ├── main.tsx              # 入口
│   ├── App.tsx               # 路由 & 布局
│   ├── routes/               # TanStack Router 路由定义
│   ├── core/                 # 基础设施
│   │   ├── api/              # HTTP 客户端、API 函数
│   │   ├── ws/               # WebSocket 连接管理
│   │   ├── auth/             # 认证状态、Token 管理
│   │   ├── sync/             # 同步状态指示器
│   │   ├── i18n/             # 国际化配置
│   │   └── theme/            # Tailwind 主题定制
│   ├── features/             # 按功能模块分组
│   │   ├── object/           # 知识对象 CRUD
│   │   │   ├── components/   # 页面级组件
│   │   │   ├── hooks/        # 自定义 hooks
│   │   │   ├── stores/       # Zustand stores
│   │   │   └── types/        # 类型定义
│   │   ├── collection/       # 数据库视图
│   │   ├── search/           # 搜索
│   │   ├── graph/            # 知识图谱
│   │   ├── capture/          # 快速剪藏
│   │   └── settings/         # 设置页面
│   ├── shared/               # 跨功能共享
│   │   ├── ui/               # 通用 UI 组件
│   │   ├── utils/            # 工具函数
│   │   └── types/            # 全局类型
│   └── styles/               # 全局样式
├── public/
├── e2e/                      # Playwright 测试
├── tailwind.config.ts
├── vite.config.ts
└── tsconfig.json
```

### 2.3 状态管理

```
                ┌──────────────┐
                │  React Query  │  ← 服务端状态（API 数据缓存）
                │  (Server)     │
                └──────┬───────┘
                       │
          ┌────────────┼────────────┐
          │            │            │
    ┌─────▼────┐ ┌────▼───┐ ┌─────▼────┐
    │ AuthStore │ │UIStore │ │ SyncStore│  ← Zustand 客户端状态
    │ (Zustand) │ │(Zust)  │ │(Zustand) │
    └──────────┘ └────────┘ └──────────┘
```

### 2.4 数据流

```
User Action → Handler → API Call → React Query Cache → UI Update
                ↓
           WebSocket ← Sync Engine ← Server
```

---

## 3. 桌面端 (Tauri 2)

### 3.1 技术栈

| 领域 | 选择 | 理由 |
|------|------|------|
| 框架 | Tauri 2 | 体积小、安全（Rust core）、跨平台 |
| 前端 | React 19（与 Web 共用） | 代码复用 |
| 系统 API | Rust 侧 + Tauri Commands | 文件系统、系统托盘、全局快捷键 |
| 窗口管理 | Tauri Window API | 多窗口（主窗口 + 剪藏窗口） |
| 原生能力 | tauri-plugin-* 生态 | 通知、shell、剪贴板、文件对话框 |

### 3.2 架构

```rust
// src-tauri/src/main.rs (概念结构)
fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_notification::init())
        .invoke_handler(tauri::generate_handler![
            capture_screenshot,
            get_system_audio,
            open_file_dialog,
        ])
        .run(tauri::generate_context!())
}
```

### 3.3 特性差异

桌面端相比 Web 端独有的功能：

- **视频截屏**：通过 Tauri Command 调用系统 API 捕获屏幕/窗口
- **系统托盘**：常驻系统托盘，全局快捷键唤出
- **离线存储**：本地 SQLite 直接读写（不经过 HTTP）
- **文件管理**：本地文件系统访问（导入/导出）

---

## 4. 移动端 (Flutter 3)

### 4.1 技术栈

| 领域 | 选择 | 理由 |
|------|------|------|
| 框架 | Flutter 3 | 跨平台 UI 高性能 |
| 状态管理 | Riverpod | 编译安全、依赖注入 |
| 路由 | GoRouter | 声明式、深度链接 |
| 网络 | Dio + Retrofit | 拦截器链式处理 |
| 本地存储 | Isar / Drift | 移动端高性能 SQLite |
| 相机/相册 | image_picker | 拍照导入 |
| 音频录制 | record | 语音速记 |
| 国际化 | Flutter i18n | ARB 文件管理 |

### 4.2 目录结构

```
frontend/mobile/
├── lib/
│   ├── main.dart
│   ├── app.dart              # 应用入口 + 主题 + 路由
│   ├── core/                 # 基础设施
│   │   ├── network/          # API 客户端、拦截器
│   │   ├── database/         # 本地数据库
│   │   ├── auth/             # 认证管理
│   │   ├── sync/             # 同步引擎客户端
│   │   └── theme/            # Material Design 3 主题
│   ├── features/             # 按功能模块分组
│   │   ├── object/           # 知识对象
│   │   ├── capture/          # 快速采集（相机、截屏）
│   │   ├── search/           # 搜索
│   │   └── settings/         # 设置
│   └── shared/               # 共享组件
│       ├── widgets/          # 通用 UI 组件
│       └── utils/            # 工具函数
├── test/
├── assets/                   # 图片、字体
└── pubspec.yaml
```

### 4.3 平台特性

| 功能 | Android | iOS |
|------|---------|-----|
| 相机捕获 | CameraX | AVFoundation |
| 本地推送 | FCM | APNs |
| 后台同步 | WorkManager | BGTaskScheduler |
| 生物认证 | Biometric | Face ID / Touch ID |
| 文件选择 | SAF | UIDocumentPicker |

---

## 5. 浏览器扩展

### 5.1 技术栈

| 领域 | 选择 |
|------|------|
| 框架 | WXT + React 19 |
| 构建 | Vite |
| 样式 | Tailwind CSS 4 |
| 存储 | Web Extension Storage API |
| 消息 | Extension Messaging API |

### 5.2 架构

```
┌─────────────────────────────────┐
│          Service Worker          │  ← 后台常驻，处理网络请求
│   消息中转、剪藏API代理          │
└──────────┬──────────────────────┘
           │  Messaging API
      ┌────┴────┐
      │ Popup   │  ← 点击工具栏图标弹出
      │ 快速保存 │
      └─────────┘
```

---

## 6. 通用 UI 设计原则

### 6.1 主题系统

- 支持亮色/暗色模式
- 基于 CSS 变量（Web） / ThemeData（Flutter）的设计 Token 系统
- 品牌色：待定

### 6.2 响应式断点

| 断点 | 宽度 | 目标设备 |
|------|------|---------|
| `sm` | < 640px | 手机 |
| `md` | 640-1024px | 平板 |
| `lg` | 1024-1440px | 桌面 |
| `xl` | > 1440px | 宽屏桌面 |

### 6.3 交互模式

- **三指操作**：触摸板手势（桌面端）
- **快捷键**：统一快捷键绑定（所有端）
- **拖拽**：拖拽文件、对象、块
- **右键菜单**：上下文操���
- **滑动操作**：移动端手势

---

## 7. 构建与发布

| 端 | 构建命令 | 产物 |
|----|---------|------|
| Web | `pnpm build` | `dist/` 静态文件 |
| Desktop | `pnpm tauri build` | `.dmg` / `.msi` / `.AppImage` |
| Mobile | `flutter build` | `.apk` / `.aab` / `.ipa` |
| Extension | `pnpm build` | `dist/` 加载包 |

---

## 8. 性能目标

| 指标 | Web | Desktop | Mobile |
|------|-----|---------|--------|
| 首屏加载 | < 2s | < 2s | < 3s |
| 交互响应 | < 100ms | < 100ms | < 200ms |
| 列表渲染（1000 项） | < 500ms | < 300ms | < 500ms |
| 搜索响应 | < 300ms | < 200ms | < 500ms |
| 包体积 | < 500KB (gzip) | < 20MB | < 50MB |
