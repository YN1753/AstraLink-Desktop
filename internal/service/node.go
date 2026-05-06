package service

import (
	"astralink/internal/model"
	"astralink/internal/repo"
	"astralink/pkg/utils"
	"bytes"
	"encoding/base64"
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
	commaIdx := strings.Index(avatar, ",")
	if commaIdx == -1 {
		return "", fmt.Errorf("invalid data URL")
	}

	// 从 header 提取 MIME 类型并转换为扩展名
	header := avatar[:commaIdx]
	ext := ".jpg"
	if strings.Contains(header, "image/png") {
		ext = ".png"
	} else if strings.Contains(header, "image/svg") {
		ext = ".svg"
	} else if strings.Contains(header, "image/webp") {
		ext = ".webp"
	}

	// 解码 base64 数据
	data, err := base64.StdEncoding.DecodeString(avatar[commaIdx+1:])
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
	//保存笔记
	if req.File != "" {
		file := strings.NewReader(req.File)
		address, err := n.base.SaveLocalFile("notes", file, node.ID+".md")
		if err != nil {
			return node.ID, err
		}
		node.Address = address
	}
	//更新逻辑路径
	parentPath := strings.TrimSuffix(strings.TrimSpace(req.ParentPath), "/")
	if parentPath == "" {
		parentPath = "root"
	}
	node.Path = parentPath + "/" + node.ID

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
	node.Path = parentPath + "/" + node.ID

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
	} else if node.Type == "user" && node.Address != "" {
		os.Remove(node.Address)
	}

	if err := n.base.DeleteRelationsByNodeID(id); err != nil {
		return err
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

func (n *NodeService) UpdateNoteContent(id string, content string) error {
	file := strings.NewReader(content)
	_, err := n.base.SaveLocalFile("notes", file, id+".md")
	return err
}

func (n *NodeService)GetRelationById(id string)([]model.Relation,error){
	var res []model.Relation
	var graph model.D3Graph
	rel,err:=n.base.GetNodeById(id)
	if err!=nil{
		return res,err
	}
	nodes,err:=n.base.GetNodeIdByPath(rel.Path,rel.ID)
	if err!=nil{
		return res,err
	}
	nodeMessage,err:=n.base.GetNodeMessageByPath(rel.Path,rel.ID)
	if err!=nil{
		return res,err
	}
	relations,err:=n.base.GetRelationByNodeId(nodes)
	if err!=nil{
		return relations,err
	}
	vailNodeMap:=make(map[string]bool)
	for _,node:=range nodeMessage{
		vailNodeMap[node.ID]=true
		graph.Nodes = append(graph.Nodes, model.D3Node{
			Id:node.ID,
			Title: node.Name,
			Type: node.Type,
		})
	}
	
//	return res,nil
//
}
