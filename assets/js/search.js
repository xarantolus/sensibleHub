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
    if (url.indexOf("album") !== -1 || url.indexOf("artist") !== -1) {
        return;
    }

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

var createSuggestionNode = function (obj) {
    var n = document.createElement("a");
    n.href = "/song/" + obj.id;
    n.className = "navbar-item autosuggest-song";
    n.dataset.title = obj.title;

    if (obj.selected) {
        n.className += " search-selected";
    }

    var span = document.createElement("span");
    span.className = "autosuggest-song-name";
    span.innerText = obj.title;

    n.append(span);

    return n;
}

function searchSuggestions(loading) {
    var searchText = document.getElementById("search-bar").value.trim();

    if (searchText == "") {
        var resultTarget = document.getElementById("search-suggestions");
        resultTarget.innerHTML = "";
        return;
    }

    loading(true);

    ajax("/api/v1/search?q=" + encodeURIComponent(searchText)).get(function (status, obj) {
        if (status !== 200) {
            return;
        }
        
        // Typing too fast / Slow network
        if (obj.query !== searchText) return;

        if (obj.results.length > 0) {
            // If the selected result is the same as last time, we keep it
            // Also if there's only one, we also select it (we would get redirected anyways)
            if (obj.results.length == 1 || obj.results[0].title == selectedText) {
                obj.results[0].selected = true;
            } else {
                selectedIndex = -1;
                selectedText = "";
            }
        }

        var resultTarget = document.getElementById("search-suggestions");
        resultTarget.innerHTML = "";

        for (var i = 0; i < obj.results.length; i++) {
            resultTarget.append(createSuggestionNode(obj.results[i]));
        }

        loading(false);
    });
}


var selectedIndex = -1;
var selectedText = "";

function searchKeypress(evt) {
    // Handle UP/DOWN arrows

    if (evt.key === "ArrowUp" || evt.which === 38) {
        // UP Arrow
        var suggestions = document.getElementById("search-suggestions");
        if (selectedIndex == 0) {
            // Deselect
            suggestions.children[selectedIndex].classList.remove("search-selected");

            selectedIndex = -1;
            selectedText = "";
        } else if (selectedIndex > 0) {
            suggestions.children[selectedIndex].classList.remove("search-selected");

            // Choose element
            selectedIndex--;
            selectedText = suggestions.children[selectedIndex].dataset.title;

            suggestions.children[selectedIndex].classList.add("search-selected");
        }

        evt.preventDefault();
        return false;
    }

    if (evt.key === "ArrowDown" || evt.which === 40) {
        // DOWN Arrow
        var suggestions = document.getElementById("search-suggestions");
        evt.preventDefault();

        if (selectedIndex == -1) {
            // Select the first item
            selectedIndex = 0;
            selectedText = suggestions.children[selectedIndex].dataset.title;

            suggestions.children[selectedIndex].classList.add("search-selected");
        } else if (selectedIndex + 1 < suggestions.children.length) {
            suggestions.children[selectedIndex].classList.remove("search-selected");

            // Choose element
            selectedIndex++;
            selectedText = suggestions.children[selectedIndex].dataset.title;

            suggestions.children[selectedIndex].classList.add("search-selected");
        }

        evt.preventDefault();
        return false;
    }

    if (evt.key == "Enter" || evt.which === 13) {
        var suggestions = document.getElementById("search-suggestions");

        if (selectedIndex > -1 && selectedIndex < suggestions.children.length) {
            evt.preventDefault();
            suggestions.children[selectedIndex].click();
            return false;
        }
    }

    return true;
}


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

    sb.oninput = function (evt) {
        searchSuggestions(setLoading);
    }
    sb.onkeydown = searchKeypress;
    sb.onfocus = function (evt) {
        searchSuggestions(setLoading);
    }

    sb.onblur = function () {
        setTimeout(function () {
            var resultTarget = document.getElementById("search-suggestions");
            resultTarget.innerHTML = "";
        }, 500)
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
