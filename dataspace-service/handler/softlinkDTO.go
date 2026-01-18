package handler

type SoftlinkDTO struct {
	Application1Id      string `json:"application1ID"`
	Application2Id      string `json:"application2ID"`
	Namespace1Id        string `json:"namespace1ID"`
	Namespace2Id        string `json:"namespace2ID"`
	DataSpaceItemPath   string `json:"dataSpaceItemPath"`
	StoredProcedurePath string `json:"storedProcedurePath"`
	JsonParameters      string `json:"jsonParameters"`
	TriggerPath         string `json:"triggerPath"`
	EventTopic          string `json:"eventTopic"`
}
