package model

type UpdateInfo struct {
	HasUpdate   bool   `json:"hasUpdate"`
	LatestVer   string `json:"latestVer"`
	CurrentVer  string `json:"currentVer"`
	ReleaseNote string `json:"releaseNote"`
	ExeURL      string `json:"exeUrl"`
	ZipURL      string `json:"zipUrl"`
	ExeSize     int64  `json:"exeSize"`
	ZipSize     int64  `json:"zipSize"`
	IsPortable  bool   `json:"isPortable"`
}
