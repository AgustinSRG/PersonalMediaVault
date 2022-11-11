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
      'vault-locked': locked,
    }"
  >
    <a
      v-if="!locked"
      href="javascript:;"
      @click="skipToMainContent"
      class="skip-to-main-content"
      >{{ $t("Skip to main content") }}</a
    >
    <SideBar
      v-model:display="displaySidebar"
      :initialayout="layout === 'initial'"
      @skip-to-content="skipToMainContent"
    ></SideBar>
    <TopBar
      @logout="logout"
      @settings="showSettings"
      @menu="toggleSidebar"
      @search-open="openSearchModal"
      @help="showHelp"
    ></TopBar>
    <PlayerContainer v-if="layout === 'media-split' || layout === 'media' || layout === 'album'" @albums-open="showAlbumList"></PlayerContainer>
    <PageContent
      :min="layout === 'media-split'"
      @album-create="createAlbum"
    ></PageContent>
    <AlbumContainer
      @album-rename="showAlbumRename"
      @album-delete="showAlbumDelete"
    ></AlbumContainer>

    <BottomBar></BottomBar>
    <div
      class="sidebar-float-overlay"
      :class="{ hidden: !displaySidebar }"
      @click="hideSidebar"
    ></div>

    <SettingsModal
      v-model:display="displaySettings"
      @goto="onGoSettings"
    ></SettingsModal>
    <LanguageModal v-model:display="displayLang"></LanguageModal>
    <ThemeModal v-model:display="displayTheme"></ThemeModal>
    <ChangeUsernameModal
      v-model:display="displayUsernameModal"
    ></ChangeUsernameModal>
    <ChangePasswordModal
      v-model:display="displayPasswordModal"
    ></ChangePasswordModal>
    <AdvancedSettingsModal
      v-model:display="displayAdvancedSettings"
    ></AdvancedSettingsModal>

    <AlbumListModal
      v-model:display="displayAlbumList"
      @album-create="createAlbum"
    ></AlbumListModal>
    <AlbumCreateModal v-model:display="displayAlbumCreate" @update:display="doneCreateAlbum"></AlbumCreateModal>
    <AlbumRenameModal v-model:display="displayAlbumRename"></AlbumRenameModal>
    <AlbumDeleteModal v-model:display="displayAlbumDelete"></AlbumDeleteModal>
    <AlbumMovePosModal
      v-model:display="displayAlbumMovePos"
    ></AlbumMovePosModal>
    <MediaDeleteModal v-model:display="displayMediaDelete"></MediaDeleteModal>
    <ResolutionConfirmationModal
      v-model:display="displayResolutionConfirmation"
    ></ResolutionConfirmationModal>
    <SubtitlesDeleteModal
      v-model:display="displaySubtitlesDelete"
    ></SubtitlesDeleteModal>
    <ReEncodeConfirmationModal
      v-model:display="displayReEncode"
    ></ReEncodeConfirmationModal>

    <AccountsAdminModal
      v-model:display="displayAccountAdmin"
    ></AccountsAdminModal>

    <TaskListModal v-model:display="displayTaskList"></TaskListModal>

    <SearchInputModal v-model:display="displaySearchModal"></SearchInputModal>

    <HelpHubModal v-model:display="displayHelpModal" @goto="onGoHelp"></HelpHubModal>

    <AboutModal v-model:display="displayAboutModal"></AboutModal>

    <KeyboardGuideModal v-model:display="displayKeyboardHelpModal"></KeyboardGuideModal>

    <LogoutModal v-model:display="displayLogout"></LogoutModal>

    <LoadingOverlay
      :display="locked || loadingAuth || loadingTags || loadingAlbums"
      :fixed="true"
    ></LoadingOverlay>
    <LoginModal :display="locked && !loadingAuth"></LoginModal>

    <SnackBar></SnackBar>
  </div>
</template>

<script lang="ts">
import { defineComponent, defineAsyncComponent } from "vue";

