// https://stackoverflow.com/a/44897454
var lastClick = 0;

function openCloseMenu() {
    // since this event might be called twice (once because of click, once because touchstart),
    // we need to make sure it only fires once, else the menu would not be visible at all 
    if (new Date() - lastClick < 750) {
        return;
    }
    lastClick = new Date();

    var burger = document.querySelector('.navbar-burger');
    var menu = document.querySelector('.navbar-menu');

    burger.classList.toggle('is-active');
    menu.classList.toggle('is-active');
}

function registerMenu() {
    var burger = document.querySelector('.navbar-burger');

    // click doesn't work in some mobile versions of chrome
    // this is why we need to use touchstart too
    burger.addEventListener('click', openCloseMenu);
    burger.addEventListener('touchstart', openCloseMenu);
}

InstantClick.on('change', registerMenu)