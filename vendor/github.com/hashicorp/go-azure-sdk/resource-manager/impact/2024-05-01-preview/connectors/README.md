
## `github.com/hashicorp/go-azure-sdk/resource-manager/impact/2024-05-01-preview/connectors` Documentation

The `connectors` SDK allows for interaction with Azure Resource Manager `impact` (API Version `2024-05-01-preview`).

This readme covers example usages, but further information on [using this SDK can be found in the project root](https://github.com/hashicorp/go-azure-sdk/tree/main/docs).

### Import Path

```go
import "github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
import "github.com/hashicorp/go-azure-sdk/resource-manager/impact/2024-05-01-preview/connectors"
```


### Client Initialization

```go
client := connectors.NewConnectorsClientWithBaseURI("https://management.azure.com")
client.Client.Authorizer = authorizer
```


### Example Usage: `ConnectorsClient.CreateOrUpdate`

```go
ctx := context.TODO()
id := connectors.NewConnectorID("12345678-1234-9876-4563-123456789012", "connectorName")

payload := connectors.Connector{
	// ...
}


if err := client.CreateOrUpdateThenPoll(ctx, id, payload); err != nil {
	// handle the error
}
```


### Example Usage: `ConnectorsClient.Delete`

```go
ctx := context.TODO()
id := connectors.NewConnectorID("12345678-1234-9876-4563-123456789012", "connectorName")

read, err := client.Delete(ctx, id)
if err != nil {
	// handle the error
}
if model := read.Model; model != nil {
	// do something with the model/response object
}
```


### Example Usage: `ConnectorsClient.Get`

```go
ctx := context.TODO()
id := connectors.NewConnectorID("12345678-1234-9876-4563-123456789012", "connectorName")

read, err := client.Get(ctx, id)
if err != nil {
	// handle the error
}
if model := read.Model; model != nil {
	// do something with the model/response object
}
```


### Example Usage: `ConnectorsClient.ListBySubscription`

```go
ctx := context.TODO()
id := commonids.NewSubscriptionID("12345678-1234-9876-4563-123456789012")

// alternatively `client.ListBySubscription(ctx, id)` can be used to do batched pagination
items, err := client.ListBySubscriptionComplete(ctx, id)
if err != nil {
	// handle the error
}
for _, item := range items {
	// do something
}
```


### Example Usage: `ConnectorsClient.Update`

```go
ctx := context.TODO()
id := connectors.NewConnectorID("12345678-1234-9876-4563-123456789012", "connectorName")

payload := connectors.ConnectorUpdate{
	// ...
}


read, err := client.Update(ctx, id, payload)
if err != nil {
	// handle the error
}
if model := read.Model; model != nil {
	// do something with the model/response object
}
```
