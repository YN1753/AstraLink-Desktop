package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	basePath, _ := os.Getwd()

	//  物理目录初始化
	dirs := []string{"data/db", "data/notes", "data/assets", "data/index"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(basePath, d), 0755)
	}

	appCtx, err := InitializeApp(basePath)
	if err != nil {
		log.Fatal("初始化失败:", err)
	}

	// 3. 运行 Wails
	err = wails.Run(&options.App{
		Title: "AstraLink - 星链笔记",
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			appCtx,
		},
		OnStartup: func(ctx context.Context) {
			log.Println("🚀 AstraLink 系统就绪")
		},
	})

	if err != nil {
		log.Fatal("应用启动失败:", err)
	}
}
