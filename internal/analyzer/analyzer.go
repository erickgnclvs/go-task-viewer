package analyzer

import (
	"log"
	"github.com/erickgnclvs/go-task-viewer/internal/types"
)

// AnalyzeData processes a slice of tasks and calculates summary statistics.
func AnalyzeData(tasks []types.Task) map[string]interface{} {
	totalTasks := 0             // Count of items explicitly marked as "Task"
	totalTasksValue := 0.0      // Sum of value for "Task" items
	totalExceededTimeValue := 0.0 // Sum of value for "Exceeded Time" items
	totalOtherValue := 0.0      // Sum of value for all other item types (Mission Reward, Operation, etc.)
	totalHours := 0.0         // Sum of duration (in hours) for all item types

	// Detailed hour breakdowns
	taskHours := 0.0         // Sum of duration (in hours) for "Task" items
	exceededTimeHours := 0.0 // Sum of duration (in hours) for "Exceeded Time" items
	otherHours := 0.0        // Sum of duration (in hours) for all other item types

	// For average calculations specific to "Task" items
	totalTaskTime := 0.0 // Sum of duration (in minutes) for "Task" items
	totalTaskCount := 0  // Count used for average calculations (same as totalTasks)

	for _, task := range tasks {
		// Debug output
		// log.Printf("Analisando: Type=%s, Value=%.2f, DurationMins=%.2f\n",
		// 	task.Type, task.Value, task.DurationMins)

		// Accumulate hours based on type
		hours := task.DurationMins / 60
		totalHours += hours

		switch task.Type {
		case "Task":
			totalTasks++
			totalTaskCount++             // Increment task count for averages
			totalTaskTime += task.DurationMins // Accumulate task time in minutes
			totalTasksValue += task.Value
			taskHours += hours
		case "Exceeded Time":
			totalExceededTimeValue += task.Value
			exceededTimeHours += hours
		case "Mission Reward", "Operation": // Group known 'Other' types
			totalOtherValue += task.Value
			otherHours += hours
		default: // Catch any unexpected types
			log.Printf("Warning: Unknown task type encountered: %s", task.Type)
			totalOtherValue += task.Value // Add value to 'Other'
			otherHours += hours        // Add hours to 'Other'
		}
	}

	totalValue := totalTasksValue + totalExceededTimeValue + totalOtherValue

	// Calculate averages
	averageHourlyRate := 0.0
	if totalHours > 0 {
		// Average hourly rate considers value from Task and Exceeded Time, divided by total hours
		averageHourlyRate = (totalTasksValue + totalExceededTimeValue) / totalHours
	}

	// Average time per task (in minutes)
	avgTimePerTask := 0.0
	if totalTaskCount > 0 {
		avgTimePerTask = totalTaskTime / float64(totalTaskCount)
	}

	// Average value per task
	avgValuePerTask := 0.0
	if totalTaskCount > 0 {
		avgValuePerTask = totalTasksValue / float64(totalTaskCount)
	}

	return map[string]interface{}{
		"TotalTasks":        totalTasks,
		"TotalHours":        totalHours, // Raw float value
		"TotalValue":        totalValue,
		"TasksValue":        totalTasksValue,
		"ExceededTimeValue": totalExceededTimeValue,
		"OtherValue":        totalOtherValue,
		"AverageHourlyRate": averageHourlyRate,
		// Detailed hour breakdowns (raw float values)
		"TaskHours":         taskHours,
		"ExceededTimeHours": exceededTimeHours,
		"OtherHours":        otherHours,
		// Average metrics (raw float values)
		"AvgTimePerTask":  avgTimePerTask, // In minutes
		"AvgValuePerTask": avgValuePerTask,
	}
}
