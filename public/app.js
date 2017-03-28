/**
 * Created by Kirill on 24.03.2017.
 */
var socket = new WebSocket("ws://localhost:8000/ws");
var state = {
    username: "",
    status: 0 //0 - before login, 1 - accepted, 2 - in queue, 3 - waiting for answer
};
var message = {
    username: "",
    text: "",
    service: false,
    time: ""
};

// Parse incoming messages (all, service and broadcast)
socket.onmessage = function(event) {
    var incomingMessage = event.data;
    var msg = JSON.parse(incomingMessage);

    if (!msg.service) {
        showMessage(msg);
    }
    else if(msg.username == "$USERLIST"){
        parseUserList(msg);
    }
    else if(msg.username == "$ERROR"){
        parseError(msg);
    }
    else if(msg.username == "$STATE"){
        parseState(msg);
    }
};
//Push alert on socket close
socket.onclose = function(event) {
    makeHidden("mainapp");
    alert('Сonnection closed. Probably, no free slots. Please, try later');
    //alert('Код: ' + event.code + ' причина: ' + event.reason);
};

function sendMessage() {
    var mi = document.getElementById("messageInput");
    var date = new Date();
    if (mi.value != ''){
        socket.send(JSON.stringify({
                username: state.username,
                text: mi.value,
                service: false,
                time: date.toLocaleString()
            })
        );
    }
    mi.value = '';
}

function showMessage(msg) {
    var container = document.createElement('div');
    var cm = document.getElementById("chatMessages");
    container.innerHTML = '<div class="card"><div class="chip light-blue">'
        + msg.username + '</div>'
        + msg.text + '<br>' + '<small>'
        + msg.time + '</small>' + '</div>';
    cm.insertBefore(container.firstChild, cm.firstChild);
} //

function loginClick() {
    var li = document.getElementById("loginInput");
    var newState = {
        username: li.value,
        status: 1
    };
    if(state.status != 3){
        askForChangeState(newState);
    }
}//

function askForChangeState(st) {
    var msg = {
        username:"$STATE",
        text: JSON.stringify(st),
        service: true,
        time: ""
    };
    socket.send(JSON.stringify(msg));
    state.status = 3;
} //

function parseState(msg) {
    //Parse STATE struct from text field
    state = JSON.parse(msg.text);
    //Change iface according to new state
    if (state.status == 0){
        showLoginState();
    }else if (state.status == 1){
        showUserState();
    }else if (state.status == 2){
        showQueueState();
    }
} //

function parseError(msg) {
    //Parse state struct from text field, error text sent in 'username'
    var err = JSON.parse(msg.text);
    alert(err.username);
}

function parseUserList(msg) {
    var ul = document.getElementById("usersList");
    var usersArray = msg.text.split(',');
    var container = document.createElement('div');
    while (ul.firstChild) {
        ul.removeChild(ul.firstChild);
    }
    usersArray.forEach(function(e){
        if (e != "") {
            container.innerHTML = '<div class="chip light-blue">' + e + '</div>';
            ul.appendChild(container.firstChild);
        }
    })
}

//Show 3 different interfaces

function showLoginState(){
    makeVisible("loginBlock");
    makeHidden("messageBlock");
    makeHidden("chatBlock");
    makeHidden("notify")
} //

function showUserState(){
    makeHidden("loginBlock");
    makeVisible("messageBlock");
    makeVisible("chatBlock");
    makeHidden("notify");

    document.getElementById("messageInput")
        .addEventListener("keyup", function(event) {
            event.preventDefault();
            if (event.keyCode == 13) {
                document.getElementById("sendButton").click();
            }
        });

} //

function showQueueState() {
    makeHidden("loginBlock");
    makeHidden("messageBlock");
    makeHidden("chatBlock");
    document.getElementById("notifyText").innerHTML = state.username + ", please wait for free slot";
    makeVisible("notify");
} //

function makeVisible(id){
    var e = document.getElementById(id);
    e.style.display = 'block';
    document.getElementById("usernameTag").innerHTML = state.username;
} //

function makeHidden(id) {
    var e = document.getElementById(id);
    e.style.display = 'none';
} //


