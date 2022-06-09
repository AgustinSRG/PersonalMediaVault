<template>
  <div
    class="main-layout"
    :class="{
      'light-theme': theme !== 'dark',
      'dark-theme': theme === 'dark',
      'layout-album': layout === 'album',
      'layout-media-split': layout === 'media-split',
      'layout-media': layout === 'media',
    }"
  >
    <TopBar @logout="logout" @settings="showSettings"></TopBar>

    <SettingsModal
      v-model:display="displaySettings"
      @goto="onGoSettings"
    ></SettingsModal>
    <LanguageModal v-model:display="displayLang"></LanguageModal>
    <ThemeModal v-model:display="displayTheme"></ThemeModal>
    <ChangeUsernameModal v-model:display="displayUsernameModal"></ChangeUsernameModal>
    <ChangePasswordModal v-model:display="displayPasswordModal"></ChangePasswordModal>
    <AdvancedSettingsModal v-model:display="displayAdvancedSettings"></AdvancedSettingsModal>

    <LogoutModal v-model:display="displayLogout"></LogoutModal>

    <LoadingOverlay :display="locked || loadingAuth"></LoadingOverlay>
    <LoginModal :display="locked && !loadingAuth"></LoginModal>

    <SnackBar></SnackBar>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import TopBar from "./TopBar.vue";
import SnackBar from "./SnackBar.vue";
import LoadingOverlay from "./LoadingOverlay.vue";
import LoginModal from "../modals/LoginModal.vue";
import LogoutModal from "../modals/LogoutModal.vue";
import SettingsModal from "../modals/SettingsModal.vue";
import LanguageModal from "../modals/LanguageModal.vue";
import ThemeModal from "../modals/ThemeModal.vue";
import ChangeUsernameModal from "../modals/ChangeUsernameModal.vue";
import ChangePasswordModal from "../modals/ChangePasswordModal.vue";
import AdvancedSettingsModal from "../modals/AdvancedSettingsModal.vue";

import { AuthController } from "../../control/auth";
import { AppEvents } from "../../control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";

export default defineComponent({
  components: {
    TopBar,
    LoadingOverlay,
    LoginModal,
    LogoutModal,
    SettingsModal,
    LanguageModal,
    ThemeModal,
    ChangeUsernameModal,
    ChangePasswordModal,
    AdvancedSettingsModal,
    SnackBar,
  },
  name: "MainLayout",
  data: function () {
    return {
      theme: AppPreferences.Theme,
      locked: AuthController.Locked,
      loadingAuth: AuthController.Loading,
      layout: AppStatus.CurrentLayout,

      displayLogout: false,
      displaySettings: false,
      displayTheme: false,
      displayLang: false,
      displayUsernameModal: false,
      displayPasswordModal: false,
      displayAdvancedSettings: false,
    };
  },
  methods: {
    logout: function () {
      this.displayLogout = true;
    },

    showSettings: function () {
      this.displaySettings = true;
    },

    onGoSettings: function (o: string) {
      switch (o) {
        case "theme":
          this.displayTheme = true;
          break;
        case "lang":
          this.displayLang = true;
          break;
        case "username":
          this.displayUsernameModal = true;
          break;
        case "password":
          this.displayPasswordModal = true;
          break;
        case "advanced":
          this.displayAdvancedSettings = true;
          break;
      }
    },
  },
  mounted: function () {
    AppEvents.AddEventListener("theme-changed", () => {
      this.theme = AppPreferences.Theme;
    });

    AppEvents.AddEventListener("app-status-update", () => {
      this.layout = AppStatus.CurrentLayout;
    });

    AppEvents.AddEventListener("auth-status-changed", (locked: boolean) => {
      this.locked = locked;
    });

    AppEvents.AddEventListener("auth-status-loading", (l: boolean) => {
      this.loadingAuth = l;
    });
  },
});
</script>

<style>
*,
*::before,
*::after {
  box-sizing: border-box;
}

.main-layout {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: #181818;
  color: white;
}
</style>
