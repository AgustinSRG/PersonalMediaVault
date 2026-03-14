// Entry point

"use strict";

import { initializeGlobalState } from "./initialization";
import { createApp } from "vue";
import App from "./App.vue";

// Initialize global state

initializeGlobalState();

// Disable default behavior for dropping files

document.addEventListener("drop", function (e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
});

document.addEventListener("dragover", function (e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();
});

// Create app

createApp(App).mount("#app");

// Register service worker

if (import.meta.env.PROD && "serviceWorker" in navigator) {
    window.addEventListener("load", () => {
        navigator.serviceWorker.register("/sw.js", { scope: "/" });
    });
}
