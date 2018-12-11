package mongodb

import (
	"os"
	"gopkg.in/mgo.v2"
	"fmt"
		"time"
		"io/ioutil"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"io"
	"strings"
	"github.com/satori/go.uuid"
	"strconv"
	"net/http"
)

type Task struct {
	UUID string `json:"uuid"`
	Body string `json:"body"`
	TimeCreated time.Time `json:"timecreated"`
	TimeCompleted time.Time `json:"timecompleted"`
	DurationExecutionEstimatedSeconds int `json:"duration_execution_estimated_seconds"`
	DurationExecutionRealSeconds int `json:"duration_execution_real_seconds"`
	TimeHardDeadline time.Time `json:"time_hard_dead_line"`
	Order int `json:"order"`
	Child []Task `json:"-"`
}

type NewTask struct {
	Body string `json:"body"`
	EstimationDays string `json:"estimationDays"`
	estimationHours string `json:"estimationHours"`
	deadlineMonth string `json:"deadlineMonth"`
	deadlineDay string `json:"deadlineDay"`
}

type TaskOld struct {
	UUID string `json:"uuid"`
	Body string `json:"body"`
	TimeCreated time.Time `json:"timecreated"`
	TimeCompleted time.Time `json:"timecompleted"`
	TimeExecutionEstimated time.Time `json:"timeexecutionestimated"`
	TimeExecutionReal time.Time `json:"timeexecutionreal "`
	Child []Task `json:"-"`
}

type TaskCompleted struct {
	UUID string `json:"uuid"`
	Body string `json:"body"`
	TimeCreated time.Time `json:"timecreated"`
	TimeCompleted time.Time `json:"timecompleted"`
	TimeExecutionEstimated time.Time `json:"timeexecutionestimated"`
	TimeExecutionReal time.Time `json:"timeexecutionreal "`
	Child []Task `json:"-"`
}

type RearrangeTasksData struct {
	sourceTaskUUID string `json:"sourceTaskUUID"`
	destinationTaskUUID string `json:"destinationTaskUUID"`
}

func DBUpgrade() string {
	//session := MongoConnect();
	//dbDone := session.DB("done")
	//cTasks := dbDone.C("tasks")
	//
	//var tasksOld []TaskOld;
	//
	//err := cTasks.Find(nil).All(&tasksOld);
	//errHandler(err);
	//
	//for i := 0; i < len(tasksOld); i++ {
	//	taskold := tasksOld[i];
	//	var task Task;
	//    task.UUID = taskold.UUID;
	//	task.Body = taskold.Body;
	//    task.TimeCompleted = taskold.TimeCompleted;
	//	task.TimeCreated = taskold.TimeCreated;
	//	task.TimeExecutionEstimated = taskold.TimeExecutionEstimated;
	//	task.TimeExecutionReal = taskold.TimeExecutionReal;
	//	task.Child = taskold.Child;
	//	task.Order = i;
	//
	//	err = cTasks.Remove(taskold);
	//	errHandler(err);
	//
	//	err = cTasks.Insert(task);
	//	errHandler(err);
	//}
	//
	//defer MongoDisconnect(session);
	//return "DBUpgrade completed.";
	return "DBUpgrade not required"
}

func errHandler (err error) {
	if err != nil {
		panic(err)
	}
}

func MongoConnect() *mgo.Session{
	uri := os.Getenv("MONGODB_URL")
	if uri == "" {
		uri = "mongodb://localhost/done"
	}
	session, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}

	return session;
}

func MongoDisconnect(session *mgo.Session) {
	session.Close();
}

