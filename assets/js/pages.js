InstantClick.on('change', function () {
    switch (location.pathname.split("/")[1]) {
        case "song":
            songPage();
            break;
        case "add":
            addPage();
            break;
        case "album":
            albumPage();
        default:
            break;
    }

    initSearch();
})
