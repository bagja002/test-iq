package router

import (
	"context"
	"log"
	"time"

	redisstorage "github.com/gofiber/storage/redis/v3"
	redis "github.com/redis/go-redis/v9"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/config"
	"test-iq-ku/apps/api/internal/handlers"
	"test-iq-ku/apps/api/internal/middleware"
	"test-iq-ku/apps/api/internal/services"
)

func New(cfg *config.Config, db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:   "Test IQ Ku API",
		BodyLimit: 2 * 1024 * 1024,
	})

	app.Use(recovermw.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))
	app.Use(middleware.Authenticate(db, cfg))

	authService := services.NewAuthService(db, cfg)
	testService := services.NewTestService(db)
	adminService := services.NewAdminService(db)
	paymentService := services.NewPaymentService(db, cfg)
	handler := handlers.New(cfg, authService, testService, adminService, paymentService)

	api := app.Group("/api/v1")
	api.Get("/health", handler.Health)

	loginLimiter := limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: time.Minute,
		Storage:    buildLimiterStorage(cfg),
	})
	submitLimiter := limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax * 2,
		Expiration: time.Minute,
		Storage:    buildLimiterStorage(cfg),
	})

	authGroup := api.Group("/auth")
	authGroup.Post("/register", loginLimiter, handler.Register)
	authGroup.Post("/login", loginLimiter, handler.Login)
	authGroup.Post("/logout", handler.Logout)
	authGroup.Post("/refresh", loginLimiter, handler.Refresh)
	authGroup.Get("/session", handler.Session)

	api.Post("/payments/midtrans/notifications", handler.MidtransNotification)

	api.Get("/test-config/active", middleware.RequireAuth(), handler.GetTestConfig)

	paymentGroup := api.Group("/payments", middleware.RequireAuth())
	paymentGroup.Post("/pro-upgrade", submitLimiter, handler.CreateProUpgradePayment)
	paymentGroup.Post("/pro-upgrade/confirm", submitLimiter, handler.ConfirmProUpgradePayment)

	testGroup := api.Group("/test-attempts", middleware.RequireAuth())
	testGroup.Post("/start", submitLimiter, handler.StartAttempt)
	testGroup.Get("/current", handler.GetCurrentAttempt)
	testGroup.Get("/:attemptId", handler.GetAttemptDetail)
	testGroup.Put("/:attemptId/answers", handler.SaveAnswers)
	testGroup.Post("/:attemptId/sections/:sectionCode/start", submitLimiter, handler.StartAttemptSection)
	testGroup.Post("/:attemptId/sections/:sectionCode/submit", submitLimiter, handler.SubmitAttemptSection)
	testGroup.Post("/:attemptId/submit", submitLimiter, handler.SubmitAttempt)

	resultGroup := api.Group("/results", middleware.RequireAuth())
	resultGroup.Get("/me/latest", handler.LatestResult)
	resultGroup.Get("/me", handler.ListMyResults)

	adminGroup := api.Group("/admin", middleware.RequireAdmin())
	adminGroup.Get("/overview", handler.AdminOverview)
	adminGroup.Get("/questions", handler.ListQuestions)
	adminGroup.Post("/questions", handler.CreateQuestion)
	adminGroup.Patch("/questions/:questionId", handler.UpdateQuestion)
	adminGroup.Delete("/questions/:questionId", handler.DeleteQuestion)
	adminGroup.Get("/users", handler.ListUsers)
	adminGroup.Post("/users", handler.CreateUser)
	adminGroup.Patch("/users/:userId", handler.UpdateUser)
	adminGroup.Get("/results", handler.ListResults)
	adminGroup.Get("/test-config", handler.GetTestConfig)
	adminGroup.Patch("/test-config", handler.UpdateTestConfig)

	return app
}

func buildLimiterStorage(cfg *config.Config) (storage fiber.Storage) {
	if cfg.RedisURL == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	clientOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("redis limiter storage disabled, invalid redis url: %v", err)
		return nil
	}

	client := redis.NewClient(clientOptions)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("redis limiter storage disabled, falling back to in-memory storage: %v", err)
		return nil
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("redis limiter storage disabled, falling back to in-memory storage: %v", recovered)
			storage = nil
		}
	}()

	return redisstorage.New(redisstorage.Config{
		URL: cfg.RedisURL,
	})
}
