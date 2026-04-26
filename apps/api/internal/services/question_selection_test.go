package services

import (
	"testing"

	"test-iq-ku/apps/api/internal/models"
)

func TestBuildQuestionSelectionPlanGuaranteesAllIndices(t *testing.T) {
	plan, err := buildQuestionSelectionPlan(10, map[models.QuestionIndex]int{
		models.QuestionIndexVCI: 5,
		models.QuestionIndexPRI: 5,
		models.QuestionIndexWMI: 3,
		models.QuestionIndexPSI: 3,
	})
	if err != nil {
		t.Fatalf("expected valid plan, got %v", err)
	}

	expected := map[models.QuestionIndex]int{
		models.QuestionIndexVCI: 3,
		models.QuestionIndexPRI: 3,
		models.QuestionIndexWMI: 2,
		models.QuestionIndexPSI: 2,
	}

	total := 0
	for _, row := range plan {
		total += row.Requested
		if row.Requested != expected[row.Code] {
			t.Fatalf("expected %s to request %d questions, got %d", row.Code, expected[row.Code], row.Requested)
		}
	}

	if total != 10 {
		t.Fatalf("expected total selected questions to be 10, got %d", total)
	}
}

func TestBuildQuestionSelectionPlanRejectsMissingIndex(t *testing.T) {
	_, err := buildQuestionSelectionPlan(8, map[models.QuestionIndex]int{
		models.QuestionIndexVCI: 4,
		models.QuestionIndexPRI: 4,
		models.QuestionIndexWMI: 0,
		models.QuestionIndexPSI: 2,
	})
	if err == nil {
		t.Fatal("expected plan to reject missing published questions for one index")
	}
}

func TestBuildQuestionSelectionPlanSupportsPreviewQuestionCounts(t *testing.T) {
	plan, err := buildQuestionSelectionPlan(2, map[models.QuestionIndex]int{
		models.QuestionIndexVCI: 4,
		models.QuestionIndexPRI: 4,
		models.QuestionIndexWMI: 4,
		models.QuestionIndexPSI: 4,
	})
	if err != nil {
		t.Fatalf("expected preview question count to be supported, got %v", err)
	}
	if len(plan) < 2 {
		t.Fatalf("expected at least 2 plan rows, got %d", len(plan))
	}
	if plan[0].Code != models.QuestionIndexVCI || plan[0].Requested != 1 {
		t.Fatalf("unexpected first preview row: %+v", plan[0])
	}
	if plan[1].Code != models.QuestionIndexPRI || plan[1].Requested != 1 {
		t.Fatalf("unexpected second preview row: %+v", plan[1])
	}
}
