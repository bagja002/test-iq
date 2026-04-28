package services

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"test-iq-ku/apps/api/internal/models"
)

type QuestionImportRow struct {
	RowNumber int             `json:"rowNumber"`
	Payload   QuestionPayload `json:"payload"`
	Error     string          `json:"error"`
}

type QuestionImportRowError struct {
	RowNumber int    `json:"rowNumber"`
	Message   string `json:"message"`
}

type QuestionImportResult struct {
	TotalRows     int                      `json:"totalRows"`
	ImportedCount int                      `json:"importedCount"`
	FailedCount   int                      `json:"failedCount"`
	Errors        []QuestionImportRowError `json:"errors"`
}

var questionImportHeaders = []string{
	"test_type",
	"question_index",
	"subtest_code",
	"room_code",
	"prompt",
	"prompt_media_url",
	"prompt_media_alt",
	"difficulty",
	"status",
	"option_a",
	"option_a_media_url",
	"option_a_media_alt",
	"option_b",
	"option_b_media_url",
	"option_b_media_alt",
	"option_c",
	"option_c_media_url",
	"option_c_media_alt",
	"option_d",
	"option_d_media_url",
	"option_d_media_alt",
	"correct_key",
}

func ParseQuestionImportSpreadsheet(filename string, raw []byte) ([]QuestionImportRow, error) {
	if len(raw) == 0 {
		return nil, errors.New("file import kosong")
	}

	extension := strings.ToLower(filepath.Ext(filename))
	switch extension {
	case ".xlsx":
		records, err := parseXLSXRows(raw)
		if err != nil {
			return nil, err
		}
		return buildQuestionImportRows(records)
	case ".csv":
		reader := csv.NewReader(bytes.NewReader(raw))
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("file CSV tidak valid: %w", err)
		}
		return buildQuestionImportRows(records)
	default:
		return nil, errors.New("format file harus .xlsx atau .csv")
	}
}

func buildQuestionImportRows(records [][]string) ([]QuestionImportRow, error) {
	if len(records) < 2 {
		return nil, errors.New("file import harus memiliki header dan minimal satu baris soal")
	}

	headerMap := map[string]int{}
	for index, header := range records[0] {
		headerMap[normalizeImportHeader(header)] = index
	}

	for _, requiredHeader := range questionImportHeaders {
		if _, ok := headerMap[requiredHeader]; !ok {
			return nil, fmt.Errorf("kolom %s belum ada di template", requiredHeader)
		}
	}

	rows := make([]QuestionImportRow, 0, len(records)-1)
	for index, record := range records[1:] {
		rowNumber := index + 2
		if importRecordIsEmpty(record) {
			continue
		}

		payload, err := buildQuestionPayloadFromRecord(record, headerMap)
		row := QuestionImportRow{
			RowNumber: rowNumber,
			Payload:   payload,
		}
		if err != nil {
			row.Error = err.Error()
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil, errors.New("tidak ada baris soal yang bisa diimport")
	}

	return rows, nil
}

func buildQuestionPayloadFromRecord(record []string, headerMap map[string]int) (QuestionPayload, error) {
	value := func(header string) string {
		index, ok := headerMap[header]
		if !ok || index >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[index])
	}

	testType := strings.ToUpper(value("test_type"))
	questionIndex := strings.ToUpper(value("question_index"))
	subtestCode := strings.ToUpper(value("subtest_code"))
	roomCode := strings.ToUpper(value("room_code"))
	correctKey := strings.ToUpper(value("correct_key"))

	if testType == "" {
		testType = string(models.TestTypeIQ)
	}
	if questionIndex == "" && testType == string(models.TestTypeSKB) {
		questionIndex = string(models.QuestionIndexSKB)
	}
	if subtestCode == "" && testType == string(models.TestTypeSKB) {
		subtestCode = roomCode
	}

	options := []QuestionOptionInput{
		buildImportOption("A", correctKey, value("option_a"), value("option_a_media_url"), value("option_a_media_alt")),
		buildImportOption("B", correctKey, value("option_b"), value("option_b_media_url"), value("option_b_media_alt")),
		buildImportOption("C", correctKey, value("option_c"), value("option_c_media_url"), value("option_c_media_alt")),
		buildImportOption("D", correctKey, value("option_d"), value("option_d_media_url"), value("option_d_media_alt")),
	}

	payload := QuestionPayload{
		Prompt:         value("prompt"),
		PromptMediaURL: value("prompt_media_url"),
		PromptMediaAlt: value("prompt_media_alt"),
		Difficulty:     value("difficulty"),
		QuestionIndex:  models.QuestionIndex(questionIndex),
		SubtestCode:    subtestCode,
		Status:         models.QuestionStatus(strings.ToUpper(value("status"))),
		Options:        options,
	}

	if payload.Status == "" {
		payload.Status = models.QuestionStatusPublished
	}
	if payload.Difficulty == "" {
		payload.Difficulty = "medium"
	}

	if testType != string(models.TestTypeIQ) && testType != string(models.TestTypeSKB) {
		return payload, errors.New("test_type harus IQ atau SKB")
	}
	if testType == string(models.TestTypeIQ) && payload.QuestionIndex == models.QuestionIndexSKB {
		return payload, errors.New("question_index IQ tidak boleh SKB")
	}
	if testType == string(models.TestTypeSKB) && payload.QuestionIndex != models.QuestionIndexSKB {
		return payload, errors.New("question_index SKB harus SKB")
	}
	if correctKey != "A" && correctKey != "B" && correctKey != "C" && correctKey != "D" {
		return payload, errors.New("correct_key harus A, B, C, atau D")
	}
	if testType == string(models.TestTypeSKB) && roomCode != "" && roomCode != subtestCode {
		return payload, errors.New("room_code SKB harus sama dengan subtest_code")
	}

	return payload, nil
}