import TopBar from "./TopBar.vue";
import BottomBar from "./BottomBar.vue";
import SideBar from "./SideBar.vue";
import SnackBar from "./SnackBar.vue";
import LoadingOverlay from "./LoadingOverlay.vue";
import PlayerContainerLoader from "./PlayerContainerLoader.vue";
import AlbumContainerLoader from "./AlbumContainerLoader.vue";
import PageContentLoader from "./PageContentLoader.vue";
import LoginModal from "../modals/LoginModal.vue";
import LogoutModal from "../modals/LogoutModal.vue";
import SettingsModal from "../modals/SettingsModal.vue";
import LanguageModal from "../modals/LanguageModal.vue";
import ThemeModal from "../modals/ThemeModal.vue";
import ChangeUsernameModal from "../modals/ChangeUsernameModal.vue";
import ChangePasswordModal from "../modals/ChangePasswordModal.vue";
import AdvancedSettingsModal from "../modals/AdvancedSettingsModal.vue";
import AlbumCreateModal from "../modals/AlbumCreateModal.vue";
import AlbumListModal from "../modals/AlbumListModal.vue";
import AlbumRenameModal from "../modals/AlbumRenameModal.vue";
import AlbumDeleteModal from "../modals/AlbumDeleteModal.vue";
import AlbumMovePosModal from "../modals/AlbumMovePosModal.vue";
import MediaDeleteModal from "../modals/MediaDeleteModal.vue";
import ResolutionConfirmationModal from "../modals/ResolutionConfirmationModal.vue";
import ReEncodeConfirmationModal from "../modals/ReEncodeConfirmationModal.vue";
import SubtitlesDeleteModal from "../modals/SubtitlesDeleteModal.vue";
import AccountsAdminModal from "../modals/AccountsAdminModal.vue";
import TaskListModal from "../modals/TaskListModal.vue";
import SearchInputModal from "../modals/SearchInputModal.vue";
import HelpHubModal from "../modals/HelpHubModal.vue";
import AboutModal from "../modals/AboutModal.vue";
import KeyboardGuideModal from "../modals/KeyboardGuideModal.vue";

import { AuthController } from "../../control/auth";
import { TagsController } from "../../control/tags";
import { AlbumsController } from "../../control/albums";
import { AppEvents } from "../../control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";

const PlayerContainer = defineAsyncComponent({
  // the loader function
  loader: () => import("@/components/layout/PlayerContainer.vue"),

  // A component to use while the async component is loading
  loadingComponent: PlayerContainerLoader,
  // Delay before showing the loading component. Default: 200ms.
  delay: 200,
});

const AlbumContainer = defineAsyncComponent({
  // the loader function
  loader: () => import("@/components/layout/AlbumContainer.vue"),

  // A component to use while the async component is loading
  loadingComponent: AlbumContainerLoader,
  // Delay before showing the loading component. Default: 200ms.
  delay: 200,
});

const PageContent = defineAsyncComponent({
  // the loader function
  loader: () => import("@/components/layout/PageContent.vue"),

  // A component to use while the async component is loading
  loadingComponent: PageContentLoader,
  // Delay before showing the loading component. Default: 200ms.
  delay: 200,
});

