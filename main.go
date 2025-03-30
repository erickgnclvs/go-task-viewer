package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Task struct {
	Date         string
	ID           string
	Category     string
	Duration     string
	Rate         float64
	Value        float64
	Type         string
	Status       string
	DurationMins float64
}

func parseTime(timeStr string) float64 {
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

// TemplateData holds data to be passed to HTML templates
type TemplateData struct {
	RawInput         string
	HasResults       bool
	TotalTasks       int
	TotalHours       string
	TotalValue       string
	TasksValue       string
	ExceededTimeValue string
	OtherValue       string
	AverageHourlyRate string
	CurrentYear      int
	// Input source (csv or text)
	InputSource      string
	// Task details
	ShowDetails      bool
	Tasks            []TaskDisplay
}

// TaskDisplay represents a task for display purposes
type TaskDisplay struct {
	Date         string
	ID           string
	Category     string
	Duration     string
	Rate         string
	Value        string
	Type         string
	Status       string
	DurationMins string
}

// parseCSV parses a CSV file with task data
func parseCSV(file io.Reader) []Task {
	var tasks []Task
	
	// Create CSV reader
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	
	// Read and skip header row
	header, err := reader.Read()
	if err != nil {
		log.Printf("Error reading CSV header: %v\n", err)
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
		switch strings.ToLower(col) {
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
		
		task := Task{}
		
		// Extract data from CSV columns
		if dateIdx >= 0 && dateIdx < len(record) {
			task.Date = strings.Trim(record[dateIdx], "\"")
		}
		
		if idIdx >= 0 && idIdx < len(record) {
			task.ID = strings.Trim(record[idIdx], "\"")
		}
		
		if durationIdx >= 0 && durationIdx < len(record) {
			task.Duration = strings.Trim(record[durationIdx], "\"")
			task.DurationMins = parseTime(task.Duration)
		}
		
		if rateIdx >= 0 && rateIdx < len(record) {
			rateStr := strings.Trim(record[rateIdx], "\"")
			if strings.Contains(rateStr, "$") && strings.Contains(rateStr, "/hr") {
				rateVal := strings.TrimSuffix(strings.TrimPrefix(rateStr, "$"), "/hr")
				task.Rate, _ = strconv.ParseFloat(rateVal, 64)
			}
		}
		
		if valueIdx >= 0 && valueIdx < len(record) {
			valueStr := strings.Trim(record[valueIdx], "\"")
			if strings.Contains(valueStr, "$") {
				valueVal := strings.TrimPrefix(valueStr, "$")
				task.Value, _ = strconv.ParseFloat(valueVal, 64)
			}
		}
		
		if typeIdx >= 0 && typeIdx < len(record) {
			payType := strings.Trim(record[typeIdx], "\"")
			switch strings.ToLower(payType) {
			case "prepay":
				task.Type = "Task"
			case "overtimepay":
				task.Type = "Exceeded Time"
			case "missionreward":
				task.Type = "Mission Reward"
			default:
				task.Type = payType
			}
		}
		
		if projectIdx >= 0 && projectIdx < len(record) {
			task.Category = strings.Trim(record[projectIdx], "\"")
		}
		
		if statusIdx >= 0 && statusIdx < len(record) {
			task.Status = strings.Trim(record[statusIdx], "\"")
		}
		
		// Debug output
		log.Printf("CSV Parsed: Date=%s, ID=%s, Type=%s, Duration=%s, Rate=%.2f, Value=%.2f, DurationMins=%.2f\n",
			task.Date, task.ID, task.Type, task.Duration, task.Rate, task.Value, task.DurationMins)
		
		tasks = append(tasks, task)
	}
	
	log.Printf("Total de %d tarefas foram analisadas do CSV.\n", len(tasks))
	return tasks
}

func main() {
	// Define PORT environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Look for templates in multiple locations (local dev vs production)
	var tmplPath string
	templateLocations := []string{
		"templates/index.html",  // Local development
		"/templates/index.html", // Railway deployment
		"./templates/index.html", // Alternative path
	}

	// Try each location until we find the template
	var tmpl *template.Template
	var err error
	for _, loc := range templateLocations {
		log.Printf("Trying to load template from: %s\n", loc)
		tmpl, err = template.ParseFiles(loc)
		if err == nil {
			tmplPath = loc
			break
		}
	}

	// Check if we found the template
	if tmpl == nil {
		log.Fatalf("Failed to load template from any location: %v\n", err)
	}

	log.Printf("Successfully loaded template from: %s\n", tmplPath)

	// Handle root path - show the form
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			CurrentYear: time.Now().Year(),
		}
		tmpl.Execute(w, data)
	})

	// Handle form submission
	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Set max file size (10MB)
		r.ParseMultipartForm(10 << 20)
		
		// Get form data
		showDetails := r.FormValue("showDetails") == "on"
		
		var tasks []Task
		var inputSource string
		var taskData string
		
		// Check for file upload first
		file, handler, err := r.FormFile("csvFile")
		if err == nil {
			defer file.Close()
			log.Printf("Uploaded File: %+v\n", handler.Filename)
			log.Printf("File Size: %+v\n", handler.Size)
			log.Printf("MIME Header: %+v\n", handler.Header)
			
			// Parse the CSV file
			tasks = parseCSV(file)
			inputSource = "csv"
			// Save CSV content for display
			file.Seek(0, 0) // Rewind file for reading again
			buffer := new(strings.Builder)
			io.Copy(buffer, file)
			taskData = buffer.String()
		} else {
			// Fall back to text input if no file was uploaded
			taskData = r.FormValue("taskData")
			if taskData != "" {
				tasks = parseInput(taskData)
				inputSource = "text"
			}
		}

		// Analyze the data if we have tasks
		results := make(map[string]interface{})
		if len(tasks) > 0 {
			results = analyzeData(tasks)
		}

		// Convert tasks for display
		var taskDisplays []TaskDisplay
		if showDetails && len(tasks) > 0 {
			for _, task := range tasks {
				taskDisplays = append(taskDisplays, TaskDisplay{
					Date:         task.Date,
					ID:           task.ID,
					Category:     task.Category,
					Duration:     task.Duration,
					Rate:         fmt.Sprintf("$%.2f/hr", task.Rate),
					Value:        fmt.Sprintf("$%.2f", task.Value),
					Type:         task.Type,
					Status:       task.Status,
					DurationMins: fmt.Sprintf("%.2f mins", task.DurationMins),
				})
			}
		}

		// Initialize data structure
		data := TemplateData{
			RawInput:    taskData,
			HasResults:  len(tasks) > 0,
			InputSource: inputSource,
			CurrentYear: time.Now().Year(),
			ShowDetails: showDetails,
			Tasks:       taskDisplays,
		}
		
		// Format results for display if we have any
		if len(tasks) > 0 {
			totalHoursValue := results["TotalHours"].(float64)
			totalHoursInt := int(totalHoursValue)
			totalMinutes := int((totalHoursValue - float64(totalHoursInt)) * 60)
			
			data.TotalTasks = results["TotalTasks"].(int)
			data.TotalHours = fmt.Sprintf("%.2f horas (%dh %dmin)", totalHoursValue, totalHoursInt, totalMinutes)
			data.TotalValue = fmt.Sprintf("%.2f", results["TotalValue"].(float64))
			data.TasksValue = fmt.Sprintf("%.2f", results["TasksValue"].(float64))
			data.ExceededTimeValue = fmt.Sprintf("%.2f", results["ExceededTimeValue"].(float64))
			data.OtherValue = fmt.Sprintf("%.2f", results["OtherValue"].(float64))
			data.AverageHourlyRate = fmt.Sprintf("%.2f", results["AverageHourlyRate"].(float64))
		}

		tmpl.Execute(w, data)
	})

	// Add health check endpoint for Railway
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve static files from the data directory
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	// Create server with reasonable timeouts
	srv := &http.Server{
		Addr:         ":"+port,
		Handler:      nil, // Use default mux
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it doesn't block
	go func() {
		log.Printf("Servidor iniciado na porta %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	sig := <-c
	log.Printf("Received signal %v, shutting down\n", sig)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	log.Println("Server gracefully stopped")
}

func parseInput(input string) []Task {
	var tasks []Task
	lines := strings.Split(input, "\n")

	for i := 0; i < len(lines); {
		// Skip empty lines
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}

		task, advance := parseTask(lines[i:])
		if task != nil {
			tasks = append(tasks, *task)
		}
		i += advance
	}

	// Only output to logs, not to web interface
	log.Printf("Total de %d tarefas foram analisadas.\n", len(tasks))
	return tasks
}

func parseTask(lines []string) (*Task, int) {
	// Skip if we don't have enough lines
	if len(lines) < 8 {
		return nil, len(lines)
	}

	// Check if this is a valid task block by looking for the pattern of data
	// Date, ID, Category, empty line, details, type, empty line, status
	if strings.TrimSpace(lines[3]) != "" || strings.TrimSpace(lines[6]) != "" {
		return nil, 1 // Not a task block, advance by 1
	}

	task := &Task{
		Date:     strings.TrimSpace(lines[0]),
		ID:       strings.TrimSpace(lines[1]),
		Category: strings.TrimSpace(lines[2]),
		Status:   strings.TrimSpace(lines[7]),
	}

	// Parse duration/rate/value line (line 4)
	durationRateValue := strings.TrimSpace(lines[4])

	// Split by tabs first to handle the tabbed format
	tabParts := strings.Split(durationRateValue, "\t")
	var parts []string

	// Clean up the parts by removing empty strings and trimming spaces
	for _, part := range tabParts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	// Handle different patterns based on the content
	if len(parts) >= 3 {
		// Parse duration (first part)
		task.Duration = parts[0]

		// Parse rate (second part, which should contain $/hr)
		if strings.Contains(parts[1], "$") && strings.Contains(parts[1], "/hr") {
			rateStr := strings.TrimSuffix(strings.TrimPrefix(parts[1], "$"), "/hr")
			task.Rate, _ = strconv.ParseFloat(rateStr, 64)
		}

		// Parse value (third part, which should start with $)
		if strings.HasPrefix(parts[2], "$") {
			valueStr := strings.TrimPrefix(parts[2], "$")
			task.Value, _ = strconv.ParseFloat(valueStr, 64)
		}
	} else if len(parts) >= 2 && parts[0] == "-" && parts[1] == "-" {
		// Special case for Mission Reward: "- - $92.75"
		if len(parts) > 2 && strings.HasPrefix(parts[2], "$") {
			valueStr := strings.TrimPrefix(parts[2], "$")
			task.Value, _ = strconv.ParseFloat(valueStr, 64)
		}
	}

	// Parse task type (line 5)
	task.Type = strings.TrimSpace(lines[5])

	// Calculate duration minutes
	if task.Duration != "" && task.Duration != "-" {
		task.DurationMins = parseTime(task.Duration)
	}

	// Debug output to logs, not web interface
	log.Printf("Parsed: Date=%s, ID=%s, Type=%s, Duration=%s, Rate=%.2f, Value=%.2f, DurationMins=%.2f\n",
		task.Date, task.ID, task.Type, task.Duration, task.Rate, task.Value, task.DurationMins)

	return task, 8 // Each task block is 8 lines
}

// We no longer need these functions as we're directly using the trimmed string content

func analyzeData(tasks []Task) map[string]interface{} {
	totalTasks := 0
	totalTasksValue := 0.0
	totalExceededTimeValue := 0.0
	totalOtherValue := 0.0
	totalHours := 0.0

	for _, task := range tasks {
		// Debug output to logs, not web interface
		log.Printf("Analisando: Type=%s, Value=%.2f, DurationMins=%.2f\n",
			task.Type, task.Value, task.DurationMins)

		// Only count actual Task items in the total tasks count
		if task.Type == "Task" {
			totalTasks++
		}

		// Include both Task and Exceeded Time in total hours calculation
		if task.Type == "Task" || task.Type == "Exceeded Time" {
			totalHours += task.DurationMins / 60
		}

		if task.Type == "Task" {
			totalTasksValue += task.Value
		} else if task.Type == "Exceeded Time" {
			totalExceededTimeValue += task.Value
		} else {
			totalOtherValue += task.Value
		}
	}

	totalValue := totalTasksValue + totalExceededTimeValue + totalOtherValue
	averageHourlyRate := 0.0
	if totalHours > 0 {
		averageHourlyRate = (totalTasksValue + totalExceededTimeValue) / totalHours
	}

	return map[string]interface{}{
		"TotalTasks":        totalTasks,
		"TotalHours":        totalHours,
		"TotalValue":        totalValue,
		"TasksValue":        totalTasksValue,
		"ExceededTimeValue": totalExceededTimeValue,
		"OtherValue":        totalOtherValue,
		"AverageHourlyRate": averageHourlyRate,
	}
}
