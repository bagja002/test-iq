package services

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"test-iq-ku/apps/api/internal/models"
)

type AttemptQuestionView struct {
	ID                 uint                    `json:"id"`
	QuestionID         uint                    `json:"questionId"`
	OrderNo            int                     `json:"orderNo"`
	QuestionIndex      models.QuestionIndex    `json:"questionIndex"`
	QuestionIndexLabel string                  `json:"questionIndexLabel"`
	SubtestCode        string                  `json:"subtestCode"`
	SubtestLabel       string                  `json:"subtestLabel"`
	Prompt             string                  `json:"prompt"`
	PromptMediaURL     string                  `json:"promptMediaUrl"`
	PromptMediaAlt     string                  `json:"promptMediaAlt"`
	Options            []models.OptionSnapshot `json:"options"`
	SelectedOptionKey  *string                 `json:"selectedOptionKey"`
}

type AttemptSectionView struct {
	ID              uint                        `json:"id"`
	Code            models.QuestionIndex        `json:"code"`
	Label           string                      `json:"label"`
	OrderNo         int                         `json:"orderNo"`
	Status          models.AttemptSectionStatus `json:"status"`
	DurationMinutes int                         `json:"durationMinutes"`
	QuestionCount   int                         `json:"questionCount"`
	AnsweredCount   int                         `json:"answeredCount"`
	StartedAt       *time.Time                  `json:"startedAt"`
	ExpiresAt       *time.Time                  `json:"expiresAt"`
	SubmittedAt     *time.Time                  `json:"submittedAt"`
}

type AttemptDetailView struct {
	Attempt   models.Attempt
	Questions []AttemptQuestionView
	Sections  []AttemptSectionView
}

type IndexScoreView struct {
	Code       models.QuestionIndex `json:"code"`
	Label      string               `json:"label"`
	Correct    int                  `json:"correct"`
	Total      int                  `json:"total"`
	Percentage float64              `json:"percentage"`
}

type AttemptResultView struct {
	Attempt             models.Attempt
	EstimatedIQ         *int
	SKBScore            *float64
	ClassificationLabel string
	IndexScores         []IndexScoreView
}

type StartAttemptInput struct {
	TestType models.TestType `json:"testType"`
	RoomCode string          `json:"roomCode"`
}

type SaveAnswerInput struct {
	AttemptQuestionID uint   `json:"attemptQuestionId"`
	SelectedOptionKey string `json:"selectedOptionKey"`
}

type TestService struct {
	db *gorm.DB
}

const freeIQQuestionLimit = 2
const iqQuestionsPerSection = 30
const iqTotalQuestionCount = iqQuestionsPerSection * 4
const skbQuestionCountPerAttempt = 30

func NewTestService(db *gorm.DB) *TestService {
	return &TestService{db: db}
}

