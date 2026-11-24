// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-azure-sdk/resource-manager/containerservice/2026-07-01/agentpools"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

func NodePoolAzureContainerLinuxCustomizeDiff(pathPrefix string) pluginsdk.CustomizeDiffFunc {
	return func(ctx context.Context, d *pluginsdk.ResourceDiff, meta any) error {
		osSKU := ""
		secureBootEnabled, vtpmEnabled := false, false

		if val, ok := d.GetOkExists(pathPrefix + "os_sku"); ok {
			osSKU = val.(string)
		} else {
			return nil
		}

		if osSKU != string(agentpools.OSSKUAzureContainerLinux) {
			return nil
		}

		if val, ok := d.GetOkExists(pathPrefix + "security.0.secure_boot_enabled"); ok {
			secureBootEnabled = val.(bool)
		} else {
			return nil
		}

		if val, ok := d.GetOkExists(pathPrefix + "security.0.vtpm_enabled"); ok {
			vtpmEnabled = val.(bool)
		} else {
			return nil
		}

		if !secureBootEnabled {
			return fmt.Errorf("`%[1]ssecurity.0.secure_boot_enabled` must be set to `true` when `%[1]s` is `AzureContainerLinux`", pathPrefix)
		}
		if !vtpmEnabled {
			return fmt.Errorf("`%[1]ssecurity.0.vtpm_enabled` must be set to `true` when `%[1]s` is `AzureContainerLinux`", pathPrefix)
		}

		return nil
	}
}
