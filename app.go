package main

import (
	"astralink/internal/model"
	"astralink/internal/repo"
	"astralink/internal/service"
	"context"
	"fmt"
)

type App struct {
	ctx             context.Context
	SqliteRepo      *repo.BaseRepo
	NodeService     *service.NodeService
	RelationService *service.RelationService
}

func NewApp(r *repo.BaseRepo, n *service.NodeService, re *service.RelationService) *App {
	return &App{
		SqliteRepo:      r,
		NodeService:     n,
		RelationService: re,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("🚀 AstraLink 后端引擎已通过 Wire 自动化装配就绪！")
	if _, err := a.NodeService.CheckUserNode(); err != nil {
		fmt.Printf("⚠️ 创建默认用户节点失败: %v\n", err)
	}
}

// ============================================================
// 1. 用户模块 (User Module)
// ============================================================

func (a *App) CheckUserNode() (*model.Node, error) {
	return a.NodeService.CheckUserNode()
}

func (a *App) MergeUserInfo(req model.MergeNodeReq) (string, error) {
	return a.NodeService.MergeUserInfo(req)
}

func (a *App) GetUserInfo(userID string) (model.Node, error) {
	return a.NodeService.GetUserInfo(userID)
}

func (a *App) GetAvatar(id string) (string, error) {
	return a.NodeService.GetAvatar(id)
}

func (a *App) UploadAvatar(id string, avatar string) (string, error) {
	return a.NodeService.UploadAvatar(id, avatar)
}

// ============================================================
// 2. 系统全局状态 (System Stats)
// ============================================================

func (a *App) GetDataSpace() (int64, error) {
	return a.NodeService.GetDataSpace()
}

func (a *App) GetTagCount() (int64, error) {
	return a.NodeService.GetTagCount()
}

func (a *App) GetNoteCount() (int64, error) {
	return a.NodeService.GetNoteCount()
}

func (a *App) GetGalaxyCount() (int64, error) {
	return a.NodeService.GetGalaxyCount()
}

// ============================================================
// 3. 核心写入与关联 (Mutations)
// ============================================================

func (a *App) CreateLink(fromID, toID string) error {
	return a.RelationService.CreateLink(fromID, toID)
}

func (a *App) CreateTag(req model.CreateTagReq) (string, error) {
	return a.NodeService.CreateTag(req)
}

func (a *App) CreateNote(req model.CreateNoteReq) (string, error) {
	return a.NodeService.CreateNote(req)
}

func (a *App) CreateGalaxy(req model.MergeNodeReq) (string, error) {
	return a.NodeService.CreateGalaxy(req)
}

func (a *App) DeleteNode(id string) error {
	return a.NodeService.DeleteNode(id)
}

// ============================================================
// 4. 查询 (Queries)
// ============================================================

func (a *App) GetNodeByType(typename string) (*[]model.Node, error) {
	return a.SqliteRepo.GetNodeByType(typename)
}

func (a *App) GetAllTag() (*[]model.Node, error) {
	return a.NodeService.GetAllTag()
}

func (a *App) GetTagsWithCount() ([]model.TagWithCount, error) {
	return a.NodeService.GetTagsWithCount()
}

func (a *App) GetNodeByPath(path string) ([]model.Node, error) {
	return a.NodeService.GetNodeByPath(path)
}

func (a *App) GetRelationById(id string) (model.D3Graph, error) {
	return a.NodeService.GetRelationById(id)
}

func (a *App) LinkTagToNode(tagID string, nodeID string) error {
	return a.RelationService.LinkTagToNode(tagID, nodeID)
}

func (a *App) GetTagsByNodeID(nodeID string) (*[]model.TagMessage, error) {
	return a.RelationService.GetTagsByNodeID(nodeID)
}

func (a *App) DeleteTagRelation(tagID string, noteID string) error {
	return a.RelationService.DeleteRelation(model.DeleteRelationReq{
		FromId: tagID,
		ToId:   noteID,
		Type:   "tag",
	})
}

func (a *App) SearchNotes(query string) ([]model.NoteSearchResult, error) {
	return a.NodeService.SearchNotes(query)
}

func (a *App) IndexNote(id string, name string, content string) error {
	return a.NodeService.IndexNote(id, name, content)
}

// ============================================================

// ============================================================
// 5. 笔记深度管理 (Note Management)
// ============================================================

func (a *App) GetNoteContent(id string) (string, error) {
	return a.NodeService.GetNoteContent(id)
}

func (a *App) GetNotesByTag(tagId string) ([]model.NoteSearchResult, error) {
	return a.NodeService.GetNotesByTag(tagId)
}

func (a *App) UpdateNoteContent(id string, content string) error {
	return a.NodeService.UpdateNoteContent(id, content)
}

func (a *App) UpdateNodeInfo(req model.UpdateNoteInfoReq) error {
	return a.NodeService.UpdateNodeInfo(req)
}

func (a *App) GetRecentNotes(num int) ([]model.Node, error) {
	return a.NodeService.GetRecentNotes(num)
}
