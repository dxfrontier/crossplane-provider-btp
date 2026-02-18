package btp

import (
	"encoding/json"
	"testing"
)

func TestUserCredential_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, uc UserCredential)
	}{
		{
			name:  "lowercase JSON fields are parsed correctly",
			input: `{"email":"user@example.com","username":"user@example.com","password":"secret","idp":"arrevqqkn.accounts.cloud.sap","origin":"arrevqqkn-platform"}`,
			validate: func(t *testing.T, uc UserCredential) {
				if uc.Email != "user@example.com" {
					t.Errorf("Email = %q, want %q", uc.Email, "user@example.com")
				}
				if uc.Username != "user@example.com" {
					t.Errorf("Username = %q, want %q", uc.Username, "user@example.com")
				}
				if uc.Password != "secret" {
					t.Errorf("Password = %q, want %q", uc.Password, "secret")
				}
				if uc.Idp != "arrevqqkn.accounts.cloud.sap" {
					t.Errorf("Idp = %q, want %q", uc.Idp, "arrevqqkn.accounts.cloud.sap")
				}
				if uc.Origin != "arrevqqkn-platform" {
					t.Errorf("Origin = %q, want %q", uc.Origin, "arrevqqkn-platform")
				}
			},
		},
		{
			name:  "origin field omitted leaves Origin empty",
			input: `{"email":"user@example.com","username":"user@example.com","password":"secret","idp":"some-idp"}`,
			validate: func(t *testing.T, uc UserCredential) {
				if uc.Idp != "some-idp" {
					t.Errorf("Idp = %q, want %q", uc.Idp, "some-idp")
				}
				if uc.Origin != "" {
					t.Errorf("Origin = %q, want empty", uc.Origin)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var uc UserCredential
			err := json.Unmarshal([]byte(tt.input), &uc)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, uc)
			}
		})
	}
}
