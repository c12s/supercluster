package model

type Application struct {
	ApplicationId     string `json:"applicationID"`
	ParentNamespaceId string `json:"parentNamespaceId"`
	DataSpaceId       string `json:"dataSpaceId"`
	FreeSpaceKB       int    `json:"FreeSpaceKB"`
}
