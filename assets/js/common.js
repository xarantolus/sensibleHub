var isReload = false;
var oldX = 0;
var oldY = 0;

// Allow programmatically visiting a page
InstantClick.go = function (url) {
    oldX = window.scrollX;
    oldY = window.scrollY;

    var link = document.createElement('a');
    link.href = url;
    document.body.appendChild(link);
    link.click();
}

InstantClick.on('change', function () {
    if (!isReload) {
        return;
    }
    isReload = false;
    window.scrollTo(oldX, oldY)
})

// Method for Requesting Data
// Based on https://gist.github.com/duanckham/e5b690178b759603b81c
// usage(POST): ajax(url, data).post(function(status, obj) { });
// usage(GET): ajax(url, data).get(function(status, obj) { });
var ajax = function (url, data) {
    var wrap = function (method, cb) {
        var xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        xhr.setRequestHeader("X-XHR", "true");

        var sendstr = null;
        if (method === "POST" && data) {
            if ('entries' in data) {
                sendstr = data;
            } else {
                xhr.setRequestHeader("Content-Type", "application/json");
                sendstr = JSON.stringify(data);
            }
        }

        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4 && xhr.status > 0) {
                try {
                    var resp = JSON.parse(xhr.responseText)
                } catch (e) {
                    cb(422, { message: xhr.responseText })
                    return;
                }
                cb(xhr.status, resp);
            }
        }

        xhr.onerror = function () {
            cb(503, { message: "Error while connecting." })
        }

        xhr.send(sendstr);

        return xhr;
    };

    return {
        get: function (cb) {
            return wrap("GET", cb);
        },
        post: function (cb) {
            return wrap("POST", cb);
        }
    };
};

// https://stackoverflow.com/a/26156806
function trimChar(string, charToRemove) {
    while (string.charAt(0) == charToRemove) {
        string = string.substring(1);
    }

    while (string.charAt(string.length - 1) == charToRemove) {
        string = string.substring(0, string.length - 1);
    }

    return string;
}

// must this page be reloaded after
function isListingPage() {
    return location.pathname.startsWith("/album/") || location.pathname.startsWith("/artist/") || ["/", "/add", "/songs", "/artists", "/years", "/incomplete", "/unsynced", "/search", "/edits"].indexOf(location.pathname) !== -1;
}

function registerCover() {
    // Show the cover image by putting it into the image container
    function renderImagePreview(evt) {
        var files = evt.target.files;
        if (files.length != 1) {
            return;
        }

        document.getElementById('cover-upload-button').innerText = files[0].name;


        var img = document.getElementById('song-cover');
        img.src = window.URL.createObjectURL(files[0]);

        // remove the styled border as no longer know the dominant color
        img.style.backgroundColor = null;
        img.style.border = 0;
        img.style.borderWidth = 0;

        // Also remove the hover-shinethrough background
        var imgBg = document.querySelector('figure.image');
        if (imgBg) {
            imgBg.style.background = "";
        }


        // we also no longer know the cover size - this is only present on song pages
        var sizeSpan = document.querySelector(".cover-image-size");
        if (sizeSpan) {
            sizeSpan.style.visibility = "hidden";
        }
    }

    document.getElementById("song-cover-input").addEventListener("change", renderImagePreview);

    // Make clicking easier, allow clicking on image to select a file
    function selectCover(evt) {
        if (evt.target.tagName == "INPUT" || evt.target.tagName === "LABEL" || evt.target.tagName == "BUTTON") {
            // Ignore other clicked elements as they will trigger it again. That's not good
            return;
        }
        document.querySelector(".file-input").click()
    }

    (document.querySelector(".song-image-container") || document.querySelector(".album-image-container")).addEventListener('click', selectCover);
}


// preloadImage preloads the image at the given URL
function preloadImage(url) {
    if (Image) {
        var img = new Image();
        img.src = url;
        img.onload = function () { };
    }
}

InstantClick.on('change', function () {
    // Preload a maximum of 10 song covers
    function loadCovers() {
        var linkedSongs = document.getElementsByClassName("song-link");
        for (var i = 0; i < Math.min(linkedSongs.length, 10); i++) {
            if (linkedSongs[i].id.startsWith("song-")) {
                preloadImage("/song/" + linkedSongs[i].id.substr(5) + "/cover")
            }
        }
    }

    // Present on album, artist & song page
    var sc = document.getElementById("song-cover");
    if (sc || document.location.pathname == "/" || document.location.pathname == "/search") {
        (sc || window).addEventListener("load", () => setTimeout(loadCovers, 2500));
    }
})

// When we click on a song link, we already preload the cover
InstantClick.on('receive', function (url, body, title) {
    // Trim URL prefix
    var pref = location.protocol + "//" + location.host;
    if (url.startsWith(pref)) {
        url = pref.substr(pref.length);
    }

    var split = url.split("/");

    if (split[0] == "song") {
        preloadImage("/song/" + split[1] + "/cover")
    }
})
