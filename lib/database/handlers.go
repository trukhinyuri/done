package database

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	database "done/lib/database/interface"
	"done/lib/utils"
	uuid "github.com/satori/go.uuid"
)

type Handler struct {
	DB database.Database
}

type addTaskRequest struct {
	UUID          string `json:"uuid"`
	Body          string `json:"body"`
	Estimation    int    `json:"estimation"`
	DeadlineMonth string `json:"deadlineMonth"`
	DeadlineDay   string `json:"deadlineDay"`
	DeadlineYear  string `json:"deadlineYear"`
}

type rearrangeTasksRequest struct {
	SourceUUID      string `json:"source_uuid"`
	DestinationUUID string `json:"destination_uuid"`
	InsertBefore    bool   `json:"insert_before"`
}

func NewHandler(db database.Database) *Handler {
	return &Handler{DB: db}
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func writeTasks(w http.ResponseWriter, db database.Database) {
	tasks, err := db.GetTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "failed to get tasks", http.StatusInternalServerError)
		return
	}
	writeJSON(w, tasks)
}

func parseOptionalInt(value, placeholder string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "0" || value == placeholder {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func parseAddTaskRequest(r *http.Request) (*addTaskRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return nil, fmt.Errorf("empty request body")
	}

	var request addTaskRequest
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal(body, &request); err != nil {
			return nil, err
		}
	} else {
		parts := strings.Split(raw, "$;")
		if len(parts) < 4 {
			return nil, fmt.Errorf("expected task body, estimation, month, and day")
		}

		estimation, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}

		request = addTaskRequest{
			Body:          parts[0],
			Estimation:    estimation,
			DeadlineMonth: parts[2],
			DeadlineDay:   parts[3],
		}
		if len(parts) > 4 {
			request.DeadlineYear = parts[4]
		}
	}

	request.Body = strings.TrimSpace(utils.CleanTaskText(request.Body))
	if request.Body == "" {
		return nil, fmt.Errorf("task body is required")
	}
	if request.Estimation < 0 {
		return nil, fmt.Errorf("estimation must be non-negative")
	}

	return &request, nil
}

func taskDeadlineFromRequest(request *addTaskRequest) (time.Time, error) {
	deadlineMonth, err := parseOptionalInt(request.DeadlineMonth, "MM")
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid deadline month")
	}

	deadlineDay, err := parseOptionalInt(request.DeadlineDay, "DD")
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid deadline day")
	}

	deadlineYear, err := parseOptionalInt(request.DeadlineYear, "YYYY")
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid deadline year")
	}

	if deadlineMonth < 0 || deadlineMonth > 12 {
		return time.Time{}, fmt.Errorf("deadline month must be between 1 and 12")
	}
	if deadlineDay < 0 || deadlineDay > 31 {
		return time.Time{}, fmt.Errorf("deadline day must be between 1 and 31")
	}

	currentYear, currentMonth, _ := time.Now().Date()
	var taskDeadlineYear int

	if deadlineMonth == 0 && deadlineDay == 0 && deadlineYear == 0 {
		taskDeadlineYear = 9999
		deadlineMonth = 1
		deadlineDay = 1
	} else {
		if deadlineYear > 0 {
			taskDeadlineYear = deadlineYear
		} else if deadlineMonth < int(currentMonth) {
			taskDeadlineYear = currentYear + 1
		} else {
			taskDeadlineYear = currentYear
		}

		if deadlineMonth == 0 {
			deadlineMonth = int(time.Now().Month())
		}
		if deadlineDay == 0 {
			deadlineDay = 1
		}
	}

	deadline := time.Date(taskDeadlineYear, time.Month(deadlineMonth), deadlineDay, 0, 0, 0, 0, time.UTC)
	if deadline.Month() != time.Month(deadlineMonth) || deadline.Day() != deadlineDay || deadline.Year() != taskDeadlineYear {
		return time.Time{}, fmt.Errorf("invalid deadline date")
	}

	return deadline, nil
}

