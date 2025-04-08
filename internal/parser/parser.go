package parser

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/erickgnclvs/go-task-viewer/internal/types"
)

// ParseTime function remains the same...
func ParseTime(timeStr string) float64 {
	// Skip empty or placeholder values
	if timeStr == "" || timeStr == "-" {
		return 0
	}

	// Clean the input
	timeStr = strings.TrimSpace(timeStr)

	// Initialize total minutes
	totalMinutes := 0.0

	// Handle hours (more robustly - find 'h' first)
	hourParts := strings.SplitN(timeStr, "h", 2)
	if len(hourParts) == 2 {
		hourStr := strings.TrimSpace(hourParts[0])
		hours, err := strconv.ParseFloat(hourStr, 64)
		if err == nil {
			totalMinutes += hours * 60
		}
		timeStr = strings.TrimSpace(hourParts[1]) // Remaining part
	}

	// Handle minutes (find 'm' first)
	minuteParts := strings.SplitN(timeStr, "m", 2)
	if len(minuteParts) == 2 {
		minuteStr := strings.TrimSpace(minuteParts[0])
		minutes, err := strconv.ParseFloat(minuteStr, 64)
		if err == nil {
			totalMinutes += minutes
		}
		timeStr = strings.TrimSpace(minuteParts[1]) // Remaining part
	}

	// Handle seconds (find 's' first)
	secondParts := strings.SplitN(timeStr, "s", 2)
	if len(secondParts) == 2 {
		secondStr := strings.TrimSpace(secondParts[0])
		seconds, err := strconv.ParseFloat(secondStr, 64)
		if err == nil {
			totalMinutes += seconds / 60
		}
		// No remaining part needed after seconds usually
	}

	// Handle cases where only numbers are present (assume minutes?) - Optional
	// if totalMinutes == 0 && strings.TrimSpace(timeStr) != "" {
	//     // Maybe treat raw numbers as minutes? Or log warning?
	// }

	return totalMinutes
}

// ParseCSV function remains the same...
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
		} else {
			task.Duration = "-" // Ensure default if column missing
			task.DurationMins = 0
		}

		if rateIdx >= 0 && rateIdx < len(record) {
			rateStr := strings.Trim(record[rateIdx], " \"")
			if rateStr == "-" || rateStr == "" {
				task.Rate = 0
			} else if strings.Contains(rateStr, "$") && strings.Contains(rateStr, "/hr") {
				rateVal := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(rateStr, "$"), "/hr"))
				task.Rate, _ = strconv.ParseFloat(rateVal, 64)
			} else if strings.HasPrefix(rateStr, "$") { // Handle rate given just as $ amount (assume per hour?)
				rateVal := strings.TrimSpace(strings.TrimPrefix(rateStr, "$"))
				task.Rate, _ = strconv.ParseFloat(rateVal, 64)
			}
		}

		if valueIdx >= 0 && valueIdx < len(record) {
			valueStr := strings.Trim(record[valueIdx], " \"")
			// Allow for "-" or empty value string
			if valueStr == "-" || valueStr == "" {
				task.Value = 0
			} else if strings.HasPrefix(valueStr, "$") {
				valueVal := strings.TrimSpace(strings.TrimPrefix(valueStr, "$"))
				// Use MustParseFloat or check error
				val, err := strconv.ParseFloat(valueVal, 64)
				if err == nil {
					task.Value = val
				} else {
					log.Printf("Warning: Could not parse value '%s' from CSV row: %v", valueStr, record)
					task.Value = 0 // Default to 0 on parse error
				}
			} else {
				// Handle case where value might be a plain number without '$'
				val, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
				if err == nil {
					task.Value = val
				} else {
					log.Printf("Warning: Could not parse value '%s' (no '$' prefix) from CSV row: %v", valueStr, record)
					task.Value = 0 // Default to 0 on parse error
				}
			}
		}

		if typeIdx >= 0 && typeIdx < len(record) {
			payType := strings.Trim(record[typeIdx], " \"")
			switch strings.ToLower(payType) {
			case "prepay", "regularpay", "task": // Added aliases
				task.Type = "Task"
			case "overtimepay", "exceeded time": // Added alias
				task.Type = "Exceeded Time"
			case "missionreward":
				task.Type = "Mission Reward"
			case "qaoperation", "operation":
				task.Type = "Operation"
			case "adjustment":
				task.Type = "Adjustment" // Handle Adjustment type
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

// ParseText function remains the same...
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
			// log.Printf("[DEBUG] Tarefa (text) encontrada: Type=%s, Value=%.2f", task.Type, task.Value) // Log value here too
			tasks = append(tasks, *task)
		}
		// If parseTextBlock returns nil, it means it wasn't a valid block starting at lines[i].
		// We should advance by 1 to check the next line as a potential start.
		// If it *did* parse a block, advance is 8. If it determined it wasn't a block
		// at the start, advance should be 1.
		if advance == 0 { // Prevent infinite loops if parseTextBlock has a bug
			log.Printf("[WARN] parseTextBlock returned advance=0, advancing by 1 to avoid loop. Line: %s", lines[i])
			advance = 1
		}
		i += advance // Advance by the number of lines consumed or skipped
	}

	log.Printf("Total de %d tarefas foram analisadas do texto.\n", len(tasks))
	return tasks
}

