package validate

import (
	"testing"
)

func TestConnectorID(t *testing.T) {
	testData := []struct {
		Value string
		Error bool
	}{
		{
			Value: "",
			Error: true,
		},
		{
			Value: "oo",
			Error: true,
		},
		{
			Value: "acctest123-east",
			Error: false,
		},
		{
			Value: "an-identifier-which-is-too-long",
			Error: true,
		},
		{
			Value: "inv4l!dch@r4ct3rs",
			Error: true,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Value)

		_, err := ConnectorID(v.Value, "unit test")
		if err != nil && !v.Error {
			t.Fatalf("Expected pass but got an error: %s", err)
		}
	}
}

func TestConnectorType(t *testing.T) {
	testData := []struct {
		Value string
		Error bool
	}{
		{
			Value: "",
			Error: true,
		},
		{
			Value: "AzureMonitor",
			Error: false,
		},
		{
			Value: "InvalidType",
			Error: true,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Value)

		_, err := ConnectorType(v.Value, "unit test")
		if err != nil && !v.Error {
			t.Fatalf("Expected pass but got an error: %s", err)
		}
	}
}
