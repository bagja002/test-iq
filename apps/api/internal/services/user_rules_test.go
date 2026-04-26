package services

import (
	"testing"

	"test-iq-ku/apps/api/internal/models"
)

func TestValidatePublicRegistration(t *testing.T) {
	name, position, email, err := validatePublicRegistration("  Siti   Aminah  ", " Pengelola   Koperasi ", " Siti@example.com ", "password123")
	if err != nil {
		t.Fatalf("expected registration payload to be valid, got error: %v", err)
	}

	if name != "Siti Aminah" {
		t.Fatalf("expected normalized name to be %q, got %q", "Siti Aminah", name)
	}

	if email != "siti@example.com" {
		t.Fatalf("expected normalized email to be %q, got %q", "siti@example.com", email)
	}

	if position != "Pengelola Koperasi" {
		t.Fatalf("expected normalized position to be %q, got %q", "Pengelola Koperasi", position)
	}
}

func TestValidatePublicRegistrationRejectsInvalidPayload(t *testing.T) {
	testCases := []struct {
		name        string
		inputName   string
		inputPos    string
		inputEmail  string
		inputPass   string
		expectError string
	}{
		{
			name:        "missing name",
			inputName:   "",
			inputPos:    "Staff Koperasi",
			inputEmail:  "user@example.com",
			inputPass:   "password123",
			expectError: "nama wajib diisi",
		},
		{
			name:        "missing position",
			inputName:   "Budi",
			inputPos:    "",
			inputEmail:  "user@example.com",
			inputPass:   "password123",
			expectError: "jabatan wajib diisi",
		},
		{
			name:        "invalid email",
			inputName:   "Budi",
			inputPos:    "Staff Koperasi",
			inputEmail:  "not-an-email",
			inputPass:   "password123",
			expectError: "format email tidak valid",
		},
		{
			name:        "weak password",
			inputName:   "Budi",
			inputPos:    "Staff Koperasi",
			inputEmail:  "budi@example.com",
			inputPass:   "1234567",
			expectError: "password minimal 8 karakter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := validatePublicRegistration(tc.inputName, tc.inputPos, tc.inputEmail, tc.inputPass)
			if err == nil || err.Error() != tc.expectError {
				t.Fatalf("expected error %q, got %v", tc.expectError, err)
			}
		})
	}
}

func TestNormalizeRole(t *testing.T) {
	role, err := normalizeRole("")
	if err != nil {
		t.Fatalf("expected empty role to default to USER, got error: %v", err)
	}
	if role != models.RoleUser {
		t.Fatalf("expected role USER, got %s", role)
	}

	role, err = normalizeRole(models.RoleAdmin)
	if err != nil {
		t.Fatalf("expected admin role to be accepted, got error: %v", err)
	}
	if role != models.RoleAdmin {
		t.Fatalf("expected role ADMIN, got %s", role)
	}
}

func TestNormalizeUserStatus(t *testing.T) {
	status, err := normalizeUserStatus("")
	if err != nil {
		t.Fatalf("expected empty status to default to ACTIVE, got error: %v", err)
	}
	if status != models.UserStatusActive {
		t.Fatalf("expected ACTIVE status, got %s", status)
	}

	if _, err := normalizeUserStatus("BLOCKED"); err == nil {
		t.Fatal("expected invalid status to return an error")
	}
}

func TestNormalizeAccountType(t *testing.T) {
	accountType, err := normalizeAccountType("")
	if err != nil {
		t.Fatalf("expected empty account type to default to FREE, got error: %v", err)
	}
	if accountType != models.AccountTypeFree {
		t.Fatalf("expected FREE account type, got %s", accountType)
	}

	accountType, err = normalizeAccountType(models.AccountTypePaid)
	if err != nil {
		t.Fatalf("expected paid account type to be accepted, got error: %v", err)
	}
	if accountType != models.AccountTypeMax {
		t.Fatalf("expected MAX account type, got %s", accountType)
	}

	if _, err := normalizeAccountType("TRIAL"); err == nil {
		t.Fatal("expected invalid account type to return an error")
	}
}
