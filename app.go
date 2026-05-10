package main

import (
	"astralink/internal/model"
	"astralink/internal/repo"
	"astralink/internal/service"
	"astralink/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

)

type App struct {
	ctx             context.Context
	SqliteRepo      *repo.BaseRepo
	NodeService     *service.NodeService
	RelationService *service.RelationService
	Version         string
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

func (a *App) GetVersion() string {
	return a.Version
}

func (a *App) SaveAttachment(noteID string, dataURL string, filename string) (string, error) {
	return a.NodeService.SaveAttachment(noteID, dataURL, filename)
}

func (a *App) GetAttachmentPath(noteID string, filename string) (string, error) {
	return a.NodeService.GetAttachmentPath(noteID, filename)
}

func (a *App) DeleteNoteAssets(noteID string) error {
	return a.NodeService.DeleteNoteAssets(noteID)
}

func (a *App) ReadFileAsDataUrl(filePath string) (string, error) {
	return a.NodeService.ReadFileAsDataUrl(filePath)
}

func (a *App) OpenURL(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

func (a *App) CheckUpdate() (model.UpdateInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/YN1753/AstraLink-Desktop/releases/latest")
	if err != nil {
		return model.UpdateInfo{}, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return model.UpdateInfo{HasUpdate: false}, nil
	}
	if resp.StatusCode != 200 {
		return model.UpdateInfo{}, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return model.UpdateInfo{}, fmt.Errorf("解析响应失败: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(a.Version, "v")

	info := model.UpdateInfo{
		HasUpdate:   latest != current,
		LatestVer:   latest,
		CurrentVer:  current,
		ReleaseNote: release.Body,
		IsPortable:  utils.IsPortable(),
	}

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.HasSuffix(name, ".zip") {
			info.ZipURL = asset.BrowserDownloadURL
			info.ZipSize = asset.Size
		} else if strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".msi") {
			isInstaller := strings.Contains(name, "setup") || strings.Contains(name, "installer")
			if isInstaller || info.ExeURL == "" {
				info.ExeURL = asset.BrowserDownloadURL
				info.ExeSize = asset.Size
				info.ExeName = asset.Name
			}
		}
	}

	return info, nil
}

func (a *App) DownloadAndInstallUpdate(url string) (string, error) {
	parts := strings.Split(url, "/")
	fileName := parts[len(parts)-1]
	if fileName == "" {
		fileName = "AstraLink-Update.exe"
	}
	savePath := os.TempDir() + string(os.PathSeparator) + fileName

	os.Remove(savePath)

	psCmd := fmt.Sprintf(
		"[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%s' -OutFile '%s' -UseBasicParsing",
		url, savePath,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("下载失败: %s", string(output))
	}

	if _, err := os.Stat(savePath); err != nil {
		return "", fmt.Errorf("下载文件不存在: %w", err)
	}

	return savePath, nil
}

func (a *App) OpenFileInExplorer(filePath string) error {
	return exec.Command("explorer", "/select,", filePath).Start()
}
