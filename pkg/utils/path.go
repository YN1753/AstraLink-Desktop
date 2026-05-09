package utils

import (
	"os"
	"path/filepath"
)

func GetAppStoragePath() string {

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// 便携版判断
	portableFlag := filepath.Join(exeDir, ".portable")
	localDataDir := filepath.Join(exeDir, "data")

	if _, err := os.Stat(portableFlag); err == nil {
		_ = os.MkdirAll(localDataDir, 0755) // 确保 data 夹存在
		return localDataDir
	}

	// 安装版模式

	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataPath := filepath.Join(configDir, "AstraLink")
		_ = os.MkdirAll(appDataPath, 0755)
		return appDataPath
	}

	// 防止获取不到 AppData
	_ = os.MkdirAll(localDataDir, 0755)
	return localDataDir
}

func IsPortable() bool {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	_, err := os.Stat(filepath.Join(exeDir, ".portable"))
	return err == nil
}
