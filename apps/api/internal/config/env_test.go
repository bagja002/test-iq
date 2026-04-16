package config

import "testing"

func TestParseEnvLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
		wantErr   bool
	}{
		{
			name:      "plain assignment",
			line:      "PORT=8080",
			wantKey:   "PORT",
			wantValue: "8080",
			wantOK:    true,
		},
		{
			name:      "mysql dsn with ampersand",
			line:      "MYSQL_DSN=root:@tcp(127.0.0.1:3306)/iq_test?charset=utf8mb4&parseTime=True&loc=Local",
			wantKey:   "MYSQL_DSN",
			wantValue: "root:@tcp(127.0.0.1:3306)/iq_test?charset=utf8mb4&parseTime=True&loc=Local",
			wantOK:    true,
		},
		{
			name:      "quoted value",
			line:      "JWT_SECRET=\"change-this-secret\"",
			wantKey:   "JWT_SECRET",
			wantValue: "change-this-secret",
			wantOK:    true,
		},
		{
			name:   "comment",
			line:   "# comment",
			wantOK: false,
		},
		{
			name:    "invalid assignment",
			line:    "BROKEN_LINE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue, gotOK, err := parseEnvLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotOK != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, gotOK)
			}
			if gotKey != tt.wantKey {
				t.Fatalf("expected key=%q, got %q", tt.wantKey, gotKey)
			}
			if gotValue != tt.wantValue {
				t.Fatalf("expected value=%q, got %q", tt.wantValue, gotValue)
			}
		})
	}
}
