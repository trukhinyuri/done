package json

//import "encoding/json"
import (
	//"github.com/satori/go.uuid"
	"time"
)

type Task struct {
	UUID          string
	Body          string
	TaskCreated   time.Time
	TaskCompleted time.Time
	Child         []Task
}