func hasAchievement(gamification *database.Gamification, achievementID string) bool {
	for _, existing := range gamification.Achievements {
		if existing == achievementID {
			return true
		}
	}
	return false
}

func addAchievement(gamification *database.Gamification, achievementID string) {
	if !hasAchievement(gamification, achievementID) {
		gamification.Achievements = append(gamification.Achievements, achievementID)
	}
}

func completedTasksOnDate(db database.Database, date time.Time) (int, error) {
	completedTasks, err := db.GetCompletedTasks()
	if err != nil {
		return 0, err
	}

	year, month, day := date.Date()
	count := 0
	for _, task := range completedTasks {
		taskYear, taskMonth, taskDay := task.TimeCompleted.Date()
		if taskYear == year && taskMonth == month && taskDay == day {
			count++
		}
	}
	return count, nil
}

func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	request, err := parseAddTaskRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deadline, err := taskDeadlineFromRequest(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks, err := h.DB.GetTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "failed to get tasks", http.StatusInternalServerError)
		return
	}

	for i := 0; i < len(tasks); i++ {
		tasks[i].Order++
		if err = h.DB.UpdateTask(&tasks[i]); err != nil {
			log.Printf("Error updating task order: %v", err)
			http.Error(w, "failed to update task order", http.StatusInternalServerError)
			return
		}
	}

	task := database.Task{
		UUID:                              uuid.NewV4().String(),
		Body:                              request.Body,
		DurationExecutionEstimatedSeconds: request.Estimation,
		TimeHardDeadline:                  deadline,
		TimeCreated:                       time.Now(),
		Order:                             0,
	}

	if err = h.DB.AddTask(&task); err != nil {
		log.Printf("Error adding task: %v", err)
		http.Error(w, "failed to add task", http.StatusInternalServerError)
		return
	}

	writeTasks(w, h.DB)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	request, err := parseAddTaskRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(request.UUID) == "" {
		http.Error(w, "task uuid is required", http.StatusBadRequest)
		return
	}

	deadline, err := taskDeadlineFromRequest(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.DB.GetTaskByUUID(request.UUID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	task.Body = request.Body
	task.DurationExecutionEstimatedSeconds = request.Estimation
	task.TimeHardDeadline = deadline

	if err = h.DB.UpdateTask(task); err != nil {
		log.Printf("Error updating task: %v", err)
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	writeTasks(w, h.DB)
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	tasks, err := h.DB.GetTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "failed to get tasks", http.StatusInternalServerError)
		return
	}

	writeJSON(w, tasks)
}

func (h *Handler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	uuid := strings.TrimSpace(string(body))
	if uuid == "" {
		http.Error(w, "task uuid is required", http.StatusBadRequest)
		return
	}

	if err = h.DB.RemoveTask(uuid); err != nil {
		log.Printf("Error removing task: %v", err)
		http.Error(w, "failed to remove task", http.StatusInternalServerError)
		return
	}

	writeTasks(w, h.DB)
}

