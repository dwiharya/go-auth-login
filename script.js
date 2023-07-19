document.getElementById("register-btn").addEventListener("click", function() {
    window.location.href = "/register";
});

document.getElementById("logout-btn").addEventListener("click", function() {
    showPopup("popup-logout");
});

document.getElementById("confirm-logout-btn").addEventListener("click", function() {
    window.location.href = "/logout";
});

document.getElementById("cancel-logout-btn").addEventListener("click", function() {
    hidePopup("popup-logout");
});

function showPopup(popupId) {
    document.getElementById(popupId).style.display = "block";
}

function hidePopup(popupId) {
    document.getElementById(popupId).style.display = "none";
}
