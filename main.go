package main

import (
	"astralink/internal/handler"
	"astralink/pkg/utils"
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

var AppVersion = "dev"

func main() {
	basePath := utils.GetAppStoragePath()

	// 物理目录初始化
	dirs := []string{"db", "index", "notes", "avatar", "assets"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(basePath, d), 0755)
	}

	appCtx, err := InitializeApp(basePath)
	if err != nil {
		log.Fatal("初始化失败:", err)
	}
	appCtx.Version = AppVersion

	// 3. 运行 Wails
	assetHandler := handler.NewAssetHandler(basePath)
	err = wails.Run(&options.App{
		Title: "AstraLink - 星链笔记",
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: assetHandler,
		},
		Bind: []any{
			appCtx,
		},
	})

	if err != nil {
		log.Fatal("应用启动失败:", err)
	}
}
