// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datafactory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/datafactory/2018-06-01/factories"
	"github.com/hashicorp/go-azure-sdk/resource-manager/datafactory/2018-06-01/linkedservices"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/datafactory/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/datafactory/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
	"github.com/jackofallops/kermit/sdk/datafactory/2018-06-01/datafactory"
)

// Typed resource definition
var _ sdk.Resource = DataFactoryLinkedCustomServiceResource{}

type DataFactoryLinkedCustomServiceResource struct{}

type DataFactoryLinkedCustomServiceResourceModel struct {
	Name                 string               `tfschema:"name"`
	DataFactoryId        string               `tfschema:"data_factory_id"`
	Type                 string               `tfschema:"type"`
	TypePropertiesJson   string               `tfschema:"type_properties_json"`
	Description          string               `tfschema:"description"`
	IntegrationRuntime   []IntegrationRuntime `tfschema:"integration_runtime"`
	Parameters           map[string]string    `tfschema:"parameters"`
	Annotations          []string             `tfschema:"annotations"`
	AdditionalProperties map[string]string    `tfschema:"additional_properties"`
}

type IntegrationRuntime struct {
	Name       string            `tfschema:"name"`
	Parameters map[string]string `tfschema:"parameters"`
}

func (DataFactoryLinkedCustomServiceResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.LinkedServiceDatasetName,
		},
		"data_factory_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: factories.ValidateFactoryID,
		},
		"type": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},
		"type_properties_json": {
			Type:             pluginsdk.TypeString,
			Required:         true,
			StateFunc:        utils.NormalizeJson,
			DiffSuppressFunc: suppressJsonOrderingDifference,
		},
		"description": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"integration_runtime": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},
					"parameters": {
						Type:     pluginsdk.TypeMap,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
				},
			},
		},
		"parameters": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
		"annotations": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
		"additional_properties": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
	}
}

func (DataFactoryLinkedCustomServiceResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (DataFactoryLinkedCustomServiceResource) ModelObject() interface{} {
	return &DataFactoryLinkedCustomServiceResourceModel{}
}

func (DataFactoryLinkedCustomServiceResource) ResourceType() string {
	return "azurerm_data_factory_linked_custom_service"
}

func (r DataFactoryLinkedCustomServiceResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.DataFactory.LinkedServiceGoAzureSdk
			subscriptionId := metadata.Client.Account.SubscriptionId

			var model DataFactoryLinkedCustomServiceResourceModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			dataFactoryId, err := factories.ParseFactoryID(model.DataFactoryId)
			if err != nil {
				return err
			}

			id := linkedservices.NewLinkedServiceID(subscriptionId, dataFactoryId.ResourceGroupName, dataFactoryId.FactoryName, model.Name)

			existing, err := client.Get(ctx, id, linkedservices.DefaultGetOperationOptions())
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			// TODO need to fix the unmarshalling logic
			linkedServiceRes := linkedservices.LinkedServiceResource{}

			linkedServiceRes.Properties = linkedservices.U

			props := map[string]interface{}{
				"type":       model.Type,
				"connectVia": expandDataFactoryLinkedServiceIntegrationRuntimeReference(model.IntegrationRuntime),
			}

			jsonDataStr := fmt.Sprintf(`{ "typeProperties": %s }`, model.TypePropertiesJson)
			if err = json.Unmarshal([]byte(jsonDataStr), &props); err != nil {
				return err
			}

			if model.Description != "" {
				props["description"] = model.Description
			}

			if len(model.Parameters) > 0 {
				props["parameters"] = expandLinkedServiceParametersStringMap(model.Parameters)
			}

			if len(model.Annotations) > 0 {
				annotations := make([]interface{}, len(model.Annotations))
				for i, v := range model.Annotations {
					annotations[i] = v
				}
				props["annotations"] = annotations
			}

			for k, v := range model.AdditionalProperties {
				props[k] = v
			}

			jsonData, err := json.Marshal(map[string]interface{}{
				"properties": props,
			})
			if err != nil {
				return err
			}

			
			if err := linkedService.UnmarshalJSON(jsonData); err != nil {
				return err
			}

			if _, err := client.CreateOrUpdate(ctx, id, linkedService, linkedservices.DefaultCreateOrUpdateOperationOptions()); err != nil {
				return fmt.Errorf("creating/updating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r DataFactoryLinkedCustomServiceResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.DataFactory.LinkedServiceGoAzureSdk
			
			id, err = linkedservices.ParseLinkedServiceID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var config DataFactoryLinkedCustomServiceResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			existing, err := client.Get(ctx, id, linkedservices.DefaultGetOperationOptions())
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			if existing.Model == nil {
				return fmt.Errorf("retrieving %s: model was nil", id)
			}
			if existing.Model.Properties == nil {
				return fmt.Errorf("retrieving %s: properties was nil", id)
			}

			if metadata.ResourceData.HasChange("type_properties_json") {
				jsonDataStr := fmt.Sprintf(`{ "typeProperties": %s }`, config.TypePropertiesJson)
				if err = json.Unmarshal([]byte(jsonDataStr), &existing.Model.Properties); err != nil {
					return err
				}
			}
		}
	}
}

func (r DataFactoryLinkedCustomServiceResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.DataFactory.LinkedServiceClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			id, err := parse.LinkedServiceID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.Get(ctx, id.ResourceGroup, id.FactoryName, id.Name, "")
			if err != nil {
				if utils.ResponseWasNotFound(resp.Response) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			state := DataFactoryLinkedCustomServiceResourceModel{
				Name:          id.Name,
				DataFactoryId: factories.NewFactoryID(subscriptionId, id.ResourceGroup, id.FactoryName).ID(),
			}

			byteArr, err := json.Marshal(resp.Properties)
			if err != nil {
				return err
			}

			var m map[string]*json.RawMessage
			if err = json.Unmarshal(byteArr, &m); err != nil {
				return err
			}

			if v, ok := m["description"]; ok && v != nil {
				_ = json.Unmarshal(*v, &state.Description)
				delete(m, "description")
			}
			if v, ok := m["type"]; ok && v != nil {
				_ = json.Unmarshal(*v, &state.Type)
				delete(m, "type")
			}
			if v, ok := m["annotations"]; ok && v != nil {
				var annotations []string
				_ = json.Unmarshal(*v, &annotations)
				state.Annotations = annotations
				delete(m, "annotations")
			}
			if v, ok := m["parameters"]; ok && v != nil {
				var parameters map[string]*datafactory.ParameterSpecification
				_ = json.Unmarshal(*v, &parameters)
				state.Parameters = flattenLinkedServiceParametersStringMap(parameters)
				delete(m, "parameters")
			}
			if v, ok := m["connectVia"]; ok && v != nil {
				var integrationRuntime *datafactory.IntegrationRuntimeReference
				_ = json.Unmarshal(*v, &integrationRuntime)
				state.IntegrationRuntime = flattenDataFactoryLinkedServiceIntegrationRuntimeV2Typed(integrationRuntime)
				delete(m, "connectVia")
			}
			delete(m, "typeProperties")

			// set "additional_properties"
			additionalProperties := make(map[string]string)
			bytes, err := json.Marshal(m)
			if err != nil {
				return err
			}
			_ = json.Unmarshal(bytes, &additionalProperties)
			state.AdditionalProperties = additionalProperties

			// type_properties_json
			if v, ok := m["typeProperties"]; ok && v != nil {
				var raw json.RawMessage
				_ = json.Unmarshal(*v, &raw)
				state.TypePropertiesJson = string(raw)
			}

			return metadata.Encode(&state)
		},
	}
}

func (r DataFactoryLinkedCustomServiceResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.DataFactory.LinkedServiceClient

			id, err := parse.LinkedServiceID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, id.ResourceGroup, id.FactoryName, id.Name); err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}
			return nil
		},
	}
}

func (DataFactoryLinkedCustomServiceResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return parse.ValidateLinkedServiceID
}

// Helper expand/flatten functions for typed model

func expandDataFactoryLinkedServiceIntegrationRuntimeReference(input []IntegrationRuntime) *linkedservices.IntegrationRuntimeReference {
	if len(input) == 0 {
		return nil
	}
	v := input[0]
	return &linkedservices.IntegrationRuntimeReference{
		ReferenceName: v.Name,
		Type:          linkedservices.IntegrationRuntimeReferenceType("IntegrationRuntimeReference"),
		Parameters:    stringMapToInterfaceMap(v.Parameters),
	}
}

func flattenDataFactoryLinkedServiceIntegrationRuntimeV2Typed(input *datafactory.IntegrationRuntimeReference) []IntegrationRuntime {
	if input == nil {
		return nil
	}
	name := ""
	if input.ReferenceName != nil {
		name = *input.ReferenceName
	}
	return []IntegrationRuntime{
		{
			Name:       name,
			Parameters: interfaceMapToStringMap(input.Parameters),
		},
	}
}

func expandLinkedServiceParametersStringMap(input map[string]string) *map[string]linkedservices.ParameterSpecification {
	out := make(map[string]linkedservices.ParameterSpecification, len(input))
	for k, v := range input {
		out[k] = linkedservices.ParameterSpecification{
			DefaultValue: &v,
			Type:         linkedservices.ParameterTypeString,
		}
	}
	return &out
}

func flattenLinkedServiceParametersStringMap(input map[string]*datafactory.ParameterSpecification) map[string]string {
	// ...implement as needed, similar to flattenLinkedServiceParameters...
	return nil // placeholder
}

func stringMapToInterfaceMap(m map[string]string) *map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return &out
}

func interfaceMapToStringMap(m map[string]interface{}) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}
