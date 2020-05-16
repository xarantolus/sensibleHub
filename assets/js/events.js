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
    }catch(e) {
        isReload = false;
    }

}


ws.onmessage = function (evt) {
    var e = JSON.parse(evt.data)

    console.log(e);


    // song-edit song-add song-delete
    if (e.type.startsWith("song-")) {
        if (isListingPage()) {
            // this listing might contain this song, so we reload 
            reload();
        } else {
            // If we are on a song page, we reload it on edit
            if (trimChar(location.pathname, "/") === "song/" + e.data.id) {
                reload();
            }
        }
    }

}
