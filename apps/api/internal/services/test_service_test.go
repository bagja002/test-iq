package services

import (
	"testing"

	"test-iq-ku/apps/api/internal/models"
)

func TestScoreAttempt(t *testing.T) {
	questions := []models.AttemptQuestion{
		{ID: 1, CorrectOptionKey: "A"},
		{ID: 2, CorrectOptionKey: "C"},
		{ID: 3, CorrectOptionKey: "B"},
	}

	answers := map[uint]string{
		1: "A",
		2: "D",
		3: "B",
	}

	rawScore, total, percentage := ScoreAttempt(questions, answers)
	if rawScore != 2 {
		t.Fatalf("expected rawScore=2, got %d", rawScore)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if percentage != 66.67 {
		t.Fatalf("expected percentage=66.67, got %.2f", percentage)
	}
}

func TestScoreAttemptByIndex(t *testing.T) {
	questions := []models.AttemptQuestion{
		{ID: 1, QuestionIndex: models.QuestionIndexVCI, CorrectOptionKey: "A"},
		{ID: 2, QuestionIndex: models.QuestionIndexVCI, CorrectOptionKey: "B"},
		{ID: 3, QuestionIndex: models.QuestionIndexPRI, CorrectOptionKey: "C"},
		{ID: 4, QuestionIndex: models.QuestionIndexWMI, CorrectOptionKey: "D"},
	}

	answers := map[uint]string{
		1: "A",
		2: "C",
		3: "C",
		4: "A",
	}

	indexScores := ScoreAttemptByIndex(questions, answers)
	if len(indexScores) != 3 {
		t.Fatalf("expected 3 index scores, got %d", len(indexScores))
	}

	if indexScores[0].Code != models.QuestionIndexVCI || indexScores[0].Correct != 1 || indexScores[0].Total != 2 || indexScores[0].Percentage != 50 {
		t.Fatalf("unexpected VCI score: %+v", indexScores[0])
	}

	if indexScores[1].Code != models.QuestionIndexPRI || indexScores[1].Correct != 1 || indexScores[1].Total != 1 || indexScores[1].Percentage != 100 {
		t.Fatalf("unexpected PRI score: %+v", indexScores[1])
	}

	if indexScores[2].Code != models.QuestionIndexWMI || indexScores[2].Correct != 0 || indexScores[2].Total != 1 || indexScores[2].Percentage != 0 {
		t.Fatalf("unexpected WMI score: %+v", indexScores[2])
	}
}

func TestEstimateScreeningIQ(t *testing.T) {
	testCases := []struct {
		name       string
		percentage float64
		expectedIQ int
	}{
		{name: "average band", percentage: 50, expectedIQ: 100},
		{name: "higher band", percentage: 80, expectedIQ: 113},
		{name: "lower band", percentage: 20, expectedIQ: 87},
		{name: "ceiling clamp", percentage: 100, expectedIQ: 139},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if actual := EstimateScreeningIQ(testCase.percentage); actual != testCase.expectedIQ {
				t.Fatalf("expected IQ=%d, got %d", testCase.expectedIQ, actual)
			}
		})
	}
}

func TestClassifyEstimatedIQ(t *testing.T) {
	testCases := []struct {
		estimatedIQ int
		expected    string
	}{
		{estimatedIQ: 78, expected: "Di bawah rata-rata"},
		{estimatedIQ: 84, expected: "Rata-rata bawah"},
		{estimatedIQ: 100, expected: "Rata-rata"},
		{estimatedIQ: 115, expected: "Di atas rata-rata"},
		{estimatedIQ: 124, expected: "Superior"},
		{estimatedIQ: 133, expected: "Sangat superior"},
	}

	for _, testCase := range testCases {
		if actual := ClassifyEstimatedIQ(testCase.estimatedIQ); actual != testCase.expected {
			t.Fatalf("expected classification %q, got %q", testCase.expected, actual)
		}
	}
}

func TestSortQuestionsForAttempt(t *testing.T) {
	questions := []models.Question{
		{ID: 1, QuestionIndex: models.QuestionIndexPSI},
		{ID: 2, QuestionIndex: models.QuestionIndexVCI},
		{ID: 3, QuestionIndex: models.QuestionIndexWMI},
		{ID: 4, QuestionIndex: models.QuestionIndexPRI},
		{ID: 5, QuestionIndex: ""},
	}

	sortQuestionsForAttempt(questions)

	expectedOrder := []uint{2, 4, 3, 1, 5}
	for index, expectedID := range expectedOrder {
		if questions[index].ID != expectedID {
			t.Fatalf("expected question at position %d to have ID %d, got %d", index, expectedID, questions[index].ID)
		}
	}
}

func TestDistributeSectionDurations(t *testing.T) {
	durations := distributeSectionDurations(35, 4)
	expected := []int{9, 9, 9, 8}

	for index, expectedValue := range expected {
		if durations[index] != expectedValue {
			t.Fatalf("expected duration at index %d to be %d, got %d", index, expectedValue, durations[index])
		}
	}
}

