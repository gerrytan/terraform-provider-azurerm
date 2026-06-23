// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package oracle_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/provider/framework"
)

func TestExascaleDatabaseVirtualMachineClusterResource_list_basic(t *testing.T) {
	skipIfExascaleGridImageOcidEnvVariableNotSpecified(t)

	r := ExascaleDatabaseVirtualMachineClusterResource{}
	listResourceAddress := "azurerm_oracle_exascale_database_virtual_machine_cluster.list"

	data := acceptance.BuildTestData(t, "azurerm_oracle_exascale_database_virtual_machine_cluster", "test")

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV5ProviderFactories: framework.ProtoV5ProviderFactoriesInit(context.Background(), "azurerm"),
		Steps: []resource.TestStep{
			{
				Config: r.basicList(data),
			},
			{
				Query:  true,
				Config: r.basicListQuery(),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast(listResourceAddress, 2),
				},
			},
			{
				Query:  true,
				Config: r.basicListQueryByResourceGroupName(data),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength(listResourceAddress, 2),
				},
			},
		},
	})
}

func (a ExascaleDatabaseVirtualMachineClusterResource) basicList(data acceptance.TestData) string {
	return fmt.Sprintf(`
  %s
resource "azurerm_oracle_exascale_database_virtual_machine_cluster" "test1" {
  location                           = "%[3]s"
  name                               = "OFakeVm1acctest%[2]d"
  zones                              = local.zones
  resource_group_name                = azurerm_resource_group.test.name
  exascale_database_storage_vault_id = azurerm_oracle_exascale_database_storage_vault.test.id
  display_name                       = "OFakeVm1acctest%[2]d"
  enabled_ecpu_count                 = 16
  grid_image_ocid                    = local.grid_image_ocid
  hostname                           = "host1"
  number_of_vms_in_cluster           = 2
  ssh_public_keys                    = [local.ssh_public_key]
  subnet_id                          = azurerm_subnet.virtual_network_subnet.id
  virtual_machine_file_system_storage {
    total_size_in_gb = 440
  }
  virtual_network_id = azurerm_virtual_network.virtual_network.id
}

resource "azurerm_oracle_exascale_database_virtual_machine_cluster" "test2" {
  location                           = "%[3]s"
  name                               = "OFakeVm2acctest%[2]d"
  zones                              = local.zones
  resource_group_name                = azurerm_resource_group.test.name
  exascale_database_storage_vault_id = azurerm_oracle_exascale_database_storage_vault.test.id
  display_name                       = "OFakeVm2acctest%[2]d"
  enabled_ecpu_count                 = 16
  grid_image_ocid                    = local.grid_image_ocid
  hostname                           = "host2"
  number_of_vms_in_cluster           = 2
  ssh_public_keys                    = [local.ssh_public_key]
  subnet_id                          = azurerm_subnet.virtual_network_subnet.id
  virtual_machine_file_system_storage {
    total_size_in_gb = 440
  }
  virtual_network_id = azurerm_virtual_network.virtual_network.id
}`, a.template(data), data.RandomInteger, data.Locations.Primary)
}

func (a ExascaleDatabaseVirtualMachineClusterResource) basicListQuery() string {
	return `
list "azurerm_oracle_exascale_database_virtual_machine_cluster" "list" {
  provider = azurerm
  config {}
}
`
}

func (a ExascaleDatabaseVirtualMachineClusterResource) basicListQueryByResourceGroupName(data acceptance.TestData) string {
	return fmt.Sprintf(`
list "azurerm_oracle_exascale_database_virtual_machine_cluster" "list" {
  provider = azurerm
  config {
    resource_group_name = "acctestRG-%[1]d"
  }
}
`, data.RandomInteger)
}
