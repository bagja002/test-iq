package services

import (
	"errors"
	"strings"

	"test-iq-ku/apps/api/internal/models"
)

type SubtestDefinition struct {
	Code          string
	Label         string
	QuestionIndex models.QuestionIndex
}

var orderedQuestionIndices = []models.QuestionIndex{
	models.QuestionIndexVCI,
	models.QuestionIndexPRI,
	models.QuestionIndexWMI,
	models.QuestionIndexPSI,
}

var questionIndexLabels = map[models.QuestionIndex]string{
	models.QuestionIndexVCI: "Verbal Comprehension Index",
	models.QuestionIndexPRI: "Perceptual Reasoning Index",
	models.QuestionIndexWMI: "Working Memory Index",
	models.QuestionIndexPSI: "Processing Speed Index",
}

var subtestCatalog = []SubtestDefinition{
	{Code: "SIMILARITIES", Label: "Similarities", QuestionIndex: models.QuestionIndexVCI},
	{Code: "VOCABULARY", Label: "Vocabulary", QuestionIndex: models.QuestionIndexVCI},
	{Code: "INFORMATION", Label: "Information", QuestionIndex: models.QuestionIndexVCI},
	{Code: "COMPREHENSION", Label: "Comprehension", QuestionIndex: models.QuestionIndexVCI},
	{Code: "BLOCK_DESIGN", Label: "Block Design", QuestionIndex: models.QuestionIndexPRI},
	{Code: "MATRIX_REASONING", Label: "Matrix Reasoning", QuestionIndex: models.QuestionIndexPRI},
	{Code: "VISUAL_PUZZLES", Label: "Visual Puzzles", QuestionIndex: models.QuestionIndexPRI},
	{Code: "PICTURE_COMPLETION", Label: "Picture Completion", QuestionIndex: models.QuestionIndexPRI},
	{Code: "FIGURE_WEIGHTS", Label: "Figure Weights", QuestionIndex: models.QuestionIndexPRI},
	{Code: "DIGIT_SPAN", Label: "Digit Span", QuestionIndex: models.QuestionIndexWMI},
	{Code: "ARITHMETIC", Label: "Arithmetic", QuestionIndex: models.QuestionIndexWMI},
	{Code: "LETTER_NUMBER_SEQUENCING", Label: "Letter-Number Sequencing", QuestionIndex: models.QuestionIndexWMI},
	{Code: "SYMBOL_SEARCH", Label: "Symbol Search", QuestionIndex: models.QuestionIndexPSI},
	{Code: "CODING", Label: "Coding", QuestionIndex: models.QuestionIndexPSI},
	{Code: "CANCELLATION", Label: "Cancellation", QuestionIndex: models.QuestionIndexPSI},
}

var subtestByCode = func() map[string]SubtestDefinition {
	items := make(map[string]SubtestDefinition, len(subtestCatalog))
	for _, item := range subtestCatalog {
		items[item.Code] = item
	}
	return items
}()

func OrderedQuestionIndices() []models.QuestionIndex {
	items := make([]models.QuestionIndex, len(orderedQuestionIndices))
	copy(items, orderedQuestionIndices)
	return items
}

func NormalizeQuestionIndex(value models.QuestionIndex) (models.QuestionIndex, error) {
	normalized := models.QuestionIndex(strings.ToUpper(strings.TrimSpace(string(value))))
	switch normalized {
	case models.QuestionIndexVCI, models.QuestionIndexPRI, models.QuestionIndexWMI, models.QuestionIndexPSI:
		return normalized, nil
	default:
		return "", errors.New("question index tidak valid")
	}
}

func NormalizeSubtestCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func ValidateQuestionTaxonomy(index models.QuestionIndex, subtestCode string) (models.QuestionIndex, string, error) {
	normalizedIndex, err := NormalizeQuestionIndex(index)
	if err != nil {
		return "", "", err
	}

	normalizedSubtestCode := NormalizeSubtestCode(subtestCode)
	definition, ok := subtestByCode[normalizedSubtestCode]
	if !ok {
		return "", "", errors.New("subtes tidak valid")
	}

	if definition.QuestionIndex != normalizedIndex {
		return "", "", errors.New("subtes tidak sesuai dengan index yang dipilih")
	}

	return normalizedIndex, normalizedSubtestCode, nil
}

func GetQuestionIndexLabel(index models.QuestionIndex) string {
	if label, ok := questionIndexLabels[index]; ok {
		return label
	}
	return "General Cognitive Screening"
}

func GetSubtestDefinition(code string) (SubtestDefinition, bool) {
	definition, ok := subtestByCode[NormalizeSubtestCode(code)]
	return definition, ok
}

func GetSubtestLabel(code string) string {
	if definition, ok := GetSubtestDefinition(code); ok {
		return definition.Label
	}
	return "Unmapped Subtest"
}
