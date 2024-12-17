package impact_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-sdk/resource-manager/impact/2024-05-01-preview/connectors"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type ImpactConnectorsTestResource struct{}

/* Only one connector resource per connector_type is supported per subscription. If all tests are run
in parallel we will get a 409 already exists error. */

func TestAccImpactConnectors_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_impact_connectors", "test")
	r := ImpactConnectorsTestResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccImpactConnectors_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_impact_connectors", "test")
	r := ImpactConnectorsTestResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_impact_connectors"),
		},
	})
}

func (ImpactConnectorsTestResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := connectors.ParseConnectorID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Impact.ConnectorsClient.Get(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	return pointer.To(resp.Model != nil), nil
}

func (ImpactConnectorsTestResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
  
resource "azurerm_impact_connectors" "test" {
  name = "conn-%d"
  connector_type = "AzureMonitor"
}
`, data.RandomInteger)
}

func (r ImpactConnectorsTestResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_impact_connectors" "import" {
  name = azurerm_impact_connectors.test.name
  connector_type = azurerm_impact_connectors.test.connector_type
}
`, template)
}

func (ImpactConnectorsTestResource) update() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
  
resource "azurerm_impact_connectors" "test" {
  name = "something-else"
  connector_type = "AzureMonitor"
}
`)
}