func (s *TestService) StartAttempt(user models.User, input StartAttemptInput) (*models.Attempt, error) {
	now := time.Now()
	if input.TestType == "" {
		input.TestType = models.TestTypeIQ
	}
	testType, err := NormalizeTestType(input.TestType)
	if err != nil {
		return nil, err
	}
	roomCode, err := NormalizeRoomCode(input.RoomCode)
	if err != nil {
		return nil, err
	}
	if testType == models.TestTypeIQ {
		roomCode = ""
	}
	if testType == models.TestTypeSKB && roomCode == "" {
		return nil, errors.New("room SKB wajib dipilih")
	}
	if err := validateAccountTestAccess(user.AccountType, testType); err != nil {
		return nil, err
	}
	roomCodes := []string{roomCode}
	if testType == models.TestTypeSKB {
		roomCodes = MatchingRoomCodes(roomCode)
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := expireAttempts(tx, user.ID, testType, roomCode, now); err != nil {
		tx.Rollback()
		return nil, err
	}

	var activeAttempt models.Attempt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND test_type = ? AND room_code IN ? AND status = ?", user.ID, testType, roomCodes, models.AttemptStatusInProgress).
		Order("id DESC").
		First(&activeAttempt).Error; err == nil {
		tx.Commit()
		return &activeAttempt, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	if err := ensureDailySubmitQuota(tx, user.ID, user.AccountType, testType, now); err != nil {
		tx.Rollback()
		return nil, err
	}

	var config models.TestConfig
	if err := tx.Where("is_active = ? AND test_type = ? AND room_code IN ?", true, testType, roomCodes).Order("id DESC").First(&config).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("belum ada test aktif")
	}

	effectiveQuestionCount := resolveQuestionCountForAccount(user.AccountType, config.TestType, config.QuestionCount)

	type availabilityRow struct {
		QuestionIndex models.QuestionIndex
		Total         int
	}

	var availabilityRows []availabilityRow
	questions := make([]models.Question, 0, config.QuestionCount)
	if config.TestType == models.TestTypeSKB {
		var chunk []models.Question
		if err := tx.Preload("Options").
			Where("status = ? AND question_index = ? AND subtest_code = ?", models.QuestionStatusPublished, models.QuestionIndexSKB, config.RoomCode).
			Order("RAND()").
			Limit(effectiveQuestionCount).
			Find(&chunk).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if len(chunk) < effectiveQuestionCount {
			tx.Rollback()
			return nil, errors.New("jumlah soal SKB published belum mencukupi untuk kamar ini")
		}
		questions = append(questions, chunk...)
	} else {
		if err := tx.Model(&models.Question{}).
			Select("question_index, COUNT(*) as total").
			Where("status = ? AND question_index <> '' AND subtest_code <> '' AND question_index <> ?", models.QuestionStatusPublished, models.QuestionIndexSKB).
			Group("question_index").
			Scan(&availabilityRows).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		available := make(map[models.QuestionIndex]int, len(availabilityRows))
		for _, row := range availabilityRows {
			available[row.QuestionIndex] = row.Total
		}

		selectionPlan, err := buildQuestionSelectionPlan(effectiveQuestionCount, available)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		for _, item := range selectionPlan {
			var chunk []models.Question
			if err := tx.Preload("Options").
				Where("status = ? AND question_index = ? AND subtest_code <> ''", models.QuestionStatusPublished, item.Code).
				Order("RAND()").
				Limit(item.Requested).
				Find(&chunk).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			if len(chunk) < item.Requested {
				tx.Rollback()
				return nil, errors.New("jumlah soal published belum mencukupi untuk distribusi per index")
			}

			questions = append(questions, chunk...)
		}
	}

	sortQuestionsForAttempt(questions)

	attempt := models.Attempt{
		UserID:          user.ID,
		TestConfigID:    config.ID,
		TestType:        config.TestType,
		RoomCode:        config.RoomCode,
		RoomLabel:       config.RoomLabel,
		Status:          models.AttemptStatusInProgress,
		StartedAt:       now,
		ExpiresAt:       now.Add(time.Duration(config.DurationMinutes) * time.Minute),
		TotalQuestions:  len(questions),
		DurationMinutes: config.DurationMinutes,
	}

	if err := tx.Create(&attempt).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	attemptQuestions := make([]models.AttemptQuestion, 0, len(questions))
	for idx, question := range questions {
		options := make([]models.OptionSnapshot, 0, len(question.Options))
		correctKey := ""
		for _, option := range question.Options {
			options = append(options, models.OptionSnapshot{
				Key:      option.Key,
				Content:  option.Content,
				MediaURL: option.MediaURL,
				MediaAlt: option.MediaAlt,
			})
			if option.IsCorrect {
				correctKey = option.Key
			}
		}

		payload, err := models.MarshalOptionsSnapshot(options)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		attemptQuestions = append(attemptQuestions, models.AttemptQuestion{
			AttemptID:        attempt.ID,
			QuestionID:       question.ID,
			QuestionIndex:    question.QuestionIndex,
			SubtestCode:      question.SubtestCode,
			OrderNo:          idx + 1,
			PromptSnapshot:   question.Prompt,
			PromptMediaURL:   question.PromptMediaURL,
			PromptMediaAlt:   question.PromptMediaAlt,
			OptionsSnapshot:  payload,
			CorrectOptionKey: correctKey,
		})
	}

	if err := tx.Create(&attemptQuestions).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	attemptSections := buildAttemptSectionRecordsFromQuestions(attempt.ID, questions, config.DurationMinutes, config.TestType)
	if len(attemptSections) == 0 {
		tx.Rollback()
		return nil, errors.New("gagal membentuk bagian test")
	}

	if err := tx.Create(&attemptSections).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &attempt, nil
}

func (s *TestService) GetCurrentAttempt(userID uint, testType models.TestType, roomCode string) (*models.Attempt, error) {
	now := time.Now()
	if err := expireAttempts(s.db, userID, testType, roomCode, now); err != nil {
		return nil, err
	}

	roomCodes := []string{roomCode}
	if testType == models.TestTypeSKB {
		roomCodes = MatchingRoomCodes(roomCode)
	}

	var attempt models.Attempt
	if err := s.db.Where("user_id = ? AND test_type = ? AND room_code IN ? AND status = ?", userID, testType, roomCodes, models.AttemptStatusInProgress).Order("id DESC").First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &attempt, nil
}

func (s *TestService) GetAttemptDetail(requester models.User, attemptID uint) (*AttemptDetailView, error) {
	now := time.Now()
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var attempt models.Attempt
	query := tx.Where("id = ?", attemptID)
	if requester.Role != models.RoleAdmin {
		query = query.Where("user_id = ?", requester.ID)
	}

	if err := query.First(&attempt).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attempt tidak ditemukan")
		}
		return nil, err
	}
	if requester.Role != models.RoleAdmin {
		if err := validateAccountTestAccess(requester.AccountType, attempt.TestType); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	var questions []models.AttemptQuestion
	if err := tx.Where("attempt_id = ?", attempt.ID).Order("order_no ASC").Find(&questions).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err := ensureAttemptSections(tx, attempt, questions)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err = syncAttemptAndSections(tx, &attempt, sections, now)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var answers []models.AttemptAnswer
	if err := tx.Where("attempt_id = ?", attempt.ID).Find(&answers).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	answerMap := make(map[uint]string, len(answers))
	for _, answer := range answers {
		answerMap[answer.AttemptQuestionID] = answer.SelectedOptionKey
	}

	items, err := buildAttemptQuestionViews(questions, answerMap)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &AttemptDetailView{
		Attempt:   attempt,
		Questions: items,
		Sections:  buildAttemptSectionViews(sections, questions, answerMap),
	}, nil
}