// **REVISED parseTextBlock**
func parseTextBlock(lines []string) (*types.Task, int) {
	// Need at least 8 lines for a potential task block
	if len(lines) < 8 {
		// log.Printf("[DEBUG] Not enough lines remaining (%d) for a text block.", len(lines))
		return nil, len(lines) // Consumed all remaining lines
	}

	// Check for the expected structure: non-empty lines 0, 1, 2, 4, 5, 7 and empty lines 3, 6
	if strings.TrimSpace(lines[0]) == "" || strings.TrimSpace(lines[1]) == "" ||
		strings.TrimSpace(lines[2]) == "" || strings.TrimSpace(lines[4]) == "" || // Line 4 must have *something*
		strings.TrimSpace(lines[5]) == "" || strings.TrimSpace(lines[7]) == "" ||
		strings.TrimSpace(lines[3]) != "" || strings.TrimSpace(lines[6]) != "" {
		// log.Printf("[DEBUG] Line structure mismatch at line starting with: %s", lines[0])
		return nil, 1 // Not a task block, advance by 1 line and try again
	}

	task := &types.Task{
		Date:     strings.TrimSpace(lines[0]),
		ID:       strings.TrimSpace(lines[1]),
		Category: strings.TrimSpace(lines[2]),
		Status:   strings.TrimSpace(lines[7]),
		// Assign Type based on line 5, handle variations
		Type:         mapTextType(strings.TrimSpace(lines[5])),
		Duration:     "-", // Default
		Rate:         0.0,
		Value:        0.0,
		DurationMins: 0.0,
	}

	// --- Robust Parsing of Line 4 ---
	durationRateValue := strings.TrimSpace(lines[4])
	parts := strings.Fields(durationRateValue)
	nParts := len(parts)

	valueIdx := -1
	rateIdx := -1
	durationEndIdx := nParts // Assume all parts are duration initially

	// 1. Find Value (last part starting with '$', not containing '/hr')
	if nParts > 0 && strings.HasPrefix(parts[nParts-1], "$") && !strings.Contains(parts[nParts-1], "/hr") {
		valueStr := strings.TrimPrefix(parts[nParts-1], "$")
		val, err := strconv.ParseFloat(valueStr, 64)
		if err == nil {
			task.Value = val
			valueIdx = nParts - 1
			durationEndIdx = valueIdx // Duration ends before value
		} else {
			log.Printf("[WARN] Text Parser: Failed to parse potential value '%s': %v", parts[nParts-1], err)
		}
	}

	// 2. Find Rate (part before Value OR last part, containing '$/hr')
	rateSearchIdx := -1
	if valueIdx > 0 {
		rateSearchIdx = valueIdx - 1 // Look before value
	} else if nParts > 0 && valueIdx == -1 { // No value found, check last part for rate
		rateSearchIdx = nParts - 1
	}

	if rateSearchIdx >= 0 && strings.Contains(parts[rateSearchIdx], "$") && strings.Contains(parts[rateSearchIdx], "/hr") {
		rateStr := strings.TrimSuffix(strings.TrimPrefix(parts[rateSearchIdx], "$"), "/hr")
		rate, err := strconv.ParseFloat(strings.TrimSpace(rateStr), 64)
		if err == nil {
			task.Rate = rate
			rateIdx = rateSearchIdx
			durationEndIdx = rateIdx // Duration ends before rate
		} else {
			log.Printf("[WARN] Text Parser: Failed to parse potential rate '%s': %v", parts[rateSearchIdx], err)
		}
	} else if rateSearchIdx >= 0 && strings.HasPrefix(parts[rateSearchIdx], "$") && valueIdx != rateSearchIdx {
		// Handle case like "$7.95 $0.00" where rate doesn't have /hr
		// Check if it looks like a rate (starts with $) and wasn't already identified as value
		rateStr := strings.TrimPrefix(parts[rateSearchIdx], "$")
		rate, err := strconv.ParseFloat(strings.TrimSpace(rateStr), 64)
		if err == nil {
			task.Rate = rate
			rateIdx = rateSearchIdx
			durationEndIdx = rateIdx // Duration ends before rate
		} else {
			log.Printf("[WARN] Text Parser: Failed to parse potential rate (no /hr) '%s': %v", parts[rateSearchIdx], err)
		}
	}

	// 3. Extract Duration (parts before rate/value)
	if durationEndIdx > 0 {
		durationParts := parts[0:durationEndIdx]
		// Filter out placeholder "-" before joining
		actualDurationParts := []string{}
		isPlaceholder := true
		for _, p := range durationParts {
			if p != "-" {
				actualDurationParts = append(actualDurationParts, p)
				isPlaceholder = false
			}
		}

		if !isPlaceholder && len(actualDurationParts) > 0 {
			task.Duration = strings.Join(actualDurationParts, " ")
		} else {
			task.Duration = "-" // Keep default "-" if only placeholders or empty
		}
	} else {
		// No parts left for duration, or Rate/Value took all parts
		task.Duration = "-"
	}

	// Ensure duration is "-" if it still looks like a money value mistakenly
	if strings.HasPrefix(task.Duration, "$") {
		log.Printf("[WARN] Text Parser: Corrected Duration from '%s' to '-'", task.Duration)
		task.Duration = "-"
	}

	// Calculate duration minutes *after* parsing duration string
	if task.Duration != "" && task.Duration != "-" {
		task.DurationMins = ParseTime(task.Duration)
	} else {
		task.Duration = "-"   // Standardize
		task.DurationMins = 0 // Ensure it's 0 if duration is missing/placeholder
	}

	// Debug output (optional)
	// log.Printf("Text Parsed: Date=%s, ID=%s, Type=%s, Category=%s, Status=%s", task.Date, task.ID, task.Type, task.Category, task.Status)
	// log.Printf("           Line 4: '%s' -> Duration='%s', Rate=%.2f, Value=%.2f, DurationMins=%.2f", durationRateValue, task.Duration, task.Rate, task.Value, task.DurationMins)

	return task, 8 // Successfully parsed a task block of 8 lines
}

