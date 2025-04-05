package parser

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/erickgnclvs/go-task-viewer/internal/types"
)

// ParseTime converts a time string (e.g., "1h 30m", "45m", "15s") into total minutes.
func ParseTime(timeStr string) float64 {
	// Skip empty or placeholder values
	if timeStr == "" || timeStr == "-" {
		return 0
	}

	// Clean the input
	timeStr = strings.TrimSpace(timeStr)

	// Initialize total minutes
	totalMinutes := 0.0

	// Handle hours
	if strings.Contains(timeStr, "h") {
		parts := strings.Split(timeStr, "h")
		hourStr := strings.TrimSpace(parts[0])
		hours, err := strconv.ParseFloat(hourStr, 64)
		if err == nil {
			totalMinutes += hours * 60
		}

		// If there's content after 'h', update timeStr to process it
		if len(parts) > 1 {
			timeStr = strings.TrimSpace(parts[1])
		} else {
			timeStr = ""
		}
	}

	// Handle minutes
	if strings.Contains(timeStr, "m") {
		parts := strings.Split(timeStr, "m")
		minuteStr := strings.TrimSpace(parts[0])
		minutes, err := strconv.ParseFloat(minuteStr, 64)
		if err == nil {
			totalMinutes += minutes
		}

		// If there's content after 'm', update timeStr to process it
		if len(parts) > 1 {
			timeStr = strings.TrimSpace(parts[1])
		} else {
			timeStr = ""
		}
	}

	// Handle seconds
	if strings.Contains(timeStr, "s") {
		parts := strings.Split(timeStr, "s")
		secondStr := strings.TrimSpace(parts[0])
		seconds, err := strconv.ParseFloat(secondStr, 64)
		if err == nil {
			totalMinutes += seconds / 60
		}
	}

	return totalMinutes
}

// ParseCSV parses a CSV file (or CSV string content) with task data.
func ParseCSV(file io.Reader) []types.Task {
	var tasks []types.Task

	// Create CSV reader
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Read and skip header row
	header, err := reader.Read()
	if err != nil {
		if err != io.EOF { // Allow empty CSVs
			log.Printf("Error reading CSV header: %v\n", err)
		}
		return tasks
	}

	// Map CSV columns to our expected structure
	dateIdx := -1
	idIdx := -1
	durationIdx := -1
	rateIdx := -1
	valueIdx := -1
	typeIdx := -1
	projectIdx := -1
	statusIdx := -1

	for i, col := range header {
		switch strings.ToLower(strings.TrimSpace(col)) { // Trim spaces from header
		case "workdate", "date":
			dateIdx = i
		case "itemid", "id":
			idIdx = i
		case "duration":
			durationIdx = i
		case "rateapplied", "rate":
			rateIdx = i
		case "payout", "value":
			valueIdx = i
		case "paytype", "type":
			typeIdx = i
		case "projectname", "project", "category":
			projectIdx = i
		case "status":
			statusIdx = i
		}
	}

	// Read all records and convert to tasks
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV row: %v\n", err)
			continue
		}

		task := types.Task{}

		// Extract data from CSV columns safely
		if dateIdx >= 0 && dateIdx < len(record) {
			task.Date = strings.Trim(record[dateIdx], " \"") // Trim spaces and quotes
		}

		if idIdx >= 0 && idIdx < len(record) {
			task.ID = strings.Trim(record[idIdx], " \"")
		}

		if durationIdx >= 0 && durationIdx < len(record) {
			task.Duration = strings.Trim(record[durationIdx], " \"")
			if task.Duration != "-" && task.Duration != "" {
				task.DurationMins = ParseTime(task.Duration)
			} else {
				task.Duration = "-" // Standardize empty values
				task.DurationMins = 0
			}
		}

		if rateIdx >= 0 && rateIdx < len(record) {
			rateStr := strings.Trim(record[rateIdx], " \"")
			if rateStr == "-" || rateStr == "" {
				task.Rate = 0
			} else if strings.Contains(rateStr, "$") && strings.Contains(rateStr, "/hr") {
				rateVal := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(rateStr, "$"), "/hr"))
				task.Rate, _ = strconv.ParseFloat(rateVal, 64)
			}
		}

		if valueIdx >= 0 && valueIdx < len(record) {
			valueStr := strings.Trim(record[valueIdx], " \"")
			if strings.Contains(valueStr, "$") {
				valueVal := strings.TrimSpace(strings.TrimPrefix(valueStr, "$"))
				task.Value, _ = strconv.ParseFloat(valueVal, 64)
			}
		}

		if typeIdx >= 0 && typeIdx < len(record) {
			payType := strings.Trim(record[typeIdx], " \"")
			switch strings.ToLower(payType) {
			case "prepay":
				task.Type = "Task"
			case "overtimepay":
				task.Type = "Exceeded Time"
			case "missionreward":
				task.Type = "Mission Reward"
			case "qaoperation", "operation":
				task.Type = "Operation"
			default:
				task.Type = payType // Keep original if unknown
			}
		}

		if projectIdx >= 0 && projectIdx < len(record) {
			task.Category = strings.Trim(record[projectIdx], " \"")
		}

		if statusIdx >= 0 && statusIdx < len(record) {
			task.Status = strings.Trim(record[statusIdx], " \"")
		}

		// Debug output
		// log.Printf("CSV Parsed: Date=%s, ID=%s, Type=%s, Duration=%s, Rate=%.2f, Value=%.2f, DurationMins=%.2f\n",
		// 	task.Date, task.ID, task.Type, task.Duration, task.Rate, task.Value, task.DurationMins)

		tasks = append(tasks, task)
	}

	log.Printf("Total de %d tarefas foram analisadas do CSV.\n", len(tasks))
	return tasks
}

