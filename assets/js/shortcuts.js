window.addEventListener('keydown', function (ev) {
    // Escape from inputs
    if (ev.keyCode === 27) {
        ev.target.blur();
        ev.preventDefault()
        return;
    }

    // Prevent the events below this from firing if the user is focusing an input element (e.g. while typing)
    if (ev.target.tagName === "INPUT" || ev.target.tagName === "TEXTAREA") {
        return false;
    }

    if (ev.keyCode === 27 && location.pathname !== "/") {
        // ESC => Return to main page if not already there
        InstantClick.go("/");
        return;
    }

    // 'n' => Load add page
    if (ev.keyCode === 78 && location.pathname != "/add") {
        ev.preventDefault();
        InstantClick.go("/add");
        return;
    }

    if (ev.key == "/") {
        ev.preventDefault();
        focusSearch();
        return;
    }
});