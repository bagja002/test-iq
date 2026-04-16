package services

import (
	"errors"
	"net/mail"
	"strings"

	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/models"
)

func normalizeName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validatePublicRegistration(name string, email string, password string) (string, string, error) {
	normalizedName := normalizeName(name)
	normalizedEmail := normalizeEmail(email)

	if normalizedName == "" {
		return "", "", errors.New("nama wajib diisi")
	}

	if len(normalizedName) < 3 {
		return "", "", errors.New("nama minimal 3 karakter")
	}

	if normalizedEmail == "" {
		return "", "", errors.New("email wajib diisi")
	}

	parsedAddress, err := mail.ParseAddress(normalizedEmail)
	if err != nil || parsedAddress.Address != normalizedEmail {
		return "", "", errors.New("format email tidak valid")
	}

	if len(strings.TrimSpace(password)) < 8 {
		return "", "", errors.New("password minimal 8 karakter")
	}

	return normalizedName, normalizedEmail, nil
}

func normalizeRole(role models.Role) (models.Role, error) {
	switch role {
	case "", models.RoleUser:
		return models.RoleUser, nil
	case models.RoleAdmin:
		return models.RoleAdmin, nil
	default:
		return "", errors.New("role tidak valid")
	}
}

func normalizeUserStatus(status models.UserStatus) (models.UserStatus, error) {
	switch status {
	case "", models.UserStatusActive:
		return models.UserStatusActive, nil
	case models.UserStatusInactive:
		return models.UserStatusInactive, nil
	default:
		return "", errors.New("status user tidak valid")
	}
}

func mapCreateUserError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return errors.New("email sudah terdaftar")
	}

	return err
}
