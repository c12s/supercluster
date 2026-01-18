package handler

type AppDTO struct {
	ApplicationId     string `json:"applicationID"`
	ParentNamespaceId string `json:"parentNamespaceId"`
	SizeKB            int    `json:"sizeKB"`
}