func AddTask(w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasks := dbDone.C("tasks")

	var tasks []Task;

	err := cTasks.Find(nil).All(&tasks);
	errHandler(err);

	for i := 0; i < len(tasks); i++ {
		task := tasks[i];
		tasks[i].Order ++
		err = cTasks.Update(task, tasks[i])
	}

	newTaskJSON, err := ioutil.ReadAll(r.Body)
	errHandler(err);

	newTaskJSONString := string(newTaskJSON);
	newTaskSplitted := strings.Split(newTaskJSONString, "$;")
	body := newTaskSplitted[0];
	durationExecutionEstimatedSeconds, err := strconv.Atoi(newTaskSplitted[1]);
	errHandler(err)
	deadlineMonth, err := strconv.Atoi(newTaskSplitted[2]);
	errHandler(err)
	deadlineDay, err := strconv.Atoi(newTaskSplitted[3])
	errHandler(err)

	var task Task
	task.UUID = uuid.NewV4().String();
	task.Body = body;
	task.DurationExecutionEstimatedSeconds = durationExecutionEstimatedSeconds;

	currentYear, currentMonth, currentDay := time.Now().Date();
	var taskDeadlineYear int

	if (deadlineMonth == 0) && (deadlineDay == 0) {
		taskDeadlineYear = 9999
		deadlineMonth = 1
		deadlineDay = 1
	} else {
		if deadlineMonth == 0 {
			deadlineMonth = int(time.Now().Month()+1);
		}

		if deadlineDay == 0 {
			deadlineDay = 1;
		}

		if deadlineMonth < int(currentMonth) {
			taskDeadlineYear = currentYear + 1;
		} else {
			taskDeadlineYear = currentYear;
		}

		if deadlineDay < int(currentDay) {
			deadlineMonth ++
		}
	}

	task.TimeHardDeadline = time.Date(taskDeadlineYear, time.Month(deadlineMonth), deadlineDay, 0, 0, 0,0, time.UTC)

	timeNow := time.Now();
	task.TimeCreated = timeNow;
	task.Order = 0;

	err = cTasks.Insert(task);
	errHandler(err);

	err = cTasks.Find(nil).All(&tasks);
	errHandler(err);

	tasksJSON, err := json.Marshal(tasks);
	errHandler(err);

	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON)

	defer MongoDisconnect(session);
}

func GetTasks (w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasks := dbDone.C("tasks")

	var tasks []Task;

	err := cTasks.Find(nil).All(&tasks);
	errHandler(err);

	tasksJSON, err := json.Marshal(tasks);
	errHandler(err);

	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON)

    defer MongoDisconnect(session);
}

func RemoveTask (w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasks := dbDone.C("tasks")

	uuid, err := ioutil.ReadAll(r.Body)
	errHandler(err);

	var taskForRemove Task;

	err = cTasks.Find(bson.M{"uuid": string(uuid)}).One(&taskForRemove)
	errHandler(err);


	err = cTasks.Remove(taskForRemove);
	errHandler(err);

	var tasks []Task;

	err = cTasks.Find(nil).All(&tasks);
	errHandler(err);

	tasksJSON, err := json.Marshal(tasks);
	errHandler(err);

	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON);
	defer MongoDisconnect(session);
}

func CompleteTask (w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasks := dbDone.C("tasks")
	cTasksCompleted := dbDone.C("tasks_completed")

	uuid, err := ioutil.ReadAll(r.Body)
	errHandler(err);

	var taskCompleted Task;
	err = cTasks.Find(bson.M{"uuid": string(uuid)}).One(&taskCompleted)
	errHandler(err);

	err = cTasks.Remove(taskCompleted);
	errHandler(err);

	taskCompleted.TimeCompleted = time.Now();

	err = cTasksCompleted.Insert(taskCompleted);
	errHandler(err);


	day := taskCompleted.TimeCompleted.Format("_2")
	month := taskCompleted.TimeCompleted.Format("Jan")
	year := taskCompleted.TimeCompleted.Format("2006")

	filenameComplete := year + "-" + month + "-" + day + ".html"

	f, err := os.OpenFile(filenameComplete, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		fmt.Println(err)
	}

	finfo, err := f.Stat();

	if finfo.Size() == 0 {
		io.WriteString(f, "<!DOCTYPE html>\n")
		io.WriteString(f, "<html>\n")
		io.WriteString(f, "<head>\n")
		io.WriteString(f, "<meta charset=\"utf-8\">\n")
		io.WriteString(f, "</head>\n")
		io.WriteString(f, "<body>\n")
		io.WriteString(f, "</body>\n")
		io.WriteString(f, "</html>\n")
	}


	n, err := io.WriteString(f, "[completed: " + taskCompleted.TimeCompleted.Format("2006–01–02 Mon 15:04:05") + "] "  + taskCompleted.Body + "<br>")
	if err != nil {
		fmt.Println(n, err)
	}
	f.Close()

	var tasks []Task;

	err = cTasks.Find(nil).All(&tasks);
	errHandler(err);

	tasksJSON, err := json.Marshal(tasks);
	errHandler(err);

	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON)
	defer MongoDisconnect(session);
}

