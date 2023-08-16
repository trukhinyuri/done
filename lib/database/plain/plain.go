package plain

import (
	"encoding/json"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

//func addTask(w rest.ResponseWriter, r *rest.Request)  {
//	body, err := ioutil.ReadAll(r.Body)
//
//	filename := "tasks.txt"
//
//	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	n, err := io.WriteString(f, string(body) + "\n")
//	if err != nil {
//		fmt.Println(n, err)
//	}
//	f.Close()
//
//	buf, err := ioutil.ReadFile(filename)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	tasks := string(buf)
//
//	w.WriteJson(tasks)
//}

func AddTask(w rest.ResponseWriter, r *rest.Request) {
	t := time.Now()

	body, err := ioutil.ReadAll(r.Body)

	filename := "tasks.txt"

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(buf), "\n")

	var newTasks string = " [created: " + t.Format("2006–01–02 Mon 15:04:05") + "] " + string(body) + "\n"

	for i := range lines {
		if len(lines)-1 != i {
			newTasks += lines[i]
			newTasks += "\n"
		} else {
			newTasks += lines[i]
		}
	}

	ioutil.WriteFile(filename, []byte(newTasks), 0660)

	w.WriteJson(newTasks)
}

func GetTasks(w rest.ResponseWriter, r *rest.Request) {
	filename := "tasks.txt"

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		fmt.Println(err)
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	f.Close()

	tasks := string(buf)

	w.WriteJson(tasks)
}

func GetTodayResults(w rest.ResponseWriter, r *rest.Request) {
	t := time.Now()
	day := t.Format("_2")
	month := t.Format("Jan")
	year := t.Format("2006")

	filenameComplete := year + "-" + month + "-" + day + ".html"

	f, err := os.OpenFile(filenameComplete, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		fmt.Println(err)
	}

	buf, err := ioutil.ReadFile(filenameComplete)
	if err != nil {
		fmt.Println(err)
	}
	f.Close()

	completeTasks := string(buf)

	w.WriteJson(completeTasks)
}

func RemoveTask(w rest.ResponseWriter, r *rest.Request) {
	body, err := ioutil.ReadAll(r.Body)
	lineNumber, err := strconv.ParseInt(string(body), 0, 64)

	filename := "tasks.txt"

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(buf), "\n")
	var newTasks string = ""

	for i := range lines {

		if (i != int(lineNumber)) && (len(lines)-1 != i) {
			newTasks += lines[i]
			newTasks += "\n"
		}

		if (i != int(lineNumber)) && (len(lines)-1 == i) {
			newTasks += lines[i]
		}

	}

	ioutil.WriteFile(filename, []byte(newTasks), 0660)

	w.WriteJson(newTasks)
}

func CloseTask(w rest.ResponseWriter, r *rest.Request) {
	body, err := ioutil.ReadAll(r.Body)
	lineNumber, err := strconv.ParseInt(string(body), 0, 64)

	filename := "tasks.txt"

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(buf), "\n")
	var newTasks string = ""

	closedTask := ""

	for i := range lines {

		if (i != int(lineNumber)) && (len(lines)-1 != i) {
			newTasks += lines[i]
			newTasks += "\n"
		}

		if (i != int(lineNumber)) && (len(lines)-1 == i) {
			newTasks += lines[i]
		}

		if i == int(lineNumber) {
			closedTask = lines[i]
		}

	}

	ioutil.WriteFile(filename, []byte(newTasks), 0660)

	t := time.Now()
	day := t.Format("_2")
	month := t.Format("Jan")
	year := t.Format("2006")

	filenameComplete := year + "-" + month + "-" + day + ".html"

	f, err := os.OpenFile(filenameComplete, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		fmt.Println(err)
	}

	finfo, err := f.Stat()

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

	n, err := io.WriteString(f, "[completed: "+t.Format("2006–01–02 Mon 15:04:05")+"] "+closedTask+"<br>")
	if err != nil {
		fmt.Println(n, err)
	}
	f.Close()

	w.WriteJson(newTasks)
}

type RearrangeTasksData struct {
	Source      string
	Destination string
}

func RearrangeTasks(w rest.ResponseWriter, r *rest.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	var rearrangeTasksData RearrangeTasksData

	if err == nil && body != nil {
		err = json.Unmarshal(body, &rearrangeTasksData)
	}

	filename := "tasks.txt"

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(buf), "\n")

	source, e := strconv.Atoi(rearrangeTasksData.Source)
	if e != nil {
		return
		fmt.Println(e)
	}

	destination, e := strconv.Atoi(rearrangeTasksData.Destination)
	if e != nil {
		return
		fmt.Println(e)
	}

	if source != destination {
		var newTasks string = ""

		for i := 0; i < len(lines); i++ {
			if i != source {
				if i == destination {
					newTasks += lines[source]
					newTasks += "\n"

				}

				if len(lines)-1 != i {
					newTasks += lines[i]
					newTasks += "\n"
				} else {
					newTasks += lines[i]
				}
			}
		}

		go ioutil.WriteFile(filename, []byte(newTasks), 0660)

		w.WriteJson(newTasks)
	} else {
		return
	}
}