// Helper to map text input types to standardized types
func mapTextType(textType string) string {
	switch strings.ToLower(textType) {
	case "task", "regular pay", "prepay":
		return "Task"
	case "exceeded time", "overtime pay":
		return "Exceeded Time"
	case "mission reward":
		return "Mission Reward"
	case "operation", "qa operation":
		return "Operation"
	case "adjustment":
		return "Adjustment"
	default:
		log.Printf("[WARN] Unknown text task type encountered: %s", textType)
		return textType // Keep original if unknown
	}
}

// isProjectCategory checks if a category string looks like a specific project name
// rather than a generic type or placeholder.
func isProjectCategory(category string) bool {
	trimmedCategory := strings.TrimSpace(category)
	// Basic check: not empty, not "-", and doesn't look like a "Mission:" description.
	// You might add more specific checks here if needed (e.g., !strings.Contains(trimmedCategory, "Operation")).
	return trimmedCategory != "" && trimmedCategory != "-" && !strings.HasPrefix(trimmedCategory, "Mission:")
}

// FillMissingCategories iterates through tasks and fills missing/generic categories
// based on the last known specific project category encountered.
func FillMissingCategories(tasks []types.Task) []types.Task {
	if len(tasks) == 0 {
		return tasks // No tasks to process
	}

	lastKnownProjectCategory := ""
	modifiedTasks := make([]types.Task, len(tasks)) // Create a new slice to avoid modifying the original directly if needed elsewhere

	for i, task := range tasks {
		// Check if the current task's category seems like a specific project
		if isProjectCategory(task.Category) {
			lastKnownProjectCategory = strings.TrimSpace(task.Category) // Update the last known project
			// log.Printf("[DEBUG] FillCategory: Found project category '%s' at index %d", lastKnownProjectCategory, i)
		} else {
			// If the category is not a specific project and we have a known project category, fill it in.
			if lastKnownProjectCategory != "" {
				// log.Printf("[DEBUG] FillCategory: Filling category for task '%s' (index %d) from '%s' to '%s'", task.ID, i, task.Category, lastKnownProjectCategory)
				task.Category = lastKnownProjectCategory // Modify the category
			} else {
				// log.Printf("[DEBUG] FillCategory: No known project category yet for task '%s' (index %d), category remains '%s'", task.ID, i, task.Category)
			}
		}
		modifiedTasks[i] = task // Add the (potentially modified) task to the new slice
	}

	// Optional: Log how many categories were potentially filled
	filledCount := 0
	for i := range tasks {
		if tasks[i].Category != modifiedTasks[i].Category {
			filledCount++
		}
	}
	if filledCount > 0 {
		log.Printf("[INFO] Filled missing/generic category for %d tasks based on preceding project context.", filledCount)
	}

	return modifiedTasks
}