func RearrangeTasks(w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasks := dbDone.C("tasks")

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

	destinationTaskUUID := strBody[destinationTaskPosition:];


	var rearrangeTasksData RearrangeTasksData
	rearrangeTasksData.sourceTaskUUID = sourceTaskUUID;
	rearrangeTasksData.destinationTaskUUID = destinationTaskUUID;


	var sourceTask Task;
	var destinationTask Task;

	var tasks []Task;

	err = cTasks.Find(nil).All(&tasks);
	errHandler(err);

	err = cTasks.Find(bson.M{"uuid": string(rearrangeTasksData.sourceTaskUUID)}).One(&sourceTask)
	errHandler(err);

	err = cTasks.Find(bson.M{"uuid": string(rearrangeTasksData.destinationTaskUUID)}).One(&destinationTask)
	errHandler(err);

	sourceOrder := sourceTask.Order;
	destinationOrder := destinationTask.Order;

	if rearrangeTasksData.sourceTaskUUID != rearrangeTasksData.destinationTaskUUID {
		if sourceOrder > destinationOrder {
			for i := 0; i < len(tasks); i++ {
				if tasks[i].UUID == sourceTaskUUID {
					tasks[i].Order = destinationOrder;
				} else {
					if tasks[i].Order >= destinationOrder {
						tasks[i].Order++;
					}
				}
			}
		} else if sourceOrder < destinationOrder {
			for i := 0; i < len(tasks); i++ {
				if tasks[i].UUID == sourceTaskUUID {
					tasks[i].Order = destinationOrder - 1;
				} else {
					if (tasks[i].Order > sourceOrder) && (tasks[i].Order < destinationOrder) {
						tasks[i].Order--
					}
				}
			}
		}
	}

	for i := 0; i < len(tasks); i++ {
		colQuerier := bson.M{"uuid": tasks[i].UUID};
		err = cTasks.Update(colQuerier, tasks[i]);
		errHandler(err);

	}

	err = cTasks.Find(nil).All(&tasks);
	errHandler(err);

	tasksJSON, err := json.Marshal(tasks);
	errHandler(err);

	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON)

	defer MongoDisconnect(session);
}

func UpdateTaskExecutionRealSeconds(w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasks := dbDone.C("tasks")

	updateTaskJSON, err := ioutil.ReadAll(r.Body)
	errHandler(err);

	updateTaskJSONString := string(updateTaskJSON);
	updateTaskSplitted := strings.Split(updateTaskJSONString, "$;")

	uuid := updateTaskSplitted[0];
	seconds, err := strconv.Atoi(updateTaskSplitted[1]);
	errHandler(err)

	var taskForUpdate Task;
	err = cTasks.Find(bson.M{"uuid": string(uuid)}).One(&taskForUpdate)
	errHandler(err);

	taskForUpdate.DurationExecutionRealSeconds = seconds;

	colQuerier := bson.M{"uuid": string(uuid)};
	err = cTasks.Update(colQuerier, taskForUpdate);
	errHandler(err);

	//w.Header().Set("Content-Type", "application/json")
	//w.Write([]byte)

	defer MongoDisconnect(session);
}

