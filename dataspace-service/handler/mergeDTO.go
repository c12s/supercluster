package handler

type MergeDTO struct {
	Application1Id string `json:"app1Id"`
	Namespace1Id   string `json:"ns1Id"`
	Application2Id string `json:"app2Id"`
	Namespace2Id   string `json:"ns2Id"`
	DeleteLinks    bool   `json:"deleteLinks"`
}
