// Entry point

// Initialize stuff

import { AppPreferences } from './control/app-preferences';
import { AppStatus } from './control/app-status';
import { PlayerPreferences } from './control/player-preferences'
import { AuthController } from './control/auth';
import { TagsController } from './control/tags';
import { AlbumsController } from './control/albums';
import { MediaController } from './control/media';
import { KeyboardManager } from './control/keyboard';
import { BusyStateController } from './control/busy-state';

AppPreferences.LoadPreferences();
PlayerPreferences.LoadPreferences();

KeyboardManager.Initialize();

AppStatus.Initialize();

AuthController.Initialize();
TagsController.Initialize();
MediaController.Initialize();
AlbumsController.Initialize();
BusyStateController.Initialize();

// Setup App

import { createApp } from 'vue'

import { i18n } from './i18n'
import './registerServiceWorker'
import App from './App.vue'

document.addEventListener("drop", function (e) {
    e.preventDefault();
    e.stopPropagation();
});

document.addEventListener("dragover", function (e) {
    e.preventDefault();
    e.stopPropagation();
})

createApp(App).use(i18n).mount('#app')