func GetTodayResults (w http.ResponseWriter, r *http.Request) {
	session := MongoConnect();
	dbDone := session.DB("done")
	cTasksCompleted := dbDone.C("tasks_completed")

	var tasksCompleted []Task;

	err := cTasksCompleted.Find(nil).All(&tasksCompleted);
	errHandler(err);

	var tasksCompletedToday []Task;

	var todayYear, todayMonth, todayDay = time.Now().Date();

	for i := 0; i < len(tasksCompleted); i++ {
		year, month, day := tasksCompleted[i].TimeCompleted.Date()
		if (year == todayYear) && (month == todayMonth) && (day == todayDay) {
			tasksCompletedToday = append(tasksCompletedToday, tasksCompleted[i]);
		}
	}

	saveReport(tasksCompletedToday);

	tasksJSON, err := json.Marshal(tasksCompletedToday);
	errHandler(err);

	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON)
	defer MongoDisconnect(session);
}

func saveReport (completedTasks []Task) {
	t := time.Now()
	day := t.Format("_2")
	month := t.Format("Jan")
	year := t.Format("2006")

	filenameComplete := year + "-" + month + "-" + day + ".html"

	f, err := os.Create(filenameComplete)
	errHandler(err)

	finfo, err := f.Stat();

	if finfo.Size() == 0 {
		io.WriteString(f, "<!DOCTYPE html>\n")
		io.WriteString(f, "<html>\n")
		io.WriteString(f, "<head>\n")
		io.WriteString(f, "<meta charset=\"utf-8\">\n")
		io.WriteString(f, "</head>\n")
		io.WriteString(f, "<body>\n")
		io.WriteString(f, "</body>\n")
		io.WriteString(f, "</html>\n")
	}

	n, err := io.WriteString(f, "<b>Отчет по выполненным задачам за "+ year + "-" + month + "-" + day + "</b><br>");
	if err != nil {
		fmt.Println(n, err)
	}

	for i := 0; i < len(completedTasks); i++ {

		realSeconds := completedTasks[i].DurationExecutionRealSeconds;
		daysForRound := realSeconds / (60*60*8)
		realSeconds -= daysForRound * (60*60*8)
		hoursForRound := realSeconds / (60*60)
		realSeconds -= hoursForRound * 60 * 60
		minutesForRount := realSeconds / 60
		realSeconds -= minutesForRount * 60

		taskReport := "<br>Задача <b>" + completedTasks[i].Body + "</b> выполнена. <br>"+
			"Затрачено: " + strconv.Itoa(daysForRound) + " дн. " + strconv.Itoa(hoursForRound) + " ч. " +
			strconv.Itoa(minutesForRount) + " мин. " + strconv.Itoa(realSeconds) + " сек.<br>"
		n, err := io.WriteString(f, taskReport)
		if err != nil {
			fmt.Println(n, err)
		}
	}

	f.Close()

}

//For correct work with > and < but not needed
//func JSONMarshal(t interface{}) ([]byte, error) {
//	buffer := &bytes.Buffer{}
//	encoder := json.NewEncoder(buffer)
//	encoder.SetEscapeHTML(false)
//	err := encoder.Encode(t)
//	return buffer.Bytes(), err
//}

//func RFC3339TimetoReadableTime(t time.Time) (readableTime string) {
//	year := string(t.Format("2006"))
//	month := string(t.Format("01"))
//	day := string(t.Format("02"))
//	dayOfTheWeek := string(t.Format("Mon"))
//	hour := string(t.Format("15"))
//	minute := string(t.Format("04"))
//	second := string(t.Format("05"))
//
//	readableTime = year + "-" + month + "-" + day + " (" + dayOfTheWeek + ") " + hour + ":" + minute + ":" + second
//	return readableTime
//}
//
//func ReadableTimeToRFC3339Time(readableTime string) (t time.Time) {
//	t, err := time.Parse(readableTime, "2006-01-02 (Mon) 15:04:03")
//	errHandler(err)
//
//	return t;
//}
