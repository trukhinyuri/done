/**
 * Done Task Manager - Main Application Logic
 * This file handles all task operations, API communication, and UI updates
 */
"use strict";

// Global namespace for exports
window.exports = window.exports || (window.exports = {});
window.exports.timerId = 0; // Currently active timer ID

(function (Done) {

    /**
     * Parse JSON response from backend
     * TODO: Replace eval with JSON.parse for better security
     */
    function makeObjectFromBackendJSON(json) {
        let obj = eval(json);
        return obj;
    }

    // function makeBackendJSONFromObject(obj) {
    //     var json = "\"" + "[" + JSON.stringify(obj).split("\"").join("\\\"").split("\'").join("\"") + "]" + "\"";
    //     return json;
    // }

    /**
     * Convert task time fields to human-readable format
     * Processes duration and deadline fields for display
     */
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

                            var result = days + " d., "
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
                            if (date.getFullYear() == 9999) {
                                objArray[i][property + "_readable"] = "";
                            } else {
                                // Month is 0-indexed in JS, so add 1 for display
                                objArray[i][property + "_readable"] = date.getDate() + " / " + (date.getMonth() + 1) + " / " + date.getFullYear();
                            }

                        } else if (property.indexOf('duration_execution_real_seconds') == 0) {
                            var seconds = objArray[i][property];

                            var days = Math.floor(seconds / (60 * 60 * 8));
                            seconds -= days * 60 * 60 * 8;
                            var hours = Math.floor(seconds / (60*60));
                            seconds -= hours * 60 * 60;
                            var minutes = Math.floor(seconds / 60);
                            seconds -= minutes * 60;

                            var result = days + " d., "
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



    /**
     * Sort array of objects by field
     * @param {string} field - Field name to sort by
     * @param {boolean} reverse - Sort in reverse order
     * @param {function} primer - Optional function to process field value
     */
    var sort_by = function(field, reverse, primer){

        var key = primer ?
            function(x) {return primer(x[field])} :
            function(x) {return x[field]};

        reverse = !reverse ? 1 : -1;

        return function (a, b) {
            return a = key(a), b = key(b), reverse * ((a > b) - (b > a));
        }
    };

    /**
     * Clean task text for display
     * Handles both legacy UUID format and new encoding
     */
    function replaceQuotes(tasksData) {
        if (tasksData != null) {
            for (var i = 0; i < tasksData.length; i++) {
                // Use the new cleaning function that handles all formats
                tasksData[i].body = window.TaskUtils.cleanTaskText(tasksData[i].body);
            }
        }
        return tasksData;
    }

    /**
     * Render all tasks in the UI
     * Handles task display, ordering, and planning calculations
     * @param {string} tasksJSON - JSON string containing task data
     */
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
                        plannerContainer.innerHTML += "📅 Will be done by "
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
                        plannerContainer.innerHTML += "📅 Will be done by "
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
                        plannerContainer.innerHTML += "📅 Will be done by "
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
            return "sunday";
                break;
            case 1:
                return "monday";
                break;
            case 2:
                return "tuesday";
                break;
            case 3:
                return "wednesday";
                break;
            case 4:
                return "thursday";
                break;
            case 5:
                return "friday";
                break;
            case 6:
                return "saturday";
                break;
        }
        
    }

    /**
     * Fetch all tasks from the API
     */
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

    /**
     * Fetch today's completed tasks from the API
     */
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

    /**
     * Render today's completed tasks in the UI
     * @param {string} tasksJSON - JSON string containing completed task data
     */
    function renderTodayResults(tasksJSON) {
        var today = document.getElementsByClassName("page_today_content")[0];
        today.innerHTML = "";
        var tasksCompletedData = makeObjectFromBackendJSON(tasksJSON);
        FromRFC3339ToJSTime(tasksCompletedData);
        replaceQuotes(tasksCompletedData);

        if (tasksCompletedData != null && tasksCompletedData.length > 0) {
            for (var i = 0; i < tasksCompletedData.length; i++) {
                var taskDiv = document.createElement('div');
                taskDiv.className = 'completed-task';
                
                var taskName = document.createElement('div');
                taskName.className = 'task-name';
                taskName.textContent = tasksCompletedData[i].body;
                
                var taskTime = document.createElement('div');
                taskTime.className = 'task-time';
                taskTime.textContent = '⏱️ Time spent: ' + tasksCompletedData[i]["duration_execution_real_seconds_readable"];
                
                taskDiv.appendChild(taskName);
                taskDiv.appendChild(taskTime);
                today.appendChild(taskDiv);
            }
        }

    }

    /**
     * Create a new task
     * Validates input and sends task data to the API
     * @param {Event} e - Form submit event (optional)
     */
    function postTask(e) {

        window.exports.timerId = 0;
        var taskTextElement = document.getElementsByClassName("taskText")[0];
        var estimationDaysElement = document.getElementsByClassName("page_newTask_estimationDays_value")[0];
        var estimationHoursElement = document.getElementsByClassName("page_newTask_estimationHours_value")[0];
        var estimationMinutesElement = document.getElementsByClassName("page_newTask_estimationMinutes_value")[0];
        var deadlineMonthElement = document.getElementsByClassName("page_newTask_timeHardDeadlineMonth_value")[0];
        var deadlineDayElement = document.getElementsByClassName("page_newTask_timeHardDeadlineDay_value")[0];
        var deadlineYearElement = document.getElementsByClassName("page_newTask_timeHardDeadlineYear_value")[0];

        // Encode special characters for safe transmission
        var taskText = window.TaskUtils.encodeTaskText(taskTextElement.value);

        // Set default estimation if not provided
        var estimationDays;
        if (estimationDaysElement.value.localeCompare("") == 0) {
            estimationDays = 0; // Default: 0 days
        } else {
            estimationDays = estimationDaysElement.value;
        }

        var estimationHours;
        if (estimationHoursElement.value.localeCompare("") == 0) {
            estimationHours = 1; // Default: 1 hour
        } else {
            estimationHours = estimationHoursElement.value;
        }

        var estimationMinutes;
        if (estimationMinutesElement.value.localeCompare("") == 0) {
            estimationMinutes = 0; // Default: 0 minutes
        } else {
            estimationMinutes = estimationMinutesElement.value;
        }


        // Smart deadline processing with auto-completion
        var currentDate = new Date();
        var currentYear = currentDate.getFullYear();
        var currentMonth = currentDate.getMonth() + 1; // JS months are 0-indexed
        var currentDay = currentDate.getDate();
        
        console.log('Current date:', currentDay, currentMonth, currentYear);
        
        // Get raw values from form fields
        var deadlineMonthRaw = deadlineMonthElement.value.trim();
        var deadlineDayRaw = deadlineDayElement.value.trim();
        var deadlineYearRaw = deadlineYearElement.value.trim();
        
        // Check if any deadline field has actual data (not empty or placeholder)
        var hasMonth = deadlineMonthRaw && deadlineMonthRaw !== "MM" && deadlineMonthRaw !== "0";
        var hasDay = deadlineDayRaw && deadlineDayRaw !== "DD" && deadlineDayRaw !== "0";
        var hasYear = deadlineYearRaw && deadlineYearRaw !== "YYYY" && deadlineYearRaw !== "0";
        
        var deadlineMonth, deadlineDay, deadlineYear;
        
        // If no deadline fields are filled, set as no deadline
        if (!hasMonth && !hasDay && !hasYear) {
            deadlineMonth = "0";
            deadlineDay = "0";
            deadlineYear = "0";
        } else {
            // Smart auto-completion with logic for partial dates
            if (hasDay && !hasMonth && !hasYear) {
                // Only day specified - use current month/year, but if day is in past, use next month
                deadlineYear = currentYear.toString();
                deadlineMonth = currentMonth.toString();
                deadlineDay = deadlineDayRaw;
                
                if (parseInt(deadlineDayRaw) < currentDay) {
                    // Day is in the past for current month, use next month
                    var nextMonth = currentMonth + 1;
                    var nextYear = currentYear;
                    if (nextMonth > 12) {
                        nextMonth = 1;
                        nextYear = currentYear + 1;
                    }
                    deadlineMonth = nextMonth.toString();
                    deadlineYear = nextYear.toString();
                }
            } else if (hasMonth && !hasDay && !hasYear) {
                // Only month specified - use current year, and first day if month is current or future
                deadlineYear = currentYear.toString();
                deadlineMonth = deadlineMonthRaw;
                
                if (parseInt(deadlineMonthRaw) < currentMonth || 
                    (parseInt(deadlineMonthRaw) == currentMonth && currentDay > 28)) {
                    // Month is in past or current month is almost over, use next year
                    deadlineYear = (currentYear + 1).toString();
                    deadlineDay = "1";
                } else if (parseInt(deadlineMonthRaw) == currentMonth) {
                    // Current month, use current day or later
                    deadlineDay = currentDay.toString();
                } else {
                    // Future month, use first day
                    deadlineDay = "1";
                }
            } else {
                // If any field is filled, auto-complete missing fields with current date
                deadlineMonth = hasMonth ? deadlineMonthRaw : currentMonth.toString();
                deadlineDay = hasDay ? deadlineDayRaw : currentDay.toString();
                deadlineYear = hasYear ? deadlineYearRaw : currentYear.toString();
            }
            
            console.log('Auto-completed deadline:', deadlineDay, deadlineMonth, deadlineYear);
            
            // Validate the constructed date
            var testDate = new Date(parseInt(deadlineYear), parseInt(deadlineMonth) - 1, parseInt(deadlineDay));
            var isValidDate = testDate.getFullYear() == parseInt(deadlineYear) &&
                             testDate.getMonth() == parseInt(deadlineMonth) - 1 &&
                             testDate.getDate() == parseInt(deadlineDay);
            
            // If date is invalid, use tomorrow
            if (!isValidDate) {
                console.log('Date is invalid, adjusting to tomorrow');
                var tomorrow = new Date(currentDate);
                tomorrow.setDate(tomorrow.getDate() + 1);
                deadlineMonth = (tomorrow.getMonth() + 1).toString();
                deadlineDay = tomorrow.getDate().toString();
                deadlineYear = tomorrow.getFullYear().toString();
                console.log('Adjusted to:', deadlineDay, deadlineMonth, deadlineYear);
            }
        }

        var newTask = {};
        newTask.body = taskText;
        newTask.estimation = estimationDays * 8 * 60 * 60 + estimationHours * 60 * 60 + estimationMinutes * 60;
        newTask.deadlineMonth = deadlineMonth;
        newTask.deadlineDay = deadlineDay;
        newTask.deadlineYear = deadlineYear;

        estimationMinutesElement.value = "";
        estimationHoursElement.value = "";
        estimationDaysElement.value = "";
        deadlineMonthElement.value = "";
        deadlineDayElement.value = "";
        deadlineYearElement.value = "";


        if ((newTask.body.localeCompare("") != 0) && (Number(estimationDays) <=31) && (Number(estimationHours) <= 23)
        && (Number(estimationMinutes) <= 59) && (Number(deadlineMonth) <=12) && (Number(deadlineDay) <= 31)) {
            var xhr = new XMLHttpRequest();
            xhr.open('POST', "/api/addTask", true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.send(newTask.body + "$;" + newTask.estimation + "$;"
                + newTask.deadlineMonth + "$;" + newTask.deadlineDay + "$;" + newTask.deadlineYear);
            taskTextElement.value = "";
            taskTextElement.focus();
            xhr.onreadystatechange = function() {
                if (xhr.readyState == XMLHttpRequest.DONE) {
                    Done.renderTasks(xhr.responseText);
                    if (window.NinstyleSounds) {
                        window.NinstyleSounds.taskCreate();
                    }
                }
            }
        } else {
            alert("Incorrect input in the task setting form")
        }
    }

    /**
     * Mark a task as completed
     * @param {string} taskUUID - UUID of the task to complete
     */
    function completeTask(taskUUID) {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', "/api/completeTask", true);
        xhr.setRequestHeader('Content-Type', 'text/plain')
        xhr.send(taskUUID);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                Done.renderTasks(xhr.responseText);
                Done.getTodayResults();
                if (window.NinstyleSounds) {
                    window.NinstyleSounds.taskComplete();
                }
            }
        }
    }

    /**
     * Delete a task
     * @param {string} taskUUID - UUID of the task to delete
     */
    function removeTask(taskUUID) {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', "/api/removeTask", true);
        xhr.setRequestHeader('Content-Type', 'text/plain')
        xhr.send(taskUUID);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                Done.renderTasks(xhr.responseText);
                if (window.NinstyleSounds) {
                    window.NinstyleSounds.taskDelete();
                }
            }
        }
    }

    /**
     * Rearrange task order (drag & drop functionality)
     * @param {string} sourceTaskUUID - UUID of task being moved
     * @param {string} destinationTaskUUID - UUID of target position task
     * TODO: Replace string concatenation with proper JSON
     */
    function rearrangeTasks(sourceTaskUUID, destinationTaskUUID) {
        console.log('Rearranging tasks:', sourceTaskUUID, '->', destinationTaskUUID);
        var rearrangeTasksData = sourceTaskUUID + "," + destinationTaskUUID;

        var xhr = new XMLHttpRequest();
        xhr.open('POST', "/api/rearrangeTasks", true);
        xhr.setRequestHeader('Content-Type', 'text/plain')
        xhr.send(rearrangeTasksData);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                console.log('Rearrange response status:', xhr.status);
                console.log('Rearrange response:', xhr.responseText);
                if (xhr.status === 200) {
                    Done.renderTasks(xhr.responseText);
                } else {
                    console.error('Failed to rearrange tasks:', xhr.status, xhr.responseText);
                    // Reload tasks to restore original order
                    Done.getTasks();
                }
            }
        }
    }

    // Export public API
    Done.renderTasks = renderTasks;
    Done.getTasks = getTasks;
    Done.postTask = postTask;
    Done.completeTask = completeTask;
    Done.removeTask = removeTask;
    Done.getTodayResults = getTodayResults;
    Done.renderTodayResults = renderTodayResults;
    Done.rearrangeTasks = rearrangeTasks;
})(window.exports.Done || (window.exports.Done = {}));

