window.addEventListener('keydown', function (ev) {
    // Escape from inputs
    if (ev.keyCode === 27) {
        ev.target.blur();
        ev.preventDefault()
    }

    // Prevent the events below this from firing if the user is focusing an input element (e.g. while typing)
    if (ev.target.tagName === "INPUT" || ev.target.tagName === "TEXTAREA") {
        return false;
    }

    // Navigation shortcuts
    var keyMap = {
        27: "", // Escape: main page
        78: "add", // 'n': add page
        83: "songs", // 's': song listing
        65: "artists", // 'a': artist listing
        89: "years", // 'y': years listing
        73: "incomplete", // 'i' incomplete songs listing
        85: "unsynced", // 'u' unsynced songs
    }

    var destination = keyMap[ev.keyCode];

    // if we know this destination and we're not already there
    // if (destination && ...) doesn't work as 27:"" is falsey
    if (destination !== undefined && location.pathname !== "/" + destination) {
        ev.preventDefault();
        InstantClick.go("/" + destination)
        return;
    }


    if (ev.key == "/") {
        ev.preventDefault();
        focusSearch();
        return;
    }
});