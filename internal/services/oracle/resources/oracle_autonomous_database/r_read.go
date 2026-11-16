// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package oracle_autonomous_database

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/oracledatabase/2025-09-01/autonomousdatabases"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/oracle/helpers"
)

func (Resource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := autonomousdatabases.ParseAutonomousDatabaseID(metadata.ResourceData.Id())
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			client := metadata.Client.Oracle.OracleClient.AutonomousDatabases
			result, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(result.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			state := AutonomousDatabaseRegularResourceModel{
				Name:              id.AutonomousDatabaseName,
				ResourceGroupName: id.ResourceGroupName,
			}
			if model := result.Model; model != nil {
				props, ok := model.Properties.(autonomousdatabases.AutonomousDatabaseProperties)
				if !ok {
					return fmt.Errorf("%s was not of type `Regular`", id)
				}
				state.AdminPassword = metadata.ResourceData.Get("admin_password").(string)
				state.AutoScalingEnabled = pointer.From(props.IsAutoScalingEnabled)
				state.MaintenancePatchLevel = pointer.FromEnum(props.AutonomousMaintenanceScheduleType)
				state.BackupRetentionPeriodInDays = pointer.From(props.BackupRetentionPeriodInDays)
				state.AutoScalingForStorageEnabled = pointer.From(props.IsAutoScalingForStorageEnabled)
				state.CharacterSet = pointer.From(props.CharacterSet)
				state.ComputeCount = pointer.From(props.ComputeCount)
				state.ComputeModel = pointer.FromEnum(props.ComputeModel)
				state.CustomerContacts = helpers.FlattenAdbsCustomerContacts(props.CustomerContacts)
				state.DataStorageSizeInGb = coalesceStorageSizeToGbs(props.DataStorageSizeInGbs, props.DataStorageSizeInTbs)
				if !features.SixPointOh() {
					state.DataStorageSizeInTbs = pointer.From(props.DataStorageSizeInTbs)
					if props.DataStorageSizeInTbs == nil && props.DataStorageSizeInGbs != nil && *props.DataStorageSizeInGbs%1024 == 0 {
						state.DataStorageSizeInTbs = *props.DataStorageSizeInGbs / 1024
					}
				}
				state.DbWorkload = pointer.FromEnum(props.DbWorkload)
				state.DbVersion = pointer.From(props.DbVersion)
				state.DisplayName = pointer.From(props.DisplayName)
				state.LicenseModel = pointer.FromEnum(props.LicenseModel)
				state.Location = result.Model.Location
				state.MtlsConnectionRequired = pointer.From(props.IsMtlsConnectionRequired)
				state.Name = pointer.From(result.Model.Name)
				state.NationalCharacterSet = pointer.From(props.NcharacterSet)
				state.SubnetId = pointer.From(props.SubnetId)
				state.Tags = pointer.From(result.Model.Tags)
				state.VnetId = pointer.From(props.VnetId)
				state.LongTermBackUpSchedule = helpers.FlattenLongTermBackUpScheduleDetails(props.LongTermBackupSchedule)
				state.AllowedIps = pointer.From(props.WhitelistedIPs)
				state.Ocid = pointer.From(props.Ocid)
			}
			return metadata.Encode(&state)
		},
	}
}

func (r Resource) hasGeneralUpdates(metadata sdk.ResourceMetaData) bool {
	// lintignore:R019 // deliberate subset: only the fields covered by the general update payload; the remaining attributes are updated via separate calls
	keys := []string{
		"tags",
		"data_storage_size_in_gb",
		"compute_count",
		"auto_scaling_enabled",
		"auto_scaling_for_storage_enabled",
		"allowed_ips",
	}
	if !features.SixPointOh() {
		keys = append(keys, "data_storage_size_in_tbs")
	}
	return metadata.ResourceData.HasChanges(keys...)
}

// coalesceStorageSizeToGbs normalizes API responses which may contain only one storage unit.
func coalesceStorageSizeToGbs(sizeInGbs, sizeInTbs *int64) int64 {
	if sizeInGbs != nil {
		return *sizeInGbs
	}
	if sizeInTbs != nil {
		return *sizeInTbs * 1024
	}
	return 0
}
