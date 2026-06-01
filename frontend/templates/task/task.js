
"use strict";
(function () {
    Modules.Loader.onTemplateLoaded("task", function (currentTemplate, currentId) {

        var taskVisible = currentTemplate.getElementsByClassName("task_visible")[0];
        taskVisible.addEventListener("click", taskVisibleClickHandler);
        taskVisible.addEventListener("keydown", taskVisibleKeydownHandler);
        currentTemplate.addEventListener('dragover', allowDrop);
        currentTemplate.addEventListener('drop', drop);
        currentTemplate.addEventListener('dragstart', drag);
        currentTemplate.addEventListener('dragend', dragEnd);
        currentTemplate.addEventListener('dragleave', dragLeave);
        currentTemplate.addEventListener('contextmenu', copyToBuffer);
        var taskCompleteButton = currentTemplate.getElementsByClassName("taskCompletedButton")[0];
        var taskRemoveButton = currentTemplate.getElementsByClassName("taskRemoveButton")[0];
        var taskChangeButton = currentTemplate.getElementsByClassName("taskChangeButton")[0];
        var taskStartButton = currentTemplate.getElementsByClassName("taskStartButton")[0];
        var taskStopButton = currentTemplate.getElementsByClassName("taskStopButton")[0];
        var taskMoveUpButton = currentTemplate.getElementsByClassName("taskMoveUpButton")[0];
        var taskMoveDownButton = currentTemplate.getElementsByClassName("taskMoveDownButton")[0];
        // Check if buttons exist
        if (!taskCompleteButton || !taskRemoveButton || !taskChangeButton) {
            console.error('Some task buttons not found!');
            return;
        }
        
        // Try multiple event binding approaches for WKWebView compatibility
        taskCompleteButton.addEventListener("click", completeTaskHandler);
        taskCompleteButton.addEventListener("touchend", completeTaskHandler);
        taskCompleteButton.onclick = completeTaskHandler;
        
        taskRemoveButton.addEventListener("click", removeTaskHandler);
        taskRemoveButton.addEventListener("touchend", removeTaskHandler);
        taskRemoveButton.onclick = removeTaskHandler;
        
        taskChangeButton.addEventListener("click", changeTaskHandler);
        taskChangeButton.addEventListener("touchend", changeTaskHandler);
        taskChangeButton.onclick = changeTaskHandler;
        taskStartButton.addEventListener("click", startTaskHandler);
        taskStopButton.addEventListener("click", stopTaskHandler);
        taskMoveUpButton.addEventListener("click", function() { moveTask(-1); });
        taskMoveDownButton.addEventListener("click", function() { moveTask(1); });

        function setActiveState(isActive) {
            var taskVisibleContent = currentTemplate.getElementsByClassName("task_visible_content")[0];
            currentTemplate.classList.toggle("task--active", isActive);
            taskVisibleContent.classList.toggle("task_visible_content_active", isActive);
            taskStartButton.setAttribute("aria-pressed", isActive ? "true" : "false");
        }
        
        function startTaskHandler() {
            if (window.exports.timerId && window.exports.timerId !== currentTemplate.dataset.uuid && window.TaskTimerRegistry) {
                window.TaskTimerRegistry.stop(window.exports.timerId);
            }

             if (window.exports.timerId == 0 || window.exports.timerId === currentTemplate.dataset.uuid) {
                 window.exports.timerId = currentTemplate.dataset.uuid;
                 startTimer();
                 setActiveState(true);
                 if (window.NinstyleSounds) {
                     window.NinstyleSounds.taskStart();
                 }
            }
        }

        function stopTaskHandler() {
            stopTimer();
            window.exports.timerId = 0;
            setActiveState(false);
        }

        var w = null;

        window.TaskTimerRegistry = window.TaskTimerRegistry || {
            stops: {},
            register: function(uuid, stopFn) {
                this.stops[uuid] = stopFn;
            },
            unregister: function(uuid) {
                delete this.stops[uuid];
            },
            stop: function(uuid) {
                if (this.stops[uuid]) {
                    this.stops[uuid]();
                }
                if (window.exports.timerId === uuid) {
                    window.exports.timerId = 0;
                }
            },
            stopAll: function() {
                var stops = this.stops;
                Object.keys(stops).forEach(function(uuid) {
                    stops[uuid]();
                });
                window.exports.timerId = 0;
            }
        };

        function startTimer()
        {

            if (w==null){
                w = new Worker("timer.js");
                window.TaskTimerRegistry.register(currentTemplate.dataset.uuid, stopTimer);
            }
            // Update timer div with output from Web Worker
            w.onmessage = function (e) {
                tick();
            };
        }


        function stopTimer()
        {
            if (w == null) {
                setActiveState(false);
                return;
            }
            w.terminate();
            w = null;
            window.TaskTimerRegistry.unregister(currentTemplate.dataset.uuid);
            setActiveState(false);
        }

        function tick() {

            currentTemplate.dataset.duration_execution_real_seconds++;
            updateRealTime(currentTemplate.dataset.uuid, currentTemplate.dataset.duration_execution_real_seconds);

        }

        function updateRealTime(uuid, duration_execution_real_seconds) {
            var task_visible_timeExcecutionReal = currentTemplate.getElementsByClassName("task_visible_timeExcecutionReal")[0];

            var xhr = new XMLHttpRequest();
            xhr.open('POST', "/api/updateTaskExecutionRealSeconds", true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.send(uuid + "$;" + duration_execution_real_seconds);
            xhr.onreadystatechange = function() {
                if (xhr.readyState == XMLHttpRequest.DONE) {
                    var duration_seconds = currentTemplate.dataset.duration_execution_real_seconds;
                    var duration_days = Math.floor(duration_seconds / (60 * 60 * 8));
                    duration_seconds -= duration_days * 60 * 60 * 8;
                    var duration_hours = Math.floor(duration_seconds / (60*60));
                    duration_seconds -= duration_hours * 60 * 60;
                    var duration_minutes = Math.floor(duration_seconds / 60);
                    duration_seconds -= duration_minutes * 60;
                    task_visible_timeExcecutionReal.textContent = "Actual "
                        + duration_days + " d., "
                        + duration_hours + " : "
                        + duration_minutes + " : "
                        + duration_seconds;
                }
            }
        }

	        function changeTaskHandler() {
	            window.exports.editingTaskUUID = currentTemplate.dataset.uuid;
	            var task_visible_content = currentTemplate.getElementsByClassName("task_visible_content")[0];
	            var taskText = document.getElementsByClassName("taskText")[0];
            var estimationDaysElement = document.getElementsByClassName("page_newTask_estimationDays_value")[0];
            var estimationHoursElement = document.getElementsByClassName("page_newTask_estimationHours_value")[0];
            var estimationMinutesElement = document.getElementsByClassName("page_newTask_estimationMinutes_value")[0];
            var deadlineMonthElement = document.getElementsByClassName("page_newTask_timeHardDeadlineMonth_value")[0];
            var deadlineDayElement = document.getElementsByClassName("page_newTask_timeHardDeadlineDay_value")[0];
            var deadlineYearElement = document.getElementsByClassName("page_newTask_timeHardDeadlineYear_value")[0];

            // Get text content preserving line breaks
            var taskContent = task_visible_content.innerText || task_visible_content.textContent;
            taskText.value = taskContent;
	            var duration_seconds = parseInt(currentTemplate.dataset.duration_execution_estimated_seconds, 10);
	            var duration_days = Math.floor(duration_seconds / (60 * 60 * 8));
	            duration_seconds -= duration_days * 60 * 60 * 8;
	            var duration_hours = Math.floor(duration_seconds / (60*60));
            duration_seconds -= duration_hours * 60 * 60;
            var duration_minutes = Math.floor(duration_seconds / 60);

            estimationDaysElement.value = duration_days;
            estimationHoursElement.value = duration_hours;
            estimationMinutesElement.value = duration_minutes;

            var deadlineDate = new Date(currentTemplate.dataset.time_hard_dead_line);
            var deadlineYear = deadlineDate.getFullYear();
            var deadMonth = deadlineDate.getMonth() + 1; // JS months are 0-indexed
            var deadDay = deadlineDate.getDate();
            if (deadlineYear != 9999) {
                deadlineMonthElement.value = deadMonth;
                deadlineDayElement.value = deadDay;
                deadlineYearElement.value = deadlineYear;
            } else {
                deadlineMonthElement.value = "";
                deadlineDayElement.value = "";
                deadlineYearElement.value = "";
            }

            if (window.setTaskComposerMode) {
                window.setTaskComposerMode('edit');
            }
            taskText.focus();
            taskText.scrollIntoView({behavior: 'smooth', block: 'center'});


        }

        function removeTaskHandler() {
            // Prevent multiple simultaneous calls
            if (removeTaskHandler.inProgress) {
                return;
            }
            removeTaskHandler.inProgress = true;
            
            var taskUUID = currentTemplate.getAttribute('data-uuid');
            var taskBody = currentTemplate.getAttribute('data-body');

            // Use custom confirmation modal
            if (window.customConfirm) {
                window.customConfirm('Delete this task? ' + taskBody, 'delete')
                    .then(function(confirmed) {
                        if (confirmed) {
                            if (window.TaskTimerRegistry) {
                                window.TaskTimerRegistry.stop(taskUUID);
                            }
                            Done.removeTask(taskUUID);
                        }
                        removeTaskHandler.inProgress = false;
                    })
                    .catch(function() {
                        removeTaskHandler.inProgress = false;
                    });
            } else {
                // Fallback to native confirm if customConfirm not available
                if (confirm('Delete this task? ' + taskBody)) {
                    if (window.TaskTimerRegistry) {
                        window.TaskTimerRegistry.stop(taskUUID);
                    }
                    Done.removeTask(taskUUID);
                }
                removeTaskHandler.inProgress = false;
            }
        }

        function completeTaskHandler() {
            // Prevent multiple simultaneous calls
            if (completeTaskHandler.inProgress) {
                return;
            }
            completeTaskHandler.inProgress = true;
            
            var taskUUID = currentTemplate.getAttribute('data-uuid');
            var taskBody = currentTemplate.getAttribute('data-body');

            // Use custom confirmation modal
            if (window.customConfirm) {
                window.customConfirm('Mark this task as done? ' + taskBody, 'complete')
                    .then(function(confirmed) {
                        if (confirmed) {
                            if (window.TaskTimerRegistry) {
                                window.TaskTimerRegistry.stop(taskUUID);
                            }
                            Done.completeTask(taskUUID);
                        }
                        completeTaskHandler.inProgress = false;
                    })
                    .catch(function() {
                        completeTaskHandler.inProgress = false;
                    });
            } else {
                // Fallback to native confirm if customConfirm not available
                if (confirm('Mark this task as done? ' + taskBody)) {
                    if (window.TaskTimerRegistry) {
                        window.TaskTimerRegistry.stop(taskUUID);
                    }
                    Done.completeTask(taskUUID);
                }
                completeTaskHandler.inProgress = false;
            }
        }

        function taskVisibleClickHandler() {
            var collapsedClassName = "collapsed";
            var taskControl = currentTemplate.getElementsByClassName("task_controls")[0];
            var classList = taskControl.className.split(" ");
            var find = false;
            for (var i = 0; i < classList.length; i++) {
                if (classList[i] == collapsedClassName) {
                    classList.splice(i, 1);
                    find = true;
                }
            }
            if (!find) {
                classList.push(collapsedClassName);
            }

            taskControl.className = classList.join(" ");
            taskVisible.setAttribute("aria-expanded", find ? "true" : "false");
            currentTemplate.classList.toggle("task--expanded", find);
        }

        function taskVisibleKeydownHandler(e) {
            if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                taskVisibleClickHandler();
            }
        }

        function moveTask(direction) {
            var tasks = Array.prototype.slice.call(document.querySelectorAll(".page_tasks_content .task"));
            var currentIndex = tasks.indexOf(currentTemplate);
            if (currentIndex === -1) {
                return;
            }

            if (direction < 0 && currentIndex > 0) {
                Done.rearrangeTasks(currentTemplate.dataset.uuid, tasks[currentIndex - 1].dataset.uuid, true);
            } else if (direction > 0 && currentIndex < tasks.length - 1) {
                Done.rearrangeTasks(currentTemplate.dataset.uuid, tasks[currentIndex + 1].dataset.uuid, false);
            }
        }

        function copyToBuffer(e) {
            // window.prompt("Задача: Ctrl+C, Enter", currentTemplate.getElementsByClassName("task_visible_content")[0].innerHTML);
            e.preventDefault();
        }

        function allowDrop(e) {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
            
            // Add visual feedback for drop target only if not the dragged element
            if (!currentTemplate.classList.contains('dragging')) {
                // Get the position of the cursor relative to the task
                var rect = currentTemplate.getBoundingClientRect();
                var y = e.clientY - rect.top;
                var height = rect.height;
                
                // Get the dragging element to adjust drop zone size
                var draggingElement = document.querySelector('.task.dragging');
                if (draggingElement) {
                    var draggedHeight = draggingElement.offsetHeight;
                    // Adjust drop zone margins based on dragged element height
                    var dropZoneHeight = draggedHeight + 20; // Add some padding
                    
                    // Apply dynamic sizing using CSS variables
                    currentTemplate.style.setProperty('--drop-zone-height', dropZoneHeight + 'px');
                    currentTemplate.style.setProperty('--drop-zone-margin', (dropZoneHeight + 10) + 'px');
                    currentTemplate.style.setProperty('--drop-zone-offset', '-' + (dropZoneHeight + 10) + 'px');
                }
                
                // Remove all drag-over classes first
                var allTasks = document.querySelectorAll('.task');
                allTasks.forEach(function(task) {
                    task.classList.remove('drag-over', 'drag-over-top');
                });
                
                // If cursor is in top half, prepare to insert before this task
                if (y < height / 2) {
                    currentTemplate.classList.add('drag-over-top');
                } else {
                    // If cursor is in bottom half, prepare to insert after
                    currentTemplate.classList.add('drag-over');
                }
            }
        }

        function drag(e) {
            var sourceTaskUUID = currentTemplate.getAttribute('data-uuid');
            e.dataTransfer.setData("sourceTaskUUID", sourceTaskUUID);
            e.dataTransfer.effectAllowed = 'move';
            
            // Add dragging class to the element being dragged
            currentTemplate.classList.add('dragging');
            
            // Add drag-active class to container
            var tasksContent = document.querySelector('.page_tasks_content');
            if (tasksContent) {
                tasksContent.classList.add('drag-active');
            }
            
            if (window.NinstyleSounds) {
                window.NinstyleSounds.dragStart();
            }
            
        }

        function drop(e) {
            e.preventDefault();
            var sourceTaskUUID = e.dataTransfer.getData("sourceTaskUUID");
            var destinationTaskUUID = currentTemplate.getAttribute('data-uuid');
            
            // Check if we should insert before or after
            var insertBefore = currentTemplate.classList.contains('drag-over-top');

            if ((destinationTaskUUID != undefined) && (destinationTaskUUID != sourceTaskUUID)) {
                Done.rearrangeTasks(sourceTaskUUID, destinationTaskUUID, insertBefore);
                if (window.NinstyleSounds) {
                    window.NinstyleSounds.drop();
                }
            }
            
            // Clean up classes
            currentTemplate.classList.remove('drag-over', 'drag-over-top');
        }
        
        function dragEnd(e) {
            // Clean up dragging class
            currentTemplate.classList.remove('dragging');
            
            // Remove drag-active from container
            var tasksContent = document.querySelector('.page_tasks_content');
            if (tasksContent) {
                tasksContent.classList.remove('drag-active');
            }
            
            // Remove drag-over from all tasks and reset CSS variables
            var allTasks = document.querySelectorAll('.task');
            allTasks.forEach(function(task) {
                task.classList.remove('drag-over', 'drag-over-top');
                // Reset CSS variables to defaults
                task.style.removeProperty('--drop-zone-height');
                task.style.removeProperty('--drop-zone-margin');
                task.style.removeProperty('--drop-zone-offset');
            });
        }
        
        function dragLeave(e) {
            // Remove drag-over class when leaving a task
            if (e.target === currentTemplate) {
                currentTemplate.classList.remove('drag-over', 'drag-over-top');
            }
        }
    });
}());
