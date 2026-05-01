package services

import (
	"errors"
	"fmt"

	"test-iq-ku/apps/api/internal/models"
)

type QuestionSelectionPlanRow struct {
	Code            models.QuestionIndex
	Label           string
	OrderNo         int
	Requested       int
	Available       int
	DurationMinutes int
}

type IQSectionRule struct {
	Code            models.QuestionIndex
	Label           string
	OrderNo         int
	QuestionCount   int
	DurationMinutes int
}

func defaultIQSectionRules() []IQSectionRule {
	return []IQSectionRule{
		{
			Code:            models.QuestionIndexVCI,
			Label:           "Verbal Reasoning Index",
			OrderNo:         1,
			QuestionCount:   50,
			DurationMinutes: 40,
		},
		{
			Code:            models.QuestionIndexPRI,
			Label:           GetQuestionIndexLabel(models.QuestionIndexPRI),
			OrderNo:         2,
			QuestionCount:   20,
			DurationMinutes: 15,
		},
		{
			Code:            models.QuestionIndexWMI,
			Label:           GetQuestionIndexLabel(models.QuestionIndexWMI),
			OrderNo:         3,
			QuestionCount:   40,
			DurationMinutes: 30,
		},
		{
			Code:            models.QuestionIndexPSI,
			Label:           GetQuestionIndexLabel(models.QuestionIndexPSI),
			OrderNo:         4,
			QuestionCount:   20,
			DurationMinutes: 15,
		},
	}
}

func iqSectionRulesFromConfig(config models.TestConfig) []IQSectionRule {
	if len(config.SectionConfigs) == 0 {
		return defaultIQSectionRules()
	}

	rulesByCode := make(map[models.QuestionIndex]IQSectionRule, len(config.SectionConfigs))
	for _, section := range config.SectionConfigs {
		rulesByCode[section.QuestionIndex] = IQSectionRule{
			Code:            section.QuestionIndex,
			Label:           section.Label,
			OrderNo:         section.OrderNo,
			QuestionCount:   section.QuestionCount,
			DurationMinutes: section.DurationMinutes,
		}
	}

	defaults := defaultIQSectionRules()
	rules := make([]IQSectionRule, 0, len(defaults))
	for _, fallback := range defaults {
		rule, ok := rulesByCode[fallback.Code]
		if !ok {
			rule = fallback
		}
		if rule.Label == "" {
			rule.Label = fallback.Label
		}
		if rule.OrderNo <= 0 {
			rule.OrderNo = fallback.OrderNo
		}
		if rule.QuestionCount <= 0 {
			rule.QuestionCount = fallback.QuestionCount
		}
		if rule.DurationMinutes <= 0 {
			rule.DurationMinutes = fallback.DurationMinutes
		}
		rules = append(rules, rule)
	}

	return rules
}

func normalizeIQSectionPayloads(sections []TestSectionConfigPayload) ([]IQSectionRule, error) {
	defaults := defaultIQSectionRules()
	if len(sections) == 0 {
		return defaults, nil
	}

	byCode := make(map[models.QuestionIndex]TestSectionConfigPayload, len(sections))
	for _, section := range sections {
		code, err := NormalizeQuestionIndex(section.QuestionIndex)
		if err != nil {
			return nil, err
		}
		if code == models.QuestionIndexSKB {
			return nil, errors.New("section IQ tidak boleh memakai index SKB")
		}
		byCode[code] = section
	}

	rules := make([]IQSectionRule, 0, len(defaults))
	for _, fallback := range defaults {
		section, ok := byCode[fallback.Code]
		if !ok {
			section = TestSectionConfigPayload{
				QuestionIndex:   fallback.Code,
				Label:           fallback.Label,
				OrderNo:         fallback.OrderNo,
				QuestionCount:   fallback.QuestionCount,
				DurationMinutes: fallback.DurationMinutes,
			}
		}
		if section.QuestionCount <= 0 || section.DurationMinutes <= 0 {
			return nil, fmt.Errorf("jumlah soal dan durasi %s harus lebih dari 0", attemptSectionLabel(fallback.Code))
		}

		label := section.Label
		if label == "" {
			label = fallback.Label
		}
		orderNo := section.OrderNo
		if orderNo <= 0 {
			orderNo = fallback.OrderNo
		}

		rules = append(rules, IQSectionRule{
			Code:            fallback.Code,
			Label:           label,
			OrderNo:         orderNo,
			QuestionCount:   section.QuestionCount,
			DurationMinutes: section.DurationMinutes,
		})
	}

	return rules, nil
}

func totalIQSectionQuestionCount(rules []IQSectionRule) int {
	total := 0
	for _, rule := range rules {
		total += rule.QuestionCount
	}
	return total
}

func totalIQSectionDurationMinutes(rules []IQSectionRule) int {
	total := 0
	for _, rule := range rules {
		total += rule.DurationMinutes
	}
	return total
}

func buildQuestionSelectionPlan(totalQuestions int, available map[models.QuestionIndex]int) ([]QuestionSelectionPlanRow, error) {
	if totalQuestions == iqTotalQuestionCount {
		return buildIQQuestionSelectionPlan(defaultIQSectionRules(), available)
	}

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
			Code:            code,
			Label:           GetQuestionIndexLabel(code),
			OrderNo:         len(rows) + 1,
			Requested:       0,
			Available:       count,
			DurationMinutes: 0,
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

func buildIQQuestionSelectionPlan(rules []IQSectionRule, available map[models.QuestionIndex]int) ([]QuestionSelectionPlanRow, error) {
	if len(rules) == 0 {
		rules = defaultIQSectionRules()
	}

	rows := make([]QuestionSelectionPlanRow, 0, len(rules))
	for _, rule := range rules {
		if rule.QuestionCount <= 0 {
			continue
		}

		count := available[rule.Code]
		if count == 0 {
			continue
		}

		requested := rule.QuestionCount
		if count < requested {
			requested = count
		}
		durationMinutes := rule.DurationMinutes
		if durationMinutes <= 0 {
			durationMinutes = 1
		}

		rows = append(rows, QuestionSelectionPlanRow{
			Code:            rule.Code,
			Label:           rule.Label,
			OrderNo:         rule.OrderNo,
			Requested:       requested,
			Available:       count,
			DurationMinutes: durationMinutes,
		})
	}

	if len(rows) == 0 {
		return nil, errors.New("bank soal IQ belum memiliki soal published")
	}

	return rows, nil
}
