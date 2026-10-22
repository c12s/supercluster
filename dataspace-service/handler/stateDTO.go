package handler

type StateDTO struct {
	ApplicationId     string `json:"applicationID"`
	DataSpaceItemPath string `json:"dsiPath"`
	State             int    `json:"state"`
	Schema            string `json:"schema"`
}
