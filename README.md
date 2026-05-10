# AstraLink （2.0.0）- 星链笔记

> 书写有序的人生

## 简介

AstraLink（2.0,0）是一款本地优先的笔记应用，采用 Go + Vue3 + Wails 构建，支持 Markdown 编辑、Galaxy 视图、全文搜索、双链笔记。

基于Astralink-1.0版本的轻量化本地程序，解决了1.0的部署复杂，占用高的bug。目前内存占用仅36MB左右（现在实测没看到超过8MB，估计是release删除了一些调试的依赖），可执行文件大小仅为24MB。

支持两种安装方式： 

​	1.便携版 （解压缩后直接双击exe文件即可，数据存储在同目录下的data文件夹）

​	2.安装版 （下载后双击exe安装程序即可，数据存储在AppData路径下）

## 技术栈

- **后端**: Go 1.25+, GORM, SQLite, Bleve (全文搜索)
- **前端**: Vue 3, TypeScript, TailwindCSS
- **框架**: Wails v2 (跨平台桌面应用框架)

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
# 普通构建
wails build

# 构建并生成 NSIS 安装包（需要安装 NSIS 并加入 PATH）
wails build -nsis
```

构建产物位于 `build/bin/` 目录。

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
│   └── src/
│       ├── components/ # Vue 组件
│       └── wailsjs/    # Wails 生成的绑定代码
└── build/
    ├── bin/            # 构建产物
    └── windows/        # Windows 图标和清单
```

## 数据存储

应用支持两种数据模式：

- **安装模式**（默认）: 数据存储在 `%APPDATA%\AstraLink\`
- **便携模式**: 在 exe 同目录放置 `.portable` 文件，数据存储在 `exe目录\data\`

## 功能特性

- [x] 创建和编辑 Markdown 笔记
- [x] Galaxy 视图浏览笔记
- [x] 双链笔记支持
- [x] 全文搜索（基于 Bleve）
- [x] 标签管理与标签云
- [x] 用户头像上传与展示
- [x] 多主题支持（深色/浅色/护眼）
- [x] 数据本地存储
- [x] 软件内版本更新检查
- [x] NSIS 安装包打包
- [x] 支持mac和windows
## License

MIT

### 作者

如果说有什么建议或者好的想法可以进群联系我(记得备注一下)，QQ群:1102628859



要是觉得好用的话可以随意打赏一些

![image-20260509153104111](img.png)
