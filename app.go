package main

import (
	"astralink/internal/model"
	"astralink/internal/repo"
	"astralink/internal/service"
	"context"
	"fmt"
)

type App struct {
	ctx         context.Context
	SqliteRepo  *repo.BaseRepo
	NodeService *service.NodeService
}

func NewApp(r *repo.BaseRepo, n *service.NodeService) *App {
	return &App{
		SqliteRepo:  r,
		NodeService: n,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("🚀 AstraLink 后端引擎已通过 Wire 自动化装配就绪！")
}

//  暴露给前端的接口

func (a *App) CreateNewNode(req model.MergeNodeReq) (string, error) {
	return a.NodeService.MergeNode(req)
}

func (a *App) GetNodes(id string) (model.Node, error) {
	return a.NodeService.GetNode(id)
}
func (a *App) GetNodeByType(nodeType string) (*[]model.Node, error) {
	return a.NodeService.GetNodeByType(nodeType)
}
func (a *App) UpdateAvatar(id string, value string) (string, error) {
	return a.NodeService.UpdateAvatar(id, value)
}
func (a *App) GetAvatar(id string) (string, error) {
	return a.NodeService.GetAvatar(id)
}
