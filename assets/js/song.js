function songPage() {
    registerCover();

    // Confirm submitting when deleting song
    function confirmDelete(evt) {
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
    document.getElementById("delete-button").addEventListener("click", confirmDelete);


    function confirmDeleteCover(evt) {
        evt.preventDefault();

        if (!confirm("Are you sure you want to delete the cover?")) {
            return false;
        }

        var formData = new FormData();
        formData.set("delete-cover", "delete-cover");

        ajax(location.pathname, formData).post(function (status, obj) {
            // If successful, a reload will be done from events.js (song-edit event)
            if (status !== 200) {
                document.getElementById("song-notif").innerText = obj.message || "Unknown error";
            }
        })
        return false;
    }
    // This one is rendered conditionally
    var dc = document.getElementById("delete-cover");
    if (dc) {
        dc.addEventListener("click", confirmDeleteCover);
    }

    // Store and restore audio volume
    function saveVolumeChange(evt) {
        localStorage.setItem("audio-volume", evt.target.volume);
    }

    // initialize mediaSession metadata
    if ('mediaSession' in navigator) {
        // if span is visible, we have a non-placeholder image
        var coverSizeSpan = document.querySelector(".cover-image-size");
        var coverImage = document.getElementById("song-cover");
        var artwork = [];
        if (coverSizeSpan && coverImage) {
            artwork.push({ src: coverImage.src + "?size=small", sizes: "120x120" });
            artwork.push({ src: coverImage.src, sizes: coverImage.dataset.size + "x" + coverImage.dataset.size });
        }
        
        navigator.mediaSession.metadata = new MediaMetadata({
            title: document.getElementById("song-title").value,
            artist: document.getElementById("song-artist").value,
            album: document.getElementById("song-album").value,
            artwork: artwork,
        });
    }


    var audioElement = document.getElementsByTagName("audio")[0];
    // If this page is loaded from InstantClick's cache, then the audio src might be removed
    if (audioElement.src == "") {
        audioElement.src = location.href + "/audio";
    }

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
