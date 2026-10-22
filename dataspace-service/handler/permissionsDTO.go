package handler

type PermissionsDTO struct {
	DataSpaceItemPath string `json:"dsiPath"`
	Permissions       string `json:"permissions"`
}
