package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"projekat/model"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	//namespace/namespaceid
	namespaceKey = "namespace/%s"

	//schema/pathToDSI
	schemaKey = "schema/%s"

	//application/namespaceId/applicationId
	applicationKey = "application/%s/%s"

	//dataspace/applicationid/dataspaceid
	dataSpaceKey = "dataspace/%s/%s"

	//dataspaceitem/path/name -> dataspaceID/Root/...
	dataSpaceItemKey = "dataspaceitem/%s/%s"

	//hardlink/applicationid/dataspaceitemid
	hardlinkKey = "hardlink/%s/%s"

	//softlink/dataspaceItemPath/applicationid
	softlinkKey = "softlink/%s/%s"
)

func key_one(id, template string) string {
	return fmt.Sprintf(template, id)
}

func key(id1, id2, template string) string {
	return fmt.Sprintf(template, id1, id2)
}

type DB struct {
	Kv     clientv3.KV
	Client *clientv3.Client
}

func New(endpoints []string, timeout time.Duration) (model.Store, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: timeout,
		Endpoints:   endpoints,
	})

	if err != nil {
		return nil, err
	}

	return &DB{
		Kv:     clientv3.NewKV(cli),
		Client: cli,
	}, nil
}

func (db *DB) Close() { db.Client.Close() }

func (db *DB) PutApp(app *model.Application) error {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	key := key(app.ParentNamespaceId, app.ApplicationId, applicationKey)

	jsonData, err := json.Marshal(app)
	if err != nil {
		return err
	}

	_, err = db.Kv.Put(ctx, key, string(jsonData))

	return err

}

func (db *DB) PutNamespace(namespace *model.Namsespace) error {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	key := key_one(namespace.Id, namespaceKey)

	jsonData, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	_, err = db.Kv.Put(ctx, key, string(jsonData))

	return err

}

func (db *DB) PutDataSpace(applicationID string, dataSpace *model.DataSpace) error {

	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	key := key(applicationID, dataSpace.DataSpaceId, dataSpaceKey)

	jsonData, err := json.Marshal(dataSpace)
	if err != nil {
		return err
	}

	_, err = db.Kv.Put(ctx, key, string(jsonData))

	return err
}

func (db *DB) PutDataSpaceItem(dataSpaceItem *model.DataSpaceItem) error {

	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	key := key(dataSpaceItem.Path, dataSpaceItem.Name, dataSpaceItemKey)

	jsonData, err := json.Marshal(dataSpaceItem)
	if err != nil {
		return err
	}

	_, err = db.Kv.Put(ctx, key, string(jsonData))

	return err
}

func (db *DB) PutHardlink(hardlink *model.Hardlink) error {

	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	key := key(hardlink.ApplicationID, hardlink.DataSpaceID, hardlinkKey)

	jsonData, err := json.Marshal(hardlink)
	if err != nil {
		return err
	}

	_, err = db.Kv.Put(ctx, key, string(jsonData))

	return err
}

func (db *DB) PutSoftlink(softlink *model.Softlink) error {

	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	key := key(softlink.DataSpaceItemPath, softlink.Application.ApplicationId, softlinkKey)

	jsonData, err := json.Marshal(softlink)
	if err != nil {
		return err
	}

	_, err = db.Kv.Put(ctx, key, string(jsonData))

	return err
}

func (db *DB) GetApp(namespaceID, appID string) (*model.Application, error) {
	key := key(namespaceID, appID, applicationKey)
	resp, err := db.Kv.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		log.Fatalf("No data found for the key -app")
	}

	var app model.Application
	err = json.Unmarshal(resp.Kvs[0].Value, &app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (db *DB) GetNamespace(namespaceID string) (*model.Namsespace, error) {
	key := key_one(namespaceID, namespaceKey)
	resp, err := db.Kv.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		log.Fatalf("No data found for the key -ns")
	}

	var ns model.Namsespace
	err = json.Unmarshal(resp.Kvs[0].Value, &ns)
	if err != nil {
		return nil, err
	}

	return &ns, nil
}

func (db *DB) GetDataSpace(applicationID string, dataSpaceId string) (*model.DataSpace, error) {
	key := key(applicationID, dataSpaceId, dataSpaceKey)
	resp, err := db.Kv.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		log.Fatalf("No data found for the key - ds")
	}

	var ds model.DataSpace
	err = json.Unmarshal(resp.Kvs[0].Value, &ds)
	if err != nil {
		return nil, err
	}

	return &ds, nil
}

func (db *DB) GetDataSpaceItem(path string) (*model.DataSpaceItem, error) {
	resp, err := db.Kv.Get(context.Background(), "dataspaceitem/"+path)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, errors.New("no data found for the given key")
	}

	var dsi model.DataSpaceItem
	err = json.Unmarshal(resp.Kvs[0].Value, &dsi)
	if err != nil {
		return nil, err
	}

	return &dsi, nil
}

func (db *DB) DeleteAllSoftlinksForDataSpaceItem(dataSpaceItemPath string) error {
	prefix := "softlink/" + dataSpaceItemPath

	resp, err := db.Kv.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	ops := []clientv3.Op{}

	for _, kv := range resp.Kvs {
		ops = append(ops, clientv3.OpDelete(string(kv.Key)))
	}

	if len(ops) > 0 {
		_, err = db.Kv.Txn(context.Background()).Then(ops...).Commit()
		if err != nil {
			return err
		}
		fmt.Printf("Deleted %d keys\n", len(ops))
	} else {
		fmt.Println("No keys with requested prefix!")
	}

	return nil
}