func buildImportOption(key string, correctKey string, content string, mediaURL string, mediaAlt string) QuestionOptionInput {
	return QuestionOptionInput{
		Key:       key,
		Content:   content,
		MediaURL:  mediaURL,
		MediaAlt:  mediaAlt,
		IsCorrect: correctKey == key,
	}
}

func normalizeImportHeader(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func importRecordIsEmpty(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parseXLSXRows(raw []byte) ([][]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("file Excel tidak valid: %w", err)
	}

	files := map[string]*zip.File{}
	for _, file := range reader.File {
		files[file.Name] = file
	}

	sharedStrings, err := readXLSXSharedStrings(files["xl/sharedStrings.xml"])
	if err != nil {
		return nil, err
	}

	sheet := files["xl/worksheets/sheet1.xml"]
	if sheet == nil {
		return nil, errors.New("sheet pertama Excel tidak ditemukan")
	}

	sheetFile, err := sheet.Open()
	if err != nil {
		return nil, err
	}
	defer sheetFile.Close()

	var worksheet xlsxWorksheet
	if err := xml.NewDecoder(sheetFile).Decode(&worksheet); err != nil {
		return nil, fmt.Errorf("sheet Excel tidak valid: %w", err)
	}

	records := make([][]string, 0, len(worksheet.Rows))
	for _, row := range worksheet.Rows {
		values := []string{}
		for cellIndex, cell := range row.Cells {
			columnIndex := xlsxColumnIndex(cell.Ref)
			if columnIndex < 0 {
				columnIndex = cellIndex
			}
			for len(values) <= columnIndex {
				values = append(values, "")
			}
			values[columnIndex] = xlsxCellValue(cell, sharedStrings)
		}
		records = append(records, values)
	}

	return records, nil
}

type xlsxWorksheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	Ref         string         `xml:"r,attr"`
	Type        string         `xml:"t,attr"`
	Value       string         `xml:"v"`
	InlineValue xlsxInlineText `xml:"is"`
}

type xlsxInlineText struct {
	Text string `xml:"t"`
}

func readXLSXSharedStrings(file *zip.File) ([]string, error) {
	if file == nil {
		return []string{}, nil
	}

	sharedFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer sharedFile.Close()

	decoder := xml.NewDecoder(sharedFile)
	items := []string{}
	var builder strings.Builder
	inSharedItem := false

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("shared strings Excel tidak valid: %w", err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "si" {
				inSharedItem = true
				builder.Reset()
				continue
			}
			if inSharedItem && element.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &element); err != nil {
					return nil, err
				}
				builder.WriteString(text)
			}
		case xml.EndElement:
			if element.Name.Local == "si" {
				items = append(items, builder.String())
				inSharedItem = false
			}
		}
	}

	return items, nil
}

func xlsxCellValue(cell xlsxCell, sharedStrings []string) string {
	switch cell.Type {
	case "s":
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return ""
		}
		return strings.TrimSpace(sharedStrings[index])
	case "inlineStr":
		return strings.TrimSpace(cell.InlineValue.Text)
	default:
		return strings.TrimSpace(cell.Value)
	}
}

func xlsxColumnIndex(reference string) int {
	letters := ""
	for _, char := range reference {
		if char >= 'A' && char <= 'Z' {
			letters += string(char)
			continue
		}
		if char >= 'a' && char <= 'z' {
			letters += strings.ToUpper(string(char))
			continue
		}
		break
	}
	if letters == "" {
		return -1
	}

	index := 0
	for _, char := range letters {
		index = index*26 + int(char-'A'+1)
	}
	return index - 1
}
