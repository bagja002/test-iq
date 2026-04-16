package services

import (
	"testing"

	"test-iq-ku/apps/api/internal/models"
)

func TestValidateQuestionPayloadAllowsMediaOnlyOption(t *testing.T) {
	payload := QuestionPayload{
		Prompt:        "Pilih gambar yang identik dengan target.",
		QuestionIndex: models.QuestionIndexPRI,
		SubtestCode:   "BLOCK_DESIGN",
		Status:        models.QuestionStatusPublished,
		Options: []QuestionOptionInput{
			{Key: "A", MediaURL: "/question-assets/pri/block-design/option-a.svg", IsCorrect: true},
			{Key: "B", MediaURL: "/question-assets/pri/block-design/option-b.svg"},
		},
	}

	if err := validateQuestionPayload(payload); err != nil {
		t.Fatalf("expected payload to be valid, got error: %v", err)
	}
}

func TestValidateQuestionPayloadRejectsEmptyOption(t *testing.T) {
	payload := QuestionPayload{
		Prompt:        "Pilih jawaban terbaik.",
		QuestionIndex: models.QuestionIndexVCI,
		SubtestCode:   "SIMILARITIES",
		Status:        models.QuestionStatusPublished,
		Options: []QuestionOptionInput{
			{Key: "A", Content: "", MediaURL: "", IsCorrect: true},
			{Key: "B", Content: "Pilihan B"},
		},
	}

	if err := validateQuestionPayload(payload); err == nil {
		t.Fatal("expected payload to be rejected when option has no content and no media")
	}
}
