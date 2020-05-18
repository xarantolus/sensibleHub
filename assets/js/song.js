function songPage() {
    // Show the cover image by putting it into the image container
    function renderImagePreview(evt) {
        var files = evt.target.files;
        if (files.length != 1) {
            return;
        }

        document.getElementById('cover-upload-button').innerHTML = files[0].name;

        var img = document.getElementById('song-cover');
        img.src = window.URL.createObjectURL(files[0]);
    }

    document.getElementById("song-cover-input").addEventListener("change", renderImagePreview);

    // Make clicking easier, allow clicking on image to select a file
    function selectCover(evt) {
        if (evt.target.tagName === "LABEL") {
            return;
        }
        document.querySelector(".file-input").click()
    }
    document.querySelector(".song-image-container").addEventListener('click', selectCover);


    // Confirm submitting if 
    function confirmSubmit(evt) {
        evt.preventDefault();

        if (!confirm("Are you sure you want to delete this song?")) {
            return false;
        }

        var formData = new FormData();
        formData.set("delete", "delete");

        ajax(location.pathname, formData).post(function (status, obj) {
            if (status === 200) {
                isReload = true;
                InstantClick.go("/");
            } else {
                document.getElementById("song-notif").innerText = obj.message || "Unknown error";
            }
        })
        return false;
    }

    document.getElementById("delete-button").addEventListener("click", confirmSubmit);

    // Store and restore audio volume
    function saveVolumeChange(evt) {
        localStorage.setItem("audio-volume", evt.target.volume);
    }

    var audioElement = document.getElementsByTagName("audio")[0];

    // Fix audio issue in firefox: sometimes it doesn't load the audio because "not all candidates could be loaded", then it disables "media loading" for the page
    audioElement.load();

    audioElement.volume = localStorage.getItem("audio-volume") || 0;
    if (audioElement.volume == 0) {
        audioElement.volume = 0.5;
    }
    audioElement.addEventListener("volumechange", saveVolumeChange);

    // Download button loading animation
    var songDownloadButton = document.getElementById("download-song-button");

    function removeLoading(evt) {
        songDownloadButton.classList.remove("is-loading");
    }
    function downloadButtonClicked() {
        songDownloadButton.classList.add("is-loading");
    }

    songDownloadButton.addEventListener('click', downloadButtonClicked);
    songDownloadButton.addEventListener('blur', removeLoading);

}