
"use strict";
(function () {
    Modules.Loader.onTemplateLoaded("task", function (currentTemplate, currentId) {

        var taskVisible = currentTemplate.getElementsByClassName("task_visible")[0];
        taskVisible.addEventListener("click", taskVisibleClickHandler);
        currentTemplate.addEventListener('dragover', allowDrop);
        currentTemplate.addEventListener('drop', drop);
        currentTemplate.addEventListener('dragstart', drag);
        currentTemplate.addEventListener('contextmenu', copyToBuffer);
        var taskCompleteButton = currentTemplate.getElementsByClassName("taskCompletedButton")[0];
        var taskRemoveButton = currentTemplate.getElementsByClassName("taskRemoveButton")[0];
        var taskChangeButton = currentTemplate.getElementsByClassName("taskChangeButton")[0];
        var taskStartButton = currentTemplate.getElementsByClassName("taskStartButton")[0];
        var taskStopButton = currentTemplate.getElementsByClassName("taskStopButton")[0];
        taskCompleteButton.addEventListener("click", completeTaskHandler);
        taskRemoveButton.addEventListener("click", removeTaskHandler);
        taskChangeButton.addEventListener("click", changeTaskHandler);
        taskStartButton.addEventListener("click", startTaskHandler);
        taskStopButton.addEventListener("click", stopTaskHandler);
        
        function startTaskHandler() {
            var taskVisibleContent = currentTemplate.getElementsByClassName("task_visible_content")[0];
            // if (window.exports.timerId == 0) {
            //     window.exports.timerId = setInterval(tick, 1000);
            //     taskVisibleContent.classList.add("task_visible_content_active");
            // }
             if (window.exports.timerId == 0) {
                 window.exports.timerId = currentTemplate.dataset.uuid;
                 startTimer();
                 taskVisibleContent.classList.add("task_visible_content_active");
            }
        }

        function stopTaskHandler() {
            stopTimer();
            window.exports.timerId = 0;
            var taskVisibleContent = currentTemplate.getElementsByClassName("task_visible_content")[0];
            taskVisibleContent.classList.remove("task_visible_content_active");
        }

        var w = null;

        function startTimer()
        {

            if (w==null){
                w = new Worker("timer.js");
            }
            // Update timer div with output from Web Worker
            w.onmessage = function (e) {
                tick();
            };
        }


        function stopTimer()
        {
            w.terminate();
            w = null;
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
                    var duration_days = Math.floor(duration_seconds / (60 * 60 * 24));
                    duration_seconds -= duration_days * 60 * 60 * 24;
                    var duration_hours = Math.floor(duration_seconds / (60*60));
                    duration_seconds -= duration_hours * 60 * 60;
                    var duration_minutes = Math.floor(duration_seconds / 60);
                    duration_seconds -= duration_minutes * 60;
                    task_visible_timeExcecutionReal.innerHTML = "[Затрачено: "
                        + duration_days + " дн., "
                        + duration_hours + " : "
                        + duration_minutes + " : "
                        + duration_seconds + " ]";
                }
            }
        }

        function changeTaskHandler() {
            var task_visible_content = currentTemplate.getElementsByClassName("task_visible_content")[0];
            var taskText = document.getElementsByClassName("taskText")[0];
            var estimationDaysElement = document.getElementsByClassName("page_newTask_estimationDays_value")[0];
            var estimationHoursElement = document.getElementsByClassName("page_newTask_estimationHours_value")[0];
            var estimationMinutesElement = document.getElementsByClassName("page_newTask_estimationMinutes_value")[0];
            var deadlineMonthElement = document.getElementsByClassName("page_newTask_timeHardDeadlineMonth_value")[0];
            var deadlineDayElement = document.getElementsByClassName("page_newTask_timeHardDeadlineDay_value")[0];

            taskText.value = task_visible_content.innerHTML;
            var duration_seconds = currentTemplate.dataset.duration_execution_estimated_seconds;
            var duration_days = Math.floor(duration_seconds / (60 * 60 * 24));
            duration_seconds -= duration_days * 60 * 60 * 24;
            var duration_hours = Math.floor(duration_seconds / (60*60));
            duration_seconds -= duration_hours * 60 * 60;
            var duration_minutes = Math.floor(duration_seconds / 60);

            estimationDaysElement.value = duration_days;
            estimationHoursElement.value = duration_hours;
            estimationMinutesElement.value = duration_minutes;

            var deadlineDate = new Date(currentTemplate.dataset.time_hard_dead_line);
            var deadlineYear = deadlineDate.getFullYear();
            var deadMonth = deadlineDate.getMonth();
            var deadDay = deadlineDate.getDate();
            if (deadlineYear != 9999) {
                deadlineMonthElement.value = deadMonth;
                deadlineDayElement.value = deadDay;
            } else {
                deadlineMonthElement.value = 0;
                deadlineDayElement.value = 0;
            }



        }

        function removeTaskHandler() {

            var taskUUID = currentTemplate.getAttribute('data-uuid');
            var taskBody = currentTemplate.getAttribute('data-body');

            if (confirm('Вы действительно хотите УДАЛИТЬ задачу: ' + taskBody )) {
                Done.removeTask(taskUUID);
            }
        }

        function completeTaskHandler() {
            var taskUUID = currentTemplate.getAttribute('data-uuid');
            var taskBody = currentTemplate.getAttribute('data-body');

            if (confirm('Вы действительно хотите отчитаться о выполнении задачи: ' + taskBody )) {
                Done.completeTask(taskUUID);
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
        }

        function copyToBuffer(e) {
            // window.prompt("Задача: Ctrl+C, Enter", currentTemplate.getElementsByClassName("task_visible_content")[0].innerHTML);
            e.preventDefault();
        }

        function allowDrop(e) {
            e.preventDefault();
        }

        function drag(e) {
            var sourceTaskUUID = currentTemplate.getAttribute('data-uuid');
            e.dataTransfer.setData("sourceTaskUUID", sourceTaskUUID);
        }

        function drop(e) {
            e.preventDefault();
            var sourceTaskUUID = e.dataTransfer.getData("sourceTaskUUID");
            var destinationTaskUUID = currentTemplate.getAttribute('data-uuid');

            if ((destinationTaskUUID != undefined) && (destinationTaskUUID != sourceTaskUUID)) {
                Done.rearrangeTasks(sourceTaskUUID, destinationTaskUUID);
            }
        }
    });
}());