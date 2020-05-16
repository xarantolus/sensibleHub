window.addEventListener('keydown', function (ev) {
    // Prevent the events below this from firing if the user is focusing an input element (e.g. while typing)
    if (ev.target.tagName === "INPUT") {
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
        return true;
    }
});