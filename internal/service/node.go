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

func (n *NodeService) MergeNode(req model.MergeNodeReq) (string, error) {
	fmt.Printf("MergeNode received:%+v\n", req)
	var node model.Node
	node.ID = req.ID
	node.Name = req.Name
	node.Type = req.Type
	node.Others = req.Others
	if node.ID == "" {
		node.ID = utils.GenerateV7UID()
	}
	parentPath := strings.TrimSuffix(strings.TrimSpace(req.ParentPath), "/")
	node.Path = parentPath + "/" + node.ID
	switch node.Type {
	case "user":
		node.Path = "root"
	case "note":
		file := strings.NewReader(req.File)
		address, err := n.base.SaveLocalFile("notes", file, node.ID+".md")
		if err != nil {
			return node.ID, err
		}
		node.Address = address
	}

	if _, err := n.base.MergeNode(&node); err != nil {
		return node.ID, err
	}
	if req.ParentID != "" {
		relation := model.Relation{
			FromID: req.ParentID,
			ToID:   node.ID,
			Type:   "link",
		}
		if err := n.base.CreateRelation(relation); err != nil {
			return node.ID, err
		}
	}

	return node.ID, nil
}

func (n *NodeService) GetNode(id string) (model.Node, error) {
	return n.base.GetNodeById(id)
}

func (n *NodeService) GetNodeByType(nodeType string) (*[]model.Node, error) {
	node, err := n.base.GetNodeByType(nodeType)
	return node, err
}
func GetExt(base64Str string) string {
	if !strings.HasPrefix(base64Str, "data:image/") {
		return ".bin"
	}

	start := len("data:image/")
	end := strings.Index(base64Str, ";")

	if end == -1 || end <= start {
		return ".bin"
	}

	ext := base64Str[start:end]

	switch ext {
	case "jpeg":
		return ".jpg"
	case "svg+xml":
		return ".svg"
	default:
		return "." + ext
	}
}
func (n *NodeService) UpdateAvatar(id string, value string) (string, error) {

	commaIndex := strings.Index(value, ",")
	if commaIndex == -1 {
		return "", errors.New("无效的 Base64 数据")
	}

	ext := GetExt(value)
	if ext == ".bin" {
		return "", errors.New("不支持的图片格式")
	}
	pureBase64 := value[commaIndex+1:]
	imageData, err := base64.StdEncoding.DecodeString(pureBase64)
	if err != nil {
		return "", errors.New("图片解码失败: " + err.Error())
	}

	path, err := n.base.SaveLocalFile("avatar", bytes.NewReader(imageData), id+ext)
	if err != nil {
		return "", errors.New("保存头像文件失败")
	}

	return path, n.base.UpdateNodeInfo(id, "address", path)
}

func (n *NodeService) GetAvatar(id string) (string, error) {

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

func (n *NodeService) GetAllNotes() (*[]model.Node, error) {
	return n.base.GetAllNotes()
}

func (n *NodeService) GetNodesByTitle(keyword string) (*[]model.Node, error) {
	return n.base.GetNodesByTitle(keyword)
}

func (n *NodeService) GetNoteContent(id string) (string, error) {
	return n.base.GetNoteContent(id)
}

// GetDirectoryTree 获取目录树结构
func (n *NodeService) GetDirectoryTree() (*model.TreeNode, error) {
	// 获取所有节点
	allNodes, err := n.base.GetAllNotes()
	if err != nil {
		return nil, err
	}

	// 同时获取用户节点
	userNodes, err := n.base.GetNodeByType("user")
	if err != nil {
		return nil, err
	}

	// 合并所有节点
	all := append(*allNodes, *userNodes...)

	// 构建 map
	nodeMap := make(map[string]*model.TreeNode)
	for i := range all {
		nodeMap[all[i].ID] = &model.TreeNode{
			ID:       all[i].ID,
			Name:     all[i].Name,
			Type:     all[i].Type,
			Children: []*model.TreeNode{},
		}
	}

	// 根节点
	root := &model.TreeNode{
		ID:       "root",
		Name:     "root",
		Type:     "folder",
		Children: []*model.TreeNode{},
	}

	// 查找 Path 为 "root" 的用户节点，作为树的根
	var rootUser *model.TreeNode
	for _, node := range all {
		if node.Path == "root" && node.Type == "user" {
			rootUser = nodeMap[node.ID]
			break
		}
	}

	// 根据 Path 构建树
	// Path 格式: root/parentId1/parentId2/nodeId
	for _, node := range all {
		// 跳过 root 本身和已经是根用户节点的
		if node.Path == "root" || node.Path == "" {
			continue
		}

		// 解析路径
		parts := strings.Split(strings.Trim(node.Path, "/"), "/")
		if len(parts) < 2 {
			continue
		}

		// 最后一个是当前节点ID
		nodeID := parts[len(parts)-1]
		// 前一个是父节点ID
		parentID := parts[len(parts)-2]

		treeNode := nodeMap[nodeID]
		if treeNode == nil {
			continue
		}

		parentNode := nodeMap[parentID]
		if parentNode != nil {
			parentNode.Children = append(parentNode.Children, treeNode)
		} else if rootUser != nil {
			// 如果有根用户节点，挂在根用户节点下
			rootUser.Children = append(rootUser.Children, treeNode)
		} else {
			// 父节点不在 map 中，挂在 root 下
			root.Children = append(root.Children, treeNode)
		}
	}

	// 如果存在根用户节点，返回根用户节点；否则返回默认 root
	if rootUser != nil {
		return rootUser, nil
	}
	return root, nil
}
