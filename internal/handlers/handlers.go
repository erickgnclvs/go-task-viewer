package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/erickgnclvs/go-task-viewer/internal/analyzer"
	"github.com/erickgnclvs/go-task-viewer/internal/parser"
	"github.com/erickgnclvs/go-task-viewer/internal/types"
)

// HomeHandler serves the main page with the input form.
func HomeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data := types.TemplateData{
			CurrentYear: time.Now().Year(),
		}
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing home template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// AnalyzeHandler handles the form submission, parses data, analyzes it, and displays results.
func AnalyzeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Set max file size (e.g., 10MB)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			http.Error(w, "Error processing form data", http.StatusBadRequest)
			return
		}

		// Get form data
		showDetails := r.FormValue("showDetails") == "on"
		log.Printf("[DEBUG] Form showDetails=%v", showDetails)

		var tasks []types.Task
		var inputSource string
		var rawInputData string // Store the raw input for display

		// Check for file upload first
		file, handler, err := r.FormFile("csvFile")
		if err == nil {
			// File uploaded
			defer file.Close()
			log.Printf("Uploaded File: %+v, Size: %+v", handler.Filename, handler.Size)

			// Read the file content to store for display *before* parsing
			fileBytes, readErr := io.ReadAll(file)
			if readErr != nil {
				log.Printf("Error reading uploaded file: %v", readErr)
				http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
				return
			}
			rawInputData = string(fileBytes)

			// Parse the CSV data (using a new reader from the read bytes)
			reader := strings.NewReader(rawInputData)
			tasks = parser.ParseCSV(reader)
			log.Printf("[DEBUG] CSV Upload: %d tasks found after initial parse", len(tasks))
			inputSource = "csv"
		} else if err != http.ErrMissingFile {
			// Handle other potential errors from FormFile
			log.Printf("Error retrieving file from form: %v", err)
			http.Error(w, "Error processing file upload", http.StatusInternalServerError)
			return
		} else {
			// No file uploaded, fall back to text input
			rawInputData = r.FormValue("taskData")
			specifiedSource := r.FormValue("inputSource") // Check if user specified format
			log.Printf("[DEBUG] Text input provided. Specified source: %s", specifiedSource)

			if rawInputData != "" {
				if specifiedSource == "csv" {
					log.Printf("[DEBUG] Processing text input as CSV")
					reader := strings.NewReader(rawInputData)
					tasks = parser.ParseCSV(reader)
					inputSource = "csv"
				} else { // Default to text format or if source is 'text'
					log.Printf("[DEBUG] Processing text input as multi-line text")
					tasks = parser.ParseText(rawInputData)
					inputSource = "text"
				}
				log.Printf("[DEBUG] Text Input: %d tasks found after initial parse", len(tasks))
			} else {
				log.Println("[DEBUG] No file uploaded and text area is empty.")
				// Optionally, redirect back with an error message?
			}
		}

		// *** NEW STEP: Fill missing categories ***
		if len(tasks) > 0 {
			log.Printf("[DEBUG] Running FillMissingCategories on %d tasks", len(tasks))
			tasks = parser.FillMissingCategories(tasks)
			log.Printf("[DEBUG] FillMissingCategories completed. Task count remains %d", len(tasks))
		}

		// Prepare data for the template
		data := types.TemplateData{
			RawInput:    rawInputData,
			HasResults:  len(tasks) > 0,
			InputSource: inputSource,
			CurrentYear: time.Now().Year(),
			ShowDetails: showDetails,
			// Tasks will be populated below if needed
		}

		// Analyze the data and format results if we have tasks
		if len(tasks) > 0 {
			log.Printf("[DEBUG] Analyzing %d tasks (post-category fill) from source '%s'", len(tasks), inputSource)
			results := analyzer.AnalyzeData(tasks) // Pass the modified tasks

			// Populate TemplateData with analysis results
			populateTemplateData(&data, results)

			// Format tasks for display if requested
			if showDetails {
				data.Tasks = formatTasksForDisplay(tasks) // Pass the modified tasks
				log.Printf("[DEBUG] Formatted %d tasks (post-category fill) for details display", len(data.Tasks))
			}
		} else {
			log.Println("[DEBUG] No tasks found to analyze.")
		}

		log.Printf("[DEBUG] Rendering template: HasResults=%v, ShowDetails=%v, TaskCount=%d", data.HasResults, data.ShowDetails, len(data.Tasks))
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing analyze template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// HealthHandler provides a simple health check endpoint.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// populateTemplateData fills the TemplateData struct with formatted analysis results.
func populateTemplateData(data *types.TemplateData, results map[string]interface{}) {
	totalHoursValue := results["TotalHours"].(float64)
	totalHoursInt := int(totalHoursValue)
	totalMinutes := int((totalHoursValue - float64(totalHoursInt)) * 60)

	taskHoursValue := results["TaskHours"].(float64)
	taskHoursInt := int(taskHoursValue)
	taskMinutes := int((taskHoursValue - float64(taskHoursInt)) * 60)

	exceededTimeHoursValue := results["ExceededTimeHours"].(float64)
	exceededTimeHoursInt := int(exceededTimeHoursValue)
	exceededTimeMinutes := int((exceededTimeHoursValue - float64(exceededTimeHoursInt)) * 60)

	otherHoursValue := results["OtherHours"].(float64)
	otherHoursInt := int(otherHoursValue)
	otherMinutes := int((otherHoursValue - float64(otherHoursInt)) * 60)

	avgTimePerTaskValue := results["AvgTimePerTask"].(float64) // This is in minutes
	avgTimeMinutes := int(avgTimePerTaskValue)
	avgTimeSeconds := int((avgTimePerTaskValue - float64(avgTimeMinutes)) * 60)

	avgValuePerTaskValue := results["AvgValuePerTask"].(float64)

	data.TotalTasks = results["TotalTasks"].(int)
	data.TotalHours = fmt.Sprintf("%.2f horas (%dh %dmin)", totalHoursValue, totalHoursInt, totalMinutes)
	data.TotalValue = fmt.Sprintf("%.2f", results["TotalValue"].(float64))
	data.TasksValue = fmt.Sprintf("%.2f", results["TasksValue"].(float64))
	data.ExceededTimeValue = fmt.Sprintf("%.2f", results["ExceededTimeValue"].(float64))
	data.OtherValue = fmt.Sprintf("%.2f", results["OtherValue"].(float64))
	data.AverageHourlyRate = fmt.Sprintf("%.2f", results["AverageHourlyRate"].(float64))

	data.TaskHours = fmt.Sprintf("%.2f horas (%dh %dmin)", taskHoursValue, taskHoursInt, taskMinutes)
	data.ExceededTimeHours = fmt.Sprintf("%.2f horas (%dh %dmin)", exceededTimeHoursValue, exceededTimeHoursInt, exceededTimeMinutes)
	data.OtherHours = fmt.Sprintf("%.2f horas (%dh %dmin)", otherHoursValue, otherHoursInt, otherMinutes)

	data.AvgTimePerTask = fmt.Sprintf("%dm %ds", avgTimeMinutes, avgTimeSeconds)
	data.AvgValuePerTask = fmt.Sprintf("$%.2f", avgValuePerTaskValue)

	// Calculate hour percentages for progress bars
	if totalHoursValue > 0 {
		taskPercentage := (taskHoursValue / totalHoursValue) * 100
		exceededPercentage := (exceededTimeHoursValue / totalHoursValue) * 100
		otherPercentage := (otherHoursValue / totalHoursValue) * 100
		data.RawHourPercentages = []float64{taskPercentage, exceededPercentage, otherPercentage}
	} else {
		data.RawHourPercentages = []float64{0, 0, 0}
	}
}

// formatTasksForDisplay converts raw Task structs into TaskDisplay structs for the HTML table.
func formatTasksForDisplay(tasks []types.Task) []types.TaskDisplay {
	var taskDisplays []types.TaskDisplay
	for _, task := range tasks {
		rateDisplay := "-"
		durationDisplay := task.Duration
		durationMinsDisplay := "-"

		if task.DurationMins > 0 {
			durationMinsDisplay = fmt.Sprintf("%.2f mins", task.DurationMins)
		}
		if task.Duration == "" { // Ensure empty duration shows as '-'
			durationDisplay = "-"
		}

		if task.Rate > 0 {
			rateDisplay = fmt.Sprintf("$%.2f/hr", task.Rate)
		}

		taskDisplays = append(taskDisplays, types.TaskDisplay{
			Date:         task.Date,
			ID:           task.ID,
			Category:     task.Category,
			Duration:     durationDisplay,
			Rate:         rateDisplay,
			Value:        fmt.Sprintf("$%.2f", task.Value),
			Type:         task.Type,
			Status:       task.Status,
			DurationMins: durationMinsDisplay,
		})
	}
	return taskDisplays
}
