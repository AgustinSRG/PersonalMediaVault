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

    <SettingsModal v-model:display="displaySettings"></SettingsModal>

    <LogoutModal v-model:display="displayLogout"></LogoutModal>

    <LoadingOverlay :display="locked || loadingAuth"></LoadingOverlay>
    <LoginModal :display="locked && !loadingAuth"></LoginModal>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import TopBar from "./TopBar.vue";
import LoadingOverlay from "./LoadingOverlay.vue";
import LoginModal from "../modals/LoginModal.vue";
import LogoutModal from "../modals/LogoutModal.vue";
import SettingsModal from "../modals/SettingsModal.vue";

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
    };
  },
  methods: {
    logout: function () {
      this.displayLogout = true;
    },

    showSettings: function () {
      this.displaySettings = true;
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
