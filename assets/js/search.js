
var isSearch = false;

var nextUrl = null;

InstantClick.on('change', function () {
    // These elements have been replaced
    isSearch = false;
    isReload = false;


    if (nextUrl) {
        history.replaceState("", "", nextUrl)
        nextUrl = null;
    } else {
        // Scroll to the top result
        window.scrollY = 0;

        // only focus search bar if we have *not* been redirected
        if (isSearch) {
            focusSearch();
        }
    }
})

InstantClick.on('change', function () {
    // Try to clear MediaSession API to make sure we can no longer use media keys
    if ('mediaSession' in navigator) {
        navigator.mediaSession.playbackState = "none"
        navigator.mediaSession.metadata = null;
        // firefox keeps playing audio, so we remove its source
        var a = document.querySelector("audio");
        if (a) {
            a.pause();
            a.src = "";
            a.src = null;
            a.removeAttribute("src");
        }
    }
})

InstantClick.on('receive', function (url, body, title) {
    // song is there, we have been redirected. Sadly instantclick.js doesn't handle this, so we need to check it here
    var song = body.querySelector("#song-id");
    if (song) {
        nextUrl = "/song/" + song.value;
    }

    if (!isSearch) {
        return;
    }

    return {
        body: body,
        title: title
    };
});



function initSearch() {
    var sb = document.getElementById("search-bar");
    var sc = document.getElementById("search-container");
    var sf = document.getElementById("search-form");

    function setLoading(l) {
        if (l) {
            sc.classList.add("is-loading")
        } else {
            sc.classList.remove("is-loading")
        }
    }

    sf.onsubmit = function (evt) {
        evt.preventDefault();

        isSearch = true;
        isReload = true;

        setLoading(true);

        InstantClick.go("/search?q=" + encodeURIComponent(sb.value))
    }
}

function focusSearch() {
    var sb = document.getElementById("search-bar");
    if (sb) {
        // If the search bar is hidden, we show it
        document.querySelector('.navbar-burger').classList.toggle('is-active');
        document.querySelector('.navbar-menu').classList.toggle('is-active');

        if (!sb.value.endsWith(" ") && sb.value.length !== 0) {
            sb.value += " "
        }

        sb.setSelectionRange(sb.value.length, sb.value.length);
        sb.focus();
    }
}
