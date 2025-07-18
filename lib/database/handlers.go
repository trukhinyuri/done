package database

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func NewHandler(db database.Database) *Handler {
	return &Handler{DB: db}
}

func errHandler(err error) {
	if err != nil {
		panic(err)
	}
}

func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.DB.GetTasks()
	errHandler(err)

	// Increment order for all existing tasks
	for i := 0; i < len(tasks); i++ {
		tasks[i].Order++
		err = h.DB.UpdateTask(&tasks[i])
		errHandler(err)
	}

	newTaskJSON, err := ioutil.ReadAll(r.Body)
	errHandler(err)

	newTaskJSONString := string(newTaskJSON)
	newTaskSplitted := strings.Split(newTaskJSONString, "$;")
	body := utils.CleanTaskText(newTaskSplitted[0]) // Decode and clean the task text
	durationExecutionEstimatedSeconds, err := strconv.Atoi(newTaskSplitted[1])
	errHandler(err)
	
	// Parse deadline fields with validation
	deadlineMonth := 0
	if newTaskSplitted[2] != "" && newTaskSplitted[2] != "MM" {
		deadlineMonth, err = strconv.Atoi(newTaskSplitted[2])
		if err != nil {
			log.Printf("Invalid deadline month: %s, using 0\n", newTaskSplitted[2])
			deadlineMonth = 0
		}
	}
	
	deadlineDay := 0
	if newTaskSplitted[3] != "" && newTaskSplitted[3] != "DD" {
		deadlineDay, err = strconv.Atoi(newTaskSplitted[3])
		if err != nil {
			log.Printf("Invalid deadline day: %s, using 0\n", newTaskSplitted[3])
			deadlineDay = 0
		}
	}
	
	// Handle year parameter if provided
	var deadlineYear int
	if len(newTaskSplitted) > 4 && newTaskSplitted[4] != "" && newTaskSplitted[4] != "YYYY" {
		deadlineYear, err = strconv.Atoi(newTaskSplitted[4])
		if err != nil {
			log.Printf("Invalid deadline year: %s, using 0\n", newTaskSplitted[4])
			deadlineYear = 0
		}
	} else {
		deadlineYear = 0
	}

	var task database.Task
	task.UUID = uuid.NewV4().String()
	task.Body = body
	task.DurationExecutionEstimatedSeconds = durationExecutionEstimatedSeconds

	currentYear, currentMonth, _ := time.Now().Date()
	var taskDeadlineYear int

	if (deadlineMonth == 0) && (deadlineDay == 0) && (deadlineYear == 0) {
		// No deadline set
		taskDeadlineYear = 9999
		deadlineMonth = 1
		deadlineDay = 1
	} else {
		// If year is explicitly provided, use it
		if deadlineYear > 0 {
			taskDeadlineYear = deadlineYear
		} else {
			// Legacy behavior: guess the year based on month
			if deadlineMonth < int(currentMonth) {
				taskDeadlineYear = currentYear + 1
			} else {
				taskDeadlineYear = currentYear
			}
		}
		
		// Set defaults for missing values
		if deadlineMonth == 0 {
			deadlineMonth = int(time.Now().Month())
		}
		if deadlineDay == 0 {
			deadlineDay = 1
		}
	}

	task.TimeHardDeadline = time.Date(taskDeadlineYear, time.Month(deadlineMonth), deadlineDay, 0, 0, 0, 0, time.UTC)

	timeNow := time.Now()
	task.TimeCreated = timeNow
	task.Order = 0

	err = h.DB.AddTask(&task)
	errHandler(err)

	tasks, err = h.DB.GetTasks()
	errHandler(err)

	tasksJSON, err := json.Marshal(tasks)
	errHandler(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(tasksJSON)
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.DB.GetTasks()
	errHandler(err)

	tasksJSON, err := json.Marshal(tasks)
	errHandler(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(tasksJSON)
}

func (h *Handler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	uuid, err := ioutil.ReadAll(r.Body)
	errHandler(err)

	err = h.DB.RemoveTask(string(uuid))
	errHandler(err)

	tasks, err := h.DB.GetTasks()
	errHandler(err)

	tasksJSON, err := json.Marshal(tasks)
	errHandler(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(tasksJSON)
}

func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	uuid, err := ioutil.ReadAll(r.Body)
	errHandler(err)

	task, err := h.DB.GetTaskByUUID(string(uuid))
	errHandler(err)

	err = h.DB.RemoveTask(string(uuid))
	errHandler(err)

	task.TimeCompleted = time.Now()

	err = h.DB.AddCompletedTask(task)
	errHandler(err)

	// Update gamification data
	gamification, err := h.DB.GetGamification()
	if err != nil {
		log.Printf("Error getting gamification data: %v", err)
		// Continue with the rest of the handler even if gamification fails
		gamification = &database.Gamification{
			Level: 1,
			TotalPoints: 0,
			CompletedTasks: 0,
		}
	}

	// Calculate points based on task complexity
	points := 10 // Base points
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
<h1>📋 Task Report - ` + year + `-` + month + `-` + day + `</h1>
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
		cleanBody := utils.CleanTaskText(task.Body)
		
		taskHTML := fmt.Sprintf(`<div class="task-item">
    <div class="task-time">✅ Completed at %s</div>
    <div class="task-body">%s</div>
    <div class="task-duration">⏱️ Time spent: %s</div>
</div>
`, task.TimeCompleted.Format("15:04:05"), cleanBody, durationStr)

		io.WriteString(f, taskHTML)
	}

	tasks, err := h.DB.GetTasks()
	errHandler(err)

	tasksJSON, err := json.Marshal(tasks)
	errHandler(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(tasksJSON)
}

func (h *Handler) RearrangeTasks(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	errHandler(err)

	strBody := string(body)
	sourceTaskUUID := ""
	destinationTaskPosition := 0

	for i := 0; i < len(strBody); i++ {
		if strBody[i] != ',' {
			sourceTaskUUID += string(strBody[i])
		} else {
			destinationTaskPosition = i + 1
			break
		}
	}

	destinationTaskUUID := strBody[destinationTaskPosition:]

	sourceTask, err := h.DB.GetTaskByUUID(sourceTaskUUID)
	errHandler(err)

	destinationTask, err := h.DB.GetTaskByUUID(destinationTaskUUID)
	errHandler(err)

	tasks, err := h.DB.GetTasks()
	errHandler(err)

	sourceOrder := sourceTask.Order
	destinationOrder := destinationTask.Order
	
	log.Printf("RearrangeTasks: Moving task %s (order %d) to position of task %s (order %d)\n", 
		sourceTaskUUID, sourceOrder, destinationTaskUUID, destinationOrder)

	if sourceTaskUUID != destinationTaskUUID {
		// Moving down in the list (to a higher order number)
		if sourceOrder < destinationOrder {
			for i := 0; i < len(tasks); i++ {
				if tasks[i].UUID == sourceTaskUUID {
					tasks[i].Order = destinationOrder
				} else if tasks[i].Order > sourceOrder && tasks[i].Order <= destinationOrder {
					tasks[i].Order--
				}
			}
		} else if sourceOrder > destinationOrder {
			// Moving up in the list (to a lower order number)
			for i := 0; i < len(tasks); i++ {
				if tasks[i].UUID == sourceTaskUUID {
					tasks[i].Order = destinationOrder
				} else if tasks[i].Order >= destinationOrder && tasks[i].Order < sourceOrder {
					tasks[i].Order++
				}
			}
		}
		
		// Log the new order
		log.Printf("RearrangeTasks: New order assigned - task %s now has order %d\n", 
			sourceTaskUUID, destinationOrder)
	}

	for i := 0; i < len(tasks); i++ {
		err = h.DB.UpdateTask(&tasks[i])
		errHandler(err)
	}

	tasks, err = h.DB.GetTasks()
	errHandler(err)
	
	// Log final task order
	log.Println("RearrangeTasks: Final task order:")
	for _, task := range tasks {
		log.Printf("  - %s: order %d\n", task.UUID, task.Order)
	}

	tasksJSON, err := json.Marshal(tasks)
	errHandler(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(tasksJSON)
}

func (h *Handler) UpdateTaskExecutionRealSeconds(w http.ResponseWriter, r *http.Request) {
	updateTaskJSON, err := ioutil.ReadAll(r.Body)
	errHandler(err)

	updateTaskJSONString := string(updateTaskJSON)
	updateTaskSplitted := strings.Split(updateTaskJSONString, "$;")

	uuid := updateTaskSplitted[0]
	seconds, err := strconv.Atoi(updateTaskSplitted[1])
	errHandler(err)

	task, err := h.DB.GetTaskByUUID(uuid)
	errHandler(err)

	task.DurationExecutionRealSeconds = seconds

	err = h.DB.UpdateTask(task)
	errHandler(err)
}

func (h *Handler) GetTodayResults(w http.ResponseWriter, r *http.Request) {
	tasksCompleted, err := h.DB.GetCompletedTasks()
	errHandler(err)

	var tasksCompletedToday []database.Task

	var todayYear, todayMonth, todayDay = time.Now().Date()

	for i := 0; i < len(tasksCompleted); i++ {
		year, month, day := tasksCompleted[i].TimeCompleted.Date()
		if (year == todayYear) && (month == todayMonth) && (day == todayDay) {
			tasksCompletedToday = append(tasksCompletedToday, tasksCompleted[i])
		}
	}

	saveReport(tasksCompletedToday)

	tasksJSON, err := json.Marshal(tasksCompletedToday)
	errHandler(err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(tasksJSON)
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
<h1>📊 Daily Task Summary</h1>
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
        <div class="stat-value">🔥</div>
        <div class="stat-label">Great Work!</div>
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
		cleanBody := utils.CleanTaskText(task.Body)
		
		taskHTML := fmt.Sprintf(`<div class="task-item">
    <div class="task-time">✅ Completed at %s</div>
    <div class="task-body">%s</div>
    <div class="task-duration">⏱️ Time spent: %s</div>
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
	var gamification database.Gamification
	
	body, err := ioutil.ReadAll(r.Body)
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
