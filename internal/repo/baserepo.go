package repo

import (
	"astralink/internal/model"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseRepo struct {
	SqlDb    *gorm.DB
	BasePath string
}

func NewBaseRepo(db *gorm.DB, path string) *BaseRepo {
	return &BaseRepo{
		SqlDb:    db,
		BasePath: path,
	}
}

func (b *BaseRepo) MergeNode(node *model.Node) (string, error) {
	if node == nil {
		return "", nil
	}
	// 使用 OnConflict 确保在 ID 冲突时执行全字段更新
	return node.ID, b.SqlDb.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(node).Error
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

	// 2. 构造文件夹路径（如：.../AppData/Roaming/AstraLink/notes）
	targetDir := filepath.Join(root, subDir)

	// 3. 构造完整文件路径
	fullPath := filepath.Join(targetDir, filename)

	// 4. 强制创建目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", errors.New("无法初始化存储空间")
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	fmt.Println(fullPath)
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
	return b.SqlDb.Model(&model.Node{}).Where("id = ?", id).UpdateColumn(tab, value).Error
}
