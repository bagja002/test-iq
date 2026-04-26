package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"test-iq-ku/apps/api/internal/auth"
	"test-iq-ku/apps/api/internal/config"
	"test-iq-ku/apps/api/internal/middleware"
	"test-iq-ku/apps/api/internal/models"
	"test-iq-ku/apps/api/internal/services"
	"test-iq-ku/apps/api/internal/utils"
)

type Handler struct {
	cfg   *config.Config
	auth  *services.AuthService
	test  *services.TestService
	admin *services.AdminService
	pay   *services.PaymentService
}

func New(cfg *config.Config, authService *services.AuthService, testService *services.TestService, adminService *services.AdminService, paymentService *services.PaymentService) *Handler {
	return &Handler{
		cfg:   cfg,
		auth:  authService,
		test:  testService,
		admin: adminService,
		pay:   paymentService,
	}
}

func (h *Handler) Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Login(c fiber.Ctx) error {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload login tidak valid", err.Error())
	}

	user, accessToken, refreshToken, err := h.auth.Login(payload.Email, payload.Password)
	if err != nil {
		return utils.RespondError(c, fiber.StatusUnauthorized, err.Error(), "")
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	return c.JSON(fiber.Map{
		"message": "login berhasil",
		"user":    serializeUser(*user),
	})
}

func (h *Handler) Register(c fiber.Ctx) error {
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload register tidak valid", err.Error())
	}

	user, accessToken, refreshToken, err := h.auth.Register(payload.Name, payload.Email, payload.Password)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "registrasi berhasil",
		"user":    serializeUser(*user),
	})
}

func (h *Handler) Logout(c fiber.Ctx) error {
	_ = h.auth.Logout(c.Cookies(auth.RefreshCookieName))
	h.clearAuthCookies(c)
	return utils.RespondMessage(c, fiber.StatusOK, "logout berhasil", nil)
}

func (h *Handler) Refresh(c fiber.Ctx) error {
	user, accessToken, refreshToken, err := h.auth.Refresh(c.Cookies(auth.RefreshCookieName))
	if err != nil {
		h.clearAuthCookies(c)
		return utils.RespondError(c, fiber.StatusUnauthorized, err.Error(), "")
	}

	h.setAuthCookies(c, accessToken, refreshToken)
	return c.JSON(fiber.Map{
		"message": "session diperbarui",
		"user":    serializeUser(*user),
	})
}

func (h *Handler) Session(c fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return c.JSON(fiber.Map{
			"authenticated": false,
			"user":          nil,
		})
	}

	return c.JSON(fiber.Map{
		"authenticated": true,
		"user":          serializeUser(user),
	})
}

func (h *Handler) CreateProUpgradePayment(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	result, err := h.pay.CreateProUpgradePayment(user)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message":     "snap token berhasil dibuat",
		"orderId":     result.OrderID,
		"snapToken":   result.SnapToken,
		"redirectUrl": result.RedirectURL,
		"amount":      result.Amount,
		"status":      result.Status,
	})
}

func (h *Handler) ConfirmProUpgradePayment(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	var payload services.ConfirmPaymentInput
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload konfirmasi pembayaran tidak valid", err.Error())
	}

	result, err := h.pay.ConfirmPayment(user, payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message":           "status pembayaran berhasil diperbarui",
		"orderId":           result.OrderID,
		"accountType":       result.AccountType,
		"paymentStatus":     result.PaymentStatus,
		"transactionStatus": result.TransactionStatus,
	})
}

func (h *Handler) MidtransNotification(c fiber.Ctx) error {
	result, err := h.pay.HandleMidtransNotification(c.Body())
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message":           "notifikasi pembayaran diproses",
		"orderId":           result.OrderID,
		"accountType":       result.AccountType,
		"paymentStatus":     result.PaymentStatus,
		"transactionStatus": result.TransactionStatus,
	})
}

func (h *Handler) StartAttempt(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	var payload services.StartAttemptInput
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload start attempt tidak valid", err.Error())
	}

	attempt, err := h.test.StartAttempt(user, payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message":   "attempt siap",
		"attemptId": attempt.ID,
		"attempt":   serializeAttemptSummary(*attempt),
	})
}

