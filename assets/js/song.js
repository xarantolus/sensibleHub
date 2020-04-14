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

document.getElementById("delete-button").addEventListener("click", confirmSubmit)