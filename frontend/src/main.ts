// Entry point

// Initialize stuff

import { AppPreferences } from './control/app-preferences';
import { AppStatus } from './control/app-status';
import { PlayerPreferences } from './control/player-preferences'
import { AuthController } from './control/auth';
import { TagsController } from './control/tags';
import { AlbumsController } from './control/albums';
import { MediaController } from './control/media';
import { SubtitlesController } from './control/subtitles';

AppPreferences.LoadPreferences();
PlayerPreferences.LoadPreferences();

AppStatus.Initialize();

AuthController.Initialize();
TagsController.Initailize();
AlbumsController.Initailize();
MediaController.Initialize();
SubtitlesController.Initialize();

// Setup App

import { createApp } from 'vue'

import i18n from './i18n'
import './registerServiceWorker'
import App from './App.vue'

document.addEventListener("drop", function(e) {
    e.preventDefault();
    e.stopPropagation();
});

document.addEventListener("dragover", function(e) {
    e.preventDefault();
    e.stopPropagation();
})

createApp(App).use(i18n).mount('#app')
