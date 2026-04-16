package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/auth"
	"test-iq-ku/apps/api/internal/config"
	"test-iq-ku/apps/api/internal/models"
	"test-iq-ku/apps/api/internal/utils"
)

const currentUserKey = "currentUser"

func Authenticate(db *gorm.DB, cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		rawToken := strings.TrimSpace(c.Cookies(auth.AccessCookieName))
		if rawToken == "" {
			return c.Next()
		}

		claims, err := auth.ParseAccessToken(rawToken, cfg)
		if err != nil {
			return c.Next()
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			return c.Next()
		}

		var user models.User
		if err := db.Where("id = ? AND status = ?", userID, models.UserStatusActive).First(&user).Error; err != nil {
			return c.Next()
		}

		c.Locals(currentUserKey, user)
		return c.Next()
	}
}

func CurrentUser(c fiber.Ctx) (models.User, bool) {
	user, ok := c.Locals(currentUserKey).(models.User)
	return user, ok
}

func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, ok := CurrentUser(c); !ok {
			return utils.RespondError(c, fiber.StatusUnauthorized, "authentication required", "")
		}

		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := CurrentUser(c)
		if !ok {
			return utils.RespondError(c, fiber.StatusUnauthorized, "authentication required", "")
		}

		if user.Role != models.RoleAdmin {
			return utils.RespondError(c, fiber.StatusForbidden, "admin access required", "")
		}

		return c.Next()
	}
}