func (h *Handler) GetCurrentAttempt(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	testType, roomCode, err := parseTestSelection(c)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	attempt, err := h.test.GetCurrentAttempt(user.ID, testType, roomCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil current attempt", err.Error())
	}

	return c.JSON(fiber.Map{
		"attempt": serializeAttemptSummaryPtr(attempt),
	})
}

func (h *Handler) GetAttemptDetail(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	attemptID, err := strconv.ParseUint(c.Params("attemptId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "attempt id tidak valid", "")
	}

	detail, err := h.test.GetAttemptDetail(user, uint(attemptID))
	if err != nil {
		return utils.RespondError(c, fiber.StatusNotFound, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"id":              detail.Attempt.ID,
		"status":          detail.Attempt.Status,
		"startedAt":       detail.Attempt.StartedAt.UTC().Format(time.RFC3339),
		"expiresAt":       detail.Attempt.ExpiresAt.UTC().Format(time.RFC3339),
		"submittedAt":     nullableTime(detail.Attempt.SubmittedAt),
		"testType":        detail.Attempt.TestType,
		"roomCode":        nullableString(detail.Attempt.RoomCode),
		"roomLabel":       nullableString(detail.Attempt.RoomLabel),
		"durationMinutes": detail.Attempt.DurationMinutes,
		"rawScore":        detail.Attempt.RawScore,
		"totalQuestions":  detail.Attempt.TotalQuestions,
		"percentage":      detail.Attempt.Percentage,
		"sections":        serializeAttemptSections(detail.Sections),
		"questions":       detail.Questions,
	})
}

func (h *Handler) SaveAnswers(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	attemptID, err := strconv.ParseUint(c.Params("attemptId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "attempt id tidak valid", "")
	}

	var payload struct {
		Answers []services.SaveAnswerInput `json:"answers"`
	}
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload jawaban tidak valid", err.Error())
	}

	if err := h.test.SaveAnswers(user.ID, uint(attemptID), payload.Answers); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return utils.RespondMessage(c, fiber.StatusOK, "jawaban tersimpan", nil)
}

func (h *Handler) StartAttemptSection(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	attemptID, err := strconv.ParseUint(c.Params("attemptId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "attempt id tidak valid", "")
	}

	sectionCode, err := services.NormalizeQuestionIndex(models.QuestionIndex(c.Params("sectionCode")))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "section code tidak valid", "")
	}

	section, err := h.test.StartSection(user.ID, uint(attemptID), sectionCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message": "bagian berhasil dimulai",
		"section": serializeAttemptSection(*section, 0),
	})
}

func (h *Handler) SubmitAttemptSection(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	attemptID, err := strconv.ParseUint(c.Params("attemptId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "attempt id tidak valid", "")
	}

	sectionCode, err := services.NormalizeQuestionIndex(models.QuestionIndex(c.Params("sectionCode")))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "section code tidak valid", "")
	}

	section, err := h.test.SubmitSection(user.ID, uint(attemptID), sectionCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message": "bagian berhasil disubmit",
		"section": serializeAttemptSection(*section, 0),
	})
}

func (h *Handler) SubmitAttempt(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	attemptID, err := strconv.ParseUint(c.Params("attemptId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "attempt id tidak valid", "")
	}

	attempt, err := h.test.SubmitAttempt(user.ID, uint(attemptID))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(fiber.Map{
		"message": "test berhasil disubmit",
		"attempt": serializeAttemptSummary(*attempt),
	})
}

func (h *Handler) LatestResult(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	testType, roomCode, err := parseOptionalTestSelection(c)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}
	if testType == "" {
		testType = models.TestTypeIQ
	}
	result, err := h.test.GetLatestResultDetail(user.ID, testType, roomCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil hasil", err.Error())
	}

	return c.JSON(fiber.Map{
		"attempt": serializeAttemptResultPtr(result),
	})
}

