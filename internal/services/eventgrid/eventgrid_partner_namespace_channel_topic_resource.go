// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/eventgrid/2025-02-15/channels"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

//go:generate go run ../../tools/generator-tests resourceidentity -resource-name eventgrid_partner_namespace_channel_topic -properties name -compare-values subscription_id:partner_namespace_id,resource_group_name:partner_namespace_id,partner_namespace_name:partner_namespace_id

var (
	_ sdk.ResourceWithUpdate        = EventGridPartnerNamespaceChannelTopicResource{}
	_ sdk.ResourceWithIdentity      = EventGridPartnerNamespaceChannelTopicResource{}
	_ sdk.ResourceWithCustomizeDiff = EventGridPartnerNamespaceChannelTopicResource{}
)

type EventGridPartnerNamespaceChannelTopicResource struct{}

type EventGridPartnerNamespaceChannelTopicResourceModel struct {
	Name                         string                       `tfschema:"name"`
	PartnerNamespaceId           string                       `tfschema:"partner_namespace_id"`
	PartnerTopic                 []ChannelPartnerTopicModel   `tfschema:"partner_topic"`
	ActivationMessage            string                       `tfschema:"activation_message"`
	EventTypes                   []ChannelTopicEventTypeModel `tfschema:"event_type"`
	ExpirationTimeIfNotActivated string                       `tfschema:"expiration_time_if_not_activated"`
	ReadinessState               string                       `tfschema:"readiness_state"`
}

type ChannelPartnerTopicModel struct {
	Name              string `tfschema:"name"`
	ResourceGroupName string `tfschema:"resource_group_name"`
	SubscriptionId    string `tfschema:"subscription_id"`
	Source            string `tfschema:"source"`
}

type ChannelTopicEventTypeModel struct {
	Name             string `tfschema:"name"`
	DataSchemaUrl    string `tfschema:"data_schema_url"`
	Description      string `tfschema:"description"`
	DisplayName      string `tfschema:"display_name"`
	DocumentationUrl string `tfschema:"documentation_url"`
}

func (r EventGridPartnerNamespaceChannelTopicResource) ResourceType() string {
	return "azurerm_eventgrid_partner_namespace_channel_topic"
}

func (r EventGridPartnerNamespaceChannelTopicResource) ModelObject() any {
	return &EventGridPartnerNamespaceChannelTopicResourceModel{}
}

func (r EventGridPartnerNamespaceChannelTopicResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return channels.ValidateChannelID
}

func (r EventGridPartnerNamespaceChannelTopicResource) Identity() resourceids.ResourceId {
	return &channels.ChannelId{}
}

func (r EventGridPartnerNamespaceChannelTopicResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile(`^[a-zA-Z0-9-]{3,50}$`),
				"must be between 3 and 50 characters and contain only letters, numbers and hyphens",
			),
		},
		"partner_namespace_id": commonschema.ResourceIDReferenceRequiredForceNew(&channels.PartnerNamespaceId{}),
		"activation_message": {
			Type: pluginsdk.TypeString,
			// Not required by the API, but when omitted Azure generates a default longer than 256 characters which cannot be round-tripped.
			// API bug: https://github.com/Azure/azure-rest-api-specs/issues/37280
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 256),
		},
		"partner_topic": {
			Type:     pluginsdk.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:     pluginsdk.TypeString,
						Required: true,
						ForceNew: true,
						ValidateFunc: validation.StringMatch(
							regexp.MustCompile(`^[a-zA-Z0-9-]{3,50}$`),
							"must be between 3 and 50 characters and contain only letters, numbers and hyphens",
						),
					},
					"resource_group_name": commonschema.ResourceGroupName(),
					"source": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},
					"subscription_id": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ForceNew:     true,
						ValidateFunc: validation.IsUUID,
					},
				},
			},
		},
		"event_type": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 128),
					},
					"data_schema_url": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.IsURLWithHTTPorHTTPS,
					},
					"description": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},
					"display_name": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},
					"documentation_url": {
						Type:         pluginsdk.TypeString,
						Optional:     true,
						ValidateFunc: validation.IsURLWithHTTPorHTTPS,
					},
				},
			},
		},
	}
}

func (r EventGridPartnerNamespaceChannelTopicResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"expiration_time_if_not_activated": {
			Type: pluginsdk.TypeString,
			// Property is mutable with 7 days from now maximum expiry, but after activation it's gone from the payload.
			// Default if not specified is 7 days from creation.
			Computed: true,
		},
		"readiness_state": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

func (r EventGridPartnerNamespaceChannelTopicResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var config EventGridPartnerNamespaceChannelTopicResourceModel
			if err := metadata.DecodeDiff(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}
			names := make(map[string]struct{})
			for _, event := range config.EventTypes {
				if event.Name == "" {
					continue
				}
				if _, exists := names[event.Name]; exists {
					return fmt.Errorf("`event_type` names must be unique: %q is specified more than once", event.Name)
				}
				names[event.Name] = struct{}{}
			}
			return nil
		},
	}
}

