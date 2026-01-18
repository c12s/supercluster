package model

type Namsespace struct {
	ParentNamspaceId string `json:"parentNamspaceId"`
	Id               string `json:"id"`
	FreeSpaceKB      int    `json:"freeSpaceKB"`
}
