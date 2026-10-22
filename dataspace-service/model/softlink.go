package model

type SLType int

const (
	Group SLType = iota
	Others
)

type Softlink struct {
	SoftlinkID          string      `json:"softlinkID"`
	Application         Application `json:"app"`
	DataSpaceItemPath   string      `json:"path"`
	StoredProcedurePath string      `json:"storedProcedurePath"`
	JsonParameters      string      `json:"jsonParameters"`
	Type                SLType      `json:"slType"`
	TriggerPath         string      `json:"triggerPath"`
	EventTopic          string      `json:"eventTopic"`
}
