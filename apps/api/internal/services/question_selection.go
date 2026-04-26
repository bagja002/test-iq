package services

import (
	"errors"
	"fmt"

	"test-iq-ku/apps/api/internal/models"
)

type QuestionSelectionPlanRow struct {
	Code      models.QuestionIndex
	Requested int
	Available int
}

func buildQuestionSelectionPlan(totalQuestions int, available map[models.QuestionIndex]int) ([]QuestionSelectionPlanRow, error) {
	indices := OrderedQuestionIndices()
	requiredDistinct := len(indices)
	if totalQuestions < requiredDistinct {
		requiredDistinct = totalQuestions
	}

	totalAvailable := 0
	rows := make([]QuestionSelectionPlanRow, 0, len(indices))
	distinctCount := 0
	for _, code := range indices {
		count := available[code]
		totalAvailable += count
		if count == 0 && totalQuestions >= len(indices) {
			return nil, fmt.Errorf("bank soal index %s belum memiliki soal published", attemptSectionLabel(code))
		}
		if count == 0 {
			continue
		}

		distinctCount++

		rows = append(rows, QuestionSelectionPlanRow{
			Code:      code,
			Requested: 0,
			Available: count,
		})
	}

	if totalAvailable < totalQuestions {
		return nil, errors.New("jumlah soal published belum mencukupi untuk konfigurasi test aktif")
	}

	if distinctCount < requiredDistinct {
		return nil, fmt.Errorf("jumlah index dengan soal published minimal %d untuk konfigurasi ini", requiredDistinct)
	}

	for index := range rows {
		if index >= requiredDistinct {
			break
		}
		rows[index].Requested = 1
	}

	remaining := totalQuestions - requiredDistinct
	for remaining > 0 {
		progressed := false
		for index := range rows {
			if rows[index].Requested == 0 {
				continue
			}
			if rows[index].Requested >= rows[index].Available {
				continue
			}
			rows[index].Requested++
			remaining--
			progressed = true
			if remaining == 0 {
				break
			}
		}

		if !progressed {
			return nil, errors.New("bank soal published belum mencukupi untuk distribusi per index")
		}
	}

	return rows, nil
}
