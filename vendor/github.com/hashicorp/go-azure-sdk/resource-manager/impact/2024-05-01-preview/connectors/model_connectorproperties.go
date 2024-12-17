package connectors

import (
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/dates"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type ConnectorProperties struct {
	ConnectorId       string             `json:"connectorId"`
	ConnectorType     Platform           `json:"connectorType"`
	LastRunTimeStamp  string             `json:"lastRunTimeStamp"`
	ProvisioningState *ProvisioningState `json:"provisioningState,omitempty"`
	TenantId          string             `json:"tenantId"`
}

func (o *ConnectorProperties) GetLastRunTimeStampAsTime() (*time.Time, error) {
	return dates.ParseAsFormat(&o.LastRunTimeStamp, "2006-01-02T15:04:05Z07:00")
}

func (o *ConnectorProperties) SetLastRunTimeStampAsTime(input time.Time) {
	formatted := input.Format("2006-01-02T15:04:05Z07:00")
	o.LastRunTimeStamp = formatted
}
