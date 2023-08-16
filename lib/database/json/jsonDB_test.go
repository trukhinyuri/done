package json

import (
	"github.com/satori/go.uuid"
	"testing"
	"time"
)

func TestTaskType(t *testing.T) {
	var task Task
	task.Body = "Сделать хорошо"
	task.UUID = uuid.NewV4().String()
	task.TaskCreated = time.Now()
	task.TaskCompleted = time.Now()

	var subtask Task
	subtask.Body = "Сперва прочитать задачу"
	subtask.UUID = uuid.NewV4().String()
	subtask.TaskCreated = time.Now()
	subtask.TaskCompleted = time.Now()

	task.Child = append(task.Child, subtask)

	if task.Child[0].UUID != subtask.UUID {
		t.Error("Expected task.Child[0] is subtask, actual: ", task.Child[0].UUID)
	}

	var subtask2 Task
	subtask2.Body = "Затем сделать ее"
	subtask2.UUID = uuid.NewV4().String()
	subtask2.TaskCreated = time.Now()
	subtask2.TaskCompleted = time.Now()

	task.Child = append(task.Child, subtask2)

	if task.Child[1].UUID != subtask2.UUID {
		t.Error("Expected task.Child[1] is subtask2, actual: ", task.Child[0].UUID)
	}

	if len(task.Child) != 2 {
		t.Error("Expected len(task.Child) is 2, actual: ", len(task.Child))
	}

	for i, task := range task.Child {
		if task.Body == "Затем сделать ее" {
			if i != 1 {
				t.Error("Expected task index is 1, actual: ", i)
			}
		}
	}
}
