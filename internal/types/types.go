package types

// Task represents a single task entry
type Task struct {
	Date         string
	ID           string
	Category     string
	Duration     string
	Rate         float64
	Value        float64
	Type         string // Task, Exceeded Time, Mission Reward, Operation
	Status       string
	DurationMins float64 // Duration converted to minutes
}

// TemplateData holds data to be passed to HTML templates
type TemplateData struct {
	RawInput          string
	HasResults        bool
	TotalTasks        int
	TotalHours        string // Formatted string (e.g., "X.XX horas (Yh Zmin)")
	TotalValue        string // Formatted string (e.g., "X.XX")
	TasksValue        string
	ExceededTimeValue string
	OtherValue        string
	AverageHourlyRate string
	CurrentYear       int
	InputSource       string // "csv" or "text"
	// Detailed hour breakdowns (formatted strings)
	TaskHours         string
	ExceededTimeHours string
	OtherHours        string
	// Average metrics (formatted strings)
	AvgTimePerTask  string // Formatted string (e.g., "Xm Ys")
	AvgValuePerTask string // Formatted string (e.g., "$X.XX")
	// For visualization (progress bars)
	RawHourPercentages []float64 // Task%, ExceededTime%, Other%
	// Task details section
	ShowDetails bool
	Tasks       []TaskDisplay // Tasks formatted for display
}

// TaskDisplay represents a task formatted for display in the HTML table
type TaskDisplay struct {
	Date         string
	ID           string
	Category     string
	Duration     string // Original duration string or "-"
	Rate         string // Formatted string (e.g., "$X.XX/hr" or "-")
	Value        string // Formatted string (e.g., "$X.XX")
	Type         string
	Status       string
	DurationMins string // Formatted string (e.g., "X.XX mins" or "-")
}
