# JR Plugins

Lists of producer plugins for JR.
Currently the following plugins are included:


- `awsdynamodb`
- `azblobstorage`
- `azcosmosdb`
- `cassandra`
- `elastic`
- `gcs`
- `http`
- `luascript`
- `mongodb`
- `redis`
- `s3`
- `wasm`


# Building the plugins

Launch the `make`command with the target `compile`  and the plugins will be built in the `build/` folder


# Creating a plugin

The JR plugins should be in the `internal/plugin` package since they are not meant to be exposed externally.

To build a plugin `someplugin` the following steps are needed:

1. create the package `internal/plugin/someplugin`
2. implement the plugin in a file (e.g. `plugin.go`) with the following requirements:
  - the `plugin.go` file should have conditional build directives:
  ```golang
  //go:build plugin_someplugin
  // +build plugin_someplugin
  ```
  - a `doc.go` without conditional build directives must be included (with the plugin documentation)
  - the plugin should implement the ´plugin.Plugin´ interface type:
  ```golang
  type Plugin interface {
    jrpc.Producer
    Init(context.Context, []byte) error
}
```
  - in the `plugin.go` file register the plugin:

  ```golang
package someplugin
...
import (
        "github.com/jrnd-io/jr-plugins/internal/plugin"
)
const (
    Name = "someplugin"
)
func init() {
    plugin.RegisterPlugin(Name, &Plugin{})
}
type Plugin struct{
...
}
func (p *Plugin) Init(ctx context.Context, cfgBytes []byte) error{
    ...
}
func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string) (*jrpc.ProduceResponse, error) {
    ...
}

```

  - add the `someplugin` package  to the import in the `run.go` file:
  ```golang
  package main

  import(
    ...
    _ "github.com/jrnd-io/jr-plugins/internal/plugin/someplugin"
  )
  ```
   - add `someplugin`  to the list of plugins in `Makefile``
```make
PLUGINS=mongodb \
        azblobstorage \
        azcosmosdb \
        luascript \
        awsdynamodb \
        s3 \
        cassandra \
        gcs \
        elastic \
        redis \
        http \
        someplugin
```

# Plugin parameters

Di seguito sono elencati i parametri che è possibile passare ai plugin da `jr` da linea di comando con: `-p <emitter name>.<parameter name>=<parameter value>`

### AWS DynamoDB Plugin

| Parameter Name | Type   | Example                       | Description                          |
|----------------|--------|-------------------------------|--------------------------------------|
| table          | string | `-p <emitter name>.table=myTable` | The name of the DynamoDB table to use. |

### Azure Blob Storage Plugin

| Parameter Name      | Type   | Example                                      | Description                                      |
|---------------------|--------|----------------------------------------------|--------------------------------------------------|
| container.name      | string | `-p <emitter name>.container.name=myContainer` | The name of the Azure Blob Storage container.    |
| metadata.<key>      | string | `-p <emitter name>.metadata.key=value`       | Metadata to be added to the blob.                |

### Azure Cosmos DB Plugin

| Parameter Name      | Type   | Example                                      | Description                                      |
|---------------------|--------|----------------------------------------------|--------------------------------------------------|
| partition_key       | string | `-p <emitter name>.partition_key=myPartition`   | The value of the partition key for the item.     |
| database            | string | `-p <emitter name>.database=myDatabase`         | The name of the database to use.                  |
| container           | string | `-p <emitter name>.container=myContainer`       | The name of the container to use.                 |

### Cassandra Plugin

| Parameter Name      | Type   | Example                                      | Description                                      |
|---------------------|--------|----------------------------------------------|--------------------------------------------------|
| keyspace            | string | `-p <emitter name>.keyspace=myKeyspace`          | The keyspace to use for the Cassandra database.   |
| table               | string | `-p <emitter name>.table=myTable`                 | The table to insert data into.                    |
| consistency_level   | string | `-p <emitter name>.consistency_level=QUORUM`     | The consistency level for the operation.          |

### GCS Plugin

| Parameter Name      | Type   | Example                                      | Description                                      |
|---------------------|--------|----------------------------------------------|--------------------------------------------------|
| bucket              | string | `-p <emitter name>.bucket=myBucket`                     | The name of the Google Cloud Storage bucket.     |

### HTTP Plugin

| Parameter Name              | Type   | Example                                      | Description                                      |
|-----------------------------|--------|----------------------------------------------|--------------------------------------------------|
| endpoint.url                | string | `-p <emitter name>.endpoint.url=http://example.com`  | The URL of the endpoint to send the request to.  |
| endpoint.method             | string | `-p <emitter name>.endpoint.method=POST`              | The HTTP method to use (e.g., POST, PUT).        |
| endpoint.timeout            | int    | `-p <emitter name>.endpoint.timeout=10s`                | The timeout duration for the request.             |
| error_handling.expect_status_code | int | `-p <emitter name>.error_handling.expect_status_code=200` | Expected status code |
| error_handling.ignore_status_code  | int | `-p <emitter name>.error_handling.ignore_status_code=404` | Ignore status code |
| tls.insecure_skip_verify    | bool   | `-p <emitter name>.tls.insecure_skip_verify=true`     | Whether to skip TLS verification.                 |

### Lua Script Plugin

| Parameter Name | Type | Example | Description |
|----------------|------|---------|-------------|
| N/A            | N/A  | N/A     | No specific parameters are passed in configParams. |

### MongoDB Plugin

| Parameter Name | Type   | Example                                      | Description                                      |
|----------------|--------|----------------------------------------------|--------------------------------------------------|
| database        | string | `-p <emitter name>.database=myDatabase`            | The name of the database to use.                  |
| collection      | string | `-p <emitter name>.collection=myCollection`        | The name of the collection to insert data into.   |

### Redis Plugin

| Parameter Name | Type   | Example                                      | Description                                      |
|----------------|--------|----------------------------------------------|--------------------------------------------------|
| ttl            | int    | `-p <emitter name>.ttl=60s`                         | The time-to-live for the key in seconds.         |

### S3 Plugin

| Parameter Name | Type   | Example                                      | Description                                      |
|----------------|--------|----------------------------------------------|--------------------------------------------------|
| bucket         | string | `-p <emitter name>.bucket=myBucket`                     | The name of the S3 bucket to use.                |

### WASM Plugin

| Parameter Name | Type | Example | Description |
|----------------|------|---------|-------------|
| N/A            | N/A  | N/A     | No specific parameters are passed in configParams. |

### Elastic Plugin

| Parameter Name | Type | Example | Description |
|----------------|------|---------|-------------|
| N/A            | N/A  | N/A     | No specific parameters are passed in configParams. |
