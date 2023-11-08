// Entry point

"use strict";

// Initialize stuff

import { AppStatus } from "./control/app-status";
import { AuthController } from "./control/auth";
import { TagsController } from "./control/tags";
import { AlbumsController } from "./control/albums";
import { MediaController } from "./control/media";
import { KeyboardManager } from "./control/keyboard";
import { BusyStateController } from "./control/busy-state";

KeyboardManager.Initialize();

AppStatus.Initialize();

AuthController.Initialize();
TagsController.Initialize();
MediaController.Initialize();
AlbumsController.Initialize();
BusyStateController.Initialize();

// Setup App

import { createApp } from "vue";

import { i18n } from "./i18n";

import App from "./App.vue";
import ModalDialogContainer from "@/components/utils/ModalDialogContainer.vue";

document.addEventListener("drop", function (e) {
    e.preventDefault();
    e.stopPropagation();
});

document.addEventListener("dragover", function (e) {
    e.preventDefault();
    e.stopPropagation();
});

// Create app

createApp(App).use(i18n).component("ModalDialogContainer", ModalDialogContainer).mount("#app");

// Register service worker

if (import.meta.env.PROD && "serviceWorker" in navigator) {
    window.addEventListener("load", () => {
        navigator.serviceWorker.register("/sw.js", { scope: "/" });
    });
}