func (h *Handler) ListMyResults(c fiber.Ctx) error {
	user, _ := middleware.CurrentUser(c)
	testType, roomCode, err := parseOptionalTestSelection(c)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}
	if testType == "" {
		testType = models.TestTypeIQ
	}

	results, err := h.test.ListResultDetails(user.ID, testType, roomCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil riwayat hasil", err.Error())
	}

	rows := make([]fiber.Map, 0, len(results))
	for _, result := range results {
		rows = append(rows, serializeAttemptResult(result))
	}

	return c.JSON(fiber.Map{
		"results": rows,
	})
}

func (h *Handler) AdminOverview(c fiber.Ctx) error {
	overview, err := h.admin.GetOverview()
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil overview admin", err.Error())
	}

	return c.JSON(fiber.Map{
		"overview": fiber.Map{
			"questionStats": fiber.Map{
				"total":     overview.QuestionStats.Total,
				"published": overview.QuestionStats.Published,
				"draft":     overview.QuestionStats.Draft,
				"archived":  overview.QuestionStats.Archived,
				"visual":    overview.QuestionStats.Visual,
			},
			"userStats": fiber.Map{
				"total":            overview.UserStats.Total,
				"active":           overview.UserStats.Active,
				"inactive":         overview.UserStats.Inactive,
				"adminCount":       overview.UserStats.AdminCount,
				"participantCount": overview.UserStats.ParticipantCount,
			},
			"attemptStats": fiber.Map{
				"inProgress": overview.AttemptStats.InProgress,
				"submitted":  overview.AttemptStats.Submitted,
				"expired":    overview.AttemptStats.Expired,
			},
			"resultStats": fiber.Map{
				"submissionCount":    overview.ResultStats.SubmissionCount,
				"averagePercentage":  overview.ResultStats.AveragePercentage,
				"highestEstimatedIq": overview.ResultStats.HighestEstimatedIQ,
			},
			"questionHealth": serializeAdminQuestionHealth(overview.QuestionHealth),
			"activeConfig":   serializeAdminConfigHealth(overview.ActiveConfig),
		},
	})
}