func (s *TestService) SaveAnswers(userID uint, attemptID uint, answers []SaveAnswerInput) error {
	if len(answers) == 0 {
		return nil
	}

	now := time.Now()
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var attempt models.Attempt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", attemptID, userID).
		First(&attempt).Error; err != nil {
		tx.Rollback()
		return errors.New("attempt tidak ditemukan")
	}
	if err := ensureAttemptAllowedForUser(tx, userID, attempt.TestType); err != nil {
		tx.Rollback()
		return err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return errors.New("attempt sudah tidak aktif")
	}

	var attemptQuestions []models.AttemptQuestion
	if err := tx.Where("attempt_id = ?", attemptID).Order("order_no ASC").Find(&attemptQuestions).Error; err != nil {
		tx.Rollback()
		return err
	}

	sections, err := ensureAttemptSections(tx, attempt, attemptQuestions)
	if err != nil {
		tx.Rollback()
		return err
	}

	sections, err = syncAttemptAndSections(tx, &attempt, sections, now)
	if err != nil {
		tx.Rollback()
		return err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return errors.New("waktu test sudah habis")
	}

	var existingAnswers []models.AttemptAnswer
	if err := tx.Where("attempt_id = ?", attemptID).Find(&existingAnswers).Error; err != nil {
		tx.Rollback()
		return err
	}

	answerMap := make(map[uint]string, len(existingAnswers))
	for _, answer := range existingAnswers {
		answerMap[answer.AttemptQuestionID] = answer.SelectedOptionKey
	}

	validOptions := make(map[uint]map[string]bool, len(attemptQuestions))
	questionLookup := make(map[uint]models.AttemptQuestion, len(attemptQuestions))
	for _, item := range attemptQuestions {
		options, err := models.UnmarshalOptionsSnapshot(item.OptionsSnapshot)
		if err != nil {
			tx.Rollback()
			return err
		}

		keys := make(map[string]bool, len(options))
		for _, option := range options {
			keys[option.Key] = true
		}

		validOptions[item.ID] = keys
		questionLookup[item.ID] = item
	}

	activeSection := activeAttemptSection(sections)
	if activeSection == nil {
		nextSection := nextPendingAttemptSection(sections)
		tx.Rollback()
		if nextSection != nil {
			return fmt.Errorf("mulai bagian %s terlebih dahulu", attemptSectionLabel(nextSection.QuestionIndex))
		}
		return errors.New("semua bagian sudah selesai, silakan submit final")
	}

	if err := validateSectionAnswers(activeSection, answers, questionLookup); err != nil {
		tx.Rollback()
		return err
	}

	records := make([]models.AttemptAnswer, 0, len(answers))
	for _, answer := range answers {
		question, ok := questionLookup[answer.AttemptQuestionID]
		if !ok {
			tx.Rollback()
			return errors.New("ada soal yang tidak valid")
		}

		if !validOptions[answer.AttemptQuestionID][answer.SelectedOptionKey] {
			tx.Rollback()
			return errors.New("opsi jawaban tidak valid")
		}

		records = append(records, models.AttemptAnswer{
			AttemptID:         attemptID,
			AttemptQuestionID: answer.AttemptQuestionID,
			QuestionID:        question.QuestionID,
			SelectedOptionKey: answer.SelectedOptionKey,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "attempt_id"}, {Name: "attempt_question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"selected_option_key", "updated_at"}),
	}).Create(&records).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *TestService) StartSection(userID uint, attemptID uint, sectionCode models.QuestionIndex) (*models.AttemptSection, error) {
	now := time.Now()
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var attempt models.Attempt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", attemptID, userID).
		First(&attempt).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("attempt tidak ditemukan")
	}
	if err := ensureAttemptAllowedForUser(tx, userID, attempt.TestType); err != nil {
		tx.Rollback()
		return nil, err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return nil, errors.New("attempt sudah tidak aktif")
	}

	var attemptQuestions []models.AttemptQuestion
	if err := tx.Where("attempt_id = ?", attemptID).Order("order_no ASC").Find(&attemptQuestions).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err := ensureAttemptSections(tx, attempt, attemptQuestions)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err = syncAttemptAndSections(tx, &attempt, sections, now)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return nil, errors.New("waktu test sudah habis")
	}

	targetSection := findAttemptSectionByCode(sections, sectionCode)
	if targetSection == nil {
		tx.Rollback()
		return nil, errors.New("bagian tidak ditemukan")
	}

	if targetSection.Status == models.AttemptSectionStatusSubmitted || targetSection.Status == models.AttemptSectionStatusExpired {
		tx.Rollback()
		return nil, fmt.Errorf("bagian %s sudah selesai", attemptSectionLabel(targetSection.QuestionIndex))
	}

	currentSection := nextPendingAttemptSection(sections)
	if currentSection == nil {
		tx.Rollback()
		return nil, errors.New("semua bagian sudah selesai")
	}

	if currentSection.QuestionIndex != sectionCode {
		tx.Rollback()
		return nil, fmt.Errorf("mulai bagian %s terlebih dahulu", attemptSectionLabel(currentSection.QuestionIndex))
	}

	if currentSection.Status == models.AttemptSectionStatusInProgress {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return currentSection, nil
	}

	expiresAt := now.Add(time.Duration(currentSection.DurationMinutes) * time.Minute)
	if expiresAt.After(attempt.ExpiresAt) {
		expiresAt = attempt.ExpiresAt
	}

	currentSection.Status = models.AttemptSectionStatusInProgress
	currentSection.StartedAt = &now
	currentSection.ExpiresAt = &expiresAt
	currentSection.SubmittedAt = nil

	if err := tx.Model(&models.AttemptSection{}).
		Where("id = ?", currentSection.ID).
		Updates(map[string]any{
			"status":       currentSection.Status,
			"started_at":   currentSection.StartedAt,
			"expires_at":   currentSection.ExpiresAt,
			"submitted_at": nil,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return currentSection, nil
}

func (s *TestService) SubmitSection(userID uint, attemptID uint, sectionCode models.QuestionIndex) (*models.AttemptSection, error) {
	now := time.Now()
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var attempt models.Attempt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", attemptID, userID).
		First(&attempt).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("attempt tidak ditemukan")
	}
	if err := ensureAttemptAllowedForUser(tx, userID, attempt.TestType); err != nil {
		tx.Rollback()
		return nil, err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return nil, errors.New("attempt sudah tidak aktif")
	}

	var attemptQuestions []models.AttemptQuestion
	if err := tx.Where("attempt_id = ?", attemptID).Order("order_no ASC").Find(&attemptQuestions).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err := ensureAttemptSections(tx, attempt, attemptQuestions)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err = syncAttemptAndSections(tx, &attempt, sections, now)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return nil, errors.New("waktu test sudah habis")
	}

	targetSection := findAttemptSectionByCode(sections, sectionCode)
	if targetSection == nil {
		tx.Rollback()
		return nil, errors.New("bagian tidak ditemukan")
	}

	if targetSection.Status == models.AttemptSectionStatusSubmitted || targetSection.Status == models.AttemptSectionStatusExpired {
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		return targetSection, nil
	}

	currentSection := nextPendingAttemptSection(sections)
	if currentSection == nil {
		tx.Rollback()
		return nil, errors.New("semua bagian sudah selesai")
	}

	if currentSection.QuestionIndex != sectionCode {
		tx.Rollback()
		return nil, fmt.Errorf("selesaikan bagian %s terlebih dahulu", attemptSectionLabel(currentSection.QuestionIndex))
	}

	if currentSection.Status != models.AttemptSectionStatusInProgress {
		tx.Rollback()
		return nil, fmt.Errorf("mulai bagian %s terlebih dahulu", attemptSectionLabel(currentSection.QuestionIndex))
	}

	currentSection.Status = models.AttemptSectionStatusSubmitted
	currentSection.SubmittedAt = &now

	if err := tx.Model(&models.AttemptSection{}).
		Where("id = ?", currentSection.ID).
		Updates(map[string]any{
			"status":       currentSection.Status,
			"submitted_at": currentSection.SubmittedAt,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return currentSection, nil
}

func (s *TestService) SubmitAttempt(userID uint, attemptID uint) (*models.Attempt, error) {
	now := time.Now()
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var attempt models.Attempt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", attemptID, userID).
		First(&attempt).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("attempt tidak ditemukan")
	}
	if err := ensureAttemptAllowedForUser(tx, userID, attempt.TestType); err != nil {
		tx.Rollback()
		return nil, err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return nil, errors.New("attempt sudah selesai")
	}

	var questions []models.AttemptQuestion
	if err := tx.Where("attempt_id = ?", attempt.ID).Order("order_no ASC").Find(&questions).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err := ensureAttemptSections(tx, attempt, questions)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	sections, err = syncAttemptAndSections(tx, &attempt, sections, now)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if attempt.Status != models.AttemptStatusInProgress {
		tx.Rollback()
		return nil, errors.New("waktu test sudah habis")
	}

	var answers []models.AttemptAnswer
	if err := tx.Where("attempt_id = ?", attempt.ID).Find(&answers).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	answerMap := make(map[uint]string, len(answers))
	for _, answer := range answers {
		answerMap[answer.AttemptQuestionID] = answer.SelectedOptionKey
	}

	if currentSection := activeAttemptSection(sections); currentSection != nil {
		tx.Rollback()
		return nil, fmt.Errorf("submit bagian %s terlebih dahulu", attemptSectionLabel(currentSection.QuestionIndex))
	}

	if nextSection := nextPendingAttemptSection(sections); nextSection != nil {
		tx.Rollback()
		return nil, fmt.Errorf("selesaikan bagian %s terlebih dahulu", attemptSectionLabel(nextSection.QuestionIndex))
	}

	if err := ensureDailySubmitQuota(tx, userID, attemptUserAccountType(tx, attempt.UserID), attempt.TestType, now); err != nil {
		tx.Rollback()
		return nil, err
	}

	rawScore, totalQuestions, percentage := ScoreAttempt(questions, answerMap)
	attempt.Status = models.AttemptStatusSubmitted
	attempt.SubmittedAt = &now
	attempt.RawScore = &rawScore
	attempt.TotalQuestions = totalQuestions
	attempt.Percentage = &percentage

	if err := tx.Save(&attempt).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &attempt, nil
}

