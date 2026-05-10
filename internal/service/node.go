package service

import (
	"astralink/internal/model"
	"astralink/internal/repo"
	"astralink/pkg/utils"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type NodeService struct {
	base *repo.BaseRepo
}

func NewNodeService(baseRepo *repo.BaseRepo) *NodeService {
	return &NodeService{
		base: baseRepo,
	}
}

func (n *NodeService) CheckUserNode() (*model.Node, error) { //检测用户是否存在，用于初始化
	nodesPtr, err := n.base.GetNodeByType("user")
	if err == nil && nodesPtr != nil && len(*nodesPtr) > 0 {
		return &((*nodesPtr)[0]), nil
	}
	newNode := &model.Node{
		ID:   utils.GenerateV7UID(),
		Path: "root",
		Type: "user",
		Others: model.JSONMap{
			"motto": "在星辰间寻找逻辑",
		},
	}
	id, err := n.base.MergeNode(newNode)
	if err != nil {
		return nil, err
	}
	newNode.ID = id
	return newNode, nil
}

func (n *NodeService) MergeUserInfo(req model.MergeNodeReq) (string, error) { //更新用户信息
	var node model.Node
	node.ID = req.ID
	node.Name = req.Name
	node.Type = "user"
	node.Others = req.Others
	if node.ID == "" {
		node.ID = utils.GenerateV7UID()
	}
	node.Path = "root"
	return n.base.MergeNode(&node)
}

func (n *NodeService) GetUserInfo(userID string) (model.Node, error) { //获取个人资料
	return n.base.GetNodeById(userID)
}
func (n *NodeService) UploadAvatar(id string, avatar string) (string, error) {
	// 解析数据 URL: "data:image/png;base64,..."
	header, dataPart, ok := strings.Cut(avatar, ",")
	if !ok {
		return "", fmt.Errorf("invalid data URL")
	}
	ext := ".jpg"
	if strings.Contains(header, "image/png") {
		ext = ".png"
	} else if strings.Contains(header, "image/svg") {
		ext = ".svg"
	} else if strings.Contains(header, "image/webp") {
		ext = ".webp"
	}

	// 解码 base64 数据
	data, err := base64.StdEncoding.DecodeString(dataPart)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %v", err)
	}

	path, err := n.base.SaveLocalFile("avatar", bytes.NewReader(data), id+ext)
	if err != nil {
		return "", err
	}
	err = n.base.UpdateNodeInfo(id, "address", path)
	if err != nil {
		return "", err
	}
	return path, nil
}
func (n *NodeService) GetAvatar(id string) (string, error) { //获取头像
	node, err := n.base.GetNodeById(id)
	if err != nil {
		return "", err
	}
	if node.Address == "" {
		return "", nil
	}

	data, err := os.ReadFile(node.Address)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(node.Address))
	mimeType := "image/jpeg"
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".webp":
		mimeType = "image/webp"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

var allowedAttachmentExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".csv":  true,
}

func (n *NodeService) SaveAttachment(noteID string, dataURL string, filename string) (string, error) {
	_, dataPart, ok := strings.Cut(dataURL, ",")
	if !ok {
		return "", fmt.Errorf("invalid data URL")
	}

	data, err := base64.StdEncoding.DecodeString(dataPart)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %v", err)
	}

	// Sanitize filename
	sanitized := filepath.Base(filename)
	sanitized = strings.ReplaceAll(sanitized, "\\", "")
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}
	// Validate extension
	ext := strings.ToLower(filepath.Ext(sanitized))
	if !allowedAttachmentExts[ext] {
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}

	subDir := filepath.Join("assets", noteID)
	_, err = n.base.SaveLocalFile(subDir, bytes.NewReader(data), sanitized)
	if err != nil {
		return "", err
	}
	return "assets/" + noteID + "/" + sanitized, nil
}

func (n *NodeService) GetAttachmentPath(noteID string, filename string) (string, error) {
	fullPath := filepath.Join(n.base.GetRootPath(), "assets", noteID, filename)
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("file not found: %v", err)
	}
	return fullPath, nil
}