func (h *Handler) ListQuestions(c fiber.Ctx) error {
	limit, offset := pagination(c)
	search := c.Query("q")
	status := c.Query("status")
	questionIndex := c.Query("index")
	questions, err := h.admin.ListQuestions(search, status, questionIndex, limit, offset)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil soal", err.Error())
	}

	ids := make([]uint, 0, len(questions))
	for _, question := range questions {
		ids = append(ids, question.ID)
	}
	counts, err := h.admin.CountQuestionOptions(ids)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal menghitung opsi soal", err.Error())
	}
	mediaCounts, err := h.admin.CountQuestionOptionMedia(ids)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal menghitung media opsi soal", err.Error())
	}

	rows := make([]fiber.Map, 0, len(questions))
	for _, question := range questions {
		var questionIndex any
		var questionIndexLabel any
		var subtestCode any
		var subtestLabel any
		if strings.TrimSpace(string(question.QuestionIndex)) != "" {
			questionIndex = question.QuestionIndex
			questionIndexLabel = services.GetQuestionIndexLabel(question.QuestionIndex)
		}
		if strings.TrimSpace(question.SubtestCode) != "" {
			subtestCode = question.SubtestCode
			subtestLabel = services.GetSubtestLabel(question.SubtestCode)
		}

		rows = append(rows, fiber.Map{
			"id":                 question.ID,
			"prompt":             question.Prompt,
			"promptMediaUrl":     nullableString(question.PromptMediaURL),
			"promptMediaAlt":     nullableString(question.PromptMediaAlt),
			"difficulty":         question.Difficulty,
			"questionIndex":      questionIndex,
			"questionIndexLabel": questionIndexLabel,
			"subtestCode":        subtestCode,
			"subtestLabel":       subtestLabel,
			"status":             question.Status,
			"hasOptionMedia":     mediaCounts[question.ID] > 0,
			"optionCount":        counts[question.ID],
			"updatedAt":          question.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return c.JSON(fiber.Map{"questions": rows})
}

func (h *Handler) CreateQuestion(c fiber.Ctx) error {
	var payload services.QuestionPayload
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload soal tidak valid", err.Error())
	}

	question, err := h.admin.CreateQuestion(payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.Status(fiber.StatusCreated).JSON(question)
}

func (h *Handler) UpdateQuestion(c fiber.Ctx) error {
	questionID, err := strconv.ParseUint(c.Params("questionId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "question id tidak valid", "")
	}

	var payload services.QuestionPayload
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload soal tidak valid", err.Error())
	}

	question, err := h.admin.UpdateQuestion(uint(questionID), payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(question)
}

func (h *Handler) DeleteQuestion(c fiber.Ctx) error {
	questionID, err := strconv.ParseUint(c.Params("questionId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "question id tidak valid", "")
	}

	if err := h.admin.ArchiveQuestion(uint(questionID)); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengarsipkan soal", err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "soal diarsipkan", nil)
}

func (h *Handler) ListUsers(c fiber.Ctx) error {
	limit, offset := pagination(c)
	search := c.Query("q")
	role := c.Query("role")
	status := c.Query("status")
	users, err := h.admin.ListUsers(search, role, status, limit, offset)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil users", err.Error())
	}

	rows := make([]fiber.Map, 0, len(users))
	for _, user := range users {
		rows = append(rows, fiber.Map{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"role":        user.Role,
			"status":      user.Status,
			"accountType": user.AccountType,
			"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return c.JSON(fiber.Map{"users": rows})
}

func (h *Handler) CreateUser(c fiber.Ctx) error {
	var payload services.UserPayload
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload user tidak valid", err.Error())
	}

	user, err := h.admin.CreateUser(payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.Status(fiber.StatusCreated).JSON(serializeUser(*user))
}

func (h *Handler) UpdateUser(c fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "user id tidak valid", "")
	}

	var payload services.UserUpdatePayload
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload user tidak valid", err.Error())
	}

	user, err := h.admin.UpdateUser(uint(userID), payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(serializeUser(*user))
}

func (h *Handler) ListResults(c fiber.Ctx) error {
	limit, offset := pagination(c)
	search := c.Query("q")
	rows, err := h.admin.ListResults(search, limit, offset)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil hasil", err.Error())
	}

	response := make([]fiber.Map, 0, len(rows))
	for _, row := range rows {
		profile := services.BuildScreeningIQProfile(row.Percentage)
		response = append(response, fiber.Map{
			"attemptId":           row.AttemptID,
			"userId":              row.UserID,
			"userName":            row.UserName,
			"email":               row.Email,
			"rawScore":            row.RawScore,
			"totalQuestions":      row.TotalQuestions,
			"percentage":          row.Percentage,
			"submittedAt":         row.SubmittedAt,
			"durationMinutes":     row.DurationMinutes,
			"estimatedIq":         profile.EstimatedIQ,
			"classificationLabel": profile.ClassificationLabel,
		})
	}

	return c.JSON(fiber.Map{"results": response})
}

func (h *Handler) GetTestConfig(c fiber.Ctx) error {
	testType, roomCode, err := parseOptionalTestSelection(c)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	if testType == "" {
		configs, err := h.admin.ListActiveTestConfigs("", "")
		if err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil konfigurasi test", err.Error())
		}
		rows := make([]fiber.Map, 0, len(configs))
		for _, config := range configs {
			rows = append(rows, serializeTestConfig(config))
		}
		return c.JSON(fiber.Map{"configs": rows})
	}

	config, err := h.admin.GetActiveTestConfig(testType, roomCode)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "gagal mengambil konfigurasi test", err.Error())
	}

	if config == nil {
		return c.JSON(fiber.Map{"config": nil})
	}

	return c.JSON(fiber.Map{"config": serializeTestConfig(*config)})
}

func (h *Handler) UpdateTestConfig(c fiber.Ctx) error {
	var payload services.TestConfigPayload
	if err := c.Bind().Body(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "payload konfigurasi test tidak valid", err.Error())
	}

	config, err := h.admin.UpsertTestConfig(payload)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error(), "")
	}

	return c.JSON(config)
}

