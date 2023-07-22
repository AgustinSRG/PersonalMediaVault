/* View in fullscreen */
export function openFullscreen() {
    const elem = document.documentElement;
    if (elem.requestFullscreen) {
        elem.requestFullscreen();
    }
}

/* Close fullscreen */
export function closeFullscreen() {
    if (document.exitFullscreen) {
        document.exitFullscreen();
    }
}