func (s *TestService) GetLatestResult(userID uint, testType models.TestType, roomCode string) (*models.Attempt, error) {
	var attempt models.Attempt
	query := s.db.Where("user_id = ? AND test_type = ? AND status = ?", userID, testType, models.AttemptStatusSubmitted)
	if testType != models.TestTypeSKB || roomCode != "" {
		roomCodes := []string{roomCode}
		if testType == models.TestTypeSKB {
			roomCodes = MatchingRoomCodes(roomCode)
		}
		query = query.Where("room_code IN ?", roomCodes)
	}

	if err := query.Order("submitted_at DESC, id DESC").First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &attempt, nil
}

func (s *TestService) GetLatestResultDetail(userID uint, testType models.TestType, roomCode string) (*AttemptResultView, error) {
	attempt, err := s.GetLatestResult(userID, testType, roomCode)
	if err != nil || attempt == nil {
		return nil, err
	}

	return s.BuildAttemptResultView(*attempt)
}

func (s *TestService) ListResultDetails(userID uint, testType models.TestType, roomCode string) ([]AttemptResultView, error) {
	var attempts []models.Attempt
	query := s.db.Where("user_id = ? AND test_type = ? AND status = ?", userID, testType, models.AttemptStatusSubmitted)
	if testType != models.TestTypeSKB || roomCode != "" {
		roomCodes := []string{roomCode}
		if testType == models.TestTypeSKB {
			roomCodes = MatchingRoomCodes(roomCode)
		}
		query = query.Where("room_code IN ?", roomCodes)
	}

	if err := query.Order("submitted_at DESC, id DESC").Find(&attempts).Error; err != nil {
		return nil, err
	}

	results := make([]AttemptResultView, 0, len(attempts))
	for _, attempt := range attempts {
		item, err := s.BuildAttemptResultView(attempt)
		if err != nil {
			return nil, err
		}
		results = append(results, *item)
	}

	return results, nil
}