func (h *Handler) setAuthCookies(c fiber.Ctx, accessToken string, refreshToken string) {
	accessCookie := new(fiber.Cookie)
	accessCookie.Name = auth.AccessCookieName
	accessCookie.Value = accessToken
	accessCookie.Path = "/"
	accessCookie.HTTPOnly = true
	accessCookie.Secure = h.cfg.CookieSecure
	accessCookie.SameSite = "lax"
	accessCookie.Expires = time.Now().Add(h.cfg.AccessTokenTTL)
	if strings.TrimSpace(h.cfg.CookieDomain) != "" {
		accessCookie.Domain = h.cfg.CookieDomain
	}
	c.Cookie(accessCookie)

	refreshCookie := new(fiber.Cookie)
	refreshCookie.Name = auth.RefreshCookieName
	refreshCookie.Value = refreshToken
	refreshCookie.Path = "/"
	refreshCookie.HTTPOnly = true
	refreshCookie.Secure = h.cfg.CookieSecure
	refreshCookie.SameSite = "lax"
	refreshCookie.Expires = time.Now().Add(h.cfg.RefreshTokenTTL)
	if strings.TrimSpace(h.cfg.CookieDomain) != "" {
		refreshCookie.Domain = h.cfg.CookieDomain
	}
	c.Cookie(refreshCookie)
}

func (h *Handler) clearAuthCookies(c fiber.Ctx) {
	expired := time.Now().Add(-time.Hour)
	for _, name := range []string{auth.AccessCookieName, auth.RefreshCookieName} {
		cookie := new(fiber.Cookie)
		cookie.Name = name
		cookie.Value = ""
		cookie.Path = "/"
		cookie.Expires = expired
		cookie.HTTPOnly = true
		cookie.Secure = h.cfg.CookieSecure
		cookie.SameSite = "lax"
		if strings.TrimSpace(h.cfg.CookieDomain) != "" {
			cookie.Domain = h.cfg.CookieDomain
		}
		c.Cookie(cookie)
	}
}

func serializeUser(user models.User) fiber.Map {
	return fiber.Map{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"role":        user.Role,
		"status":      user.Status,
		"accountType": user.AccountType,
	}
}

func serializeAttemptSummary(attempt models.Attempt) fiber.Map {
	return fiber.Map{
		"id":             attempt.ID,
		"testType":       attempt.TestType,
		"roomCode":       nullableString(attempt.RoomCode),
		"roomLabel":      nullableString(attempt.RoomLabel),
		"status":         attempt.Status,
		"startedAt":      attempt.StartedAt.UTC().Format(time.RFC3339),
		"expiresAt":      attempt.ExpiresAt.UTC().Format(time.RFC3339),
		"submittedAt":    nullableTime(attempt.SubmittedAt),
		"rawScore":       attempt.RawScore,
		"totalQuestions": attempt.TotalQuestions,
		"percentage":     attempt.Percentage,
	}
}

func serializeAttemptSummaryPtr(attempt *models.Attempt) any {
	if attempt == nil {
		return nil
	}
	return serializeAttemptSummary(*attempt)
}

func serializeAttemptSections(sections []services.AttemptSectionView) []fiber.Map {
	rows := make([]fiber.Map, 0, len(sections))
	for _, section := range sections {
		rows = append(rows, serializeAttemptSectionView(section))
	}
	return rows
}

func serializeAdminQuestionHealth(rows []services.AdminQuestionHealthRow) []fiber.Map {
	response := make([]fiber.Map, 0, len(rows))
	for _, row := range rows {
		response = append(response, fiber.Map{
			"code":      row.Code,
			"label":     row.Label,
			"published": row.Published,
			"total":     row.Total,
		})
	}
	return response
}

func serializeAdminConfigHealth(config *services.AdminConfigHealth) any {
	if config == nil {
		return nil
	}

	return fiber.Map{
		"id":                     config.ID,
		"title":                  config.Title,
		"testType":               config.TestType,
		"roomCode":               nullableString(config.RoomCode),
		"roomLabel":              nullableString(config.RoomLabel),
		"durationMinutes":        config.DurationMinutes,
		"questionCount":          config.QuestionCount,
		"publishedQuestionCount": config.PublishedQuestionCount,
		"canStartAttempt":        config.CanStartAttempt,
		"readinessMessage":       config.ReadinessMessage,
		"questionHealth":         serializeAdminQuestionHealth(config.QuestionHealth),
	}
}

