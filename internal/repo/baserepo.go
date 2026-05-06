package repo

import (
	"astralink/internal/model"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/blevesearch/bleve/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"path/filepath"
)

type BaseRepo struct {
	SqlDb      *gorm.DB
	BasePath   string
	BleveIndex bleve.Index
}

func NewBaseRepo(db *gorm.DB, path string) *BaseRepo {
	return &BaseRepo{
		SqlDb:    db,
		BasePath: path,
		//BleveIndex: bleve,
	}
}

func (b *BaseRepo) MergeNode(node *model.Node) (string, error) {
	if node == nil {
		return "", nil
	}
	// 使用 OnConflict 确保在 ID 冲突时执行全字段更新
	return node.ID, b.SqlDb.Save(node).Error
}

func (b *BaseRepo) GetNodeById(id string) (model.Node, error) {
	var node model.Node
	err := b.SqlDb.Where("id = ?", id).First(&node).Error
	return node, err
}
func (b *BaseRepo) GetRootPath() string {
	// 在开发模式（wails dev）下，我们通常希望数据就在项目根目录
	// 你可以根据环境变量来切换，或者简单地判断是否在正式安装目录运行
	// 这里演示标准做法：获取系统分配给本应用的配置目录
	return "data"
	//dir, err := os.UserConfigDir() // 获取 Roaming 或 Application Support 路径
	//if err != nil {
	//	return "data" // 降级方案：存本地
	//}
	//
	//// 拼接应用名，建议叫 "AstraLink"
	//appDataPath := filepath.Join(dir, "AstraLink")
	//
	//// 确保这个根目录存在
	//_ = os.MkdirAll(appDataPath, 0755)
	//
	//return appDataPath
}

func (b *BaseRepo) SaveLocalFile(subDir string, file io.Reader, filename string) (string, error) {
	root := b.GetRootPath()

	targetDir := filepath.Join(root, subDir)

	fullPath := filepath.Join(targetDir, filename)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", errors.New("无法初始化存储空间")
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return fullPath, err
}

func (b *BaseRepo) GetNodeByType(nodeType string) (*[]model.Node, error) {
	var node []model.Node
	err := b.SqlDb.Where("type = ?", nodeType).Find(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}
func (b *BaseRepo) UpdateNodeInfo(id string, tab string, value string) error {
	return b.SqlDb.Model(&model.Node{}).Where("id = ?", id).Update(tab, value).Error
}

func (b *BaseRepo) UpsertRelation(req model.Relation) error {
	return b.SqlDb.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&req).Error
}

func (b *BaseRepo) GetAllNodeByType(typename string) (*[]model.Node, error) {
	var nodes []model.Node
	err := b.SqlDb.Where("type = ?", typename).Find(&nodes).Error
	return &nodes, err
}

func (b *BaseRepo) GetAllRelationsByType(typename string) (*[]model.Relation, error) {
	var relations []model.Relation
	var err error
	if typename != "" {
		err = b.SqlDb.Where("type = ?", typename).Find(&relations).Error
	} else {
		err = b.SqlDb.Find(&relations).Error
	}
	if err != nil {
		return &[]model.Relation{}, err
	}
	return &relations, nil
}
func (b *BaseRepo) GetRelationsByID(ID string, typename string) (*[]model.Relation, error) {
	var relations []model.Relation
	var err error
	if typename == "from_id" {
		err = b.SqlDb.Where("from_id = ?", ID).Find(&relations).Error
	} else if typename == "to_id" {
		err = b.SqlDb.Where("to_id = ?", ID).Find(&relations).Error
	} else {
		err = b.SqlDb.Where("to_id = ? OR from_id = ?", ID).Find(&relations).Error
	}
	if err != nil {
		return &[]model.Relation{}, err
	}
	return &relations, nil
}

func (b *BaseRepo) GetNoteContent(id string) (string, error) {
	node, err := b.GetNodeById(id)
	if err != nil {
		return "", err
	}
	if node.Address == "" {
		return "", nil
	}
	data, err := os.ReadFile(node.Address)
	return string(data), err
}

func (b *BaseRepo) CountNodesByType(nodeType string) (int64, error) {
	var count int64
	err := b.SqlDb.Model(&model.Node{}).Where("type = ?", nodeType).Count(&count).Error
	return count, err
}

func (b *BaseRepo) GetAllNodes() (*[]model.Node, error) {
	var nodes []model.Node
	err := b.SqlDb.Find(&nodes).Error
	return &nodes, err
}

func (b *BaseRepo) DeleteRelationsByNodeID(nodeID string) error {
	return b.SqlDb.Where("from_id = ? OR to_id = ?", nodeID, nodeID).Delete(&model.Relation{}).Error
}

func (b *BaseRepo) DeleteNodeById(id string) error {
	return b.SqlDb.Where("id = ?", id).Delete(&model.Node{}).Error
}
func (b *BaseRepo) GetNodeIdByPath(path string, id string) ([]string, error) {
	var nodes []string
	err := b.SqlDb.Model(model.Node{}).Where("path LIKE=?", path+"/"+id+"%").Error
	return nodes, err
}
func (b *BaseRepo) GetRelationByNodeId(nodes []string) ([]model.Relation, error) {
	var relation []model.Relation
	err := b.SqlDb.Model(model.Relation{}).Where("from_id IN ? AND to_id IN ?", nodes, nodes).Find(&relation).Error
	return relation, err
}
func (b *BaseRepo) GetNodeMessageByPath(path string, id string) ([]model.Node, error) {
	var nodes []model.Node
	if path != "root" {
		path = path + "/" + id + "%"
	} else {
		path = path + "%"
	}
	err := b.SqlDb.Model(model.Node{}).Where("path LIKE ?", path).Find(&nodes).Error
	return nodes, err
}
func (b *BaseRepo) GetNodesByRootPath(rootPath string) ([]model.Node, error) {
	var nodes []model.Node
	// 根节点自身 (path == "root") + 所有子节点 (path LIKE "root/%")
	err := b.SqlDb.Where("path = ? OR path LIKE ?", rootPath, rootPath+"/%").Find(&nodes).Error
	return nodes, err
}
func (n *BaseRepo) GetRecentNotes(limit int) ([]model.Node, error) {
	var notes []model.Node
	err := n.SqlDb.Where("type = ?", "note").
		Order("update_time DESC").
		Limit(limit).
		Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (n *BaseRepo) GetTagById(nodes []string) ([]model.TagMessage, error) {
	var res []model.TagMessage
	err := n.SqlDb.Model(model.Node{}).Where("id IN ?", nodes).Find(&res).Error
	return res, err
}

func (b *BaseRepo) GetTagRelationCounts() (map[string]int, error) {
	var relations []model.Relation
	err := b.SqlDb.Where("type = ?", "tag").Find(&relations).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, rel := range relations {
		counts[rel.ToID]++
	}
	return counts, nil
}

func (b *BaseRepo) GetTagsByNodeID(nodeID string) (*[]model.Relation, error) {
	var relations []model.Relation
	err := b.SqlDb.Where("to_id = ? AND type = ?", nodeID, "tag").Find(&relations).Error
	if err != nil {
		return nil, err
	}
	return &relations, nil
}