func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	uuid := strings.TrimSpace(string(body))
	if uuid == "" {
		http.Error(w, "task uuid is required", http.StatusBadRequest)
		return
	}

	task, err := h.DB.CompleteTask(uuid, time.Now())
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	// Update gamification data
	gamification, err := h.DB.GetGamification()
	if err != nil {
		log.Printf("Error getting gamification data: %v", err)
		// Continue with the rest of the handler even if gamification fails
		gamification = &database.Gamification{
			Level:          1,
			TotalPoints:    0,
			CompletedTasks: 0,
		}
	}

	// Calculate points based on task complexity
	points := 10                                       // Base points
	if task.DurationExecutionEstimatedSeconds > 3600 { // More than 1 hour
		points = 25
	}
	if task.DurationExecutionEstimatedSeconds > 7200 { // More than 2 hours
		points = 50
	}

	// Bonus points for completing on time
	if task.TimeHardDeadline.Year() != 9999 && task.TimeCompleted.Before(task.TimeHardDeadline) {
		points += 10
	}

	// Update gamification stats
	gamification.TotalPoints += points
	gamification.CompletedTasks++

	// Update first task date if not set
	if gamification.FirstTaskDate == nil {
		now := time.Now()
		gamification.FirstTaskDate = &now
	}

	// Calculate level (100 points per level)
	gamification.Level = (gamification.TotalPoints / 100) + 1

	// Update streak
	today := time.Now().Truncate(24 * time.Hour)
	if gamification.LastCompletionDate != nil {
		lastDate := gamification.LastCompletionDate.Truncate(24 * time.Hour)
		daysSince := int(today.Sub(lastDate).Hours() / 24)

		if daysSince == 0 {
			// Same day, streak continues
		} else if daysSince == 1 {
			// Next day, increment streak
			gamification.CurrentStreak++
		} else {
			// Streak broken
			gamification.CurrentStreak = 1
		}
	} else {
		// First task
		gamification.CurrentStreak = 1
	}

	// Update longest streak
	if gamification.CurrentStreak > gamification.LongestStreak {
		gamification.LongestStreak = gamification.CurrentStreak
	}

	// Update last completion date
	now := time.Now()
	gamification.LastCompletionDate = &now

	if gamification.CompletedTasks == 1 {
		addAchievement(gamification, "firstTask")
	}
	if gamification.CurrentStreak >= 3 {
		addAchievement(gamification, "streak3")
	}
	if gamification.CurrentStreak >= 7 {
		addAchievement(gamification, "streak7")
	}
	if gamification.CurrentStreak >= 30 {
		addAchievement(gamification, "streak30")
	}
	if gamification.TotalPoints >= 1000 {
		addAchievement(gamification, "points1000")
	}
	if gamification.TotalPoints >= 5000 {
		addAchievement(gamification, "points5000")
	}
	if task.TimeHardDeadline.Year() != 9999 && task.TimeCompleted.Before(task.TimeHardDeadline) {
		addAchievement(gamification, "earlyBird")
	}
	if completedToday, err := completedTasksOnDate(h.DB, now); err != nil {
		log.Printf("Error checking today's achievements: %v", err)
	} else if completedToday >= 5 {
		addAchievement(gamification, "speedDemon")
	}

	// Save gamification data
	err = h.DB.UpdateGamification(gamification)
	if err != nil {
		log.Printf("Error updating gamification data: %v", err)
		// Continue even if gamification update fails
	}

	// Save to file in ~/tasksReport/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home directory: %v", err)
		homeDir = "."
	}

	reportDir := filepath.Join(homeDir, "tasksReport")
	err = os.MkdirAll(reportDir, 0755)
	if err != nil {
		log.Printf("Error creating report directory: %v", err)
	}

	day := task.TimeCompleted.Format("02")
	month := task.TimeCompleted.Format("Jan")
	year := task.TimeCompleted.Format("2006")

	filenameComplete := filepath.Join(reportDir, year+"-"+month+"-"+day+".html")

	f, err := os.OpenFile(filenameComplete, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Printf("Error opening report file: %v", err)
		// Continue even if file save fails
	} else {
		defer f.Close()

		finfo, err := f.Stat()
		if err == nil && finfo.Size() == 0 {
			// Write HTML header with modern styling
			header := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Task Report - ` + year + `-` + month + `-` + day + `</title>
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    margin: 0;
    padding: 20px;
    background: #0a0e27;
    color: #ffffff;
    line-height: 1.6;
}
.container {
    max-width: 800px;
    margin: 0 auto;
}
h1 {
    color: #00ff41;
    font-size: 28px;
    margin-bottom: 30px;
    padding-bottom: 10px;
    border-bottom: 2px solid #00ff41;
}
.task-item {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 15px 20px;
    margin-bottom: 10px;
    transition: all 0.3s ease;
}
.task-item:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: #00ff41;
}
.task-time {
    color: #00ff41;
    font-size: 12px;
    font-weight: 600;
    margin-bottom: 5px;
}
.task-body {
    color: #ffffff;
    font-size: 16px;
}
.task-duration {
    color: #888;
    font-size: 14px;
    margin-top: 8px;
}
.summary {
    background: rgba(0, 255, 65, 0.1);
    border: 2px solid #00ff41;
    border-radius: 8px;
    padding: 20px;
    margin-top: 30px;
}
.summary h2 {
    color: #00ff41;
    margin-top: 0;
}
</style>
</head>
<body>
<div class="container">
<h1>Task Report - ` + year + `-` + month + `-` + day + `</h1>
`
			io.WriteString(f, header)
		}

		// Format task completion time and duration
		hours := task.DurationExecutionRealSeconds / 3600
		minutes := (task.DurationExecutionRealSeconds % 3600) / 60
		seconds := task.DurationExecutionRealSeconds % 60

		durationStr := ""
		if hours > 0 {
			durationStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
		} else if minutes > 0 {
			durationStr = fmt.Sprintf("%dm %ds", minutes, seconds)
		} else {
			durationStr = fmt.Sprintf("%ds", seconds)
		}

		// Clean task body to remove any encoding artifacts
		cleanBody := html.EscapeString(utils.CleanTaskText(task.Body))

		taskHTML := fmt.Sprintf(`<div class="task-item">
    <div class="task-time">Completed at %s</div>
    <div class="task-body">%s</div>
    <div class="task-duration">Time spent: %s</div>