func (s *TestService) BuildAttemptResultView(attempt models.Attempt) (*AttemptResultView, error) {
	questions, answers, err := s.loadAttemptQuestionsAndAnswers(attempt.ID)
	if err != nil {
		return nil, err
	}

	answerMap := make(map[uint]string, len(answers))
	for _, answer := range answers {
		answerMap[answer.AttemptQuestionID] = answer.SelectedOptionKey
	}

	var estimatedIQ *int
	var skbScore *float64
	classificationLabel := ""
	if attempt.TestType == models.TestTypeIQ && attempt.Percentage != nil {
		profile := BuildScreeningIQProfile(*attempt.Percentage)
		estimatedIQ = &profile.EstimatedIQ
		classificationLabel = profile.ClassificationLabel
	} else if attempt.TestType == models.TestTypeSKB {
		if attempt.Percentage != nil {
			score := math.Round(*attempt.Percentage*100) / 100
			skbScore = &score
		}
		classificationLabel = fmt.Sprintf("SKB %s", attempt.RoomLabel)
	}

	return &AttemptResultView{
		Attempt:             attempt,
		EstimatedIQ:         estimatedIQ,
		SKBScore:            skbScore,
		ClassificationLabel: classificationLabel,
		IndexScores:         ScoreAttemptByIndex(questions, answerMap),
	}, nil
}

func ScoreAttempt(questions []models.AttemptQuestion, answers map[uint]string) (int, int, float64) {
	if len(questions) == 0 {
		return 0, 0, 0
	}

	rawScore := 0
	for _, question := range questions {
		if answers[question.ID] == question.CorrectOptionKey {
			rawScore++
		}
	}

	total := len(questions)
	percentage := math.Round((float64(rawScore)/float64(total))*10000) / 100
	return rawScore, total, percentage
}

