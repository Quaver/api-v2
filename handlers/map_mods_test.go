package handlers

import "testing"

func TestValidateEditableMapModComment(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		valid   bool
	}{
		{name: "valid", comment: "Updated explanation", valid: true},
		{name: "empty", comment: "", valid: false},
		{name: "too long", comment: string(make([]byte, 5001)), valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apiErr := validateEditableMapModComment(test.comment)

			if test.valid && apiErr != nil {
				t.Fatalf("expected comment to be valid, got error: %v", apiErr.Message)
			}

			if !test.valid && apiErr == nil {
				t.Fatal("expected comment validation to fail")
			}
		})
	}
}

func TestNormalizeEditableMapModTimestamp(t *testing.T) {
	validTimestamp := "12345|2,23456|4"

	tests := []struct {
		name      string
		timestamp *string
		want      *string
		valid     bool
	}{
		{name: "omitted", timestamp: nil, want: nil, valid: true},
		{name: "empty", timestamp: stringPointer(""), want: nil, valid: true},
		{name: "valid", timestamp: &validTimestamp, want: &validTimestamp, valid: true},
		{name: "invalid syntax", timestamp: stringPointer("12345|"), valid: false},
		{name: "too long", timestamp: stringPointer(string(make([]byte, 5001))), valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, apiErr := normalizeEditableMapModTimestamp(test.timestamp)

			if test.valid && apiErr != nil {
				t.Fatalf("expected timestamp to be valid, got error: %v", apiErr.Message)
			}

			if !test.valid && apiErr == nil {
				t.Fatal("expected timestamp validation to fail")
			}

			if test.want == nil && got != nil {
				t.Fatalf("expected timestamp to be cleared, got %q", *got)
			}

			if test.want != nil && (got == nil || *got != *test.want) {
				t.Fatalf("expected timestamp %q, got %v", *test.want, got)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
