package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"projekat/model"
	"projekat/proto/message"
	"strings"

	"github.com/nats-io/nats.go"
)

type EventDTO struct {
	EventTopic string `json:"event"`
	AddEvent   bool   `json:"add"`
}

type ApplicationService struct {
	store    model.Store
	conn     *nats.Conn
	meridian message.MeridianClient
}

func NewApplicationService(store model.Store, conn *nats.Conn, meridian message.MeridianClient) *ApplicationService {
	return &ApplicationService{
		store:    store,
		conn:     conn,
		meridian: meridian,
	}
}

func (service *ApplicationService) RunApplication(applicationId, parentNamespaceId string, sizeKB int) (*model.Application, error) {
	app := model.Application{
		ApplicationId:     applicationId,
		ParentNamespaceId: parentNamespaceId,
		DataSpaceId:       applicationId,
		FreeSpaceKB:       sizeKB / 2,
	}

	err := service.store.PutApp(&app)

	if err != nil {
		return nil, err
	}

	fmt.Printf("ApplicationId: %s, ParentNamespaceId: %s\n", app.ApplicationId, app.ParentNamespaceId)
	root := model.DataSpaceItem{Name: "Root", Path: app.ApplicationId, SizeKB: 1, IsLeaf: true, State: model.Custom, HasSchema: false}
	ds := model.DataSpace{
		DataSpaceId: app.ApplicationId,
		SizeKB:      sizeKB / 2,
		UsedKB:      0,
		Root:        root.Path + "/" + root.Name,
		OpenItems:   []string{},
	}

	service.CreateDataSpaceItem(app.ApplicationId, &root, "", true)

	err = service.store.PutDataSpace(app.ApplicationId, &ds)
	if err != nil {
		return nil, err
	}
	fmt.Printf("DataSpace ds: %d;\n", ds.SizeKB)

	//after dataspace creation, the hard link is created between application and ds
	hardlink := model.Hardlink{
		ApplicationID: app.ApplicationId,
		DataSpaceID:   ds.DataSpaceId,
	}
	err = service.store.PutHardlink(&hardlink)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (service *ApplicationService) CreateDataSpaceItem(appID string, dsi *model.DataSpaceItem, schema string, root bool) (*model.DataSpaceItem, error) {

	if !root {
		ds, err := service.store.GetDataSpace(appID, strings.Split(dsi.Path, "/")[0])
		if err != nil {
			return nil, err
		}

		dsiParent, err := service.store.GetDataSpaceItem(dsi.Path)
		if err != nil {
			return nil, err
		}

		// If the parent is not in custom mode, the child inherits the parent's state.
		if dsiParent.State != model.Custom {
			dsi.State = dsiParent.State
		}

		dsi.HasSchema = schema != ""

		if dsi.State != model.Open {
			dsi.SetDefaultPermissions()
		}

		if dsi.State == model.Open && dsi.HasSchema {
			ds.OpenItems = append(ds.OpenItems, dsi.GetFullPath())
		}

		if ds.UsedKB+dsi.SizeKB > ds.SizeKB {
			log.Fatal("cannot add dataSpaceItem - no available resources")
		}

		ds.UsedKB += dsi.SizeKB
		err = service.store.PutDataSpace(appID, ds)
		if err != nil {
			return nil, err
		}

		if dsiParent.IsLeaf {
			dsiParent.IsLeaf = false
			err = service.store.PutDataSpaceItem(dsiParent)
			if err != nil {
				return nil, err
			}
		}
	}

	if dsi.HasSchema {
		err := service.store.PutSchema(dsi.Path+"/"+dsi.Name, schema)
		if err != nil {
			return nil, err
		}
	}
	err := service.store.PutDataSpaceItem(dsi)
	if err != nil {
		return nil, err
	}

	return dsi, nil

}

// app2 wants to create sl to app1's dataa
func (service *ApplicationService) CreateSoftlink(app1, app2 *model.Application, dataSpaceItemPath string, storedProcedurePath string, jsonParams string, triggerPath string, eventTopic string, slId string) (*string, error) {
	dsi, err := service.store.GetDataSpaceItem(dataSpaceItemPath)
	sltype := model.Others

	if err != nil {
		return nil, err
	}
	//apps are not in the same namespace
	if app1.ParentNamespaceId != app2.ParentNamespaceId {

		if dsi.Permissions[8] != 's' {
			return nil, fmt.Errorf("no privilages for this data - type others")
		}

		//no permissions for execution - empty strings
		if dsi.Permissions[9] != 'x' {
			storedProcedurePath = ""
			jsonParams = ""
			triggerPath = ""
			eventTopic = ""
		}
	}

	// apps are in the same namespace
	if app1.ParentNamespaceId == app2.ParentNamespaceId {
		if dsi.Permissions[5] != 's' {
			return nil, fmt.Errorf("no privilages for this data - type group")
		}

		//no permissions for execution - empty strings
		if dsi.Permissions[6] != 'x' {
			storedProcedurePath = ""
			jsonParams = ""
			triggerPath = ""
			eventTopic = ""
		}

		sltype = model.Group
	}

	if triggerPath != "" && eventTopic != "" {
		err = RegisterEventForTrigger(triggerPath, eventTopic, true)
		if err != nil {
			return nil, err
		}
	}

	if slId == "" {
		slId = app2.ApplicationId + "+" + dataSpaceItemPath
	}

	softlink := model.Softlink{
		SoftlinkID:          slId,
		Application:         *app2,
		DataSpaceItemPath:   dataSpaceItemPath,
		StoredProcedurePath: storedProcedurePath,
		JsonParameters:      jsonParams,
		Type:                sltype,
		TriggerPath:         triggerPath,
		EventTopic:          eventTopic,
	}
	err = service.store.PutSoftlink(&softlink)
	if err != nil {
		return nil, err
	}

	err = service.createTopicForSoftLink(&softlink)
	if err != nil {
		return nil, err
	}
	return &softlink.SoftlinkID, nil
}

func (service *ApplicationService) createTopicForSoftLink(softlink *model.Softlink) error {

	_, err := service.conn.QueueSubscribe(softlink.SoftlinkID, "softlinks", func(message *nats.Msg) {
		fmt.Printf("RECEIVED MESSAGE: %s\n", string(message.Data))

		sl, err := service.store.GetSoftlink(softlink.DataSpaceItemPath, softlink.Application.ApplicationId)

		if err != nil {
			return
		}

		if sl.StoredProcedurePath == "" {
			fmt.Println("no path for stored procedure")
			return
		}

		url := sl.StoredProcedurePath

		if sl.JsonParameters == "" {
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error making GET request:", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}
			fmt.Println("Response:", string(body))
		} else {
			if !isValidJSON(sl.JsonParameters) {
				// TODO: videti ovde za te povratne vrednosti sta zezaju ovi errori
				fmt.Println("JSON is not valid!")
				return
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(sl.JsonParameters)))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request:", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}
			fmt.Println("Response:", string(body))

		}
	})

	return err
}

