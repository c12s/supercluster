package model

type State int

const (
	Open State = iota
	Closed
	Custom
)

type DataSpaceItem struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	SizeKB      int    `json:"sizeKb"`
	State       State  `json:"state"`
	HasSchema   bool   `json:"hasSchema"`
	Permissions string `json:"permissions"`
	IsLeaf      bool   `json:"isLeaf"`
}

func (dsi DataSpaceItem) GetFullPath() string {
	return dsi.Path + "/" + dsi.Name
}

func (dsi *DataSpaceItem) SetDefaultPermissions() {
	dsi.Permissions = "-rwx------"
}

type DataSpace struct {
	DataSpaceId string   `json:"dataSpaceId"`
	SizeKB      int      `json:"sizeKB"`
	UsedKB      int      `json:"usedKB"`
	Root        string   `json:"root"`
	OpenItems   []string `json:"openItems"`
}