</div>
`, task.TimeCompleted.Format("15:04:05"), cleanBody, durationStr)

		io.WriteString(f, taskHTML)
	}

	writeTasks(w, h.DB)
}

func (h *Handler) RearrangeTasks(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var request rearrangeTasksRequest
	raw := strings.TrimSpace(string(body))
	if strings.HasPrefix(raw, "{") {
		if err := json.Unmarshal(body, &request); err != nil {
			http.Error(w, "invalid rearrange request", http.StatusBadRequest)
			return
		}
	} else {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 {
			http.Error(w, "expected source and destination uuid", http.StatusBadRequest)
			return
		}
		request.SourceUUID = strings.TrimSpace(parts[0])
		request.DestinationUUID = strings.TrimSpace(parts[1])
		request.InsertBefore = false
	}

	if request.SourceUUID == "" || request.DestinationUUID == "" {
		http.Error(w, "source and destination uuid are required", http.StatusBadRequest)
		return
	}
	if request.SourceUUID == request.DestinationUUID {
		writeTasks(w, h.DB)
		return
	}

	tasks, err := h.DB.GetTasks()
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		http.Error(w, "failed to get tasks", http.StatusInternalServerError)
		return
	}

	var source *database.Task
	remaining := make([]database.Task, 0, len(tasks))
	for i := range tasks {
		task := tasks[i]
		if task.UUID == request.SourceUUID {
			source = &task
			continue
		}
		remaining = append(remaining, task)
	}
	if source == nil {
		http.Error(w, "source task not found", http.StatusNotFound)
		return
	}

	destinationIndex := -1
	for i := range remaining {
		if remaining[i].UUID == request.DestinationUUID {
			destinationIndex = i
			break
		}
	}
	if destinationIndex == -1 {
		http.Error(w, "destination task not found", http.StatusNotFound)
		return
	}

	insertIndex := destinationIndex
	if !request.InsertBefore {
		insertIndex = destinationIndex + 1
	}

	reordered := append(remaining[:insertIndex], append([]database.Task{*source}, remaining[insertIndex:]...)...)
	for i := range reordered {
		reordered[i].Order = i
		if err = h.DB.UpdateTask(&reordered[i]); err != nil {
			log.Printf("Error updating task order: %v", err)
			http.Error(w, "failed to update task order", http.StatusInternalServerError)
			return
		}
	}

	writeTasks(w, h.DB)
}

func (h *Handler) UpdateTaskExecutionRealSeconds(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	updateTaskJSON, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	updateTaskJSONString := string(updateTaskJSON)
	updateTaskSplitted := strings.Split(updateTaskJSONString, "$;")
	if len(updateTaskSplitted) != 2 {
		http.Error(w, "expected uuid and seconds", http.StatusBadRequest)
		return
	}

	uuid := strings.TrimSpace(updateTaskSplitted[0])
	seconds, err := strconv.Atoi(updateTaskSplitted[1])
	if err != nil || seconds < 0 {
		http.Error(w, "invalid seconds", http.StatusBadRequest)
		return
	}

	task, err := h.DB.GetTaskByUUID(uuid)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	task.DurationExecutionRealSeconds = seconds

	err = h.DB.UpdateTask(task)
	if err != nil {
		log.Printf("Error updating task timer: %v", err)
		http.Error(w, "failed to update task timer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetTodayResults(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	tasksCompleted, err := h.DB.GetCompletedTasks()
	if err != nil {
		log.Printf("Error getting completed tasks: %v", err)
		http.Error(w, "failed to get completed tasks", http.StatusInternalServerError)
		return
	}

	var tasksCompletedToday []database.Task

	var todayYear, todayMonth, todayDay = time.Now().Date()

	for i := 0; i < len(tasksCompleted); i++ {
		year, month, day := tasksCompleted[i].TimeCompleted.Date()
		if (year == todayYear) && (month == todayMonth) && (day == todayDay) {
			tasksCompletedToday = append(tasksCompletedToday, tasksCompleted[i])
		}
	}

	sort.Slice(tasksCompletedToday, func(i, j int) bool {
		return tasksCompletedToday[i].TimeCompleted.Before(tasksCompletedToday[j].TimeCompleted)
	})

	writeJSON(w, tasksCompletedToday)
}

func saveReport(completedTasks []database.Task) {
	// Get home directory for report path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home directory: %v", err)
		homeDir = "."
	}

	reportDir := filepath.Join(homeDir, "tasksReport")
	err = os.MkdirAll(reportDir, 0755)
	if err != nil {
		log.Printf("Error creating report directory: %v", err)
	}

	t := time.Now()
	day := t.Format("02")
	month := t.Format("Jan")
	year := t.Format("2006")

	filenameComplete := filepath.Join(reportDir, year+"-"+month+"-"+day+"-summary.html")

	f, err := os.Create(filenameComplete)
	if err != nil {
		log.Printf("Error creating summary report: %v", err)
		return
	}
	defer f.Close()

	// Calculate totals
	totalTasks := len(completedTasks)
	totalSeconds := 0
	for _, task := range completedTasks {
		totalSeconds += task.DurationExecutionRealSeconds
	}

	totalHours := totalSeconds / 3600
	totalMinutes := (totalSeconds % 3600) / 60

	// Write styled HTML report
	header := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Daily Summary - ` + year + `-` + month + `-` + day + `</title>
<style>
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    margin: 0;
    padding: 20px;
    background: #0a0e27;
    color: #ffffff;
    line-height: 1.6;
}
.container {
    max-width: 800px;
    margin: 0 auto;
}
h1 {
    color: #00ff41;
    font-size: 32px;
    margin-bottom: 30px;
    padding-bottom: 10px;
    border-bottom: 2px solid #00ff41;
}
.summary {
    background: rgba(0, 255, 65, 0.1);
    border: 2px solid #00ff41;
    border-radius: 12px;
    padding: 25px;
    margin-bottom: 30px;
}
.summary h2 {
    color: #00ff41;
    margin-top: 0;
    font-size: 24px;
}
.stat-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
    margin-top: 20px;
}
.stat-box {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 20px;
    text-align: center;
}
.stat-value {
    font-size: 36px;
    font-weight: bold;
    color: #00ff41;
    margin-bottom: 5px;
}
.stat-label {
    color: #888;
    font-size: 14px;
    text-transform: uppercase;
}
.task-list {
    margin-top: 30px;
}
.task-list h2 {
    color: #00ff41;
    font-size: 24px;
    margin-bottom: 20px;
}
.task-item {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 15px 20px;
    margin-bottom: 10px;
    transition: all 0.3s ease;
}
.task-item:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: #00ff41;
}
.task-time {
    color: #00ff41;
    font-size: 12px;
    font-weight: 600;
    margin-bottom: 5px;
}
.task-body {
    color: #ffffff;
    font-size: 16px;
}
.task-duration {
    color: #888;
    font-size: 14px;
    margin-top: 8px;
}
.footer {
    text-align: center;
    margin-top: 50px;
    padding-top: 20px;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    color: #666;
    font-size: 14px;
}
</style>
</head>
<body>
<div class="container">
<h1>Daily Task Summary</h1>
<div class="summary">
<h2>` + t.Format("Monday, January 2, 2006") + `</h2>
<div class="stat-grid">
    <div class="stat-box">
        <div class="stat-value">` + strconv.Itoa(totalTasks) + `</div>
        <div class="stat-label">Tasks Completed</div>
    </div>
    <div class="stat-box">
        <div class="stat-value">` + fmt.Sprintf("%dh %dm", totalHours, totalMinutes) + `</div>
        <div class="stat-label">Total Time</div>
    </div>
    <div class="stat-box">
        <div class="stat-value">OK</div>
        <div class="stat-label">Daily Summary</div>
    </div>
</div>
</div>
<div class="task-list">
<h2>Completed Tasks</h2>
`
	io.WriteString(f, header)

	// Write each task
	for _, task := range completedTasks {
		hours := task.DurationExecutionRealSeconds / 3600
		minutes := (task.DurationExecutionRealSeconds % 3600) / 60
		seconds := task.DurationExecutionRealSeconds % 60

		durationStr := ""
		if hours > 0 {
			durationStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
		} else if minutes > 0 {
			durationStr = fmt.Sprintf("%dm %ds", minutes, seconds)
		} else {
			durationStr = fmt.Sprintf("%ds", seconds)
		}

		// Clean task body to remove any encoding artifacts
		cleanBody := html.EscapeString(utils.CleanTaskText(task.Body))

		taskHTML := fmt.Sprintf(`<div class="task-item">
    <div class="task-time">Completed at %s</div>
    <div class="task-body">%s</div>
    <div class="task-duration">Time spent: %s</div>
</div>
`, task.TimeCompleted.Format("15:04:05"), cleanBody, durationStr)
		io.WriteString(f, taskHTML)
	}

	// Write footer
	footer := `</div>
<div class="footer">
    Generated by Done Task Manager
</div>
</div>
</body>
</html>`
	io.WriteString(f, footer)
}

func (h *Handler) GetGamification(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	gamification, err := h.DB.GetGamification()
	if err != nil {
		log.Printf("Error getting gamification: %v", err)
		http.Error(w, "Failed to get gamification data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(gamification)
}

func (h *Handler) UpdateGamification(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var gamification database.Gamification

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &gamification)
	if err != nil {
		log.Printf("Error unmarshaling gamification: %v", err)
		http.Error(w, "Invalid gamification data", http.StatusBadRequest)
		return
	}

	if gamification.TotalPoints < 0 ||
		gamification.CurrentStreak < 0 ||
		gamification.LongestStreak < 0 ||
		gamification.Level < 1 ||
		gamification.CompletedTasks < 0 {
		http.Error(w, "Invalid gamification data", http.StatusBadRequest)
		return
	}

	err = h.DB.UpdateGamification(&gamification)
	if err != nil {
		log.Printf("Error updating gamification: %v", err)
		http.Error(w, "Failed to update gamification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
