package database

import "github.com/ant0ine/go-json-rest/rest"

type IDatabase interface {
	addTask(w rest.ResponseWriter, r *rest.Request)
	getTasks(w rest.ResponseWriter, r *rest.Request)
	removeTask(w rest.ResponseWriter, r *rest.Request)
	closeTask(w rest.ResponseWriter, r *rest.Request)
	rearrangeTasks(w rest.ResponseWriter, r *rest.Request)
}
