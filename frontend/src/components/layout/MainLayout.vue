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
      :initialLayout="layout === 'initial'"
      @skip-to-content="skipToMainContent"
    ></SideBar>
    <TopBar
      @logout="logout"
      @settings="showSettings"
      @menu="toggleSidebar"
      @search-open="openSearchModal"
      @help="showHelp"
    ></TopBar>
    <PlayerContainer
      v-if="
        layout === 'media-split' || layout === 'media' || layout === 'album'
      "
    ></PlayerContainer>
    <PageContent
      v-if="layout === 'initial' || layout === 'media-split'"
      :min="layout === 'media-split'"
    ></PageContent>
    <AlbumContainer v-if="layout === 'album'"></AlbumContainer>

    <BottomBar
      v-if="layout === 'media-split' || layout === 'album'"
    ></BottomBar>
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

    <AccountsAdminModal
      v-model:display="displayAccountAdmin"
    ></AccountsAdminModal>

    <TaskListModal v-model:display="displayTaskList"></TaskListModal>

    <SearchInputModal v-model:display="displaySearchModal"></SearchInputModal>

    <HelpHubModal
      v-model:display="displayHelpModal"
      @goto="onGoHelp"
    ></HelpHubModal>

    <AboutModal v-model:display="displayAboutModal"></AboutModal>

    <KeyboardGuideModal
      v-model:display="displayKeyboardHelpModal"
    ></KeyboardGuideModal>

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

      displayAccountAdmin: false,

      displayTaskList: false,

      displaySearchModal: false,

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
  },
});
</script>
