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
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Password string            `json:"password"`
	Role     models.Role       `json:"role"`
	Status   models.UserStatus `json:"status"`
}

type UserUpdatePayload struct {
	Name   *string            `json:"name"`
	Role   *models.Role       `json:"role"`
	Status *models.UserStatus `json:"status"`
}

type TestConfigPayload struct {
	Title           string `json:"title"`
	DurationMinutes int    `json:"durationMinutes"`
	QuestionCount   int    `json:"questionCount"`
	Active          bool   `json:"active"`
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

func (s *AdminService) ListQuestions(search string, status string, limit int, offset int) ([]models.Question, error) {
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

func (s *AdminService) ListUsers(limit int, offset int) ([]models.User, error) {
	var users []models.User
	if err := s.db.Model(&models.User{}).Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *AdminService) CreateUser(payload UserPayload) (*models.User, error) {
	normalizedName, normalizedEmail, err := validatePublicRegistration(payload.Name, payload.Email, payload.Password)
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

	passwordHash, err := auth.HashPassword(strings.TrimSpace(payload.Password))
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:         normalizedName,
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
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

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AdminService) ListResults(limit int, offset int) ([]AdminResultRow, error) {
	var rows []AdminResultRow
	err := s.db.Table("attempts").
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
		Where("attempts.status = ?", models.AttemptStatusSubmitted).
		Order("attempts.percentage DESC, attempts.raw_score DESC, attempts.duration_minutes ASC, attempts.submitted_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *AdminService) GetActiveTestConfig() (*models.TestConfig, error) {
	var config models.TestConfig
	if err := s.db.Where("is_active = ?", true).Order("id DESC").First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}

func (s *AdminService) UpsertTestConfig(payload TestConfigPayload) (*models.TestConfig, error) {
	if strings.TrimSpace(payload.Title) == "" || payload.DurationMinutes <= 0 || payload.QuestionCount <= 0 {
		return nil, errors.New("title, durasi, dan jumlah soal harus valid")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Model(&models.TestConfig{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	config := models.TestConfig{
		Title:           strings.TrimSpace(payload.Title),
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