func ScoreAttemptByIndex(questions []models.AttemptQuestion, answers map[uint]string) []IndexScoreView {
	type rawIndexScore struct {
		correct int
		total   int
	}

	aggregate := make(map[models.QuestionIndex]*rawIndexScore, len(orderedQuestionIndices))
	for _, question := range questions {
		if question.QuestionIndex == "" {
			continue
		}

		if _, ok := aggregate[question.QuestionIndex]; !ok {
			aggregate[question.QuestionIndex] = &rawIndexScore{}
		}

		aggregate[question.QuestionIndex].total++
		if answers[question.ID] == question.CorrectOptionKey {
			aggregate[question.QuestionIndex].correct++
		}
	}

	rows := make([]IndexScoreView, 0, len(aggregate))
	displayOrder := append(OrderedQuestionIndices(), models.QuestionIndexSKB)
	for _, questionIndex := range displayOrder {
		score, ok := aggregate[questionIndex]
		if !ok || score.total == 0 {
			continue
		}

		label := GetQuestionIndexLabel(questionIndex)
		if questionIndex == models.QuestionIndexSKB {
			label = "SKB"
		}
		percentage := math.Round((float64(score.correct)/float64(score.total))*10000) / 100
		rows = append(rows, IndexScoreView{
			Code:       questionIndex,
			Label:      label,
			Correct:    score.correct,
			Total:      score.total,
			Percentage: percentage,
		})
	}

	return rows
}

func expireAttempts(db *gorm.DB, userID uint, testType models.TestType, roomCode string, now time.Time) error {
	roomCodes := []string{roomCode}
	if testType == models.TestTypeSKB {
		roomCodes = MatchingRoomCodes(roomCode)
	}

	return db.Model(&models.Attempt{}).
		Where("user_id = ? AND test_type = ? AND room_code IN ? AND status = ? AND expires_at < ?", userID, testType, roomCodes, models.AttemptStatusInProgress, now).
		Update("status", models.AttemptStatusExpired).Error
}

func (s *TestService) loadAttemptQuestionsAndAnswers(attemptID uint) ([]models.AttemptQuestion, []models.AttemptAnswer, error) {
	var questions []models.AttemptQuestion
	if err := s.db.Where("attempt_id = ?", attemptID).Order("order_no ASC").Find(&questions).Error; err != nil {
		return nil, nil, err
	}

	var answers []models.AttemptAnswer
	if err := s.db.Where("attempt_id = ?", attemptID).Find(&answers).Error; err != nil {
		return nil, nil, err
	}

	return questions, answers, nil
}

func valueOrZero(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func validateAccountTestAccess(accountType models.AccountType, testType models.TestType) error {
	if hasPaidAccess(accountType) {
		return nil
	}
	if testType == models.TestTypeSKB {
		return errors.New("akun gratis belum bisa mengakses SKB, silakan upgrade ke akun Pro atau Max")
	}
	return nil
}

func resolveQuestionCountForAccount(accountType models.AccountType, testType models.TestType, configuredQuestionCount int) int {
	if canonicalAccountType(accountType) == models.AccountTypeFree && testType == models.TestTypeIQ {
		if configuredQuestionCount > 0 && configuredQuestionCount < freeIQQuestionLimit {
			return configuredQuestionCount
		}
		return freeIQQuestionLimit
	}

	switch testType {
	case models.TestTypeIQ:
		return iqTotalQuestionCount
	case models.TestTypeSKB:
		return skbQuestionCountPerAttempt
	default:
		return configuredQuestionCount
	}
}

func ensureDailySubmitQuota(tx *gorm.DB, userID uint, accountType models.AccountType, testType models.TestType, now time.Time) error {
	limit := dailySubmitLimit(accountType, testType)
	if limit <= 0 {
		return nil
	}

	startOfDay, endOfDay := jakartaDayRange(now)
	var submittedCount int64
	if err := tx.Model(&models.Attempt{}).
		Where("user_id = ? AND test_type = ? AND status = ? AND submitted_at >= ? AND submitted_at < ?", userID, testType, models.AttemptStatusSubmitted, startOfDay, endOfDay).
		Count(&submittedCount).Error; err != nil {
		return err
	}
	if submittedCount >= int64(limit) {
		return fmt.Errorf("batas submit harian untuk %s sudah tercapai (%d kali)", testType, limit)
	}
	return nil
}

func attemptUserAccountType(tx *gorm.DB, userID uint) models.AccountType {
	var user models.User
	if err := tx.Select("id, account_type").First(&user, userID).Error; err != nil {
		return models.AccountTypeFree
	}
	return user.AccountType
}

func jakartaDayRange(now time.Time) (time.Time, time.Time) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.FixedZone("WIB", 7*60*60)
	}

	localNow := now.In(location)
	start := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	return start, start.Add(24 * time.Hour)
}

func ensureAttemptAllowedForUser(tx *gorm.DB, userID uint, testType models.TestType) error {
	var user models.User
	if err := tx.Select("id, account_type").First(&user, userID).Error; err != nil {
		return errors.New("user tidak ditemukan")
	}
	return validateAccountTestAccess(user.AccountType, testType)
}

type attemptSectionSeed struct {
	Code          models.QuestionIndex
	OrderNo       int
	QuestionCount int
}

