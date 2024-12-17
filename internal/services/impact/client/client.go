package client

import (
	"fmt"

	"github.com/hashicorp/go-azure-sdk/resource-manager/impact/2024-05-01-preview/connectors"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
)

type Client struct {
	ConnectorsClient *connectors.ConnectorsClient
}

func NewClient(o *common.ClientOptions) (*Client, error) {
	connectorsClient, err := connectors.NewConnectorsClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("building ConnectorsClient: %+v", err)
	}
	o.Configure(connectorsClient.Client, o.Authorizers.ResourceManager)

	return &Client{
		ConnectorsClient: connectorsClient,
	}, nil
}
