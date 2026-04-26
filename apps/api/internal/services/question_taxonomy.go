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

var orderedSKBRoomCodes = []string{
	"MANAJER_KOPERASI_KDMP",
	"MANAGER_OPERASIONAL_KNMP",
	"KEPALA_PRODUKSI",
	"PENGELOLA_KEUANGAN",
	"PENJAMIN_MUTU",
}

var skbRoomCodeAliases = map[string]string{
	"MANAGER_OPERASIONAL": "MANAGER_OPERASIONAL_KNMP",
	"KEPALA_KOPERASI":     "MANAJER_KOPERASI_KDMP",
}

var questionIndexLabels = map[models.QuestionIndex]string{
	models.QuestionIndexVCI: "Verbal Comprehension Index",
	models.QuestionIndexPRI: "Perceptual Reasoning Index",
	models.QuestionIndexWMI: "Working Memory Index",
	models.QuestionIndexPSI: "Processing Speed Index",
	models.QuestionIndexSKB: "Tes SKB",
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
	{Code: "MANAJER_KOPERASI_KDMP", Label: "Manajer Kopreasi (KDMP)", QuestionIndex: models.QuestionIndexSKB},
	{Code: "MANAGER_OPERASIONAL_KNMP", Label: "Manager Operesial (KNMP)", QuestionIndex: models.QuestionIndexSKB},
	{Code: "KEPALA_PRODUKSI", Label: "Kepala Produksi (KNMP)", QuestionIndex: models.QuestionIndexSKB},
	{Code: "PENGELOLA_KEUANGAN", Label: "Pengelola Keuangan (KNMP)", QuestionIndex: models.QuestionIndexSKB},
	{Code: "PENJAMIN_MUTU", Label: "Penjamin Mutu (KNMP)", QuestionIndex: models.QuestionIndexSKB},
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
	case models.QuestionIndexVCI, models.QuestionIndexPRI, models.QuestionIndexWMI, models.QuestionIndexPSI, models.QuestionIndexSKB:
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

func OrderedSKBRoomCodes() []string {
	items := make([]string, len(orderedSKBRoomCodes))
	copy(items, orderedSKBRoomCodes)
	return items
}

func NormalizeTestType(value models.TestType) (models.TestType, error) {
	normalized := models.TestType(strings.ToUpper(strings.TrimSpace(string(value))))
	switch normalized {
	case models.TestTypeIQ, models.TestTypeSKB:
		return normalized, nil
	default:
		return "", errors.New("jenis test tidak valid")
	}
}

func NormalizeRoomCode(value string) (string, error) {
	normalized := NormalizeSubtestCode(value)
	if normalized == "" {
		return "", nil
	}
	if alias, ok := skbRoomCodeAliases[normalized]; ok {
		normalized = alias
	}
	definition, ok := subtestByCode[normalized]
	if !ok || definition.QuestionIndex != models.QuestionIndexSKB {
		return "", errors.New("room SKB tidak valid")
	}
	return normalized, nil
}

func MatchingRoomCodes(value string) []string {
	normalized, err := NormalizeRoomCode(value)
	if err != nil || normalized == "" {
		return nil
	}

	items := []string{normalized}
	for legacyCode, canonicalCode := range skbRoomCodeAliases {
		if canonicalCode == normalized {
			items = append(items, legacyCode)
		}
	}

	return items
}

func GetRoomLabel(code string) string {
	return GetSubtestLabel(code)
}

func GetSubtestLabel(code string) string {
	if definition, ok := GetSubtestDefinition(code); ok {
		return definition.Label
	}
	return "Unmapped Subtest"
}
