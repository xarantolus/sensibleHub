function createWebSocket(path) {
    var protocolPrefix = (window.location.protocol === 'https:') ? 'wss:' : 'ws:';
    return new ReconnectingWebSocket(protocolPrefix + '//' + location.host + path, null, { reconnectDecay: 1 });
}

var firstConnect = true;

var ws = createWebSocket("/api/v1/events/ws")

// detect reconnects
ws.onopen = function () {
    if (!firstConnect) {
        location.reload();
    }
    firstConnect = false;
}


function reload() {
    // this makes sure that instantclick cannot scroll, see common.js
    isReload = true;

    try {
        InstantClick.go(location.toString())
    } catch (e) {
        isReload = false;
    }
}

var changedItems = {};

var lastProgress = "progress-end"; // default: don't show
ws.onmessage = function (evt) {
    var e = JSON.parse(evt.data)

    console.log(e);


    // song-edit song-add song-delete
    if (e.type.startsWith("song-")) {
        if (isListingPage()) {
            var selem = document.getElementById("song-" + e.data.id);
            if (selem && e.type == "song-delete") {
                selem.remove();
            } else {
                // this listing might contain this song, so we reload 
                reload();
            }
        } else if (e.type !== "song-delete" && !isReload) {
            // If we are on a song page, we reload it on edit
            if (trimChar(location.pathname, "/") === "song/" + e.data.id) {
                reload();
            }
        }
        
        if (e.type !== "song-delete") {
            changedItems[e.data.id] = Math.random();
        }
    }

    if (e.type.startsWith("progress-")) {
        setProgressbar(e.type, e.data)
        lastProgress = e.type;

        if (location.pathname === "/add" && document.getElementById("searchTerm").value.trim() === "") {
            reload();
        }
    }
}

InstantClick.on('receive', function (url, body, title) {
    // Replace all image references to the last changed song - they would not be updated otherwise

    var selem = null;

    // For song page
    var sid = body.querySelector("#song-id")

    // If it has been changed before we have reloaded
    if (sid && changedItems.hasOwnProperty(sid.value)) {
        selem = body.querySelector("#song-cover");
        if (selem) {
            selem.src = selem.src + "#" + changedItems[sid.value];
        }
    }

    // In listings. We need to do this every time
    Object.keys(changedItems).forEach(function(id){
        var i = body.querySelector("#img-" + id)
        if (i) {
            i.src = i.src + "#" + changedItems[id];
            return;
        }
    })

    return {
        body: body,
        title: title
    }
})

function setProgressbar(event, data) {
    var progressBar = document.getElementById("main-progress");
    switch (event) {
        case "progress-start":
            progressBar.style.display = "block";
            break;
        case "progress-end":
            progressBar.style.display = "none";
            break;
        default:
            break;
    }
}

InstantClick.on('change', function () {
    setProgressbar(lastProgress)
})