func (r EventGridPartnerNamespaceChannelTopicResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.EventGrid.Channels
			var config EventGridPartnerNamespaceChannelTopicResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}
			parent, err := channels.ParsePartnerNamespaceID(config.PartnerNamespaceId)
			if err != nil {
				return err
			}
			id := channels.NewChannelID(parent.SubscriptionId, parent.ResourceGroupName, parent.PartnerNamespaceName, config.Name)
			if !metadata.Client.Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
				existing, err := client.Get(ctx, id)
				if err != nil && !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
				}
				if !response.WasNotFound(existing.HttpResponse) {
					return metadata.ResourceRequiresImport(r.ResourceType(), id)
				}
			}
			topic := config.PartnerTopic[0]
			properties := channels.ChannelProperties{
				ChannelType:          pointer.To(channels.ChannelTypePartnerTopic),
				MessageForActivation: pointer.To(config.ActivationMessage),
				PartnerTopicInfo: &channels.PartnerTopicInfo{
					AzureSubscriptionId: pointer.To(topic.SubscriptionId),
					ResourceGroupName:   pointer.To(topic.ResourceGroupName),
					Name:                pointer.To(topic.Name),
					Source:              pointer.To(topic.Source),
					EventTypeInfo:       expandChannelTopicEventTypes(config.EventTypes),
				},
			}
			if _, err := client.CreateOrUpdate(ctx, id, channels.Channel{Properties: &properties}); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}
			metadata.SetID(id)
			return pluginsdk.SetResourceIdentityData(metadata.ResourceData, &id)
		},
	}
}

func (r EventGridPartnerNamespaceChannelTopicResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.EventGrid.Channels
			id, err := channels.ParseChannelID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}
			var config EventGridPartnerNamespaceChannelTopicResourceModel
			if err := metadata.Decode(&config); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}
			existing, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			if existing.Model == nil || existing.Model.Properties == nil {
				return fmt.Errorf("retrieving %s: `model` or `properties` was nil", *id)
			}
			props := existing.Model.Properties
			if props.ChannelType == nil || *props.ChannelType != channels.ChannelTypePartnerTopic || props.PartnerTopicInfo == nil {
				return fmt.Errorf("%s is not a Partner Topic channel", *id)
			}
			if metadata.ResourceData.HasChange("activation_message") {
				props.MessageForActivation = pointer.To(config.ActivationMessage)
			}
			if metadata.ResourceData.HasChange("event_type") {
				props.PartnerTopicInfo.EventTypeInfo = expandChannelTopicEventTypes(config.EventTypes)
			}
			props.ProvisioningState = nil
			props.ReadinessState = nil
			if _, err := client.CreateOrUpdate(ctx, *id, *existing.Model); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}
			return nil
		},
	}
}

func (r EventGridPartnerNamespaceChannelTopicResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := channels.ParseChannelID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}
			resp, err := metadata.Client.EventGrid.Channels.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}
			return r.flatten(metadata, id, resp.Model)
		},
	}
}

func (r EventGridPartnerNamespaceChannelTopicResource) flatten(metadata sdk.ResourceMetaData, id *channels.ChannelId, model *channels.Channel) error {
	state := EventGridPartnerNamespaceChannelTopicResourceModel{
		Name:               id.ChannelName,
		PartnerNamespaceId: channels.NewPartnerNamespaceID(id.SubscriptionId, id.ResourceGroupName, id.PartnerNamespaceName).ID(),
	}
	if model != nil {
		if props := model.Properties; props != nil {
			if props.ChannelType == nil || *props.ChannelType != channels.ChannelTypePartnerTopic {
				return fmt.Errorf("%s is not a Partner Topic channel", *id)
			}
			state.ActivationMessage = pointer.From(props.MessageForActivation)
			state.ExpirationTimeIfNotActivated = pointer.From(props.ExpirationTimeIfNotActivatedUtc)
			state.ReadinessState = pointer.FromEnum(props.ReadinessState)
			if topic := props.PartnerTopicInfo; topic != nil {
				state.PartnerTopic = []ChannelPartnerTopicModel{{
					Name:              pointer.From(topic.Name),
					ResourceGroupName: pointer.From(topic.ResourceGroupName),
					SubscriptionId:    pointer.From(topic.AzureSubscriptionId),
					Source:            pointer.From(topic.Source),
				}}
				state.EventTypes = flattenChannelTopicEventTypes(topic.EventTypeInfo)
			}
		}
	}
	if err := pluginsdk.SetResourceIdentityData(metadata.ResourceData, id); err != nil {
		return err
	}
	return metadata.Encode(&state)
}

func (r EventGridPartnerNamespaceChannelTopicResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := channels.ParseChannelID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}
			if err := metadata.Client.EventGrid.Channels.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", *id, err)
			}
			return nil
		},
	}
}

func expandChannelTopicEventTypes(input []ChannelTopicEventTypeModel) *channels.EventTypeInfo {
	events := make(map[string]channels.InlineEventProperties, len(input))
	for _, event := range input {
		events[event.Name] = channels.InlineEventProperties{
			DataSchemaURL:    pointer.To(event.DataSchemaUrl),
			Description:      pointer.To(event.Description),
			DisplayName:      pointer.To(event.DisplayName),
			DocumentationURL: pointer.To(event.DocumentationUrl),
		}
	}
	return &channels.EventTypeInfo{Kind: pointer.To(channels.EventDefinitionKindInline), InlineEventTypes: &events}
}

func flattenChannelTopicEventTypes(input *channels.EventTypeInfo) []ChannelTopicEventTypeModel {
	output := make([]ChannelTopicEventTypeModel, 0)
	if input == nil || input.InlineEventTypes == nil {
		return output
	}
	for name, event := range *input.InlineEventTypes {
		output = append(output, ChannelTopicEventTypeModel{
			Name:             name,
			DataSchemaUrl:    pointer.From(event.DataSchemaURL),
			Description:      pointer.From(event.Description),
			DisplayName:      pointer.From(event.DisplayName),
			DocumentationUrl: pointer.From(event.DocumentationURL),
		})
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Name < output[j].Name })
	return output
}