func RegisterEventForTrigger(triggerPath string, eventTopic string, add bool) error {
	// if add is true, register, if false, unregister
	event := EventDTO{EventTopic: eventTopic, AddEvent: add}
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", triggerPath, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}
	fmt.Println("Response:", string(body))

	return nil
}
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func (service *ApplicationService) MergeDataSpaces(app1 model.Application, app2 model.Application, deleteLinks bool) error {
	// merge app1's ds with app2's ds
	ds1, err := service.store.GetDataSpace(app1.ApplicationId, app1.DataSpaceId)
	if err != nil {
		return err
	}

	ds2, err := service.store.GetDataSpace(app2.ApplicationId, app2.DataSpaceId)
	if err != nil {
		return err
	}

	dsis1, err := service.store.GetAllDataSpaceItemsForDataSpace(ds1.DataSpaceId)

	if err != nil {
		return err
	}
	//false root in order to avoid conflict names TODO: check / after root
	falseRoot := model.DataSpaceItem{Name: "Root", Path: ds2.DataSpaceId + "/Root/" + ds1.DataSpaceId, SizeKB: 1, IsLeaf: true, State: model.Custom, HasSchema: false}
	service.CreateDataSpaceItem(app2.ApplicationId, &falseRoot, "", true)
	for _, dsiPath := range dsis1 {
		dsi, err := service.store.GetDataSpaceItem(dsiPath)
		if err != nil {
			return err
		}

		oldPath := dsi.GetFullPath()
		dsi.Path = ds2.DataSpaceId + "/Root/" + dsi.Path
		//dsi2id/root/dsi1id/root/....

		if deleteLinks {
			fmt.Println("brisanje")
			service.store.DeleteAllSoftlinksForDataSpaceItem(dsiPath)
			dsi.State = model.Closed
		} else {
			//recreated softlinks, now they are binded to new ds
			softlinks, err := service.store.GetAllSoftLinksForDataSpaceItem(dsiPath)
			if err != nil {
				return err
			}

			for _, sl := range softlinks {
				app, err := service.store.GetApp(sl.Application.ParentNamespaceId, sl.Application.ApplicationId)
				if err != nil {
					return err
				}
				_, err = service.CreateSoftlink(&app2, app, dsi.GetFullPath(), sl.StoredProcedurePath, sl.JsonParameters, sl.TriggerPath, sl.EventTopic, sl.SoftlinkID)
				if err != nil {
					return err
				}
			}

			err = service.store.DeleteAllSoftlinksForDataSpaceItem(oldPath)
			if err != nil {
				return err
			}
			if dsi.State == model.Open {
				ds2.OpenItems = append(ds2.OpenItems, dsiPath)
			}
		}

		//replace dsi and schema if exists
		err = service.store.ReplaceDataSpaceItemAndSchema(oldPath, dsi)
		if err != nil {
			return err
		}

	}
	//save ds2 because list openItems is changed
	err = service.store.PutDataSpace(app2.ApplicationId, ds2)
	if err != nil {
		return err
	}
	//ds1 and app1 can be deleted now
	err = service.store.DeleteAppDefault(&app1)
	if err != nil {
		return err
	}

	return nil
}

func (service *ApplicationService) DeleteAppWithMerge(app1Id, app2Id string, ns1Id, ns2Id string, deleteLinks bool) error {

	// app which is going to be deleted
	app1, err := service.store.GetApp(ns1Id, app1Id)
	if err != nil {
		return err
	}

	// new owner of the data
	app2, err := service.store.GetApp(ns2Id, app2Id)
	if err != nil {
		return err
	}

	ds1, err := service.store.GetDataSpace(app1Id, app1.DataSpaceId)
	if err != nil {
		return err
	}

	//videti za koje resurse da se pita
	resp, err := service.meridian.BorrowResources(context.Background(), &message.BorrowResourcesReq{App1Id: app1Id, App2Id: app2Id, Namespace1Id: ns1Id, Namespace2Id: ns2Id, DiskResources: float64(ds1.SizeKB)})
	if err != nil {
		return err
	}

	if resp.Done {
		fmt.Println(resp.Reply)
		service.MergeDataSpaces(*app1, *app2, deleteLinks)
	} else {
		fmt.Println(resp.Reply)
		return errors.New(resp.Reply)
	}
	return nil
}