// ParseText parses the multi-line text input format.
func ParseText(input string) []types.Task {
	var tasks []types.Task
	log.Printf("[DEBUG] Iniciando ParseText com %d caracteres de texto", len(input))
	lines := strings.Split(input, "\n")
	log.Printf("[DEBUG] Dividido em %d linhas", len(lines))

	for i := 0; i < len(lines); {
		// Skip empty lines that might separate task blocks
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}

		task, advance := parseTextBlock(lines[i:])
		if task != nil {
			log.Printf("[DEBUG] Tarefa (text) encontrada: Type=%s", task.Type)
			tasks = append(tasks, *task)
		}
		i += advance // Advance by the number of lines consumed by parseTextBlock
	}

	log.Printf("Total de %d tarefas foram analisadas do texto.\n", len(tasks))
	return tasks
}

// parseTextBlock parses a single task block from the multi-line text format.
// It expects a specific 8-line structure.
func parseTextBlock(lines []string) (*types.Task, int) {
	// Need at least 8 lines for a potential task block
	if len(lines) < 8 {
		return nil, len(lines) // Consumed all remaining lines
	}

	// Check for the expected structure: non-empty lines 0, 1, 2, 4, 5, 7 and empty lines 3, 6
	if strings.TrimSpace(lines[0]) == "" || strings.TrimSpace(lines[1]) == "" ||
		strings.TrimSpace(lines[2]) == "" || strings.TrimSpace(lines[4]) == "" ||
		strings.TrimSpace(lines[5]) == "" || strings.TrimSpace(lines[7]) == "" ||
		strings.TrimSpace(lines[3]) != "" || strings.TrimSpace(lines[6]) != "" {
		return nil, 1 // Not a task block, advance by 1 line and try again
	}

	task := &types.Task{
		Date:     strings.TrimSpace(lines[0]),
		ID:       strings.TrimSpace(lines[1]),
		Category: strings.TrimSpace(lines[2]),
		Status:   strings.TrimSpace(lines[7]),
		Type:     strings.TrimSpace(lines[5]), // Assign Type directly from line 5
	}

	// Parse duration/rate/value line (line 4)
	durationRateValue := strings.TrimSpace(lines[4])

	// Split by multiple spaces or tabs - more robust parsing
	parts := strings.Fields(durationRateValue)

	// Handle different patterns based on the content and number of parts
	if len(parts) >= 3 {
		// Assume format: Duration Rate Value (e.g., "1h 30m $10.00/hr $15.00")
		task.Duration = parts[0]
		if strings.Contains(parts[1], "$") && strings.Contains(parts[1], "/hr") {
			rateStr := strings.TrimSuffix(strings.TrimPrefix(parts[1], "$"), "/hr")
			task.Rate, _ = strconv.ParseFloat(rateStr, 64)
		}
		if strings.HasPrefix(parts[2], "$") {
			valueStr := strings.TrimPrefix(parts[2], "$")
			task.Value, _ = strconv.ParseFloat(valueStr, 64)
		}
	} else if len(parts) == 3 && parts[0] == "-" && parts[1] == "-" {
		// Special case for Mission Reward/Operation: "- - $Value" (e.g., "- - $92.75")
		task.Duration = "-"
		task.Rate = 0
		if strings.HasPrefix(parts[2], "$") {
			valueStr := strings.TrimPrefix(parts[2], "$")
			task.Value, _ = strconv.ParseFloat(valueStr, 64)
		}
	} else if len(parts) > 0 {
		// Fallback: might just be duration, or something else unexpected
		task.Duration = parts[0]
		// Assign default Rate/Value if not parsed
		if task.Rate == 0 {
			task.Rate = 0
		}
		if task.Value == 0 {
			task.Value = 0
		}
	}

	// Calculate duration minutes *after* parsing duration string
	if task.Duration != "" && task.Duration != "-" {
		task.DurationMins = ParseTime(task.Duration)
	} else {
		task.DurationMins = 0 // Ensure it's 0 if duration is missing/placeholder
	}

	// Debug output
	// log.Printf("Text Parsed: Date=%s, ID=%s, Type=%s, Duration=%s, Rate=%.2f, Value=%.2f, DurationMins=%.2f\n",
	// 	task.Date, task.ID, task.Type, task.Duration, task.Rate, task.Value, task.DurationMins)

	return task, 8 // Successfully parsed a task block of 8 lines
}
