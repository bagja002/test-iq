package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/auth"
	"test-iq-ku/apps/api/internal/config"
	"test-iq-ku/apps/api/internal/models"
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

func (s *AuthService) Login(email string, password string, userAgent string) (*models.User, string, string, error) {
	var user models.User
	if err := s.db.Where("email = ? AND status = ?", normalizeEmail(email), models.UserStatusActive).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errors.New("email atau password salah")
		}
		return nil, "", "", err
	}
	if !auth.VerifyPassword(user.PasswordHash, strings.TrimSpace(password)) {
		return nil, "", "", errors.New("email atau password salah")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, "", "", tx.Error
	}

	if err := s.registerUserAgent(tx, user, userAgent); err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	accessToken, refreshToken, err := s.issueSessionTx(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

func (s *AuthService) Register(name string, position string, email string, password string) (*models.User, string, string, error) {
	normalizedName, normalizedPosition, normalizedEmail, err := validatePublicRegistration(name, position, email, password)
	if err != nil {
		return nil, "", "", err
	}

	passwordHash, err := auth.HashPassword(strings.TrimSpace(password))
	if err != nil {
		return nil, "", "", err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, "", "", tx.Error
	}

	user := models.User{
		Name:         normalizedName,
		Position:     normalizedPosition,
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		Status:       models.UserStatusActive,
		AccountType:  models.AccountTypeFree,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, "", "", mapCreateUserError(err)
	}

	accessToken, refreshToken, err := s.issueSessionTx(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(rawRefreshToken string) (*models.User, string, string, error) {
	lookupHash := auth.HashToken(strings.TrimSpace(rawRefreshToken))
	now := time.Now()

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, "", "", tx.Error
	}

	var stored models.RefreshToken
	if err := tx.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", lookupHash, now).First(&stored).Error; err != nil {
		tx.Rollback()
		return nil, "", "", errors.New("refresh token tidak valid")
	}

	var user models.User
	if err := tx.Where("id = ? AND status = ?", stored.UserID, models.UserStatusActive).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	if err := tx.Model(&stored).Update("revoked_at", now).Error; err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	newRefreshToken, newRefreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	if err := tx.Create(&models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newRefreshHash,
		ExpiresAt: now.Add(s.cfg.RefreshTokenTTL),
	}).Error; err != nil {
		tx.Rollback()
		return nil, "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, "", "", err
	}

	accessToken, err := auth.IssueAccessToken(user, s.cfg)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil
	}

	now := time.Now()
	return s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", auth.HashToken(rawRefreshToken)).
		Update("revoked_at", now).
		Error
}

func (s *AuthService) issueSession(user models.User) (string, string, error) {
	return s.issueSessionTx(s.db, user)
}

func (s *AuthService) issueSessionTx(db *gorm.DB, user models.User) (string, string, error) {
	accessToken, err := auth.IssueAccessToken(user, s.cfg)
	if err != nil {
		return "", "", err
	}

	refreshToken, refreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := db.Create(&models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}).Error; err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) registerUserAgent(tx *gorm.DB, user models.User, rawUserAgent string) error {
	if !requiresDeviceRegistration(user.AccountType) {
		return nil
	}

	normalizedUserAgent := normalizeLoginUserAgent(rawUserAgent)
	userAgentHash := hashLoginUserAgent(normalizedUserAgent)
	now := time.Now()

	var device models.UserDevice
	if err := tx.Where("user_id = ? AND user_agent_hash = ?", user.ID, userAgentHash).First(&device).Error; err == nil {
		return tx.Model(&device).Updates(map[string]any{
			"user_agent":   normalizedUserAgent,
			"last_seen_at": now,
		}).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var registeredCount int64
	if err := tx.Model(&models.UserDevice{}).Where("user_id = ?", user.ID).Count(&registeredCount).Error; err != nil {
		return err
	}
	if registeredCount >= maxRegisteredDevices {
		return errors.New("akun ini sudah mencapai batas maksimal 3 browser/perangkat")
	}

	return tx.Create(&models.UserDevice{
		UserID:        user.ID,
		UserAgent:     normalizedUserAgent,
		UserAgentHash: userAgentHash,
		FirstSeenAt:   now,
		LastSeenAt:    now,
	}).Error
}

func normalizeLoginUserAgent(value string) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if normalized == "" {
		return "UNKNOWN"
	}
	if len(normalized) > 512 {
		return normalized[:512]
	}
	return normalized
}

func hashLoginUserAgent(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
