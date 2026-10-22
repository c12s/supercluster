package model

type Store interface {
	GetApp(namespaceID, appID string) (*Application, error)
	GetDataSpace(applicationID string, dataSpaceId string) (*DataSpace, error)
	GetDataSpaceItem(path string) (*DataSpaceItem, error)
	PutApp(application *Application) error
	PutDataSpace(applicationID string, dataSpace *DataSpace) error
	PutDataSpaceItem(dataSpaceItem *DataSpaceItem) error
	PutHardlink(hardlink *Hardlink) error
	PutSoftlink(softlink *Softlink) error
	DeleteAllSoftlinksForDataSpaceItem(dataSpaceItemPath string) error
	GetAllAppsForNamespace(namespaceId string) ([]Application, error)
	GetAllDataSpaceItemsForDataSpace(dataSpaceId string) ([]string, error)
	DeleteAppDefault(app *Application) error
	GetNamespace(namespaceID string) (*Namsespace, error)
	PutSchema(path string, schema string) error
	GetAllSchemas(schemas []string) ([]string, error)
	ChangeStateForAllChildren(dataSpaceItemPath string, state State, schema bool) ([]string, error)
	GetSoftlink(dsiPath string, appId string) (*Softlink, error)
	GetAllSoftLinksForDataSpaceItem(dsiPath string) ([]Softlink, error)
	DeleteAllSoftlinksFromList(softlinks []Softlink) error
	ReplaceDataSpaceItemAndSchema(oldPath string, dsi *DataSpaceItem) error
}
