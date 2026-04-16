/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"testing"
)

func TestValidateClusterName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		// Valid names
		{"valid lowercase", "my-cluster", false},
		{"valid short", "abc", false},
		{"valid with numbers", "cluster123", false},
		{"valid hyphenated", "my-prod-cluster", false},
		{"valid with underscores", "my_cluster", false},
		{"valid mixed", "my-prod_cluster", false},
		{"valid max length", "a12345678901234567890123456789012345678901234567890123456789012", false},

		// Invalid - too short
		{"too short", "ab", true},

		// Invalid - too long
		{"too long", "a123456789012345678901234567890123456789012345678901234567890123", true},

		// Invalid - path traversal
		{"path traversal dots", "abc../prod", true},
		{"path traversal double dots", "abc..prod", true},
		{"path traversal slash", "abc/prod", true},
		{"path traversal backslash", "abc\\prod", true},
		{"path traversal relative", "../etc/passwd", true},

		// Invalid - format violations
		{"uppercase", "MyCluster", true},
		{"starts with hyphen", "-cluster", true},
		{"ends with hyphen", "cluster-", true},
		{"starts with underscore", "_cluster", true},
		{"ends with underscore", "cluster_", true},
		{"special chars", "cluster@name", true},
		{"spaces", "my cluster", true},
		{"dots", "cluster.name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClusterName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("validateClusterName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}