func serializeAttemptSection(section models.AttemptSection, answeredCount int) fiber.Map {
	return fiber.Map{
		"id":              section.ID,
		"code":            section.QuestionIndex,
		"label":           services.GetQuestionIndexLabel(section.QuestionIndex),
		"orderNo":         section.OrderNo,
		"status":          section.Status,
		"durationMinutes": section.DurationMinutes,
		"questionCount":   section.QuestionCount,
		"answeredCount":   answeredCount,
		"startedAt":       nullableTime(section.StartedAt),
		"expiresAt":       nullableTime(section.ExpiresAt),
		"submittedAt":     nullableTime(section.SubmittedAt),
	}
}

func serializeTestConfig(config models.TestConfig) fiber.Map {
	return fiber.Map{
		"id":              config.ID,
		"title":           config.Title,
		"testType":        config.TestType,
		"roomCode":        nullableString(config.RoomCode),
		"roomLabel":       nullableString(config.RoomLabel),
		"durationMinutes": config.DurationMinutes,
		"questionCount":   config.QuestionCount,
		"isActive":        config.IsActive,
		"updatedAt":       config.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func serializeAttemptSectionView(section services.AttemptSectionView) fiber.Map {
	return fiber.Map{
		"id":              section.ID,
		"code":            section.Code,
		"label":           section.Label,
		"orderNo":         section.OrderNo,
		"status":          section.Status,
		"durationMinutes": section.DurationMinutes,
		"questionCount":   section.QuestionCount,
		"answeredCount":   section.AnsweredCount,
		"startedAt":       nullableTime(section.StartedAt),
		"expiresAt":       nullableTime(section.ExpiresAt),
		"submittedAt":     nullableTime(section.SubmittedAt),
	}
}

func serializeAttemptResult(result services.AttemptResultView) fiber.Map {
	indexScores := make([]fiber.Map, 0, len(result.IndexScores))
	for _, score := range result.IndexScores {
		indexScores = append(indexScores, fiber.Map{
			"code":       score.Code,
			"label":      score.Label,
			"correct":    score.Correct,
			"total":      score.Total,
			"percentage": score.Percentage,
		})
	}

	payload := serializeAttemptSummary(result.Attempt)
	payload["estimatedIq"] = result.EstimatedIQ
	payload["skbScore"] = result.SKBScore
	payload["classificationLabel"] = result.ClassificationLabel
	payload["indexScores"] = indexScores
	return payload
}

func serializeAttemptResultPtr(result *services.AttemptResultView) any {
	if result == nil {
		return nil
	}
	return serializeAttemptResult(*result)
}

func pagination(c fiber.Ctx) (int, int) {
	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	return limit, (page - 1) * limit
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func parseOptionalTestSelection(c fiber.Ctx) (models.TestType, string, error) {
	testType := strings.TrimSpace(strings.ToUpper(c.Query("testType")))
	roomCode := strings.TrimSpace(strings.ToUpper(c.Query("roomCode")))
	if testType == "" {
		return "", "", nil
	}

	normalizedType, err := services.NormalizeTestType(models.TestType(testType))
	if err != nil {
		return "", "", err
	}
	normalizedRoomCode, err := services.NormalizeRoomCode(roomCode)
	if err != nil {
		return "", "", err
	}
	if normalizedType == models.TestTypeIQ {
		normalizedRoomCode = ""
	}
	return normalizedType, normalizedRoomCode, nil
}

func parseTestSelection(c fiber.Ctx) (models.TestType, string, error) {
	testType, roomCode, err := parseOptionalTestSelection(c)
	if err != nil {
		return "", "", err
	}
	if testType == "" {
		return models.TestTypeIQ, "", nil
	}
	if testType == models.TestTypeSKB && roomCode == "" {
		return "", "", fiber.NewError(fiber.StatusBadRequest, "room SKB wajib dipilih")
	}
	return testType, roomCode, nil
}