func (n *NodeService) ReadFileAsDataUrl(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := "application/octet-stream"
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".pdf":
		mimeType = "application/pdf"
	case ".doc":
		mimeType = "application/msword"
	case ".docx":
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		mimeType = "application/vnd.ms-excel"
	case ".xlsx":
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		mimeType = "application/vnd.ms-powerpoint"
	case ".pptx":
		mimeType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".csv":
		mimeType = "text/csv"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func (n *NodeService) DeleteNoteAssets(noteID string) error {
	dirPath := filepath.Join(n.base.GetRootPath(), "assets", noteID)
	err := os.RemoveAll(dirPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (n *NodeService) GetDataSpace() (int64, error) { //获取占用空间大小
	return utils.GetStorageSize(n.base.GetRootPath())
}

func (n *NodeService) GetTagCount() (int64, error) { //获取tag个数
	return n.base.CountNodesByType("tag")
}

func (n *NodeService) GetNoteCount() (int64, error) { //获取节点个数
	return n.base.CountNodesByType("note")
}

func (n *NodeService) GetGalaxyCount() (int64, error) { //获取星系个数
	return n.base.CountNodesByType("galaxy")
}

func (n *NodeService) CreateTag(req model.CreateTagReq) (string, error) { //创建tag
	node := model.Node{
		ID:     utils.GenerateV7UID(),
		Name:   req.Name,
		Type:   "tag",
		Others: req.Others,
	}

	id, err := n.base.MergeNode(&node)
	if err != nil {
		return id, err
	}
	if req.ParentID != "" {
		if err := n.base.UpsertRelation(model.Relation{FromID: req.ParentID, ToID: id, Type: "tag"}); err != nil {
			return id, err
		}
	}
	return id, nil
}

func (n *NodeService) CreateNote(req model.CreateNoteReq) (string, error) { //创建笔记
	node := model.Node{
		ID:     utils.GenerateV7UID(),
		Name:   req.Name,
		Type:   "note",
		Others: req.Others,
	}

	file := strings.NewReader(req.File)
	address, err := n.base.SaveLocalFile("notes", file, node.ID+".md")
	if err != nil {
		return node.ID, err
	}
	node.Address = address

	//更新逻辑路径
	parentPath := strings.TrimSuffix(strings.TrimSpace(req.ParentPath), "/")
	if parentPath == "" {
		parentPath = "root"
	}
	node.Path = parentPath + "/" + req.ParentID

	id, err := n.base.MergeNode(&node)
	if err != nil {
		return id, err
	}
	//创建逻辑
	if req.ParentID != "" {
		if err := n.base.UpsertRelation(model.Relation{FromID: req.ParentID, ToID: id, Type: "note"}); err != nil {
			return id, err
		}
	}

	// Index note content in bleve
	_ = n.base.IndexNote(node.ID, req.Name, req.File)

	return id, nil
}

func (n *NodeService) CreateGalaxy(req model.MergeNodeReq) (string, error) { //创建星系
	node := model.Node{
		ID:     utils.GenerateV7UID(),
		Name:   req.Name,
		Type:   "galaxy",
		Others: req.Others,
	}
	if req.ID != "" {
		node.ID = req.ID
	}

	parentPath := strings.TrimSuffix(strings.TrimSpace(req.ParentPath), "/")
	if parentPath == "" {
		parentPath = "root"
	}
	node.Path = parentPath + "/" + req.ParentID

	id, err := n.base.MergeNode(&node)
	if err != nil {
		return id, err
	}

	if req.ParentID != "" {
		if err := n.base.UpsertRelation(model.Relation{FromID: req.ParentID, ToID: id, Type: "galaxy"}); err != nil {
			return id, err
		}
	}

	return id, nil
}
func (n *NodeService) GetNodeByPath(path string) ([]model.Node, error) {
	var nodes []model.Node
	err := n.base.SqlDb.Where("path LIKE ?", path+"%").Find(&nodes).Error
	return nodes, err
}
func (n *NodeService) DeleteNode(id string) error { //删除节点包括本地文件和关系
	node, err := n.base.GetNodeById(id)
	if err != nil {
		return err
	}

	if node.Type == "note" && node.Address != "" {
		os.Remove(node.Address)
		_ = n.DeleteNoteAssets(id)
	} else if node.Type == "user" && node.Address != "" {
		os.Remove(node.Address)
	}

	if err := n.base.DeleteRelationsByNodeID(id); err != nil {
		return err
	}

	// Remove from bleve index if it's a note
	if node.Type == "note" {
		_ = n.base.DeleteNoteFromIndex(id)
	}

	return n.base.DeleteNodeById(id)
}

func (n *NodeService) GetAllTag() (*[]model.Node, error) { //获取所有tag
	tags, err := n.base.GetNodeByType("tag")
	if err != nil {
		return nil, err
	}
	return tags, nil
}
func (n *NodeService) GetNoteContent(id string) (string, error) { //获取笔记context
	return n.base.GetNoteContent(id)
}

func (n *NodeService) GetNotesByTag(tagId string) ([]model.NoteSearchResult, error) {
	return n.base.GetNotesByTag(tagId)
}
func (n *NodeService) UpdateNodeInfo(req model.UpdateNoteInfoReq) error {
	node, err := n.base.GetNodeById(req.Id)
	if err != nil {
		return err
	}
	if req.Path != "" && req.Path != node.Path {
		node.Path = req.Path
	}
	if req.Title != "" && req.Title != node.Name {
		node.Name = req.Title
	}
	if req.Others != nil {
		node.Others = req.Others
	}
	_, err = n.base.MergeNode(&node)
	return err

}
func (n *NodeService) UpdateNoteContent(id string, content string) error {
	file := strings.NewReader(content)
	_, err := n.base.SaveLocalFile("notes", file, id+".md")
	if err != nil {
		return err
	}
	// Re-index note content in bleve
	node, err := n.base.GetNodeById(id)
	if err == nil {
		_ = n.base.IndexNote(id, node.Name, content)
	}
	return nil
}

func (n *NodeService) GetRelationById(id string) (model.D3Graph, error) {
	graph := model.D3Graph{
		Nodes: make([]model.D3Node, 0),
		Links: make([]model.D3Link, 0),
	}
	rel, err := n.base.GetNodeById(id)
	if err != nil {
		return graph, err
	}
	nodeMessage, err := n.base.GetNodeMessageByPath(rel.Path, rel.ID)
	if err != nil {
		return graph, err
	}
	var nodes []string
	for _, id := range nodeMessage {
		nodes = append(nodes, id.ID)
	}
	relations, err := n.base.GetRelationByNodeId(nodes)
	if err != nil {
		return graph, err
	}
	vailNodeMap := make(map[string]bool)
	for _, node := range nodeMessage {
		vailNodeMap[node.ID] = true
		graph.Nodes = append(graph.Nodes, model.D3Node{
			Id:    node.ID,
			Title: node.Name,
			Type:  node.Type,
		})
	}
	for _, rel := range relations {
		if vailNodeMap[rel.FromID] && vailNodeMap[rel.ToID] {
			graph.Links = append(graph.Links, model.D3Link{
				Source: rel.FromID,
				Target: rel.ToID,
				Type:   rel.Type,
			})
		}
	}
	return graph, nil
}
func (n *NodeService) GetRecentNotes(num int) ([]model.Node, error) {
	if num <= 0 || num >= 9999999 {
		return nil, errors.New("invalid num")
	}
	return n.base.GetRecentNotes(num)
}

func (n *NodeService) GetTagsWithCount() ([]model.TagWithCount, error) {
	tags, err := n.base.GetNodeByType("tag")
	if err != nil {
		return nil, err
	}

	counts, err := n.base.GetTagRelationCounts()
	if err != nil {
		return nil, err
	}

	result := make([]model.TagWithCount, 0, len(*tags))
	for _, tag := range *tags {
		result = append(result, model.TagWithCount{
			ID:        tag.ID,
			Name:      tag.Name,
			NoteCount: counts[tag.ID],
		})
	}
	return result, nil
}

func (n *NodeService) SearchNotes(query string) ([]model.NoteSearchResult, error) {
	return n.base.SearchNotes(query)
}

func (n *NodeService) IndexNote(id string, name string, content string) error {
	return n.base.IndexNote(id, name, content)
}

func (n *NodeService) DeleteNoteIndex(id string) error {
	return n.base.DeleteNoteFromIndex(id)
}

func (n *NodeService) GetNodeByTag(tags []string) ([]model.Node, error) {
	var nodes []model.Node
	var nodeId []string
	rel, err := n.base.GetNodeIdByTags(tags)
	if err != nil {
		return nodes, err
	}
	for _, id := range rel {
		nodeId = append(nodeId, id.ToID)
	}
	nodes, err = n.base.GetNodeByNodeId(tags)
	return nodes, nil
}
