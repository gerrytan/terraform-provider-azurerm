package impact

import "github.com/hashicorp/terraform-provider-azurerm/internal/sdk"

var _ sdk.TypedServiceRegistration = Registration{}

type Registration struct{}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}

func (r Registration) Name() string {
	return "impact"
}

func (r Registration) WebsiteCategories() []string {
	return []string{"Impact"}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		ImpactConnectorsResource{},
	}
}
