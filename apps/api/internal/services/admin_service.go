package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/auth"
	"test-iq-ku/apps/api/internal/models"
)

type QuestionPayload struct {
	Prompt         string                `json:"prompt"`
	PromptMediaURL string                `json:"promptMediaUrl"`
	PromptMediaAlt string                `json:"promptMediaAlt"`
	Difficulty     string                `json:"difficulty"`
	QuestionIndex  models.QuestionIndex  `json:"questionIndex"`
	SubtestCode    string                `json:"subtestCode"`
	Status         models.QuestionStatus `json:"status"`
	Options        []QuestionOptionInput `json:"options"`
}

type UserPayload struct {
	Name        string             `json:"name"`
	Position    string             `json:"position"`
	Email       string             `json:"email"`
	Password    string             `json:"password"`
	Role        models.Role        `json:"role"`
	Status      models.UserStatus  `json:"status"`
	AccountType models.AccountType `json:"accountType"`
}

type UserUpdatePayload struct {
	Name        *string             `json:"name"`
	Position    *string             `json:"position"`
	Role        *models.Role        `json:"role"`
	Status      *models.UserStatus  `json:"status"`
	AccountType *models.AccountType `json:"accountType"`
}

type TestConfigPayload struct {
	Title           string `json:"title"`
	TestType        string `json:"testType"`
	RoomCode        string `json:"roomCode"`
	RoomLabel       string `json:"roomLabel"`
	DurationMinutes int    `json:"durationMinutes"`
	QuestionCount   int    `json:"questionCount"`
	Active          bool   `json:"active"`
}

type AdminOverview struct {
	QuestionStats  AdminQuestionStats       `json:"questionStats"`
	UserStats      AdminUserStats           `json:"userStats"`
	AttemptStats   AdminAttemptStats        `json:"attemptStats"`
	ResultStats    AdminResultStats         `json:"resultStats"`
	QuestionHealth []AdminQuestionHealthRow `json:"questionHealth"`
	ActiveConfig   *AdminConfigHealth       `json:"activeConfig"`
}

type AdminQuestionStats struct {
	Total     int64 `json:"total"`
	Published int64 `json:"published"`
	Draft     int64 `json:"draft"`
	Archived  int64 `json:"archived"`
	Visual    int64 `json:"visual"`
}

type AdminUserStats struct {
	Total            int64 `json:"total"`
	Active           int64 `json:"active"`
	Inactive         int64 `json:"inactive"`
	AdminCount       int64 `json:"adminCount"`
	ParticipantCount int64 `json:"participantCount"`
}

type AdminAttemptStats struct {
	InProgress int64 `json:"inProgress"`
	Submitted  int64 `json:"submitted"`
	Expired    int64 `json:"expired"`
}

type AdminResultStats struct {
	SubmissionCount    int64   `json:"submissionCount"`
	AveragePercentage  float64 `json:"averagePercentage"`
	HighestEstimatedIQ int     `json:"highestEstimatedIq"`
}

type AdminQuestionHealthRow struct {
	Code      models.QuestionIndex `json:"code"`
	Label     string               `json:"label"`
	Published int64                `json:"published"`
	Total     int64                `json:"total"`
}

type AdminConfigHealth struct {
	ID                     uint                     `json:"id"`
	Title                  string                   `json:"title"`
	TestType               models.TestType          `json:"testType"`
	RoomCode               string                   `json:"roomCode"`
	RoomLabel              string                   `json:"roomLabel"`
	DurationMinutes        int                      `json:"durationMinutes"`
	QuestionCount          int                      `json:"questionCount"`
	PublishedQuestionCount int64                    `json:"publishedQuestionCount"`
	CanStartAttempt        bool                     `json:"canStartAttempt"`
	ReadinessMessage       string                   `json:"readinessMessage"`
	QuestionHealth         []AdminQuestionHealthRow `json:"questionHealth"`
}

type AdminResultRow struct {
	AttemptID       uint    `json:"attemptId"`
	UserID          uint    `json:"userId"`
	UserName        string  `json:"userName"`
	Email           string  `json:"email"`
	RawScore        int     `json:"rawScore"`
	TotalQuestions  int     `json:"totalQuestions"`
	Percentage      float64 `json:"percentage"`
	SubmittedAt     string  `json:"submittedAt"`
	DurationMinutes int     `json:"durationMinutes"`
}

