package service

import "testing"

func TestSanitizeCSVField(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"normal text", "hello world", "hello world"},
		{"numeric", "12345", "12345"},
		{"equals prefix", "=SUM(A1)", "'=SUM(A1)"},
		{"plus prefix", "+cmd|'/C calc'!A0", "'+cmd|'/C calc'!A0"},
		{"minus prefix", "-1+1", "'-1+1"},
		{"at prefix", "@SUM(A1)", "'@SUM(A1)"},
		{"tab prefix", "\tdata", "'\tdata"},
		{"cr prefix", "\rdata", "'\rdata"},
		{"newline prefix", "\ndata", "'\ndata"},
		{"safe special chars", "hello=world", "hello=world"},
		{"safe plus in middle", "hello+world", "hello+world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeCSVField(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeCSVField(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
