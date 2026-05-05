//go:build wireinject
// +build wireinject

package main

import (
	"astralink/internal/db"
	"astralink/internal/repo"
	"astralink/internal/service"

	"github.com/google/wire"
)

// InitializeApp 依赖注入入口
func InitializeApp(basePath string) (*App, error) {
	wire.Build(
		// 1. 基础连接
		db.InitDB,
		//2. 仓储层
		repo.NewBaseRepo,

		// 3. 所有业务服务层
		service.NewNodeService,
		service.NewRelationService,

		// 4. 组装 App
		NewApp,
	)
	return nil, nil
}
