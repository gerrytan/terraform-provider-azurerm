// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
)

func TestCoalesceStorageSizeToGbs(t *testing.T) {
	tests := []struct {
		name      string
		sizeInGbs *int64
		sizeInTbs *int64
		expected  int64
	}{
		{
			name:      "sizeInGbs provided",
			sizeInGbs: pointer.To(int64(100)),
			sizeInTbs: nil,
			expected:  100,
		},
		{
			name:      "sizeInTbs provided",
			sizeInGbs: nil,
			sizeInTbs: pointer.To(int64(2)),
			expected:  2048,
		},
		{
			name:      "both provided (Gbs win, but this is not a happy path, it's an API bug if inconsistent sizes are returned)",
			sizeInGbs: pointer.To(int64(100)),
			sizeInTbs: pointer.To(int64(2)),
			expected:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coalesceStorageSizeToGbs(tt.sizeInGbs, tt.sizeInTbs)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
