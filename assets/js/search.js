function initSearch() {
    var sb = document.getElementById("search-bar");
    var sc = document.getElementById("search-container");
    var sf = document.getElementById("search-form");
    
    function setLoading(l) {
        if (l) {
            sc.classList.add("is-loading")
        } else {
            sc.classList.remove("is-loading")
        }
    }

    var isSearch = false;

    InstantClick.on('change', function () {
        isSearch = false;

        sb.focusSearch();
    })

    sf.onsubmit = function (evt) {
        evt.preventDefault();

        InstantClick.go("/search?q=" + encodeURIComponent(sb.value))
    }
}

function focusSearch() {
    var sb = document.getElementById("search-bar");
    if (sb) {
        if (!sb.value.endsWith(" ") && sb.value.length !== 0) {
            sb.value +=" "
        }
        
        sb.setSelectionRange(sb.value.length, sb.value.length);
        sb.focus();
    }
}