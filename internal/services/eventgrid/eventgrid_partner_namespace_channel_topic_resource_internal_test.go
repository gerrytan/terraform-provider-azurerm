// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-sdk/resource-manager/eventgrid/2025-02-15/channels"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

func TestChannelTopicFieldValidation(t *testing.T) {
	args := EventGridPartnerNamespaceChannelTopicResource{}.Arguments()
	topic := args["partner_topic"].Elem.(*pluginsdk.Resource).Schema
	event := args["event_type"].Elem.(*pluginsdk.Resource).Schema
	cases := []struct {
		name      string
		validator pluginsdk.SchemaValidateFunc
		value     string
		valid     bool
	}{
		{"empty activation message", args["activation_message"].ValidateFunc, "", false},
		{"maximum activation message", args["activation_message"].ValidateFunc, strings.Repeat("a", 256), true},
		{"long activation message", args["activation_message"].ValidateFunc, strings.Repeat("a", 257), false},
		{"short channel", args["name"].ValidateFunc, "ab", false},
		{"minimum channel", args["name"].ValidateFunc, "abc", true},
		{"maximum channel", args["name"].ValidateFunc, strings.Repeat("a", 50), true},
		{"long channel", args["name"].ValidateFunc, strings.Repeat("a", 51), false},
		{"underscore channel", args["name"].ValidateFunc, "bad_name", false},
		{"period channel", args["name"].ValidateFunc, "bad.name", false},
		{"space channel", args["name"].ValidateFunc, "bad name", false},
		{"portal allows hyphens and uppercase", args["name"].ValidateFunc, "-Example-", true},
		{"invalid topic", topic["name"].ValidateFunc, "bad_name", false},
		{"valid topic", topic["name"].ValidateFunc, "Example-topic", true},
		{"invalid subscription", topic["subscription_id"].ValidateFunc, "not-a-guid", false},
		{"valid subscription", topic["subscription_id"].ValidateFunc, "00000000-0000-0000-0000-000000000000", true},
		{"invalid resource group", topic["resource_group_name"].ValidateFunc, "bad/group", false},
		{"empty source", topic["source"].ValidateFunc, "", false},
		{"free text source", topic["source"].ValidateFunc, "contains spaces !", true},
		{"empty event name", event["name"].ValidateFunc, "", false},
		{"maximum event name", event["name"].ValidateFunc, strings.Repeat("a", 128), true},
		{"long event name", event["name"].ValidateFunc, strings.Repeat("a", 129), false},
		{"free text event name", event["name"].ValidateFunc, "Order Created!", true},
		{"invalid documentation URL", event["documentation_url"].ValidateFunc, "not-a-url", false},
		{"valid documentation URL", event["documentation_url"].ValidateFunc, "https://example.com/events", true},
		{"invalid schema scheme", event["data_schema_url"].ValidateFunc, "ftp://example.com/schema", false},
		{"valid schema URL", event["data_schema_url"].ValidateFunc, "http://example.com/schema", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errors := tc.validator(tc.value, tc.name)
			if (len(errors) == 0) != tc.valid {
				t.Fatalf("expected valid=%t, got %v", tc.valid, errors)
			}
		})
	}
}

func TestChannelTopicPlanValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		events    []any
		wantError string
	}{
		{name: "defaults"},
		{name: "unique event names", events: []any{map[string]any{"name": "one"}, map[string]any{"name": "two"}}},
		{name: "unknown event name", events: []any{map[string]any{"name": "74D93920-ED26-11E3-AC10-0800200C9A66"}, map[string]any{"name": "two"}}},
		{name: "duplicate event names", events: []any{map[string]any{"name": "one", "display_name": "First"}, map[string]any{"name": "one", "display_name": "Second"}}, wantError: "names must be unique"},
		{name: "identical duplicate events", events: []any{map[string]any{"name": "one"}, map[string]any{"name": "one"}}, wantError: "names must be unique"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := sdk.WrappedResource(EventGridPartnerNamespaceChannelTopicResource{})
			config := channelTopicTestConfig()
			if tc.events != nil {
				config["event_type"] = tc.events
			}
			_, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(config), &clients.Client{})
			if tc.wantError == "" && err != nil {
				t.Fatalf("unexpected plan error: %v", err)
			}
			if tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
				t.Fatalf("expected error containing %q, got %v", tc.wantError, err)
			}
		})
	}
}

func TestChannelTopicReplacement(t *testing.T) {
	for _, field := range []string{"name", "partner_namespace_id", "topic_name", "resource_group_name", "subscription_id", "source", "activation_message", "event_type"} {
		t.Run(field, func(t *testing.T) {
			r := sdk.WrappedResource(EventGridPartnerNamespaceChannelTopicResource{})
			config := channelTopicTestConfig()
			initial, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(config), &clients.Client{})
			if err != nil {
				t.Fatal(err)
			}
			attrs, err := initial.Apply(nil, r.CoreConfigSchema())
			if err != nil {
				t.Fatal(err)
			}
			state := &terraform.InstanceState{ID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test/providers/Microsoft.EventGrid/partnerNamespaces/test/channels/test", Attributes: attrs}
			switch field {
			case "name":
				config[field] = "new-channel"
			case "partner_namespace_id":
				config[field] = strings.ReplaceAll(config[field].(string), "partnerNamespaces/test", "partnerNamespaces/other")
			case "activation_message":
				config[field] = "Updated message"
			case "event_type":
				config[field] = []any{map[string]any{"name": "Contoso.OrderCreated"}}
			default:
				key := field
				if key == "topic_name" {
					key = "name"
				}
				value := "updated"
				if key == "subscription_id" {
					value = "11111111-1111-1111-1111-111111111111"
				}
				config["partner_topic"].([]any)[0].(map[string]any)[key] = value
			}
			diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), &clients.Client{})
			if err != nil {
				t.Fatal(err)
			}
			wantNew := field != "activation_message" && field != "event_type"
			if diff.RequiresNew() != wantNew {
				t.Fatalf("expected replacement=%t, got %t", wantNew, diff.RequiresNew())
			}
		})
	}
}

