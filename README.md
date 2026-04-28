# AstraLink - 星链笔记

> 书写有序的人生

## 简介

AstraLink（2.0） 是一款本地优先的笔记应用，采用 Go + Vue3 + Wails 构建，支持 Markdown 编辑、Galaxy 视图和个性化头像设置。

## 技术栈

- **后端**: Go, GORM, SQLite
- **前端**: Vue 3, TypeScript, TailwindCSS
- **框架**: Wails (跨平台桌面应用框架)

## 开发环境

### 前置依赖

- Go 1.21+
- Node.js 18+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 运行项目

```bash
# 安装前端依赖
cd frontend
npm install

# 启动开发模式
cd ..
wails dev
```

### 构建

```bash
wails build
```

## 项目结构

```
AstraLink/
├── app.go              # 应用入口和接口定义
├── main.go             # 主程序入口
├── wire.go             # Wire 依赖注入
├── internal/
│   ├── db/             # 数据库初始化
│   ├── model/          # 数据模型
│   ├── repo/           # 数据仓储层
│   └── service/        # 业务逻辑层
├── pkg/                # 公共工具包
├── frontend/           # 前端源码
│   ├── src/            # Vue 组件
│   └── wailsjs/        # Wails 生成的绑定代码
└── data/               # 数据存储目录（运行时生成）
    ├── avatar/         # 用户头像
    ├── notes/          # 笔记文件
    └── db/             # SQLite 数据库
```

## 功能特性

- [x] 创建和编辑 Markdown 笔记
- [x] Galaxy 视图浏览笔记
- [x] 用户头像上传与展示
- [x] 多主题支持（深色/浅色/护眼）
- [x] 数据本地存储

## License

MIT