export default defineComponent({
  components: {
    TopBar,
    BottomBar,
    SideBar,
    PageContent,
    AlbumContainer,
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
    AlbumRenameModal,
    AlbumDeleteModal,
    MediaDeleteModal,
    AlbumMovePosModal,
    ResolutionConfirmationModal,
    ReEncodeConfirmationModal,
    SubtitlesDeleteModal,
    AccountsAdminModal,
    TaskListModal,
    SearchInputModal,
    HelpHubModal,
    AboutModal,
    KeyboardGuideModal,
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
      displayAlbumRename: false,
      displayAlbumDelete: false,
      displayAlbumMovePos: false,

      createAlbumFromList: false,

      displayMediaDelete: false,

      displayResolutionConfirmation: false,

      displaySubtitlesDelete: false,

      displayAccountAdmin: false,

      displayTaskList: false,

      displaySearchModal: false,

      displayReEncode: false,

      displaySidebar: true,

      displayHelpModal: false,

      displayAboutModal: false,
      displayKeyboardHelpModal: false,
    };
  },
  methods: {
    logout: function () {
      this.displayLogout = true;
    },

    showSettings: function () {
      this.displaySettings = !this.displaySettings;
      this.displayHelpModal = false;
      this.displayTheme = false;
      this.displayLang = false;
    },

    showHelp: function () {
      this.displayHelpModal = !this.displayHelpModal;
      this.displaySettings = false;
      this.displayTheme = false;
      this.displayLang = false;
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
        case "tasks":
          this.displayTaskList = true;
          break;
        case "admin":
          this.displayAccountAdmin = true;
          break;
      }
    },

    onGoHelp: function (o: string) {
      switch (o) {
        case "about":
          this.displayAboutModal = true;
          break;
        case "keyboard":
          this.displayKeyboardHelpModal = true;
          break;
      }
    },

    toggleSidebar: function () {
      this.displaySidebar = !this.displaySidebar;
    },

    createAlbum: function (fromList: boolean) {
      this.displayAlbumCreate = true;
      this.createAlbumFromList = !!fromList;
    },

    doneCreateAlbum: function () {
      if (!this.displayAlbumCreate &&  this.createAlbumFromList) {
        this.createAlbumFromList = false;
        this.displayAlbumList = true;
      }
    },

    showAlbumList: function () {
      this.displayAlbumList = true;
    },

    showAlbumRename: function () {
      this.displayAlbumRename = true;
    },

    showAlbumDelete: function () {
      this.displayAlbumDelete = true;
    },

    showMediaDelete: function () {
      this.displayMediaDelete = true;
    },

    hideSidebar: function () {
      this.displaySidebar = false;
    },

    openSearchModal: function () {
      this.displaySearchModal = true;
    },

    skipToMainContent: function (event) {
      if (event) {
        event.preventDefault();
      }
      let skipTo = null;
      switch (AppStatus.CurrentLayout) {
        case "media":
        case "media-split":
        case "album":
          skipTo = this.$el.querySelector(".player-container");
          break;
        default:
          skipTo = this.$el.querySelector(".page-content");
      }
      if (skipTo) {
        skipTo.focus();
      }
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

      if (this.locked) {
        // Close all modals
        this.displayLogout = false;
        this.displaySettings = false;
        this.displayTheme = false;
        this.displayLang = false;
        this.displayUsernameModal = false;
        this.displayPasswordModal = false;
        this.displayAdvancedSettings = false;

        this.displayAlbumCreate = false;
        this.displayAlbumList = false;
        this.displayAlbumRename = false;
        this.displayAlbumDelete = false;

        this.displayMediaDelete = false;

        this.displayResolutionConfirmation = false;

        this.displaySubtitlesDelete = false;

        this.displayAccountAdmin = false;

        this.displaySearchModal = false;
      }
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

    AppEvents.AddEventListener("media-delete-request", () => {
      this.showMediaDelete();
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
}

.main-layout.light-theme {
  background: #f9f9f9;
}

.light-theme {
  color: black;
  --theme-btn-hover-color: black;
  --theme-btn-color: rgba(0, 0, 0, 0.75);

  --theme-border-color: rgba(0, 0, 0, 0.1);

  --theme-option-hover-color: rgba(0, 0, 0, 0.1);
  --theme-option-selected-color: rgba(0, 0, 0, 0.2);

  --theme-fg-color: black;
  --theme-fg-color-opposite: white;
  --theme-bg-color: #f9f9f9;

  --modal-overlay-bg-color: rgba(0, 0, 0, 0.4);
  --modal-bg-color: white;
  --bar-bg-color: white;

  --modal-shadow: 0 16px 24px 2px rgb(255 255 255 / 14%),
    0 6px 30px 5px rgb(255 255 255 / 12%),
    0 8px 10px -5px rgb(255 255 255 / 40%);

  --hover-color: rgba(0, 0, 0, 0.1);
  --selected-color: rgba(0, 0, 0, 0.2);

  --input-bg-color: white;

  --switch-color: black;
  --switch-shadow-color: rgba(255, 255, 255, 0.33);
  --switch-disabled-color: lightgray;
}

.main-layout.dark-theme {
  background: #181818;
}

.dark-theme {
  color: white;

  --theme-btn-hover-color: white;
  --theme-btn-color: rgba(255, 255, 255, 0.75);

  --theme-border-color: rgba(255, 255, 255, 0.1);

  --theme-option-hover-color: rgba(255, 255, 255, 0.1);
  --theme-option-selected-color: rgba(255, 255, 255, 0.2);

  --theme-fg-color: white;
  --theme-fg-color-opposite: black;
  --theme-bg-color: #181818;

  --modal-overlay-bg-color: rgba(0, 0, 0, 0.4);
  --modal-bg-color: #212121;
  --bar-bg-color: #212121;

  --modal-shadow: 0 16px 24px 2px rgb(0 0 0 / 14%), 0 6px 30px 5px rgb(0 0 0 / 12%),
    0 8px 10px -5px rgb(0 0 0 / 40%);

  --hover-color: rgba(255, 255, 255, 0.1);
  --selected-color: rgba(255, 255, 255, 0.2);

  --input-bg-color: hsl(0, 0%, 7%);

  --switch-color: white;
  --switch-shadow-color: rgba(0, 0, 0, 0.33);
  --switch-disabled-color: gray;
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
  z-index: 11;
}

.sidebar-float-overlay.hidden {
  opacity: 0;
  pointer-events: none;
}

@media (min-width: 1000px) {
  .layout-initial .sidebar-float-overlay {
    display: none;
  }
}

.main-content-skip:focus {
  outline: none;
}

.skip-to-main-content {
  padding: 1rem;
  border: solid 1px var(--theme-border-color);
  background: var(--bar-bg-color);
  z-index: 13;

  position: fixed;
  top: -100%;
  left: 1rem;
}

.skip-to-main-content:focus {
  top: 1rem;
}

/* Layout - PlayerContainer */

.player-container {
  position: absolute;
  top: 57px;
  height: calc(100% - 57px);
  left: 0;
  width: 100%;
  overflow: auto;
}

.player-container:focus {
  outline: none;
}

.vault-locked .player-container {
  visibility: hidden;
}

.layout-media-split .player-container {
  width: calc(100% - 500px);
}

.layout-album .player-container {
  width: calc(100% - 500px);
}

@media (max-width: 1000px) {
  .layout-media-split .player-container,
  .layout-album .player-container {
    width: calc(100%);
    height: calc(100% - 57px - 40px);
  }

  .layout-media-split.focus-right .player-container {
    display: none;
  }

  .layout-album.focus-right .player-container {
    display: none;
  }
}

.layout-initial .player-container {
  display: none;
}

/* Layout - PageContent */

.page-content {
  position: absolute;
  top: 57px;
  height: calc(100% - 57px);
  left: 240px;
  width: calc(100% - 240px);
  display: flex;
  flex-direction: column;
}

.page-content:focus {
  outline: none;
}

.vault-locked .page-content {
  visibility: hidden;
}

.sidebar-hidden .page-content {
  left: 0;
  width: 100%;
}

@media (max-width: 1000px) {
  .page-content {
    left: 0;
    width: 100%;
  }
}

.layout-media-split .page-content,
.sidebar-hidden .layout-media-split .page-content {
  left: auto;
  right: 0;
  width: 500px;
  border-left: solid 1px var(--theme-border-color);
}

@media (max-width: 1000px) {
  .layout-media-split .page-content {
    width: calc(100%);
    height: calc(100% - 57px - 40px);
  }

  .layout-media-split.focus-left .page-content {
    display: none;
  }
}

.layout-album .page-content {
  display: none;
}

.layout-media .page-content {
  display: none;
}


/* Layout - AlbumContainer */

.album-container {
  position: absolute;
  top: 57px;
  height: calc(100% - 57px);
  right: 0;
  width: 500px;
  border-left: solid 1px var(--theme-border-color);
  display: none;
}

.vault-locked .album-container {
  visibility: hidden;
}

.layout-album .album-container {
  display: block;
}

@media (max-width: 1000px) {
  .album-container {
    width: calc(100%);
    height: calc(100% - 57px - 40px);
  }

  .layout-album.focus-left .album-container {
    display: none;
  }
}
</style>
