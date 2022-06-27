<template>
  <div
    class="main-layout"
    :class="{
      'light-theme': theme !== 'dark',
      'dark-theme': theme === 'dark',
      'layout-initial': layout === 'initial',
      'layout-album': layout === 'album',
      'layout-media-split': layout === 'media-split',
      'layout-media': layout === 'media',
      'sidebar-hidden': !displaySidebar,
      'focus-left': focus === 'left',
      'focus-right': focus === 'right',
    }"
  >
    <PageContent :min="layout === 'media-split'" @album-create="createAlbum"></PageContent>
    <PlayerContainer @albums-open="showAlbumList"></PlayerContainer>

    <TopBar @logout="logout" @settings="showSettings" @menu="toggleSidebar"></TopBar>
    <BottomBar></BottomBar>
    <div class="sidebar-float-overlay" :class="{hidden: !displaySidebar}" @click="hideSidebar"></div>
    <SideBar v-model:display="displaySidebar"></SideBar>

    <SettingsModal
      v-model:display="displaySettings"
      @goto="onGoSettings"
    ></SettingsModal>
    <LanguageModal v-model:display="displayLang"></LanguageModal>
    <ThemeModal v-model:display="displayTheme"></ThemeModal>
    <ChangeUsernameModal v-model:display="displayUsernameModal"></ChangeUsernameModal>
    <ChangePasswordModal v-model:display="displayPasswordModal"></ChangePasswordModal>
    <AdvancedSettingsModal v-model:display="displayAdvancedSettings"></AdvancedSettingsModal>

    <AlbumListModal v-model:display="displayAlbumList" @album-create="createAlbum"></AlbumListModal>
    <AlbumCreateModal v-model:display="displayAlbumCreate"></AlbumCreateModal>

    <LogoutModal v-model:display="displayLogout"></LogoutModal>

    <LoadingOverlay :display="locked || loadingAuth || loadingTags || loadingAlbums"></LoadingOverlay>
    <LoginModal :display="locked && !loadingAuth"></LoginModal>

    <SnackBar></SnackBar>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import TopBar from "./TopBar.vue";
import BottomBar from "./BottomBar.vue";
import SideBar from "./SideBar.vue";
import PageContent from "./PageContent.vue";
import PlayerContainer from "./PlayerContainer.vue";
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
import AlbumCreateModal from "../modals/AlbumCreateModal.vue"
import AlbumListModal from "../modals/AlbumListModal.vue"

import { AuthController } from "../../control/auth";
import { TagsController } from "../../control/tags";
import { AlbumsController } from "../../control/albums";
import { AppEvents } from "../../control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";

export default defineComponent({
  components: {
    TopBar,
    BottomBar,
    SideBar,
    PageContent,
    PlayerContainer,
    LoadingOverlay,
    LoginModal,
    LogoutModal,
    SettingsModal,
    LanguageModal,
    ThemeModal,
    ChangeUsernameModal,
    ChangePasswordModal,
    AdvancedSettingsModal,
    AlbumCreateModal,
    AlbumListModal,
    SnackBar,
  },
  name: "MainLayout",
  data: function () {
    return {
      theme: AppPreferences.Theme,
      locked: AuthController.Locked,
      loadingAuth: AuthController.Loading,
      layout: AppStatus.CurrentLayout,
      focus: AppStatus.CurrentFocus,

      loadingTags: TagsController.Loading,
      loadingAlbums: AlbumsController.Loading,

      displayLogout: false,
      displaySettings: false,
      displayTheme: false,
      displayLang: false,
      displayUsernameModal: false,
      displayPasswordModal: false,
      displayAdvancedSettings: false,

      displayAlbumCreate: false,
      displayAlbumList: false,

      displaySidebar: true,
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

    toggleSidebar: function () {
      this.displaySidebar = !this.displaySidebar;
    },

    createAlbum: function () {
      this.displayAlbumCreate = true;
    },

    showAlbumList: function () {
      this.displayAlbumList = true;
    },

    hideSidebar: function () {
      this.displaySidebar = false;
    },
  },
  mounted: function () {
    AppEvents.AddEventListener("theme-changed", () => {
      this.theme = AppPreferences.Theme;
    });

    AppEvents.AddEventListener("app-status-update", () => {
      this.layout = AppStatus.CurrentLayout;
      this.focus = AppStatus.CurrentFocus;
    });

    AppEvents.AddEventListener("auth-status-changed", (locked: boolean) => {
      this.locked = locked;
    });

    AppEvents.AddEventListener("auth-status-loading", (l: boolean) => {
      this.loadingAuth = l;
    });

    AppEvents.AddEventListener("tags-loading", (l: boolean) => {
      this.loadingTags = l;
    });

    AppEvents.AddEventListener("albums-loading", (l: boolean) => {
      this.loadingAlbums = l;
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

.sidebar-float-overlay {
  position: fixed;

  top: 0;
  left: 0;
  width: 100%;
  height: 100%;

  opacity: 1;
  transition: opacity 0.2s;
  background: rgba(0, 0, 0, 0.4);
}

.sidebar-float-overlay.hidden {
  opacity: 0;
  pointer-events: none;
}

.layout-initial .sidebar-float-overlay {
  display: none;
}
</style>