func buildAttemptQuestionViews(questions []models.AttemptQuestion, answerMap map[uint]string) ([]AttemptQuestionView, error) {
	items := make([]AttemptQuestionView, 0, len(questions))
	for _, question := range questions {
		options, err := models.UnmarshalOptionsSnapshot(question.OptionsSnapshot)
		if err != nil {
			return nil, err
		}

		var selected *string
		if value, ok := answerMap[question.ID]; ok {
			copyValue := value
			selected = &copyValue
		}

		items = append(items, AttemptQuestionView{
			ID:                 question.ID,
			QuestionID:         question.QuestionID,
			OrderNo:            question.OrderNo,
			QuestionIndex:      question.QuestionIndex,
			QuestionIndexLabel: GetQuestionIndexLabel(question.QuestionIndex),
			SubtestCode:        question.SubtestCode,
			SubtestLabel:       GetSubtestLabel(question.SubtestCode),
			Prompt:             question.PromptSnapshot,
			PromptMediaURL:     question.PromptMediaURL,
			PromptMediaAlt:     question.PromptMediaAlt,
			Options:            options,
			SelectedOptionKey:  selected,
		})
	}

	return items, nil
}

func buildAttemptSectionViews(
	sections []models.AttemptSection,
	questions []models.AttemptQuestion,
	answerMap map[uint]string,
) []AttemptSectionView {
	answeredCountByIndex := make(map[models.QuestionIndex]int, len(sections))
	questionCountByIndex := make(map[models.QuestionIndex]int, len(sections))
	skbLabel := "Tes SKB"
	for _, question := range questions {
		questionCountByIndex[question.QuestionIndex]++
		if answerMap[question.ID] != "" {
			answeredCountByIndex[question.QuestionIndex]++
		}
		if question.QuestionIndex == models.QuestionIndexSKB && question.SubtestCode != "" {
			skbLabel = GetSubtestLabel(question.SubtestCode)
		}
	}

	rows := make([]AttemptSectionView, 0, len(sections))
	for _, section := range sections {
		questionCount := section.QuestionCount
		if questionCount == 0 {
			questionCount = questionCountByIndex[section.QuestionIndex]
		}

		label := GetQuestionIndexLabel(section.QuestionIndex)
		if section.QuestionIndex == models.QuestionIndexSKB {
			label = skbLabel
		}

		rows = append(rows, AttemptSectionView{
			ID:              section.ID,
			Code:            section.QuestionIndex,
			Label:           label,
			OrderNo:         section.OrderNo,
			Status:          section.Status,
			DurationMinutes: section.DurationMinutes,
			QuestionCount:   questionCount,
			AnsweredCount:   answeredCountByIndex[section.QuestionIndex],
			StartedAt:       section.StartedAt,
			ExpiresAt:       section.ExpiresAt,
			SubmittedAt:     section.SubmittedAt,
		})
	}

	return rows
}

