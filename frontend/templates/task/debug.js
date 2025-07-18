// Debug script to check template loading
console.log('Task template debug loaded');

// Check if Modules is available
if (typeof Modules !== 'undefined') {
    console.log('Modules available:', Modules);
    
    // Add event listener for template loaded
    Modules.Events.addItemLoadedListener(Modules.TEMPLATE, 'task', function(e) {
        console.log('Task template loaded event fired', e);
    });
    
    // Check onTemplateLoaded
    setTimeout(function() {
        console.log('Checking for task elements...');
        var tasks = document.getElementsByClassName('task');
        console.log('Found tasks:', tasks.length);
        
        for (var i = 0; i < tasks.length; i++) {
            console.log('Task ' + i + ':', {
                uuid: tasks[i].getAttribute('data-uuid'),
                draggable: tasks[i].querySelector('.task_visible').draggable,
                hasClickHandler: tasks[i]._listeners ? true : false
            });
        }
    }, 2000);
} else {
    console.error('Modules not available!');
}