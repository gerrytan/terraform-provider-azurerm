package validate

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/hashicorp/go-azure-sdk/resource-manager/impact/2024-05-01-preview/connectors"
)

func ConnectorID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9-]{3,24}$`).MatchString(v) {
		errors = append(errors, fmt.Errorf("%q must be between 3 and 24 alphanumeric characters", k))
	}

	return warnings, errors
}

func ConnectorType(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	possibleValues := connectors.PossibleValuesForPlatform()

	if !slices.Contains(possibleValues, v) {
		errors = append(errors, fmt.Errorf("%q invalid value, must be one of: %+v", k, possibleValues))
	}

	return warnings, errors
}
