# Gridea Pro - 项目总结

> 简单而强大的跨平台静态博客写作客户端

## 基本信息

| 项 | 值 |
|---|---|
| 作者 | Eliauk (tespera@foxmail.com) |
| 版本 | 1.0.0 |
| 许可证 | MIT |
| 平台 | macOS / Windows / Linux |
| 语言 | Go 1.25.5 + TypeScript |
| 桌面框架 | Wails v2 |

## 技术栈

### 后端 (Go)

- **Wails v2** — Go 与 Web 技术桥接的桌面应用框架
- **go-git** — 纯 Go 实现的 Git 操作（无系统 Git 依赖）
- **Pongo2** — Jinja2 模板引擎
- **Goldmark** — Markdown 解析，集成 KaTeX 数学公式
- **goja** — 嵌入式 JavaScript 运行时（EJS 模板渲染）
- **fsnotify** — 文件系统事件监听
- **mark3labs/mcp-go** — MCP 协议集成

### 前端 (Vue 3)

- **Vue 3** + **TypeScript** + **Vite (rolldown-vite)**
- **Pinia** — 状态管理
- **Tailwind CSS** — 样式系统
- **Monaco Editor** — 代码编辑器（VS Code 引擎）
- **shadcn/ui** — UI 组件库
- **Vue I18n** — 国际化（11 种语言）

### 项目规模

- Go 源文件：~150 个
- Vue/TS 前端文件：~208 个

## 目录结构

```
gridea-pro/
├── main.go                  # 应用入口
├── wails.json               # Wails 构建配置
├── Makefile                 # 构建自动化
├── .github/workflows/       # CI/CD（多平台发布）
│
├── backend/
│   ├── cmd/
│   │   └── mcp/             # MCP Server 独立入口
│   ├── internal/
│   │   ├── app/             # 应用核心（Wails 绑定）
│   │   ├── config/          # 配置管理
│   │   ├── comment/         # 评论系统集成
│   │   ├── deploy/          # 部署器（Git/FTP/SFTP/Vercel/Netlify）
│   │   ├── domain/          # 领域模型
│   │   ├── engine/          # 渲染引擎（多模板引擎策略）
│   │   ├── facade/          # 门面层（简化 Wails 绑定接口）
│   │   ├── mcp/             # MCP 工具定义
│   │   ├── render/          # 模板渲染
│   │   ├── repository/      # 数据仓储层
│   │   ├── service/         # 业务服务层
│   │   │   ├── ai/          # AI 相关服务
│   │   │   ├── credential/  # 凭证管理
│   │   │   └── oauth/       # OAuth 认证
│   │   ├── template/        # 模板管理
│   │   ├── utils/           # 工具函数
│   │   └── version/         # 版本管理
│   └── pkg/
│       └── boot/            # 应用启动引导
│
├── frontend/
│   └── src/
│       ├── components/      # 通用组件 + UI 组件库
│       ├── views/           # 页面视图
│       │   ├── posts/       # 文章管理
│       │   ├── memos/       # 便签/短记
│       │   ├── tags/        # 标签管理
│       │   ├── categories/  # 分类管理
│       │   ├── links/       # 友链管理
│       │   ├── menu/        # 菜单管理
│       │   ├── theme/       # 主题配置
│       │   ├── comments/    # 评论设置
│       │   ├── settings/    # 站点设置
│       │   └── preferences/ # 偏好设置
│       ├── router/          # 路由配置
│       ├── stores/          # Pinia 状态仓库
│       ├── locales/         # 国际化翻译文件（11 种语言）
│       ├── composables/     # Vue 3 组合式函数
│       └── utils/           # 工具函数
│
├── build/                   # 构建输出
└── images/                  # 项目图片资源
```

## 核心功能

### 内容管理

- **文章编辑器** — Monaco Editor 驱动，支持 Markdown + KaTeX 数学公式实时预览
- **便签 (Memos)** — 快速记录灵感，支持热力图统计
- **标签 & 分类** — 灵活的内容组织，slug 格式校验 + 唯一性检查
- **友链管理** — 友情链接的增删改查
- **菜单管理** — 自定义导航菜单

### 主题系统

- **9 款内置主题** 开箱即用
- **多模板引擎支持** — Jinja2 (Pongo2)、EJS、Go Templates
- **可视化主题配置** — 自动生成配置表单
- **深色模式** 支持

### 部署能力

| 平台 | 方式 |
|---|---|
| GitHub Pages | 内置 Git 推送 |
| Vercel | API 部署 |
| Netlify | API 部署 |
| Gitee Pages | 内置 Git 推送 |
| Coding Pages | 内置 Git 推送 |
| SFTP/FTP | 文件传输部署 |
| 自定义服务器 | Git + SSH |

### 评论系统

集成主流评论平台：Gitalk、Giscus、Disqus、Valine、Utterances、Waline 等。

### AI 集成 (MCP)

通过 **Model Context Protocol** 提供 25+ 工具，支持 Claude、Cursor 等 AI 助手直接管理博客：
- 文章/便签的增删改查
- 标签与分类管理
- 站点配置修改
- 渲染与部署触发
- 内置工作流：写作助手、内容审核、便签整理

### SEO 与输出

- 自动生成 sitemap.xml
- robots.txt 配置
- RSS/Atom 订阅源
- 静态资源压缩 (tdewolff/minify)

## 架构设计

### 后端分层

```
facade（门面）→ service（业务）→ repository（数据）→ domain（模型）
```

- **Clean Architecture** — 领域模型与基础设施分离
- **Facade Pattern** — 简化 Wails 绑定接口
- **Repository Pattern** — 数据访问抽象与缓存
- **Strategy Pattern** — 多模板引擎可插拔
- **Single Flight** — 渲染请求合并，防止并发重复

### 前端架构

- **Composition API** — Vue 3 组合式开发
- **Feature-based** — 按功能模块组织 views/components/composables
- **shadcn/ui** — 统一设计系统

### 数据流

```
用户操作 → Vue 组件 → Wails Binding → Facade → Service → Repository → 文件系统
                                                                   ↓
                                                              渲染引擎 → 静态站点输出
```

## 国际化

支持 11 种语言：中文（简/繁）、英语、日语、韩语、法语、德语、西班牙语、意大利语、葡萄牙语（巴西）、俄语。

## 构建与发布

- **CI/CD**: GitHub Actions 自动化多平台构建
- **分发格式**: macOS (.dmg/.zip)、Windows (.exe 安装包/便携版)、Linux (.deb/.rpm/.AppImage)
- **自动更新**: 集成 minio/selfupdate

## 特色亮点

1. **零外部依赖** — 内置 Git 实现，无需安装系统 Git
2. **AI 原生** — MCP 协议让 AI 助手直接参与博客管理
3. **多模板引擎** — 主题开发者可自由选择 Jinja2/EJS/Go Templates
4. **纯静态输出** — 生成的站点可部署到任何静态托管平台
5. **跨平台一致体验** — Wails 驱动的原生桌面体验
