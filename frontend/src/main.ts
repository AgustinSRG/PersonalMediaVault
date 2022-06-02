import { createApp } from 'vue'
import App from './App.vue'
import './registerServiceWorker'
import i18n from './i18n'
import { PlayerPreferences } from './control/player-preferences'

PlayerPreferences.LoadPreferences();

document.addEventListener("drop", function(e) {
    e.preventDefault();
    e.stopPropagation();
});

document.addEventListener("dragover", function(e) {
    e.preventDefault();
    e.stopPropagation();
})

createApp(App).use(i18n).mount('#app')