func TestNextPendingAttemptSection(t *testing.T) {
	sections := []models.AttemptSection{
		{ID: 1, QuestionIndex: models.QuestionIndexVCI, Status: models.AttemptSectionStatusSubmitted},
		{ID: 2, QuestionIndex: models.QuestionIndexPRI, Status: models.AttemptSectionStatusInProgress},
		{ID: 3, QuestionIndex: models.QuestionIndexWMI, Status: models.AttemptSectionStatusNotStarted},
	}

	current := nextPendingAttemptSection(sections)
	if current == nil || current.QuestionIndex != models.QuestionIndexPRI {
		t.Fatalf("expected PRI as next pending section, got %+v", current)
	}

	active := activeAttemptSection(sections)
	if active == nil || active.QuestionIndex != models.QuestionIndexPRI {
		t.Fatalf("expected PRI as active section, got %+v", active)
	}
}

func TestValidateSectionAnswers(t *testing.T) {
	activeSection := &models.AttemptSection{
		ID:            1,
		QuestionIndex: models.QuestionIndexVCI,
		Status:        models.AttemptSectionStatusInProgress,
	}
	questionLookup := map[uint]models.AttemptQuestion{
		1: {ID: 1, QuestionIndex: models.QuestionIndexVCI},
		2: {ID: 2, QuestionIndex: models.QuestionIndexPRI},
	}

	if err := validateSectionAnswers(
		activeSection,
		[]SaveAnswerInput{{AttemptQuestionID: 1, SelectedOptionKey: "A"}},
		questionLookup,
	); err != nil {
		t.Fatalf("expected active section answer to be accepted, got %v", err)
	}

	if err := validateSectionAnswers(
		activeSection,
		[]SaveAnswerInput{{AttemptQuestionID: 2, SelectedOptionKey: "B"}},
		questionLookup,
	); err == nil {
		t.Fatal("expected future section answer to be rejected")
	}
}

func TestBuildAttemptSectionRecordsFromAttemptQuestions(t *testing.T) {
	questions := []models.AttemptQuestion{
		{ID: 1, OrderNo: 1, QuestionIndex: models.QuestionIndexVCI},
		{ID: 2, OrderNo: 2, QuestionIndex: models.QuestionIndexVCI},
		{ID: 3, OrderNo: 3, QuestionIndex: models.QuestionIndexPRI},
		{ID: 4, OrderNo: 4, QuestionIndex: models.QuestionIndexWMI},
		{ID: 5, OrderNo: 5, QuestionIndex: models.QuestionIndexPSI},
	}

	records := buildAttemptSectionRecordsFromAttemptQuestions(99, questions, 40)
	if len(records) != 4 {
		t.Fatalf("expected 4 section records, got %d", len(records))
	}

	if records[0].QuestionIndex != models.QuestionIndexVCI || records[0].QuestionCount != 2 {
		t.Fatalf("unexpected first section record: %+v", records[0])
	}

	if records[1].QuestionIndex != models.QuestionIndexPRI || records[1].DurationMinutes != 10 {
		t.Fatalf("unexpected second section record: %+v", records[1])
	}

	if records[3].QuestionIndex != models.QuestionIndexPSI || records[3].OrderNo != 4 {
		t.Fatalf("unexpected last section record: %+v", records[3])
	}
}

func TestResolveQuestionCountForAccount(t *testing.T) {
	if total := resolveQuestionCountForAccount(models.AccountTypeFree, models.TestTypeIQ, 20); total != 20 {
		t.Fatalf("expected free IQ to use configured full test count, got %d", total)
	}
	if total := resolveQuestionCountForAccount(models.AccountTypeMax, models.TestTypeIQ, 130); total != 130 {
		t.Fatalf("expected paid IQ to use configured question count, got %d", total)
	}
	if total := resolveQuestionCountForAccount(models.AccountTypeFree, models.TestTypeSKB, 3); total != 3 {
		t.Fatalf("expected SKB to use configured question count, got %d", total)
	}
	if total := resolveQuestionCountForAccount(models.AccountTypeMax, models.TestTypeSKB, 50); total != 50 {
		t.Fatalf("expected paid SKB to use configured question count, got %d", total)
	}
}

func TestValidateAccountTestAccess(t *testing.T) {
	if err := validateAccountTestAccess(models.AccountTypeFree, models.TestTypeIQ); err != nil {
		t.Fatalf("expected free account to access IQ, got %v", err)
	}
	if err := validateAccountTestAccess(models.AccountTypeFree, models.TestTypeSKB); err == nil {
		t.Fatal("expected free account to be blocked from SKB")
	}
	if err := validateAccountTestAccess(models.AccountTypeMax, models.TestTypeSKB); err != nil {
		t.Fatalf("expected paid account to access SKB, got %v", err)
	}
}