type AdminService struct {
	db *gorm.DB
}

type QuestionOptionInput struct {
	Key       string `json:"key"`
	Content   string `json:"content"`
	MediaURL  string `json:"mediaUrl"`
	MediaAlt  string `json:"mediaAlt"`
	IsCorrect bool   `json:"isCorrect"`
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) ListQuestions(search string, status string, questionIndex string, limit int, offset int) ([]models.Question, error) {
	var questions []models.Question
	query := s.db.Model(&models.Question{})
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		query = query.Where("prompt LIKE ?", "%"+trimmed+"%")
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		query = query.Where("status = ?", trimmed)
	} else {
		query = query.Where("status <> ?", models.QuestionStatusArchived)
	}
	if trimmed := strings.TrimSpace(questionIndex); trimmed != "" {
		query = query.Where("question_index = ?", strings.ToUpper(trimmed))
	}

	if err := query.
		Order("status = 'PUBLISHED' DESC").
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&questions).Error; err != nil {
		return nil, err
	}

	return questions, nil
}

func (s *AdminService) CountQuestionOptions(questionIDs []uint) (map[uint]int64, error) {
	if len(questionIDs) == 0 {
		return map[uint]int64{}, nil
	}

	type row struct {
		QuestionID uint
		Total      int64
	}

	var rows []row
	if err := s.db.Model(&models.QuestionOption{}).
		Select("question_id, COUNT(*) as total").
		Where("question_id IN ?", questionIDs).
		Group("question_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]int64, len(rows))
	for _, item := range rows {
		result[item.QuestionID] = item.Total
	}

	return result, nil
}

func (s *AdminService) CountQuestionOptionMedia(questionIDs []uint) (map[uint]int64, error) {
	if len(questionIDs) == 0 {
		return map[uint]int64{}, nil
	}

	type row struct {
		QuestionID uint
		Total      int64
	}

	var rows []row
	if err := s.db.Model(&models.QuestionOption{}).
		Select("question_id, COUNT(*) as total").
		Where("question_id IN ? AND media_url IS NOT NULL AND TRIM(media_url) <> ''", questionIDs).
		Group("question_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]int64, len(rows))
	for _, item := range rows {
		result[item.QuestionID] = item.Total
	}

	return result, nil
}

func (s *AdminService) CreateQuestion(payload QuestionPayload) (*models.Question, error) {
	if err := validateQuestionPayload(payload); err != nil {
		return nil, err
	}
	questionIndex, subtestCode, _ := ValidateQuestionTaxonomy(payload.QuestionIndex, payload.SubtestCode)

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	question := models.Question{
		Prompt:         strings.TrimSpace(payload.Prompt),
		PromptMediaURL: strings.TrimSpace(payload.PromptMediaURL),
		PromptMediaAlt: strings.TrimSpace(payload.PromptMediaAlt),
		Difficulty:     normalizeDifficulty(payload.Difficulty),
		QuestionIndex:  questionIndex,
		SubtestCode:    subtestCode,
		Status:         payload.Status,
	}

	if err := tx.Create(&question).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	options := make([]models.QuestionOption, 0, len(payload.Options))
	for _, option := range payload.Options {
		options = append(options, models.QuestionOption{
			QuestionID: question.ID,
			Key:        strings.ToUpper(strings.TrimSpace(option.Key)),
			Content:    strings.TrimSpace(option.Content),
			MediaURL:   strings.TrimSpace(option.MediaURL),
			MediaAlt:   strings.TrimSpace(option.MediaAlt),
			IsCorrect:  option.IsCorrect,
		})
	}

	if err := tx.Create(&options).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &question, nil
}

func (s *AdminService) UpdateQuestion(questionID uint, payload QuestionPayload) (*models.Question, error) {
	if err := validateQuestionPayload(payload); err != nil {
		return nil, err
	}
	questionIndex, subtestCode, _ := ValidateQuestionTaxonomy(payload.QuestionIndex, payload.SubtestCode)

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var question models.Question
	if err := tx.First(&question, questionID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("soal tidak ditemukan")
	}

	question.Prompt = strings.TrimSpace(payload.Prompt)
	question.PromptMediaURL = strings.TrimSpace(payload.PromptMediaURL)
	question.PromptMediaAlt = strings.TrimSpace(payload.PromptMediaAlt)
	question.Difficulty = normalizeDifficulty(payload.Difficulty)
	question.QuestionIndex = questionIndex
	question.SubtestCode = subtestCode
	question.Status = payload.Status

	if err := tx.Save(&question).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Where("question_id = ?", questionID).Delete(&models.QuestionOption{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	options := make([]models.QuestionOption, 0, len(payload.Options))
	for _, option := range payload.Options {
		options = append(options, models.QuestionOption{
			QuestionID: question.ID,
			Key:        strings.ToUpper(strings.TrimSpace(option.Key)),
			Content:    strings.TrimSpace(option.Content),
			MediaURL:   strings.TrimSpace(option.MediaURL),
			MediaAlt:   strings.TrimSpace(option.MediaAlt),
			IsCorrect:  option.IsCorrect,
		})
	}

	if err := tx.Create(&options).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &question, nil
}

func (s *AdminService) ArchiveQuestion(questionID uint) error {
	return s.db.Model(&models.Question{}).
		Where("id = ?", questionID).
		Update("status", models.QuestionStatusArchived).
		Error
}

func (s *AdminService) ListUsers(search string, role string, status string, limit int, offset int) ([]models.User, error) {
	var users []models.User
	query := s.db.Model(&models.User{})
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+trimmed+"%", "%"+trimmed+"%")
	}
	if trimmed := strings.TrimSpace(role); trimmed != "" {
		query = query.Where("role = ?", strings.ToUpper(trimmed))
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		query = query.Where("status = ?", strings.ToUpper(trimmed))
	}
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *AdminService) CreateUser(payload UserPayload) (*models.User, error) {
	normalizedName, normalizedPosition, normalizedEmail, err := validatePublicRegistration(payload.Name, payload.Position, payload.Email, payload.Password)
	if err != nil {
		return nil, err
	}

	role, err := normalizeRole(payload.Role)
	if err != nil {
		return nil, err
	}

	status, err := normalizeUserStatus(payload.Status)
	if err != nil {
		return nil, err
	}

	accountType, err := normalizeAccountType(payload.AccountType)
	if err != nil {
		return nil, err
	}

	passwordHash, err := auth.HashPassword(strings.TrimSpace(payload.Password))
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:         normalizedName,
		Position:     normalizedPosition,
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
		AccountType:  accountType,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, mapCreateUserError(err)
	}

	return &user, nil
}

func (s *AdminService) UpdateUser(userID uint, payload UserUpdatePayload) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	if payload.Name != nil {
		normalizedName := normalizeName(*payload.Name)
		if normalizedName == "" {
			return nil, errors.New("nama wajib diisi")
		}
		user.Name = normalizedName
	}
	if payload.Position != nil {
		normalizedPosition := normalizePosition(*payload.Position)
		if normalizedPosition == "" {
			return nil, errors.New("jabatan wajib diisi")
		}
		user.Position = normalizedPosition
	}
	if payload.Role != nil {
		role, err := normalizeRole(*payload.Role)
		if err != nil {
			return nil, err
		}
		user.Role = role
	}
	if payload.Status != nil {
		status, err := normalizeUserStatus(*payload.Status)
		if err != nil {
			return nil, err
		}
		user.Status = status
	}
	if payload.AccountType != nil {
		accountType, err := normalizeAccountType(*payload.AccountType)
		if err != nil {
			return nil, err
		}
		user.AccountType = accountType
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AdminService) ListResults(search string, limit int, offset int) ([]AdminResultRow, error) {
	var rows []AdminResultRow
	query := s.db.Table("attempts").
		Select(`
			attempts.id as attempt_id,
			attempts.user_id,
			users.name as user_name,
			users.email,
			COALESCE(attempts.raw_score, 0) as raw_score,
			attempts.total_questions,
			COALESCE(attempts.percentage, 0) as percentage,
			DATE_FORMAT(attempts.submitted_at, '%Y-%m-%dT%H:%i:%sZ') as submitted_at,
			attempts.duration_minutes
		`).
		Joins("JOIN users ON users.id = attempts.user_id").
		Where("attempts.status = ?", models.AttemptStatusSubmitted)
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		query = query.Where("users.name LIKE ? OR users.email LIKE ?", "%"+trimmed+"%", "%"+trimmed+"%")
	}
	err := query.
		Order("attempts.percentage DESC, attempts.raw_score DESC, attempts.duration_minutes ASC, attempts.submitted_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *AdminService) GetActiveTestConfig(testType models.TestType, roomCode string) (*models.TestConfig, error) {
	configs, err := s.ListActiveTestConfigs(testType, roomCode)
	if err != nil || len(configs) == 0 {
		return nil, err
	}
	return &configs[0], nil
}

func (s *AdminService) ListActiveTestConfigs(testType models.TestType, roomCode string) ([]models.TestConfig, error) {
	var configs []models.TestConfig
	query := s.db.Where("is_active = ?", true)
	if testType != "" {
		query = query.Where("test_type = ?", testType)
	}
	if roomCode != "" {
		query = query.Where("room_code IN ?", MatchingRoomCodes(roomCode))
	}
	if err := query.Order("test_type ASC, room_label ASC, id DESC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *AdminService) GetOverview() (*AdminOverview, error) {
	questionStats, err := s.loadQuestionStats()
	if err != nil {
		return nil, err
	}
	userStats, err := s.loadUserStats()
	if err != nil {
		return nil, err
	}
	attemptStats, err := s.loadAttemptStats()
	if err != nil {
		return nil, err
	}
	resultStats, err := s.loadResultStats()
	if err != nil {
		return nil, err
	}
	questionHealth, err := s.loadQuestionHealth()
	if err != nil {
		return nil, err
	}
	activeConfig, err := s.GetActiveTestConfig(models.TestTypeIQ, "")
	if err != nil {
		return nil, err
	}

	var configHealth *AdminConfigHealth
	if activeConfig != nil {
		activeQuestionHealth, _, err := s.loadQuestionHealthForConfig(*activeConfig)
		if err != nil {
			return nil, err
		}
		configHealth = buildAdminConfigHealth(*activeConfig, activeQuestionHealth)
	}

	return &AdminOverview{
		QuestionStats:  questionStats,
		UserStats:      userStats,
		AttemptStats:   attemptStats,
		ResultStats:    resultStats,
		QuestionHealth: questionHealth,
		ActiveConfig:   configHealth,
	}, nil
}

func (s *AdminService) UpsertTestConfig(payload TestConfigPayload) (*models.TestConfig, error) {
	if strings.TrimSpace(payload.Title) == "" || payload.DurationMinutes <= 0 || payload.QuestionCount <= 0 {
		return nil, errors.New("title, durasi, dan jumlah soal harus valid")
	}
	testType, err := NormalizeTestType(models.TestType(payload.TestType))
	if err != nil {
		return nil, err
	}

	roomCode, err := NormalizeRoomCode(payload.RoomCode)
	if err != nil {
		return nil, err
	}

	roomLabel := strings.TrimSpace(payload.RoomLabel)
	if testType == models.TestTypeIQ {
		roomCode = ""
		roomLabel = ""
	}
	if testType == models.TestTypeSKB {
		if roomCode == "" {
			return nil, errors.New("room SKB wajib dipilih")
		}
		if roomLabel == "" {
			roomLabel = GetRoomLabel(roomCode)
		}
	}

	if testType == models.TestTypeIQ && payload.QuestionCount < len(orderedQuestionIndices) {
		return nil, errors.New("jumlah soal minimal harus mencakup seluruh 4 index utama")
	}
	if testType == models.TestTypeSKB && payload.QuestionCount < 1 {
		return nil, errors.New("jumlah soal SKB minimal 1")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Model(&models.TestConfig{}).
		Where("is_active = ? AND test_type = ? AND room_code = ?", true, testType, roomCode).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	config := models.TestConfig{
		Title:           strings.TrimSpace(payload.Title),
		TestType:        testType,
		RoomCode:        roomCode,
		RoomLabel:       roomLabel,
		DurationMinutes: payload.DurationMinutes,
		QuestionCount:   payload.QuestionCount,
		IsActive:        true,
	}

	if err := tx.Create(&config).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &config, nil
}

func validateQuestionPayload(payload QuestionPayload) error {
	if strings.TrimSpace(payload.Prompt) == "" {
		return errors.New("prompt wajib diisi")
	}
	if _, _, err := ValidateQuestionTaxonomy(payload.QuestionIndex, payload.SubtestCode); err != nil {
		return err
	}
	if len(payload.Options) < 2 {
		return errors.New("minimal harus ada dua opsi jawaban")
	}
	if payload.Status == "" {
		return errors.New("status soal wajib diisi")
	}

	seen := make(map[string]bool, len(payload.Options))
	correctCount := 0
	for _, option := range payload.Options {
		key := strings.ToUpper(strings.TrimSpace(option.Key))
		content := strings.TrimSpace(option.Content)
		mediaURL := strings.TrimSpace(option.MediaURL)
		if key == "" {
			return errors.New("setiap opsi harus memiliki key")
		}
		if content == "" && mediaURL == "" {
			return errors.New("setiap opsi harus memiliki content atau media")
		}
		if seen[key] {
			return errors.New("key opsi jawaban harus unik")
		}
		seen[key] = true
		if option.IsCorrect {
			correctCount++
		}
	}

	if correctCount != 1 {
		return errors.New("harus ada tepat satu jawaban benar")
	}

	return nil
}

func normalizeDifficulty(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "medium"
	}
	return trimmed
}

func (s *AdminService) loadQuestionStats() (AdminQuestionStats, error) {
	type row struct {
		Total     int64
		Published int64
		Draft     int64
		Archived  int64
		Visual    int64
	}

	var result row
	err := s.db.Model(&models.Question{}).
		Select(`
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN status = 'PUBLISHED' THEN 1 ELSE 0 END), 0) as published,
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0) as draft,
			COALESCE(SUM(CASE WHEN status = 'ARCHIVED' THEN 1 ELSE 0 END), 0) as archived,
			COALESCE(SUM(CASE WHEN (prompt_media_url IS NOT NULL AND TRIM(prompt_media_url) <> '') THEN 1 ELSE 0 END), 0) as visual
		`).
		Scan(&result).Error
	if err != nil {
		return AdminQuestionStats{}, err
	}

	return AdminQuestionStats{
		Total:     result.Total,
		Published: result.Published,
		Draft:     result.Draft,
		Archived:  result.Archived,
		Visual:    result.Visual,
	}, nil
}

func (s *AdminService) loadUserStats() (AdminUserStats, error) {
	type row struct {
		Total            int64
		Active           int64
		Inactive         int64
		AdminCount       int64
		ParticipantCount int64
	}

	var result row
	err := s.db.Model(&models.User{}).
		Select(`
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END), 0) as active,
			COALESCE(SUM(CASE WHEN status = 'INACTIVE' THEN 1 ELSE 0 END), 0) as inactive,
			COALESCE(SUM(CASE WHEN role = 'ADMIN' THEN 1 ELSE 0 END), 0) as admin_count,
			COALESCE(SUM(CASE WHEN role = 'USER' THEN 1 ELSE 0 END), 0) as participant_count
		`).
		Scan(&result).Error
	if err != nil {
		return AdminUserStats{}, err
	}

	return AdminUserStats{
		Total:            result.Total,
		Active:           result.Active,
		Inactive:         result.Inactive,
		AdminCount:       result.AdminCount,
		ParticipantCount: result.ParticipantCount,
	}, nil
}

func (s *AdminService) loadAttemptStats() (AdminAttemptStats, error) {
	type row struct {
		InProgress int64
		Submitted  int64
		Expired    int64
	}

	var result row
	err := s.db.Model(&models.Attempt{}).
		Select(`
			COALESCE(SUM(CASE WHEN status = 'IN_PROGRESS' THEN 1 ELSE 0 END), 0) as in_progress,
			COALESCE(SUM(CASE WHEN status = 'SUBMITTED' THEN 1 ELSE 0 END), 0) as submitted,
			COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0) as expired
		`).
		Scan(&result).Error
	if err != nil {
		return AdminAttemptStats{}, err
	}

	return AdminAttemptStats{
		InProgress: result.InProgress,
		Submitted:  result.Submitted,
		Expired:    result.Expired,
	}, nil
}

func (s *AdminService) loadResultStats() (AdminResultStats, error) {
	type row struct {
		SubmissionCount   int64
		AveragePercentage float64
		MaxPercentage     float64
	}

	var result row
	err := s.db.Model(&models.Attempt{}).
		Select(`
			COUNT(*) as submission_count,
			COALESCE(AVG(percentage), 0) as average_percentage,
			COALESCE(MAX(percentage), 0) as max_percentage
		`).
		Where("status = ?", models.AttemptStatusSubmitted).
		Scan(&result).Error
	if err != nil {
		return AdminResultStats{}, err
	}

	return AdminResultStats{
		SubmissionCount:    result.SubmissionCount,
		AveragePercentage:  result.AveragePercentage,
		HighestEstimatedIQ: BuildScreeningIQProfile(result.MaxPercentage).EstimatedIQ,
	}, nil
}

func (s *AdminService) loadQuestionHealth() ([]AdminQuestionHealthRow, error) {
	type row struct {
		Code      models.QuestionIndex
		Published int64
		Total     int64
	}

	var rows []row
	err := s.db.Model(&models.Question{}).
		Select(`
			question_index as code,
			SUM(CASE WHEN status = 'PUBLISHED' THEN 1 ELSE 0 END) as published,
			COUNT(*) as total
		`).
		Where("question_index <> ''").
		Group("question_index").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	byCode := make(map[models.QuestionIndex]row, len(rows))
	for _, item := range rows {
		byCode[item.Code] = item
	}

	response := make([]AdminQuestionHealthRow, 0, len(orderedQuestionIndices))
	for _, code := range OrderedQuestionIndices() {
		item := byCode[code]
		response = append(response, AdminQuestionHealthRow{
			Code:      code,
			Label:     GetQuestionIndexLabel(code),
			Published: item.Published,
			Total:     item.Total,
		})
	}

	return response, nil
}

func (s *AdminService) loadQuestionHealthForConfig(config models.TestConfig) ([]AdminQuestionHealthRow, int64, error) {
	if config.TestType == models.TestTypeSKB {
		var row struct {
			Published int64
			Total     int64
		}
		if err := s.db.Model(&models.Question{}).
			Select(`
				SUM(CASE WHEN status = 'PUBLISHED' THEN 1 ELSE 0 END) as published,
				COUNT(*) as total
			`).
			Where("question_index = ? AND subtest_code = ?", models.QuestionIndexSKB, config.RoomCode).
			Scan(&row).Error; err != nil {
			return nil, 0, err
		}

		rows := []AdminQuestionHealthRow{{
			Code:      models.QuestionIndexSKB,
			Label:     config.RoomLabel,
			Published: row.Published,
			Total:     row.Total,
		}}
		return rows, row.Published, nil
	}

	rows, err := s.loadQuestionHealth()
	if err != nil {
		return nil, 0, err
	}

	var published int64
	for _, item := range rows {
		published += item.Published
	}

	return rows, published, nil
}

func buildAdminConfigHealth(
	config models.TestConfig,
	questionHealth []AdminQuestionHealthRow,
) *AdminConfigHealth {
	available := make(map[models.QuestionIndex]int, len(questionHealth))
	for _, item := range questionHealth {
		available[item.Code] = int(item.Published)
	}

	publishedQuestionCount := int64(0)
	for _, item := range questionHealth {
		publishedQuestionCount += item.Published
	}

	var err error
	effectiveQuestionCount := resolveQuestionCountForAccount(models.AccountTypeMax, config.TestType, config.QuestionCount)
	if config.TestType == models.TestTypeSKB {
		if available[models.QuestionIndexSKB] < effectiveQuestionCount {
			err = errors.New("bank soal SKB untuk kamar ini belum mencukupi")
		}
	} else {
		_, err = buildQuestionSelectionPlan(effectiveQuestionCount, available)
	}
	canStartAttempt := err == nil
	readinessMessage := "Konfigurasi aktif siap dipakai untuk attempt baru."
	if err != nil {
		readinessMessage = err.Error()
	}

	return &AdminConfigHealth{
		ID:                     config.ID,
		Title:                  config.Title,
		TestType:               config.TestType,
		RoomCode:               config.RoomCode,
		RoomLabel:              config.RoomLabel,
		DurationMinutes:        config.DurationMinutes,
		QuestionCount:          effectiveQuestionCount,
		PublishedQuestionCount: publishedQuestionCount,
		CanStartAttempt:        canStartAttempt,
		ReadinessMessage:       readinessMessage,
		QuestionHealth:         questionHealth,
	}
}
