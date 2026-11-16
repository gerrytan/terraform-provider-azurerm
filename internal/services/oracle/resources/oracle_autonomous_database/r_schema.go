// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package oracle_autonomous_database

import (
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-sdk/resource-manager/oracledatabase/2025-09-01/autonomousdatabases"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/oracle/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

func (Resource) Arguments() map[string]*pluginsdk.Schema {
	args := map[string]*pluginsdk.Schema{
		"location": commonschema.Location(),

		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: validate.AutonomousDatabaseName,
			ForceNew:     true,
		},

		"resource_group_name": commonschema.ResourceGroupName(),

		// Required
		"admin_password": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			Sensitive:    true,
			ValidateFunc: validate.AutonomousDatabasePassword,
		},

		"backup_retention_period_in_days": {
			Type:         pluginsdk.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 60),
		},

		"character_set": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"compute_count": {
			Type:         pluginsdk.TypeFloat,
			Required:     true,
			ValidateFunc: validation.FloatBetween(2.0, 512.0),
		},

		"compute_model": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.AdbsComputeModel,
		},

		"data_storage_size_in_gb": {
			Type:         pluginsdk.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(20, 393216),
		},

		"db_version": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"db_workload": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice(autonomousdatabases.PossibleValuesForWorkloadType(), false),
		},

		"display_name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.AutonomousDatabaseName,
		},

		"auto_scaling_enabled": {
			Type:     pluginsdk.TypeBool,
			Required: true,
		},

		"auto_scaling_for_storage_enabled": {
			Type:     pluginsdk.TypeBool,
			Required: true,
		},

		"license_model": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice(autonomousdatabases.PossibleValuesForLicenseModel(), false),
		},

		"long_term_backup_schedule": {
			Type:     pluginsdk.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"repeat_cadence": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice(autonomousdatabases.PossibleValuesForRepeatCadenceType(), false),
					},
					"time_of_backup": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.IsRFC3339Time,
					},
					"retention_period_in_days": {
						Type:         pluginsdk.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(90, 2558),
					},
					"enabled": {
						Type:     pluginsdk.TypeBool,
						Required: true,
					},
				},
			},
		},

		"national_character_set": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		// Optional
		"maintenance_patch_level": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Default:      autonomousdatabases.AutonomousMaintenanceScheduleTypeRegular,
			ValidateFunc: validation.StringInSlice(autonomousdatabases.PossibleValuesForAutonomousMaintenanceScheduleType(), false),
		},

		"customer_contacts": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true, // azignore:AZS007 - pre-existing violation
			ForceNew: true,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.IsEmailAddress,
			},
		},

		"mtls_connection_required": {
			Type:     pluginsdk.TypeBool,
			Required: true,
			ForceNew: true,
		},

		"subnet_id": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: commonids.ValidateSubnetID,
		},

		"virtual_network_id": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: commonids.ValidateVirtualNetworkID,
		},

		"allowed_ips": {
			Type:     pluginsdk.TypeSet,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.IsIPv4Address,
			},
		},

		"tags": commonschema.Tags(),
	}

	if !features.SixPointOh() {
		args["data_storage_size_in_gb"] = &pluginsdk.Schema{
			Type:     pluginsdk.TypeInt,
			Optional: true,
			// NOTE: O+C _tbs is deprecated in favour of _gb per guide-breaking-changes.md, API accepts either as PUT payload
			// and values will be coalesced before storing into state
			Computed:     true,
			ExactlyOneOf: []string{"data_storage_size_in_gb", "data_storage_size_in_tbs"},
			ValidateFunc: validation.IntBetween(20, 393216),
		}
		args["data_storage_size_in_tbs"] = &pluginsdk.Schema{
			Type:     pluginsdk.TypeInt,
			Optional: true,
			// NOTE: O+C _tbs is deprecated in favour of _gb per guide-breaking-changes.md, API accepts either as PUT payload
			// and values will be coalesced before storing into state
			Computed:     true,
			ExactlyOneOf: []string{"data_storage_size_in_gb", "data_storage_size_in_tbs"},
			ValidateFunc: validation.IntBetween(1, 384),
			Deprecated:   "`data_storage_size_in_tbs` has been deprecated in favour of `data_storage_size_in_gb` and will be removed in v6.0 of the AzureRM Provider",
		}
	}

	return args
}

func (Resource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"ocid": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}
