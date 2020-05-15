// Show the cover image by putting it into the image container
function renderImagePreview(evt) {
    if (evt.files.length < 1) {
        return;
    }

    document.getElementById('cover-upload-button').innerHTML = evt.files[0].name;

    var img = document.getElementById('song-cover');
    img.src = window.URL.createObjectURL(evt.files[0]);
}

// Make clicking easier, allow clicking on image to select a file
function selectCover() {
    document.querySelector(".file-input").click()
}
document.querySelector(".song-image-container").addEventListener('click', selectCover);


// Confirm submitting if 
function confirmSubmit(evt) {
    if (!confirm("Are you sure you want to delete this song?")) {
        evt.preventDefault();
        return false;
    }
    return true;
}

document.getElementById("delete-button").addEventListener("click", confirmSubmit);

// Store and restore audio volume
function saveVolumeChange(evt) {
    localStorage.setItem("audio-volume", evt.target.volume);
}

var audioElement = document.getElementsByTagName("audio")[0];
audioElement.volume = localStorage.getItem("audio-volume") || 1;
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
songDownloadButton.addEventListener('blur', removeLoading)