func (db *DB) GetAllAppsForNamespace(namespaceId string) ([]model.Application, error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	prefix := "application/" + namespaceId
	resp, err := db.Kv.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var applications []model.Application

	for _, kv := range resp.Kvs {
		var app model.Application
		err = json.Unmarshal(kv.Value, &app)
		if err != nil {
			return nil, err
		}
		applications = append(applications, app)
	}
	return applications, nil
}

func (db *DB) GetAllDataSpaceItemsForDataSpace(dataSpaceId string) ([]string, error) {
	prefix := "dataspaceitem/" + dataSpaceId
	resp, err := db.Kv.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var itemsPaths []string
	for _, kv := range resp.Kvs {
		itemsPaths = append(itemsPaths, string(kv.Key))
	}

	return itemsPaths, nil
}

func (db *DB) DeleteAppDefault(app *model.Application) error {
	ops := []clientv3.Op{}
	prefix := "dataspaceitem/" + app.DataSpaceId
	resp, err := db.Kv.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		ops = append(ops, clientv3.OpDelete(string(kv.Key)))
	}

	prefix = "schema/" + app.DataSpaceId
	resp, err = db.Kv.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		ops = append(ops, clientv3.OpDelete(string(kv.Key)))
	}

	ops = append(ops, clientv3.OpDelete(key(app.ParentNamespaceId, app.ApplicationId, applicationKey)))
	ops = append(ops, clientv3.OpDelete(key(app.ApplicationId, app.DataSpaceId, dataSpaceKey)))

	_, err = db.Kv.Txn(context.Background()).Then(ops...).Commit()
	if err != nil {
		return err
	}

	fmt.Printf("%d \n", len(ops))
	return nil
}

func (db *DB) PutSchema(path string, schema string) error {

	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	key := key_one(path, schemaKey)

	_, err := db.Kv.Put(ctx, key, schema)

	return err
}

func (db *DB) GetAllSchemas(schemas []string) ([]string, error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	var schemaValues []string
	for _, schema := range schemas {
		key := key_one(schema, schemaKey)
		resp, err := db.Kv.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		if len(resp.Kvs) == 0 {
			log.Fatalf("No data found for the key -dd")
		}
		schemaValues = append(schemaValues, string(resp.Kvs[0].Value))
	}

	return schemaValues, nil
}

func (db *DB) ChangeStateForAllChildren(dataSpaceItemPath string, state model.State, schema bool) ([]string, error) {

	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	prefix := "dataspaceitem/" + dataSpaceItemPath
	resp, err := db.Kv.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var dsis []string
	ops := []clientv3.Op{}

	for _, kv := range resp.Kvs {
		var dsi model.DataSpaceItem
		err = json.Unmarshal(kv.Value, &dsi)
		if err != nil {
			return nil, err
		}
		dsi.State = state
		if schema {
			dsi.HasSchema = true
		}
		jsonData, err := json.Marshal(dsi)
		if err != nil {
			return nil, err
		}
		ops = append(ops, clientv3.OpPut(key(dsi.Path, dsi.Name, dataSpaceItemKey), string(jsonData)))
		dsis = append(dsis, dsi.GetFullPath())
	}

	_, err = db.Kv.Txn(context.Background()).Then(ops...).Commit()
	if err != nil {
		return nil, err
	}

	return dsis, nil
}

func (db *DB) GetSoftlink(dsiPath string, appId string) (*model.Softlink, error) {
	key := key(dsiPath, appId, softlinkKey)
	resp, err := db.Kv.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, errors.New("no softlink for given key")
	}

	var sl model.Softlink
	err = json.Unmarshal(resp.Kvs[0].Value, &sl)
	if err != nil {
		return nil, err
	}

	return &sl, nil
}

func (db *DB) GetAllSoftLinksForDataSpaceItem(dsiPath string) ([]model.Softlink, error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	prefix := "softlink/" + dsiPath
	resp, err := db.Kv.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var softlinks []model.Softlink

	for _, kv := range resp.Kvs {
		var sl model.Softlink
		err = json.Unmarshal(kv.Value, &sl)
		if err != nil {
			return nil, err
		}

		softlinks = append(softlinks, sl)
	}
	return softlinks, nil
}

func (db *DB) DeleteAllSoftlinksFromList(softlinks []model.Softlink) error {
	ops := []clientv3.Op{}

	for _, sl := range softlinks {
		ops = append(ops, clientv3.OpDelete(key(sl.DataSpaceItemPath, sl.Application.ApplicationId, softlinkKey)))
	}

	if len(ops) > 0 {
		_, err := db.Kv.Txn(context.Background()).Then(ops...).Commit()
		if err != nil {
			return err
		}
		fmt.Printf("Deleted %d keys\n", len(ops))
	} else {
		fmt.Println("No keys with requested prefix!")
	}

	return nil
}

func (db *DB) ReplaceDataSpaceItemAndSchema(oldPath string, dsi *model.DataSpaceItem) error {
	ops := []clientv3.Op{}

	jsonData, err := json.Marshal(dsi)
	if err != nil {
		return err
	}

	ops = append(ops, clientv3.OpDelete("dataspaceitem/"+oldPath))
	ops = append(ops, clientv3.OpPut(key(dsi.Path, dsi.Name, dataSpaceItemKey), string(jsonData)))
	if dsi.HasSchema {
		oldKey := key_one(oldPath, schemaKey)
		resp, err := db.Kv.Get(context.Background(), oldKey)
		if err != nil {
			return err
		}
		if len(resp.Kvs) == 0 {
			return errors.New("no schema for the given key")
		}
		schema := resp.Kvs[0].Value

		ops = append(ops, clientv3.OpDelete(oldKey))
		ops = append(ops, clientv3.OpPut(key_one(dsi.GetFullPath(), schemaKey), string(schema)))

	}

	_, err = db.Kv.Txn(context.Background()).Then(ops...).Commit()
	if err != nil {
		return err
	}
	return nil
}
