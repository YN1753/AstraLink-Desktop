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

func NewBaseRepo(db *gorm.DB, path string, bleveIndex bleve.Index) *BaseRepo {
	return &BaseRepo{
		SqlDb:      db,
		BasePath:   path,
		BleveIndex: bleveIndex,
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
	return b.BasePath
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
func (b *BaseRepo) GetRecentNotes(limit int) ([]model.Node, error) {
	var notes []model.Node
	err := b.SqlDb.Where("type = ?", "note").
		Order("update_time DESC").
		Limit(limit).
		Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (b *BaseRepo) GetTagById(nodes []string) ([]model.TagMessage, error) {
	var res []model.TagMessage
	err := b.SqlDb.Model(model.Node{}).Where("id IN ?", nodes).Find(&res).Error
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
		counts[rel.FromID]++ // from_id is the tag ID (from LinkTagToNode)
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

func (b *BaseRepo) DeleteRelation(rel model.DeleteRelationReq) error {
	return b.SqlDb.Model(model.Relation{}).Where("from_id = ? AND to_id = ? AND type = ?", rel.FromId, rel.ToId, rel.Type).Delete(&model.Relation{}).Error
}

// GetNoteIdsByTag 获取指定标签关联的所有笔记ID
func (b *BaseRepo) GetNoteIdsByTag(tagId string) ([]string, error) {
	var relations []model.Relation
	// from_id = tag, to_id = note (from LinkTagToNode: FromID=tagID, ToID=nodeID)
	err := b.SqlDb.Where("from_id = ? AND type = ?", tagId, "tag").Find(&relations).Error
	if err != nil {
		return nil, err
	}
	var noteIds []string
	for _, rel := range relations {
		noteIds = append(noteIds, rel.ToID)
	}
	return noteIds, nil
}

// GetNotesByTag 获取指定标签关联的所有笔记
func (b *BaseRepo) GetNotesByTag(tagId string) ([]model.NoteSearchResult, error) {
	noteIds, err := b.GetNoteIdsByTag(tagId)
	if err != nil {
		return nil, err
	}
	if len(noteIds) == 0 {
		return []model.NoteSearchResult{}, nil
	}
	var notes []model.Node
	err = b.SqlDb.Where("id IN ?", noteIds).Find(&notes).Error
	if err != nil {
		return nil, err
	}
	var results []model.NoteSearchResult
	for _, note := range notes {
		content, _ := b.GetNoteContent(note.ID)
		results = append(results, model.NoteSearchResult{
			ID:      note.ID,
			Title:   note.Name,
			Content: content,
		})
	}
	return results, nil
}

// IndexNote indexes a note's content in bleve
func (b *BaseRepo) IndexNote(id string, name string, content string) error {
	if b.BleveIndex == nil {
		return nil
	}
	doc := map[string]any{
		"id":      id,
		"title":   name,
		"content": content,
	}
	return b.BleveIndex.Index(id, doc)
}

// SearchNotes searches notes using bleve - supports prefix matching (p matches python, py, etc.)
func (b *BaseRepo) SearchNotes(queryStr string) ([]model.NoteSearchResult, error) {
	if b.BleveIndex == nil {
		return nil, errors.New("bleve index not initialized")
	}
	if queryStr == "" {
		return nil, nil
	}

	// 使用 QueryStringQuery 实现字段匹配
	query := bleve.NewQueryStringQuery("title:*" + queryStr + "* OR content:*" + queryStr + "*")
	req := bleve.NewSearchRequest(query)
	req.Size = 20
	req.Fields = []string{"title", "content"}

	results, err := b.BleveIndex.Search(req)
	if err != nil {
		return nil, err
	}

	var searchResults []model.NoteSearchResult
	for _, hit := range results.Hits {
		title := ""
		content := ""

		if hit.Fields != nil {
			if t, ok := hit.Fields["title"].(string); ok {
				title = t
			}
			if c, ok := hit.Fields["content"].(string); ok {
				content = c
			}
		}

		searchResults = append(searchResults, model.NoteSearchResult{
			ID:      hit.ID,
			Title:   title,
			Content: content,
		})
	}
	return searchResults, nil
}

// DeleteNoteFromIndex removes a note from the bleve index
func (b *BaseRepo) DeleteNoteFromIndex(id string) error {
	if b.BleveIndex == nil {
		return nil
	}
	return b.BleveIndex.Delete(id)
}

func (b *BaseRepo) GetNodeIdByTags(tags []string) ([]model.Relation, error) {
	var relations []model.Relation
	err := b.SqlDb.Model(model.Relation{}).Where("from_id IN ? AND type = ?", tags, "tag").Find(&relations).Error
	return relations, err
}

func (b *BaseRepo) GetNodeByNodeId(nodeId []string) ([]model.Node, error) {
	var nodes []model.Node
	err := b.SqlDb.Model(model.Node{}).Where("id IN ?", nodeId).Find(&nodes).Error
	return nodes, err
}
