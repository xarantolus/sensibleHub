function addPage() {
    var linkInput = document.getElementById("searchTerm");
    // We're handling empty inputs in JavaScript. Since we're using progressive enhancement,
    // we should remove it here (so people with javascript disabled can still use the site)
    linkInput.required = false; 
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
            return setError("Link must not be empty");
        }

        ajax("/add?format=json", {
            "searchTerm": link,
        }).post(function (status, obj) {
            if (status === 200) {
                InstantClick.go("/");
                return;
            }

            setError(obj.message || "Unknown error");
        });
    })

    var abortForm = document.querySelector(".abort-form");
    if (abortForm) {
        abortForm.addEventListener('submit', function (evt) {
            evt.preventDefault();

            if (!confirm("Are you sure you want stop this download?")) {
                return false;
            }

            ajax("/abort?format=json", {}).post(function (status, obj) {
                if (status === 200) {
                    InstantClick.go("/add");
                    return;
                }


                setError(obj.message || "Unknown error");
            });
        })
    }
}
