package service

import (
	"astralink/internal/model"
	"astralink/internal/repo"
	"astralink/pkg/utils"
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type NodeService struct {
	base *repo.BaseRepo
}

func NewNodeService(baserepo *repo.BaseRepo) *NodeService {
	return &NodeService{
		base: baserepo,
	}
}

func (n *NodeService) MergeNode(req model.MergeNodeReq) (string, error) {
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

	return n.base.MergeNode(&node)
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

	// 找到 "image/" 之后的位置
	start := len("data:image/")
	// 找到分号 ";" 的位置（Data URL 格式：data:image/png;base64,
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
	// 1. 查找 Base64 数据开始的位置（逗号后面）
	commaIndex := strings.Index(value, ",")
	if commaIndex == -1 {
		return "", errors.New("无效的 Base64 数据")
	}

	// 2. 获取后缀名（使用你写的 GetExt）

	ext := GetExt(value)
	if ext == ".bin" {
		return "", errors.New("不支持的图片格式")
	}

	// 3. 提取并解码纯 Base64 数据
	pureBase64 := value[commaIndex+1:]
	imageData, err := base64.StdEncoding.DecodeString(pureBase64)
	if err != nil {
		return "", errors.New("图片解码失败: " + err.Error())
	}
	// 4. 保存二进制流
	path, err := n.base.SaveLocalFile("avatar", bytes.NewReader(imageData), id+ext)
	if err != nil {
		return "", errors.New("保存头像文件失败")
	}

	// 5. 更新数据库中的地址
	return path, n.base.UpdateNodeInfo(id, "address", path)
}

func (n *NodeService) GetAvatar(id string) (string, error) {
	// GetAvatar 读取头像文件并返回 base64
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
