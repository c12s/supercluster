package handler

type DataSpaceItemDTO struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	SizeKB      int    `json:"sizeKb"`
	State       int    `json:"state"`
	HasSchema   bool   `json:"hasSchema"`
	Permissions string `json:"permissions"`
	AppID       string `json:"appID"`
	Schema      string `json:"schema"`
}