func ensureAttemptSections(
	tx *gorm.DB,
	attempt models.Attempt,
	questions []models.AttemptQuestion,
) ([]models.AttemptSection, error) {
	var sections []models.AttemptSection
	if err := tx.Where("attempt_id = ?", attempt.ID).Order("order_no ASC").Find(&sections).Error; err != nil {
		return nil, err
	}
	if len(sections) > 0 {
		return sections, nil
	}

	records := buildAttemptSectionRecordsFromAttemptQuestions(attempt.ID, questions, attempt.DurationMinutes)
	if len(records) == 0 {
		return nil, errors.New("bagian test tidak ditemukan")
	}

	if err := tx.Create(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

func buildAttemptSectionRecordsFromQuestions(
	attemptID uint,
	questions []models.Question,
	durationMinutes int,
	testType models.TestType,
) []models.AttemptSection {
	if testType == models.TestTypeSKB {
		return buildAttemptSectionRecords(attemptID, []attemptSectionSeed{{
			Code:          models.QuestionIndexSKB,
			OrderNo:       1,
			QuestionCount: len(questions),
		}}, durationMinutes)
	}

	seeds := make([]attemptSectionSeed, 0, len(orderedQuestionIndices))
	for _, question := range questions {
		if len(seeds) == 0 || seeds[len(seeds)-1].Code != question.QuestionIndex {
			seeds = append(seeds, attemptSectionSeed{
				Code:          question.QuestionIndex,
				OrderNo:       len(seeds) + 1,
				QuestionCount: 1,
			})
			continue
		}

		seeds[len(seeds)-1].QuestionCount++
	}

	return buildAttemptSectionRecords(attemptID, seeds, durationMinutes)
}

func buildAttemptSectionRecordsFromAttemptQuestions(
	attemptID uint,
	questions []models.AttemptQuestion,
	durationMinutes int,
) []models.AttemptSection {
	if len(questions) > 0 && questions[0].QuestionIndex == models.QuestionIndexSKB {
		return buildAttemptSectionRecords(attemptID, []attemptSectionSeed{{
			Code:          models.QuestionIndexSKB,
			OrderNo:       1,
			QuestionCount: len(questions),
		}}, durationMinutes)
	}

	seeds := make([]attemptSectionSeed, 0, len(orderedQuestionIndices))
	for _, question := range questions {
		if len(seeds) == 0 || seeds[len(seeds)-1].Code != question.QuestionIndex {
			seeds = append(seeds, attemptSectionSeed{
				Code:          question.QuestionIndex,
				OrderNo:       len(seeds) + 1,
				QuestionCount: 1,
			})
			continue
		}

		seeds[len(seeds)-1].QuestionCount++
	}

	return buildAttemptSectionRecords(attemptID, seeds, durationMinutes)
}

func buildAttemptSectionRecords(
	attemptID uint,
	seeds []attemptSectionSeed,
	durationMinutes int,
) []models.AttemptSection {
	durations := distributeSectionDurations(durationMinutes, len(seeds))
	records := make([]models.AttemptSection, 0, len(seeds))
	for index, seed := range seeds {
		records = append(records, models.AttemptSection{
			AttemptID:       attemptID,
			QuestionIndex:   seed.Code,
			OrderNo:         seed.OrderNo,
			Status:          models.AttemptSectionStatusNotStarted,
			DurationMinutes: durations[index],
			QuestionCount:   seed.QuestionCount,
		})
	}

	return records
}

func distributeSectionDurations(totalMinutes int, sectionCount int) []int {
	if sectionCount == 0 {
		return nil
	}

	base := 1
	remainder := 0
	if totalMinutes > 0 {
		base = totalMinutes / sectionCount
		remainder = totalMinutes % sectionCount
		if base == 0 {
			base = 1
			remainder = 0
		}
	}

	durations := make([]int, sectionCount)
	for index := range durations {
		durations[index] = base
		if remainder > 0 {
			durations[index]++
			remainder--
		}
	}

	return durations
}

func syncAttemptAndSections(
	tx *gorm.DB,
	attempt *models.Attempt,
	sections []models.AttemptSection,
	now time.Time,
) ([]models.AttemptSection, error) {
	attemptExpired := false
	if attempt.Status == models.AttemptStatusInProgress && now.After(attempt.ExpiresAt) {
		if err := tx.Model(&models.Attempt{}).Where("id = ?", attempt.ID).Update("status", models.AttemptStatusExpired).Error; err != nil {
			return nil, err
		}
		attempt.Status = models.AttemptStatusExpired
		attemptExpired = true
	}

	for index := range sections {
		if sections[index].Status != models.AttemptSectionStatusInProgress {
			continue
		}
		if !attemptExpired && (sections[index].ExpiresAt == nil || !now.After(*sections[index].ExpiresAt)) {
			continue
		}

		sections[index].Status = models.AttemptSectionStatusExpired
		sections[index].SubmittedAt = &now

		if err := tx.Model(&models.AttemptSection{}).
			Where("id = ?", sections[index].ID).
			Updates(map[string]any{
				"status":       sections[index].Status,
				"submitted_at": sections[index].SubmittedAt,
			}).Error; err != nil {
			return nil, err
		}
	}

	return sections, nil
}

func activeAttemptSection(sections []models.AttemptSection) *models.AttemptSection {
	for index := range sections {
		if sections[index].Status == models.AttemptSectionStatusInProgress {
			return &sections[index]
		}
	}
	return nil
}

func nextPendingAttemptSection(sections []models.AttemptSection) *models.AttemptSection {
	for index := range sections {
		if !isCompletedAttemptSectionStatus(sections[index].Status) {
			return &sections[index]
		}
	}
	return nil
}

func findAttemptSectionByCode(sections []models.AttemptSection, code models.QuestionIndex) *models.AttemptSection {
	for index := range sections {
		if sections[index].QuestionIndex == code {
			return &sections[index]
		}
	}
	return nil
}

func isCompletedAttemptSectionStatus(status models.AttemptSectionStatus) bool {
	return status == models.AttemptSectionStatusSubmitted || status == models.AttemptSectionStatusExpired
}

func validateSectionAnswers(
	activeSection *models.AttemptSection,
	incoming []SaveAnswerInput,
	questionLookup map[uint]models.AttemptQuestion,
) error {
	for _, answer := range incoming {
		question, ok := questionLookup[answer.AttemptQuestionID]
		if !ok {
			return errors.New("ada soal yang tidak valid")
		}

		if question.QuestionIndex != activeSection.QuestionIndex {
			return fmt.Errorf("selesaikan bagian %s terlebih dahulu", attemptSectionLabel(activeSection.QuestionIndex))
		}
	}

	return nil
}

func attemptSectionLabel(index models.QuestionIndex) string {
	if label := GetQuestionIndexLabel(index); label != "" {
		return label
	}
	return "saat ini"
}

func sortQuestionsForAttempt(questions []models.Question) {
	sort.SliceStable(questions, func(i int, j int) bool {
		leftRank := questionIndexRank(questions[i].QuestionIndex)
		rightRank := questionIndexRank(questions[j].QuestionIndex)
		return leftRank < rightRank
	})
}

func questionIndexRank(index models.QuestionIndex) int {
	for position, item := range OrderedQuestionIndices() {
		if item == index {
			return position
		}
	}
	if index == models.QuestionIndexSKB {
		return len(orderedQuestionIndices)
	}
	return len(orderedQuestionIndices) + 1
}
