package service

import (
	"fmt"
	"log"
	"projekat/model"
	"slices"
	"strings"
)

type NamespaceService struct {
	store model.Store
}

func NewNamespaceService(store model.Store) *NamespaceService {
	return &NamespaceService{
		store: store,
	}
}

func (service *NamespaceService) RunDataDiscovery(namespaceId string) []string {
	apps, err := service.store.GetAllAppsForNamespace(namespaceId)
	evaluateError(err)
	var dataSpaceItemSchemas []string

	for _, app := range apps {

		ds, err := service.store.GetDataSpace(app.ApplicationId, app.DataSpaceId)
		evaluateError(err)
		itemSchemas, err := service.store.GetAllSchemas(ds.OpenItems)
		evaluateError(err)
		dataSpaceItemSchemas = append(dataSpaceItemSchemas, itemSchemas...)

	}

	return dataSpaceItemSchemas
}

func (service *NamespaceService) DeleteAppDefault(nsId, appId string) error {
	// default delete - app and dataspace are both deleted
	app, err := service.store.GetApp(nsId, appId)
	if err != nil {
		return err
	}
	ds, err := service.store.GetDataSpace(app.ApplicationId, app.DataSpaceId)
	if err != nil {
		return err
	}

	// delete all softlinks, checking only open items since softlink cannot be binded to any other type
	// TODO: fali odjava sa trigera?
	for _, item := range ds.OpenItems {
		err = service.store.DeleteAllSoftlinksForDataSpaceItem(item)
		if err != nil {
			return err
		}
	}

	// delete app, ds and all dsi and related schemas
	err = service.store.DeleteAppDefault(app)
	if err != nil {
		return err
	}

	return nil
}

func (service *NamespaceService) ChangeDSIState(appId string, dataSpaceItemPath string, state model.State, schema string) error {
	dsi, err := service.store.GetDataSpaceItem(dataSpaceItemPath)
	if err != nil {
		return err
	}

	if dsi.State == state {
		return nil
	}

	dsiParent, err := service.store.GetDataSpaceItem(dsi.Path)
	if err != nil {
		return err
	}

	if dsiParent.State != model.Custom {
		return fmt.Errorf("cannot change state because parent's state is not custom")
	}

	dsId := strings.Split(dsi.Path, "/")[0]
	ds, err := service.store.GetDataSpace(appId, dsId)
	if err != nil {
		return err
	}

	if state == model.Open {
		if schema == "" {
			return fmt.Errorf("no schema")
		}
		// Changes the state of the node and all its descendants.
		children, err := service.store.ChangeStateForAllChildren(dsi.GetFullPath(), state, true)
		if err != nil {
			return err
		}

		ds.OpenItems = append(ds.OpenItems, children...)
		// All children use the same schema as this DSI, since their state is enforced by this node.
		for _, child := range children {
			service.store.PutSchema(child, schema)
		}
		service.store.PutDataSpace(appId, ds)

	} else if state == model.Closed {
		// Changes the state of the node and all its descendants.
		children, err := service.store.ChangeStateForAllChildren(dsi.GetFullPath(), state, false)
		if err != nil {
			return err
		}
		var newOpenItems []string
		for _, item := range ds.OpenItems {
			if !strings.HasPrefix(item, dsi.GetFullPath()) {
				newOpenItems = append(newOpenItems, item)
			}
		}
		ds.OpenItems = newOpenItems
		for _, child := range children {
			sls, err := service.store.GetAllSoftLinksForDataSpaceItem(child)
			if err != nil {
				return err
			}
			for _, sl := range sls {
				if sl.TriggerPath != "" && sl.EventTopic != "" {
					err = RegisterEventForTrigger(sl.TriggerPath, sl.EventTopic, false)
					if err != nil {
						return err
					}
				}
			}
			err = service.store.DeleteAllSoftlinksForDataSpaceItem(child)
			if err != nil {
				return err
			}
		}
		service.store.PutDataSpace(appId, ds)

	} else {
		dsi.State = state
		err = service.store.PutDataSpaceItem(dsi)
		if err != nil {
			return err
		}

		for indx, item := range ds.OpenItems {
			if strings.HasPrefix(item, dsi.GetFullPath()) {
				ds.OpenItems = slices.Delete(ds.OpenItems, indx, indx)
			}
		}
		service.store.PutDataSpace(appId, ds)
	}
	return nil
}

func (service *NamespaceService) PutSchema(dataSpaceItemPath string, schema string) error {
	if schema == "" {
		return fmt.Errorf("no schema")
	}
	dsi, err := service.store.GetDataSpaceItem(dataSpaceItemPath)
	if err != nil {
		return err
	}
	if !dsi.HasSchema {
		dsi.HasSchema = true
		err = service.store.PutDataSpaceItem(dsi)
		if err != nil {
			return err
		}
	}

	err = service.store.PutSchema(dataSpaceItemPath, schema)
	if err != nil {
		return err
	}

	return nil
}

func (service *NamespaceService) ChangePermissions(dataSpaceItemPath string, permissions string) error {
	dsi, err := service.store.GetDataSpaceItem(dataSpaceItemPath)
	if err != nil {
		return nil
	}

	dsi.Permissions = permissions

	softlinks, err := service.store.GetAllSoftLinksForDataSpaceItem(dsi.GetFullPath())

	if err != nil {
		return err
	}

	var toDelete []model.Softlink
	for _, sl := range softlinks {
		var base int
		if sl.Type == model.Group {
			base = 0

		} else if sl.Type == model.Others {
			base = 3
		}

		if dsi.Permissions[5+base] != 's' {
			toDelete = append(toDelete, sl)
			if sl.TriggerPath != "" && sl.EventTopic != "" {
				err = RegisterEventForTrigger(sl.TriggerPath, sl.EventTopic, false)
				if err != nil {
					return err
				}
			}
		}
		if dsi.Permissions[6+base] != 'x' {
			sl.StoredProcedurePath = ""
			sl.JsonParameters = ""
			if sl.TriggerPath != "" && sl.EventTopic != "" {
				err = RegisterEventForTrigger(sl.TriggerPath, sl.EventTopic, false)
				if err != nil {
					return err
				}
			}
			sl.TriggerPath = ""
			sl.EventTopic = ""
			service.store.PutSoftlink(&sl)
		}
	}

	if len(toDelete) > 0 {
		err = service.store.DeleteAllSoftlinksFromList(toDelete)
		if err != nil {
			return err
		}
	}

	err = service.store.PutDataSpaceItem(dsi)
	if err != nil {
		return err
	}

	return nil
}

func evaluateError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