// Make Done available globally
var Done = window.exports.Done;


/**
 * Application initialization and event handlers
 */
(function () {
    /**
     * Initialize the application
     * Sets up default values, fetches data, and attaches event handlers
     */
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
        
        // Fetch and display build info
        fetchBuildInfo();

        var addTaskButton = document.getElementsByClassName("taskButton")[0];

        addTaskButton.addEventListener('click', Done.postTask);
        
        // Add drop zone for empty task list
        var tasksContent = document.querySelector('.page_tasks_content');
        if (tasksContent) {
            tasksContent.addEventListener('dragover', function(e) {
                if (e.target === tasksContent && tasksContent.children.length === 0) {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';
                }
            });
            
            tasksContent.addEventListener('drop', function(e) {
                if (e.target === tasksContent && tasksContent.children.length === 0) {
                    e.preventDefault();
                    var sourceTaskUUID = e.dataTransfer.getData("sourceTaskUUID");
                    // Handle drop on empty list
                    console.log('Dropped on empty list:', sourceTaskUUID);
                }
            });
        }
        
        // Add keyboard shortcuts and clipboard support for textarea
        var taskTextElement = document.getElementsByClassName("taskText")[0];
        taskTextElement.addEventListener('keydown', function(e) {
            if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                e.preventDefault();
                Done.postTask();
            }
        });
    }
    
    /**
     * Fetch build information from the API and update footer
     * Displays the actual build date in the footer
     */
    function fetchBuildInfo() {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', "/api/version", true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.send(null);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == XMLHttpRequest.DONE && xhr.status === 200) {
                try {
                    var info = JSON.parse(xhr.responseText);
                    var buildDate = info.buildDate || new Date().toISOString().split('T')[0].replace(/-/g, '.');
                    
                    // Update footer with build date
                    var footerContent = document.querySelector('.footer_content');
                    if (footerContent) {
                        footerContent.textContent = 'Done. The task manager. © Yuri Trukhin, 2016–2025. Build ' + buildDate + '. Database: BoltDB.';
                    }
                } catch (e) {
                    console.error('Failed to parse build info:', e);
                }
            }
        };
    }
    
    /**
     * Custom confirmation modal for WKWebView compatibility
     * @param {string} message - Confirmation message
     * @param {string} type - Modal type: 'delete', 'complete', or default
     * @returns {Promise<boolean>} - Promise that resolves with user choice
     */
    function customConfirm(message, type = 'default') {
        return new Promise((resolve) => {
            // Clean up any existing modals first
            const existingModals = document.querySelectorAll('.confirm-modal');
            existingModals.forEach(modal => {
                if (modal.parentNode) {
                    modal.parentNode.removeChild(modal);
                }
            });
            
            // Generate unique ID for this modal
            const uniqueId = 'customConfirmModal_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            const cancelId = 'confirmCancel_' + uniqueId;
            const okId = 'confirmOk_' + uniqueId;
            
            // Create modal HTML with unique IDs
            const modalHtml = `
                <div class="confirm-modal ${type}" id="${uniqueId}">
                    <div class="confirm-modal-content">
                        <div class="confirm-modal-title">
                            ${type === 'delete' ? '🗑️ Delete Task' : 
                              type === 'complete' ? '✅ Complete Task' : '❓ Confirm Action'}
                        </div>
                        <div class="confirm-modal-message">${message}</div>
                        <div class="confirm-modal-buttons">
                            <button class="confirm-modal-button cancel" id="${cancelId}">Cancel</button>
                            <button class="confirm-modal-button confirm" id="${okId}">
                                ${type === 'delete' ? 'Delete' : 
                                  type === 'complete' ? 'Complete' : 'OK'}
                            </button>
                        </div>
                    </div>
                </div>
            `;
            
            // Add modal to DOM
            document.body.insertAdjacentHTML('beforeend', modalHtml);
            const modal = document.getElementById(uniqueId);
            const cancelButton = document.getElementById(cancelId);
            const okButton = document.getElementById(okId);
            
            // Show modal with animation
            setTimeout(() => modal.classList.add('show'), 10);
            
            // Handle button clicks
            const handleChoice = (confirmed) => {
                modal.classList.remove('show');
                // Remove event listeners to prevent memory leaks
                document.removeEventListener('keydown', escapeHandler);
                
                setTimeout(() => {
                    if (modal.parentNode) {
                        modal.parentNode.removeChild(modal);
                    }
                    resolve(confirmed);
                }, 300);
            };
            
            // Attach event listeners
            okButton.addEventListener('click', () => handleChoice(true));
            cancelButton.addEventListener('click', () => handleChoice(false));
            
            // Close on backdrop click
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    handleChoice(false);
                }
            });
            
            // Close on Escape key
            const escapeHandler = (e) => {
                if (e.key === 'Escape') {
                    handleChoice(false);
                }
            };
            document.addEventListener('keydown', escapeHandler);
        });
    }
    
    // Make customConfirm available globally
    window.customConfirm = customConfirm;

    document.addEventListener("DOMContentLoaded", function() {
        run();
        // Initialize gamification system
        if (window.Gamification) {
            window.Gamification.init();
        }
    });
}());
