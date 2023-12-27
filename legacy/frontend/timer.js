let timerStart = true;

function tick()
{
    postMessage("")
}

if (timerStart){
    let myVar = setInterval(tick,1000);
    // timer should not start anymore since it has been started
    timerStart = false;
}