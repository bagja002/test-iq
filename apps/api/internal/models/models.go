package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

type Role string
type AttemptStatus string
type AttemptSectionStatus string
type QuestionStatus string
type UserStatus string
type QuestionIndex string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"

	QuestionIndexVCI QuestionIndex = "VCI"
	QuestionIndexPRI QuestionIndex = "PRI"
	QuestionIndexWMI QuestionIndex = "WMI"
	QuestionIndexPSI QuestionIndex = "PSI"

	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"

	AttemptStatusInProgress AttemptStatus = "IN_PROGRESS"
	AttemptStatusSubmitted  AttemptStatus = "SUBMITTED"
	AttemptStatusExpired    AttemptStatus = "EXPIRED"

	AttemptSectionStatusNotStarted AttemptSectionStatus = "NOT_STARTED"
	AttemptSectionStatusInProgress AttemptSectionStatus = "IN_PROGRESS"
	AttemptSectionStatusSubmitted  AttemptSectionStatus = "SUBMITTED"
	AttemptSectionStatusExpired    AttemptSectionStatus = "EXPIRED"

	QuestionStatusDraft     QuestionStatus = "DRAFT"
	QuestionStatusPublished QuestionStatus = "PUBLISHED"
	QuestionStatusArchived  QuestionStatus = "ARCHIVED"
)

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:120;not null" json:"name"`
	Email        string     `gorm:"size:191;not null;uniqueIndex" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         Role       `gorm:"size:20;not null;index" json:"role"`
	Status       UserStatus `gorm:"size:20;not null;index" json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type Question struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Prompt         string         `gorm:"type:text;not null" json:"prompt"`
	PromptMediaURL string         `gorm:"size:255" json:"promptMediaUrl"`
	PromptMediaAlt string         `gorm:"size:255" json:"promptMediaAlt"`
	Difficulty     string         `gorm:"size:40;not null;default:'medium'" json:"difficulty"`
	QuestionIndex  QuestionIndex  `gorm:"size:16;index" json:"questionIndex"`
	SubtestCode    string         `gorm:"size:64;index" json:"subtestCode"`
	Status         QuestionStatus `gorm:"size:20;not null;index" json:"status"`
	Options        []QuestionOption
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type QuestionOption struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	QuestionID uint      `gorm:"not null;index" json:"questionId"`
	Key        string    `gorm:"size:8;not null" json:"key"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	MediaURL   string    `gorm:"size:255" json:"mediaUrl"`
	MediaAlt   string    `gorm:"size:255" json:"mediaAlt"`
	IsCorrect  bool      `gorm:"not null;default:false" json:"isCorrect"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type TestConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:191;not null" json:"title"`
	DurationMinutes int       `gorm:"not null" json:"durationMinutes"`
	QuestionCount   int       `gorm:"not null" json:"questionCount"`
	IsActive        bool      `gorm:"not null;default:true;index" json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Attempt struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null;index:idx_attempt_user_status,priority:1" json:"userId"`
	TestConfigID    uint          `gorm:"not null;index:idx_attempt_config_status,priority:1" json:"testConfigId"`
	Status          AttemptStatus `gorm:"size:20;not null;index:idx_attempt_user_status,priority:2;index:idx_attempt_config_status,priority:2" json:"status"`
	StartedAt       time.Time     `gorm:"not null" json:"startedAt"`
	ExpiresAt       time.Time     `gorm:"not null" json:"expiresAt"`
	SubmittedAt     *time.Time    `json:"submittedAt"`
	RawScore        *int          `json:"rawScore"`
	TotalQuestions  int           `gorm:"not null" json:"totalQuestions"`
	Percentage      *float64      `json:"percentage"`
	DurationMinutes int           `gorm:"not null" json:"durationMinutes"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type AttemptQuestion struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	AttemptID        uint           `gorm:"not null;index:idx_attempt_question_order,priority:1" json:"attemptId"`
	QuestionID       uint           `gorm:"not null" json:"questionId"`
	QuestionIndex    QuestionIndex  `gorm:"size:16;index" json:"questionIndex"`
	SubtestCode      string         `gorm:"size:64;index" json:"subtestCode"`
	OrderNo          int            `gorm:"not null;index:idx_attempt_question_order,priority:2" json:"orderNo"`
	PromptSnapshot   string         `gorm:"type:text;not null" json:"promptSnapshot"`
	PromptMediaURL   string         `gorm:"size:255" json:"promptMediaUrl"`
	PromptMediaAlt   string         `gorm:"size:255" json:"promptMediaAlt"`
	OptionsSnapshot  datatypes.JSON `gorm:"type:json;not null" json:"optionsSnapshot"`
	CorrectOptionKey string         `gorm:"size:8;not null" json:"-"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

type AttemptSection struct {
	ID              uint                 `gorm:"primaryKey" json:"id"`
	AttemptID       uint                 `gorm:"not null;uniqueIndex:uidx_attempt_section_code,priority:1;index:idx_attempt_section_order,priority:1" json:"attemptId"`
	QuestionIndex   QuestionIndex        `gorm:"size:16;not null;uniqueIndex:uidx_attempt_section_code,priority:2" json:"questionIndex"`
	OrderNo         int                  `gorm:"not null;index:idx_attempt_section_order,priority:2" json:"orderNo"`
	Status          AttemptSectionStatus `gorm:"size:20;not null;index" json:"status"`
	DurationMinutes int                  `gorm:"not null" json:"durationMinutes"`
	QuestionCount   int                  `gorm:"not null" json:"questionCount"`
	StartedAt       *time.Time           `json:"startedAt"`
	ExpiresAt       *time.Time           `json:"expiresAt"`
	SubmittedAt     *time.Time           `json:"submittedAt"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

type AttemptAnswer struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	AttemptID         uint      `gorm:"not null;uniqueIndex:uidx_attempt_question_answer,priority:1;index:idx_attempt_answer_lookup,priority:1" json:"attemptId"`
	AttemptQuestionID uint      `gorm:"not null;uniqueIndex:uidx_attempt_question_answer,priority:2" json:"attemptQuestionId"`
	QuestionID        uint      `gorm:"not null;index:idx_attempt_answer_lookup,priority:2" json:"questionId"`
	SelectedOptionKey string    `gorm:"size:8;not null" json:"selectedOptionKey"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	TokenHash string    `gorm:"size:64;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OptionSnapshot struct {
	Key      string `json:"key"`
	Content  string `json:"content"`
	MediaURL string `json:"mediaUrl"`
	MediaAlt string `json:"mediaAlt"`
}

func MarshalOptionsSnapshot(options []OptionSnapshot) (datatypes.JSON, error) {
	raw, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(raw), nil
}

func UnmarshalOptionsSnapshot(value datatypes.JSON) ([]OptionSnapshot, error) {
	if len(value) == 0 {
		return []OptionSnapshot{}, nil
	}

	var options []OptionSnapshot
	if err := json.Unmarshal(value, &options); err != nil {
		return nil, err
	}

	return options, nil
}
