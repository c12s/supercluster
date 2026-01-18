# supercluster
Service to manage dataspaces for c12s distributed cloud computing platform.

## API Endpoints

### Run Application
- **POST** `/runApp`  

Triggers execution of an application and creates dataspaces for it.

The request body must be a JSON object matching the following structure:

```json
{
  "applicationID": "app-123",
  "parentNamespaceId": "namespace-456",
  "sizeKB": 1024
}
```

### Data Discovery
- **GET** `/dataDiscovery/{nsId}`  

Runs data discovery and retrieves information (schemas) about all dataspace items in the open state for the given namespace ID. Returns a list of strings representing the schemas of all open dataspace items.

### Add DataSpace Item
- **POST** `/addDataSpaceItem`  

Adds a new dataspace item to the dataspace. 

The request body must be a JSON object matching the following structure. The `state` field can have values `0`, `1`, or `2`, representing the open, closed, and custom states, respectively.


```json
{
    "path": "appId/Root/folder",
    "name": "example file",
    "sizeKb": 1,
    "state": 0,
    "hasSchema": true,
    "permissions": "-rw-rsxrsx",
    "appID": "appId",
    "schema": "example schema"
}
```

### Delete Application
- **DELETE** `/deleteApp`  

Deletes an existing application and its dataspace, along with all associated dataspace items. This is the default way of deleting applications.

The request body must be a JSON object matching the following structure:

```json
{
    "NamespaceId": "namespaceId",
	"AplicationID": "appId"
}
```

### Create Softlink 
- **POST** `/softlink`  

Creates a soft link on specific dataspace item.

The request body must be a JSON object matching the following structure.

Application `1` represents the owner of the dataspace item, and namespace `1` is its corresponding namespace.  
Application `2` is the application that wants to create a soft link and belongs to namespace `2`.

If a stored procedure should be binded to the soft link, its path must be provided in the `storedProcedurePath` field. If specific parameters need to be passed to the stored procedure, they must be provided as a JSON-formatted string in the `jsonParameters` field. imilarly, if a trigger-event application is used, the paths to the trigger and the event topic must be provided.  

If the application does not have permission to execute collaborative applications, these fields will be ignored.

```json
{
  "application1ID": "app1",
  "application2ID": "app2",
  "namespace1ID": "ns1",
  "namespace2ID": "ns2",
  "dataSpaceItemPath": "app1/Root/file1",
  "storedProcedurePath": "http://localhost:8001",
  "jsonParameters": "",
  "triggerPath": "http://localhost:8000",
  "eventTopic": "eventTopic"
}
```

### Change DataSpaceItem State
- **PUT** `/changeState`  

Updates the state of a dataspace item. If the new state is `open`, a schema is expected to be provided in the `schema` field.

The request body must be a JSON object matching the following structure.

```json
{
    "applicationID": "appId",
    "dsiPath": "appId/Root/folder",
    "state": 0,
    "schema": "example schema" 
}
```

### Change Permissions
- **PUT** `/changePermissions`  

Modifies permissions for a dataspace item.

The request body must be a JSON object matching the following structure.

```json
{
    "dsiPath": "app1/Root/folder0",
    "permissions": "-rwxr-xr-x"
}
```

### Put Schema
- **PUT** `/putSchema`  

Uploads or updates a data schema.

The request body must be a JSON object matching the following structure.

```json
{
    "dsiPath": "app1/Root/folder0",
    "schema": "new schema"
}
```


### Delete Application with Merge
- **DELETE** `/deleteAppMerge`  

Deletes an application and merges related resources. 

The request body must be a JSON object matching the following structure. Application `1` from namespace `1` will be deleted, and application `2` from namespace `2` will take over the data. After the request is processed, their data spaces will be merged.

The `deleteLinks` field specifies whether soft links should be deleted before merging the data spaces.


```json
{
  "app1Id": "application-1",
  "ns1Id": "namespace-1",
  "app2Id": "application-2",
  "ns2Id": "namespace-2",
  "deleteLinks": true
}
```

## Unikernel examples

The `unikernels` folder contains examples for stored procedure, event, and trigger applications, which are intended to be run as unikernels.

To run these examples, **Unikraft** and its **kraft CLI** are required. Each example includes a `run.sh` script, so any unikernel can be started by running:

```bash
./run.sh
```

The command supports the following optional arguments:

- `--plat`: defaults to `qemu` if omitted.
- `--arch`: defaults to your system's CPU architecture if omitted.


```shell
kraft run \
  --rm \
  -p 8000:8000 \
  --plat qemu \
  --arch x86_64 \
  .
  ```

In these examples, the host is accessed via the IP `10.0.2.2`, which is the default gateway provided by QEMU when using user-mode networking (SLIRP).  
Note that this address may be different if you run the unikernel in another environment or with a different network configuration.
