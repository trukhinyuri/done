"use strict";

window.exports = window.exports || (window.exports = {});
window.exports.timerId = 0;

(function (Done) {

    function makeObjectFromBackendJSON(json) {
        // let obj = eval(json.substring(1, json.length - 1).split('\\').join('').split("\"").join("\'").split('u003e').join('>').split('u003c').join("<"));
        let obj = eval(json);
        //TOTO insert correct work with u003e u003c text
        return obj;
    }

    // function makeBackendJSONFromObject(obj) {
    //     var json = "\"" + "[" + JSON.stringify(obj).split("\"").join("\\\"").split("\'").join("\"") + "]" + "\"";
    //     return json;
    // }

    function FromRFC3339ToJSTime(objArray) {
        if (objArray != null) {
            for (var i = 0; i < objArray.length; i++) {
                for (var property in objArray[i]) {
                    if (objArray[i].hasOwnProperty(property)) {
                        if (property.indexOf('duration_execution_estimated_seconds') == 0) {
                            var seconds = objArray[i][property];

                            var days = Math.floor(seconds / (60 * 60 * 8));
                            seconds -= days * 60 * 60 * 8;
                            var hours = Math.floor(seconds / (60*60));
                            seconds -= hours * 60 * 60;
                            var minutes = Math.floor(seconds / 60);

                            var result = days + " дн., "
                            + hours + " : "
                            + minutes;
                            objArray[i][property + "_readable"] = result;
                            // var date = new Date(objArray[i][property]);
                            // var durationDays = date.getDate();
                            // var durationHours = date.getHours() - 3;
                            // var durationMinutes = date.getMinutes() - 0;

                            // objArray[i][property + "readable"] =
                            //     durationDays + " дн., " + durationHours + " ч., " + durationMinutes + " мин.";

                            // objArray[i][property + "readableDayTime"] =
                            //     date.getFullYear() + "-" + date.getMonth() + "-" + date.getDate()
                            //     + " (" + date.getDay() + ") "
                            //     + date.getHours() + ":" + date.getMinutes();
                        } else if(property.indexOf('time_hard_dead_line') == 0) {
                            var date = new Date(objArray[i][property]);
                            if (date.getMonth() == 0) {
                                objArray[i][property + "_readable"] = "";
                            } else {
                                objArray[i][property + "_readable"] = date.getFullYear() + " / " + date.getMonth() + " / " + date.getDate();
                            }

                        } else if (property.indexOf('duration_execution_real_seconds') == 0) {
                            var seconds = objArray[i][property];

                            var days = Math.floor(seconds / (60 * 60 * 8));
                            seconds -= days * 60 * 60 * 8;
                            var hours = Math.floor(seconds / (60*60));
                            seconds -= hours * 60 * 60;
                            var minutes = Math.floor(seconds / 60);
                            seconds -= minutes * 60;

                            var result = days + " дн., "
                                + hours + " : "
                                + minutes + " : " + seconds;
                            objArray[i][property + "_readable"] = result;
                        }
                    }
                }
            }
        } else {
            return objArray;
        }


    }



    var sort_by = function(field, reverse, primer){

        var key = primer ?
            function(x) {return primer(x[field])} :
            function(x) {return x[field]};

        reverse = !reverse ? 1 : -1;

        return function (a, b) {
            return a = key(a), b = key(b), reverse * ((a > b) - (b > a));
        }
    };

    function replaceQuotes(tasksData) {
        if (tasksData != null) {
            for (var i = 0; i < tasksData.length; i++) {
                tasksData[i].body = tasksData[i].body.split("280d382c-f23e-4631-8551-f43661405497").join("\"").split("e6f23f57-6cad-451b-8306-7939e25542dc").join("\'");
            }
        }
        return tasksData;
    }

    function renderTasks(tasksJSON) {

        var tasksData = makeObjectFromBackendJSON(tasksJSON);
        // var secondInDay = 8*60*60;
        // var dateSeconds = new Date().getUTCSeconds();
        // alert(dateSeconds);
        FromRFC3339ToJSTime(tasksData);
        replaceQuotes(tasksData);

        var workDaySeconds = 8 * 60 * 60;
        var plannedDate = new Date();
        var dayOfWeek = plannedDate.getDay();
        var unPlannedSeconds = 0;
        var dayHolidayPlanned = false;

        var tasksElements = document.getElementsByClassName("page_tasks_content")[0];

        console.log(tasksData);
        if (tasksData != null) {
            tasksData.sort(sort_by("order", false, parseInt));


            tasksElements.innerHTML = "";
            for (var i = 0; i < tasksData.length; i++) {

                if (tasksData[i]["duration_execution_estimated_seconds"] > tasksData[i]["duration_execution_real_seconds"]) {
                    unPlannedSeconds += tasksData[i]["duration_execution_estimated_seconds"];
                } else {
                    unPlannedSeconds += tasksData[i]["duration_execution_real_seconds"];
                }

                if (dayOfWeek == 0) {
                    if (dayHolidayPlanned == false) {
                        plannedDate.setDate(plannedDate.getDate() + 2);
                        dayHolidayPlanned = true;
                    }
                } else if (dayOfWeek == 6) {
                    if (dayHolidayPlanned == false) {
                        plannedDate.setDate(plannedDate.getDate() + 3);
                        dayHolidayPlanned = true;
                    }
                } else {
                    if (dayHolidayPlanned == false) {
                        plannedDate.setDate(plannedDate.getDate() + 1);
                        dayHolidayPlanned = true;
                    }
                }

                if ((unPlannedSeconds > workDaySeconds) && (i != 0)) {
                        var plannerContainer = document.createElement("div");
                        plannerContainer.className = "page_tasks_content_taskPlanner";
                        plannerContainer.innerHTML += "↑ будет сделано до "
                            + plannedDate.getFullYear() + "-" + getCorrectMonthNumber(plannedDate.getMonth())
                            + "-" + plannedDate.getDate() + " ( " + getCorrectDayOfWeek(plannedDate.getDay()) + " )";
                        tasksElements.appendChild(plannerContainer);
                }

                var taskContainer = document.createElement("div");
                taskContainer.className = "page_tasks_content_taskContainer " + i;
                tasksElements.appendChild(taskContainer);
                Modules.Loader.loadTemplate("templates", "task", "page_tasks_content_taskContainer " + i, tasksData[i]);

                if (unPlannedSeconds > workDaySeconds) {
                    while ((unPlannedSeconds / workDaySeconds) >= 1) {
                        plannedDate.setDate(plannedDate.getDate() + 1);
                        if (plannedDate.getDay() == 0) {
                            plannedDate.setDate(plannedDate.getDate() + 1);
                        } else if (plannedDate.getDay() == 6) {
                            plannedDate.setDate(plannedDate.getDate() + 2);
                        }
                        unPlannedSeconds = unPlannedSeconds - workDaySeconds;
                    }
                }

                if ((tasksData.length - 1) == i) {
                    if (unPlannedSeconds <= workDaySeconds) {
                        var plannerContainer = document.createElement("div");
                        plannerContainer.className = "page_tasks_content_taskPlanner";
                        plannerContainer.innerHTML += "↑ будет сделано до "
                            + plannedDate.getFullYear() + "-" + getCorrectMonthNumber(plannedDate.getMonth())
                            + "-" + plannedDate.getDate() + " ( " + getCorrectDayOfWeek(plannedDate.getDay()) + " )";
                        tasksElements.appendChild(plannerContainer);
                    } else {
                        plannedDate.setDate(plannedDate.getDate() + 1);
                        if (plannedDate.getDay() == 0) {
                            plannedDate.setDate(plannedDate.getDate() + 1);
                        } else if (plannedDate.getDay() == 6) {
                            plannedDate.setDate(plannedDate.getDate() + 2);
                        }

                        var plannerContainer = document.createElement("div");
                        plannerContainer.className = "page_tasks_content_taskPlanner";
                        plannerContainer.innerHTML += "↑ будет сделано до "
                            + plannedDate.getFullYear() + "-" + getCorrectMonthNumber(plannedDate.getMonth())
                            + "-" + plannedDate.getDate() + " ( " + getCorrectDayOfWeek(plannedDate.getDay()) + " )";
                        tasksElements.appendChild(plannerContainer);
                    }

                }
            }

        } else {
            tasksElements.innerHTML = "";
        }

    }

    function getCorrectMonthNumber(jsMonth) {
        var monthNumber = parseInt(jsMonth);
        return monthNumber + 1;
    }
    
    function getCorrectDayOfWeek(jsDay) {
        switch(parseInt(jsDay)) {
            case 0:
            return "воскресенье";
                break;
            case 1:
                return "понедельник";
                break;
            case 2:
                return "вторник";
                break;
            case 3:
                return "среда";
                break;
            case 4:
                return "четверг";
                break;
            case 5:
                return "пятница";
                break;
            case 6:
                return "суббота";
                break;
        }
        
    }

    function getTasks() {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', "/api/getTasks", true);
        xhr.setRequestHeader('Content-Type', 'application/json')
        xhr.send(null);
        xhr.onreadystatechange = function(data) {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                renderTasks(xhr.responseText);
            }
        }
    }

    function getTodayResults() {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', "/api/getTodayResults", true);
        xhr.setRequestHeader('Content-Type', 'application/json')
        xhr.send(null);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                Done.renderTodayResults(xhr.responseText);
            }
        }
    }

    function renderTodayResults(tasksJSON) {
        var today = document.getElementsByClassName("page_today_content")[0];
        today.innerHTML = "";
        var tasksCompletedData = makeObjectFromBackendJSON(tasksJSON);
        FromRFC3339ToJSTime(tasksCompletedData);
        replaceQuotes(tasksCompletedData);

        if (tasksCompletedData  != null) {
            for (var i = 0; i < tasksCompletedData.length; i++) {
                today.innerHTML +=
                    tasksCompletedData[i].body
                    + " ( Затрачено: "
                    + tasksCompletedData[i]["duration_execution_real_seconds_readable"]
                    + " ) "
                    + "<br>";
            }
        }

    }

    function postTask(e) {

        window.exports.timerId = 0;
        var taskTextElement = document.getElementsByClassName("taskText")[0];
        var estimationDaysElement = document.getElementsByClassName("page_newTask_estimationDays_value")[0];
        var estimationHoursElement = document.getElementsByClassName("page_newTask_estimationHours_value")[0];
        var estimationMinutesElement = document.getElementsByClassName("page_newTask_estimationMinutes_value")[0];
        var deadlineMonthElement = document.getElementsByClassName("page_newTask_timeHardDeadlineMonth_value")[0];
        var deadlineDayElement = document.getElementsByClassName("page_newTask_timeHardDeadlineDay_value")[0];

        var taskText = taskTextElement.value;
        taskText = taskText.split("\"").join("280d382c-f23e-4631-8551-f43661405497").split("\'").join("e6f23f57-6cad-451b-8306-7939e25542dc");

        var estimationDays;
        if (estimationDaysElement.value.localeCompare("") == 0) {
            estimationDays = estimationDaysElement.placeholder;
        } else {
            estimationDays = estimationDaysElement.value;
        }

        var estimationHours;
        if (estimationHoursElement.value.localeCompare("") == 0) {
            estimationHours = estimationHoursElement.placeholder;
        } else {
            estimationHours = estimationHoursElement.value;
        }

        var estimationMinutes;
        if (estimationMinutesElement.value.localeCompare("") == 0) {
            estimationMinutes = estimationMinutesElement.placeholder;
        } else {
            estimationMinutes = estimationMinutesElement.value;
        }


        var deadlineMonth;
        if (deadlineMonthElement.value.localeCompare("") == 0) {
            deadlineMonth = deadlineMonthElement.placeholder;
        } else {
            deadlineMonth = deadlineMonthElement.value;
        }

        var deadlineDay;
        if (deadlineDayElement.value.localeCompare("") == 0) {
            deadlineDay = deadlineDayElement.placeholder;
        } else {
            deadlineDay = deadlineDayElement.value;
        }


        var newTask = {};
        newTask.body = taskText;
        newTask.estimation = estimationDays * 8 * 60 * 60 + estimationHours * 60 * 60 + estimationMinutes * 60;
        newTask.deadlineMonth = deadlineMonth;
        newTask.deadlineDay = deadlineDay;

        estimationMinutesElement.value = "";
        estimationHoursElement.value = "";
        estimationDaysElement.value = "";
        deadlineMonthElement.value = "";
        deadlineDayElement.value = "";


        if ((newTask.body.localeCompare("") != 0) && (estimationDays <=31) && (estimationHours <= 23)
        && (estimationMinutes <= 59) && (deadlineMonth <=12) && (deadlineDay <= 31)) {
            var xhr = new XMLHttpRequest();
            xhr.open('POST', "/api/addTask", true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.send(newTask.body + "$;" + newTask.estimation + "$;"
                + newTask.deadlineMonth + "$;" + newTask.deadlineDay);
            taskTextElement.value = "";
            taskTextElement.focus();
            xhr.onreadystatechange = function() {
                if (xhr.readyState == XMLHttpRequest.DONE) {
                    Done.renderTasks(xhr.responseText);
                }
            }
        } else {
            alert("Некорректный ввод в форму постановки задачи")
        }
    }

    function completeTask(taskUUID) {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', "/api/completeTask", true);
        xhr.setRequestHeader('Content-Type', 'application/json')
        xhr.send(taskUUID);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                Done.renderTasks(xhr.responseText);
                Done.getTodayResults();
            }
        }
    }

    function removeTask(taskUUID) {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', "/api/removeTask", true);
        xhr.setRequestHeader('Content-Type', 'application/json')
        xhr.send(taskUUID);
        xhr.onreadystatechange = function() {
            xhr.onreadystatechange = function() {
                if (xhr.readyState == XMLHttpRequest.DONE) {
                    Done.renderTasks(xhr.responseText);
                }
            }
        }
    }

    //TODO: remove dirty hack with single string data transfer and move back correct JSON unmarshalling
    function rearrangeTasks(sourceTaskUUID, destinationTaskUUID) {
        var rearrangeTasksData = sourceTaskUUID + "," + destinationTaskUUID;

        var xhr = new XMLHttpRequest();
        xhr.open('POST', "/api/rearrangeTasks", true);
        xhr.setRequestHeader('Content-Type', 'application/json')
        xhr.send(rearrangeTasksData);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                Done.renderTasks(xhr.responseText);
            }
        }
    }

    Done.renderTasks = renderTasks;
    Done.getTasks = getTasks;
    Done.postTask = postTask;
    Done.completeTask = completeTask;
    Done.removeTask = removeTask;
    Done.getTodayResults = getTodayResults;
    Done.renderTodayResults = renderTodayResults;
    Done.rearrangeTasks = rearrangeTasks;
})(window.exports.Done || (window.exports.Done = {}));
var Done = window.exports.Done;


(function () {
    function run() {
        var currentDate = new Date();
        var estimationDaysElement = document.getElementsByClassName("page_newTask_estimationDays_value")[0];
        var estimationHoursElement = document.getElementsByClassName("page_newTask_estimationHours_value")[0];
        var estimationMinutesElement = document.getElementsByClassName("page_newTask_estimationMinutes_value")[0];
        var deadlineMonthElement = document.getElementsByClassName("page_newTask_timeHardDeadlineMonth_value")[0];
        var deadlineDayElement = document.getElementsByClassName("page_newTask_timeHardDeadlineDay_value")[0];
        estimationDaysElement.placeholder = 0;
        estimationHoursElement.placeholder = 0;
        estimationMinutesElement.placeholder = 0;
        deadlineMonthElement.placeholder = 0;
        deadlineDayElement.placeholder = 0;

        Done.getTasks();
        Done.getTodayResults();

        var addTaskButton = document.getElementsByClassName("taskButton")[0];

        addTaskButton.addEventListener('click', Done.postTask);
    }
    document.addEventListener("DOMContentLoaded", run);
}());