func channelTopicTestConfig() map[string]any {
	return map[string]any{
		"activation_message":   "Please activate this topic.",
		"name":                 "test-channel",
		"partner_namespace_id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test/providers/Microsoft.EventGrid/partnerNamespaces/test",
		"partner_topic": []any{map[string]any{
			"name": "test-topic", "resource_group_name": "test", "subscription_id": "00000000-0000-0000-0000-000000000000", "source": "Contoso.Accounts.User1",
		}},
	}
}

func TestChannelTopicUnchangedExpiredTimestamp(t *testing.T) {
	r := sdk.WrappedResource(EventGridPartnerNamespaceChannelTopicResource{})
	config := channelTopicTestConfig()
	initial, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(config), &clients.Client{})
	if err != nil {
		t.Fatal(err)
	}
	attrs, err := initial.Apply(nil, r.CoreConfigSchema())
	if err != nil {
		t.Fatal(err)
	}
	const expired = "2020-01-01T00:00:00Z"
	attrs["expiration_time_if_not_activated"] = expired
	state := &terraform.InstanceState{ID: "test", Attributes: attrs}
	config["activation_message"] = "Update an activated channel"
	if _, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), &clients.Client{}); err != nil {
		t.Fatalf("unchanged expiration must not prevent other updates: %v", err)
	}
}

func TestChannelTopicFlatten(t *testing.T) {
	r := EventGridPartnerNamespaceChannelTopicResource{}
	id := channels.NewChannelID("00000000-0000-0000-0000-000000000000", "test", "namespace", "channel")
	metadata := sdk.NewResourceMetaData(&clients.Client{}, r)
	metadata.SetID(id)
	model := &channels.Channel{Properties: &channels.ChannelProperties{
		ExpirationTimeIfNotActivatedUtc: pointer.To("2026-10-25T12:00:00Z"),
		ChannelType:                     pointer.To(channels.ChannelTypePartnerTopic),
		PartnerTopicInfo: &channels.PartnerTopicInfo{
			AzureSubscriptionId: pointer.To("11111111-1111-1111-1111-111111111111"),
			ResourceGroupName:   pointer.To("subscriber"), Name: pointer.To("topic"), Source: pointer.To("source"),
			EventTypeInfo: expandChannelTopicEventTypes([]ChannelTopicEventTypeModel{{Name: "Contoso.OrderCreated", DisplayName: "Order created"}}),
		},
		MessageForActivation: pointer.To("Activate this topic"),
	}}
	if err := r.flatten(metadata, &id, model); err != nil {
		t.Fatal(err)
	}
	var state EventGridPartnerNamespaceChannelTopicResourceModel
	if err := metadata.Decode(&state); err != nil {
		t.Fatal(err)
	}
	if state.Name != "channel" || state.PartnerTopic[0].ResourceGroupName != "subscriber" || len(state.EventTypes) != 1 || state.EventTypes[0].DisplayName != "Order created" {
		t.Fatalf("unexpected state: %#v", state)
	}
	if state.ExpirationTimeIfNotActivated != "2026-10-25T12:00:00Z" {
		t.Fatalf("unexpected expiration: %q", state.ExpirationTimeIfNotActivated)
	}
	model.Properties.ExpirationTimeIfNotActivatedUtc = nil
	model.Properties.PartnerTopicInfo.EventTypeInfo = nil
	model.Properties.MessageForActivation = nil
	if err := r.flatten(metadata, &id, model); err != nil {
		t.Fatal(err)
	}
	if err := metadata.Decode(&state); err != nil {
		t.Fatal(err)
	}
	if len(state.EventTypes) != 0 || state.ActivationMessage != "" || state.ExpirationTimeIfNotActivated != "" {
		t.Fatalf("removed optional fields were not cleared: %#v", state)
	}
	model.Properties.ChannelType = pointer.To(channels.ChannelType("PartnerDestination"))
	if err := r.flatten(metadata, &id, model); err == nil {
		t.Fatal("expected a non-topic channel to be rejected")
	}
}

func TestChannelTopicEventTypesSortedInState(t *testing.T) {
	r := EventGridPartnerNamespaceChannelTopicResource{}
	id := channels.NewChannelID("00000000-0000-0000-0000-000000000000", "test", "namespace", "channel")
	metadata := sdk.NewResourceMetaData(&clients.Client{}, r)
	metadata.SetID(id)
	model := &channels.Channel{Properties: &channels.ChannelProperties{
		ChannelType: pointer.To(channels.ChannelTypePartnerTopic),
		PartnerTopicInfo: &channels.PartnerTopicInfo{
			EventTypeInfo: expandChannelTopicEventTypes([]ChannelTopicEventTypeModel{
				{Name: "Contoso.Updated"},
				{Name: "Contoso.Created"},
				{Name: "Contoso.Deleted"},
			}),
		},
	}}
	if err := r.flatten(metadata, &id, model); err != nil {
		t.Fatal(err)
	}
	events := metadata.ResourceData.Get("event_type").([]any)
	for i, expected := range []string{"Contoso.Created", "Contoso.Deleted", "Contoso.Updated"} {
		if actual := events[i].(map[string]any)["name"]; actual != expected {
			t.Fatalf("event_type[%d].name: expected %q, got %q", i, expected, actual)
		}
	}
}
