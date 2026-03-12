// Entry point

"use strict";

// Initialize stuff

import { AppStatus } from "./global-state/app-status";
import { AuthController } from "./global-state/auth";

AuthController.Initialize();
AppStatus.Initialize();

// Setup App

import { createApp } from "vue";

import App from "./App.vue";

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
