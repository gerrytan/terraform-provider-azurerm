package impact

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/impact/2024-05-01-preview/connectors"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/impact/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

var _ sdk.Resource = ImpactConnectorsResource{}

type ImpactConnectorsResource struct{}

type ImpactConnectorsResourceModel struct {
	Name          string `tfschema:"name"`
	ConnectorType string `tfschema:"connector_type"`
}

func (ImpactConnectorsResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.ConnectorID,
		},
		"connector_type": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validate.ConnectorType,
		},
	}
}

func (ImpactConnectorsResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (ImpactConnectorsResource) ModelObject() interface{} {
	return &ImpactConnectorsResourceModel{}
}

func (ImpactConnectorsResource) ResourceType() string {
	return "azurerm_impact_connectors"
}

func (r ImpactConnectorsResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Impact.ConnectorsClient

			var config ImpactConnectorsResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			id := connectors.NewConnectorID(metadata.Client.Account.SubscriptionId, config.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			connector := connectors.Connector{
				Properties: &connectors.ConnectorProperties{
					ConnectorType: connectors.Platform(config.ConnectorType),
					// The TypeSpec of this model does not mark properties as optional, hence the generated
					// go SDK model struct does not have 'omitempty' tag.
					// To avoid API validation error on creation we intentionally set mock time value here
					// This won't affect the actual value stored in the system.
					// Refer to the TypeSpec: https://github.com/Azure/azure-rest-api-specs/blob/96c44f99e15d99f1ae43a793f1845f912f664964/specification/impact/Impact.Management/connectors.tsp#L26
					LastRunTimeStamp: time.Unix(0, 0).Format(time.RFC3339),
				},
			}

			if _, err := client.CreateOrUpdate(ctx, id, connector); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r ImpactConnectorsResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Impact.ConnectorsClient

			id, err := connectors.ParseConnectorID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config ImpactConnectorsResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			existing, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			if existing.Model == nil {
				return fmt.Errorf("retrieving %s: `model` was nil", id)
			}
			if existing.Model.Properties == nil {
				return fmt.Errorf("retrieving %s: `properties` was nil", id)
			}

			if metadata.ResourceData.HasChanges("connector_type") {
				existing.Model.Properties.ConnectorType = connectors.Platform(metadata.ResourceData.Get("connector_type").(string))
			}

			if _, err := client.CreateOrUpdate(ctx, *id, *existing.Model); err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}

			return nil
		},
	}
}

func (ImpactConnectorsResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Impact.ConnectorsClient

			id, err := connectors.ParseConnectorID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			state := ImpactConnectorsResourceModel{
				Name:          id.ConnectorName,
				ConnectorType: string(resp.Model.Properties.ConnectorType),
			}

			return metadata.Encode(&state)
		},
	}
}

func (ImpactConnectorsResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Impact.ConnectorsClient

			id, err := connectors.ParseConnectorID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			return nil
		},
	}
}

func (ImpactConnectorsResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return connectors.ValidateConnectorID
}
