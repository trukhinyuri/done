package main

import (
	"done/lib/database/mongodb"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

var dbUpgradePtr *bool
var servicePortPtr *int
var versionPtr *bool

func init() {
	dbUpgradePtr = flag.Bool("dbupgrade", false, "Upgrade database for new version compatibility")
	servicePortPtr = flag.Int("port", 3001, "Service port")
	versionPtr = flag.Bool("version", false, "Show app version")
}

func main() {
	flag.Parse()
	submain(flag.Args())
}

func submain(args []string) {

	if *dbUpgradePtr == true {
		launchDBUpgrade()

	} else if *versionPtr == true {

		printServiceVersion()

	} else {
		startService()
	}
}

func launchDBUpgrade() {
	result := mongodb.DBUpgrade()
	log.Println(result)
}

func printServiceVersion() {
	fmt.Println(BuildVersion)
	fmt.Println(BuildTime)
}

func startService() {
	mux := http.NewServeMux()

	apiPath := "/api"
	mux.HandleFunc(apiPath+"/version", testApi)
	mux.HandleFunc(apiPath+"/getTasks", mongodb.GetTasks)
	mux.HandleFunc(apiPath+"/getTodayResults", mongodb.GetTodayResults)
	mux.HandleFunc(apiPath+"/addTask", mongodb.AddTask)
	mux.HandleFunc(apiPath+"/removeTask", mongodb.RemoveTask)
	mux.HandleFunc(apiPath+"/rearrangeTasks", mongodb.RearrangeTasks)
	mux.HandleFunc(apiPath+"/completeTask", mongodb.CompleteTask)
	mux.HandleFunc(apiPath+"/updateTaskExecutionRealSeconds", mongodb.UpdateTaskExecutionRealSeconds)

	fileServer := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", http.StripPrefix("/", fileServer))

	servicePortString := strconv.Itoa(*servicePortPtr)
	log.Println("Starting server on :" + servicePortString)
	log.Fatal(http.ListenAndServe(":"+servicePortString, mux))
}

type Test struct {
	Message  string
	CreateAt time.Time
}

func testApi(w http.ResponseWriter, r *http.Request) {
	testObj := Test{}
	testObj.Message = BuildVersion
	testObj.CreateAt = time.Now().Local()

	testObjJSON, err := json.Marshal(testObj)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(testObjJSON)
}
