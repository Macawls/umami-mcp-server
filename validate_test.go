package main

import "testing"

func TestValidateWebsiteID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid compact UUID", "550e8400e29b41d4a716446655440000", false},
		{"valid short hex", "abc123", false},
		{"empty", "", true},
		{"path traversal", "../../admin/users", true},
		{"slash", "abc/def", true},
		{"too long", "550e8400-e29b-41d4-a716-446655440000x", true},
		{"special chars", "abc<script>", true},
		{"spaces", "abc def", true},
		{"dots", "abc..def", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebsiteID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebsiteID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}
