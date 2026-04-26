package services

import "test-iq-ku/apps/api/internal/models"

const (
	proDailySubmitLimit  = 10
	maxRegisteredDevices = 3
)

func canonicalAccountType(accountType models.AccountType) models.AccountType {
	switch accountType {
	case models.AccountTypePro:
		return models.AccountTypePro
	case models.AccountTypeMax, models.AccountTypePaid:
		return models.AccountTypeMax
	default:
		return models.AccountTypeFree
	}
}

func hasPaidAccess(accountType models.AccountType) bool {
	return canonicalAccountType(accountType) != models.AccountTypeFree
}

func requiresDeviceRegistration(accountType models.AccountType) bool {
	switch canonicalAccountType(accountType) {
	case models.AccountTypePro, models.AccountTypeMax:
		return true
	default:
		return false
	}
}

func dailySubmitLimit(accountType models.AccountType, testType models.TestType) int {
	if canonicalAccountType(accountType) != models.AccountTypePro {
		return 0
	}

	switch testType {
	case models.TestTypeIQ, models.TestTypeSKB:
		return proDailySubmitLimit
	default:
		return 0
	}
}
