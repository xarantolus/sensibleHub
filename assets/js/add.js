function addPage() {
    var linkInput = document.getElementById("url");
    linkInput.focus();

    var notification = document.getElementById("add-notif");

    var form = document.querySelector(".add-form");

    function setError(text) {
        notification.innerText = text
    }

    form.addEventListener('submit', function (evt) {
        var link = linkInput.value;

        evt.preventDefault();

        if (link.trim() == "") {
            return setError;
        }

        ajax("/add?format=json", {
            "url": link,
        }).post(function (status, obj) {
            if (status === 200) {
                InstantClick.go("/");
                return;
            }

            setError(obj.message || "Unknown error");
        });
    })
}