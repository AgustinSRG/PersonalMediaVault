<template>
  <div class="main-layout" :class="{
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
  }">
    <a v-if="!locked" href="javascript:;" @click="skipToMainContent" class="skip-to-main-content">{{ $t("Skip to main content") }}</a>
    <SideBar v-model:display="displaySidebar" :initialLayout="layout === 'initial'" @skip-to-content="skipToMainContent"></SideBar>
    <TopBar @logout="logout" @settings="showSettings" @menu="toggleSidebar" @search-open="openSearchModal" @help="showHelp"></TopBar>
    <PlayerContainer v-if="
      layout === 'media-split' || layout === 'media' || layout === 'album'
    "></PlayerContainer>
    <PageContent v-if="layout === 'initial' || layout === 'media-split'" :min="layout === 'media-split'"></PageContent>
    <AlbumContainer v-if="layout === 'album'"></AlbumContainer>

    <BottomBar v-if="layout === 'media-split' || layout === 'album'"></BottomBar>
    <div class="sidebar-float-overlay" :class="{ hidden: !displaySidebar }" @click="hideSidebar"></div>

    <SettingsDropdown v-if="displaySettings" v-model:display="displaySettings" @goto="onGoSettings"></SettingsDropdown>
    <LanguageDropdown v-if="displayLang" v-model:display="displayLang"></LanguageDropdown>
    <ThemeDropdown v-if="displayTheme" v-model:display="displayTheme"></ThemeDropdown>
    <ChangeUsernameModal v-if="displayUsernameModal" v-model:display="displayUsernameModal"></ChangeUsernameModal>
    <ChangePasswordModal v-if="displayPasswordModal" v-model:display="displayPasswordModal"></ChangePasswordModal>
    <AdvancedSettingsModal v-if="displayAdvancedSettings" v-model:display="displayAdvancedSettings"></AdvancedSettingsModal>
    <BatchOperationModal v-if="displayBatchOperation" v-model:display="displayBatchOperation"></BatchOperationModal>

    <AccountsAdminModal v-if="displayAccountAdmin" v-model:display="displayAccountAdmin"></AccountsAdminModal>

    <TaskListModal v-if="displayTaskList" v-model:display="displayTaskList"></TaskListModal>

    <SearchInputModal v-if="displaySearchModal" v-model:display="displaySearchModal"></SearchInputModal>

    <HelpHubDropdown v-if="displayHelpModal" v-model:display="displayHelpModal" @goto="onGoHelp"></HelpHubDropdown>

    <AboutModal v-if="displayAboutModal" v-model:display="displayAboutModal"></AboutModal>

    <KeyboardGuideModal v-if="displayKeyboardHelpModal" v-model:display="displayKeyboardHelpModal"></KeyboardGuideModal>

    <LogoutModal v-if="displayLogout" v-model:display="displayLogout"></LogoutModal>

    <LoadingOverlay :display="locked || loadingAuth || loadingTags || loadingAlbums" :fixed="true"></LoadingOverlay>

    <LoginModal v-if="locked && !loadingAuth" :display="locked && !loadingAuth"></LoginModal>

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

import { AuthController } from "../../control/auth";
import { TagsController } from "../../control/tags";
import { AlbumsController } from "../../control/albums";
import { AppEvents } from "../../control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";

const PlayerContainer = defineAsyncComponent({
    loader: () => import("@/components/layout/PlayerContainer.vue"),
    loadingComponent: PlayerContainerLoader,
    delay: 200,
});

const AlbumContainer = defineAsyncComponent({
    loader: () => import("@/components/layout/AlbumContainer.vue"),
    loadingComponent: AlbumContainerLoader,
    delay: 200,
});

const PageContent = defineAsyncComponent({
    loader: () => import("@/components/layout/PageContent.vue"),
    loadingComponent: PageContentLoader,
    delay: 200,
});

const LoginModal = defineAsyncComponent({
    loader: () => import("@/components/modals/LoginModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const LogoutModal = defineAsyncComponent({
    loader: () => import("@/components/modals/LogoutModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const SettingsDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/SettingsDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const LanguageDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/LanguageDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const ThemeDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/ThemeDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const ChangeUsernameModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ChangeUsernameModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const ChangePasswordModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ChangePasswordModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AdvancedSettingsModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AdvancedSettingsModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountsAdminModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountsAdminModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const TaskListModal = defineAsyncComponent({
    loader: () => import("@/components/modals/TaskListModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const SearchInputModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SearchInputModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const HelpHubDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/HelpHubDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AboutModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AboutModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const KeyboardGuideModal = defineAsyncComponent({
    loader: () => import("@/components/modals/KeyboardGuideModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const BatchOperationModal = defineAsyncComponent({
    loader: () => import("@/components/modals/BatchOperationModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
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
        SettingsDropdown,
        LanguageDropdown,
        ThemeDropdown,
        ChangeUsernameModal,
        ChangePasswordModal,
        AdvancedSettingsModal,
        AccountsAdminModal,
        TaskListModal,
        SearchInputModal,
        HelpHubDropdown,
        AboutModal,
        KeyboardGuideModal,
        BatchOperationModal,
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
            displayBatchOperation: false,

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
            case "batch":
                this.displayBatchOperation = true;
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

        onThemeChanged: function () {
            this.theme = AppPreferences.Theme;
        },

        onAppStatusUpdate: function () {
            this.layout = AppStatus.CurrentLayout;
            this.focus = AppStatus.CurrentFocus;
        },

        onAuthStatusChanged: function (locked: boolean) {
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
                this.displayBatchOperation = false;

                this.displayAlbumCreate = false;

                this.displayAccountAdmin = false;

                this.displaySearchModal = false;
            }
        },

        onAuthStatusLoading: function (l: boolean) {
            this.loadingAuth = l;
        },

        onTagsLoading: function (l: boolean) {
            this.loadingTags = l && !TagsController.InitiallyLoaded;
        },

        onAlbumsLoading: function (l: boolean) {
            this.loadingAlbums = l && !AlbumsController.InitiallyLoaded;
        },
    },
    mounted: function () {
        this.$options.onThemeChangedH = this.onThemeChanged.bind(this);
        AppEvents.AddEventListener("theme-changed", this.$options.onThemeChangedH);

        this.$options.onAppStatusUpdateH = this.onAppStatusUpdate.bind(this);
        AppEvents.AddEventListener("app-status-update", this.$options.onAppStatusUpdateH);

        this.$options.onAuthStatusChangedH = this.onAuthStatusChanged.bind(this);
        AppEvents.AddEventListener("auth-status-changed", this.$options.onAuthStatusChangedH);

        this.$options.onAuthStatusLoadingH = this.onAuthStatusLoading.bind(this);
        AppEvents.AddEventListener("auth-status-loading", this.$options.onAuthStatusLoadingH);

        this.$options.onTagsLoadingH = this.onTagsLoading.bind(this);
        AppEvents.AddEventListener("tags-loading", this.$options.onTagsLoadingH);

        this.$options.onAlbumsLoadingH = this.onAlbumsLoading.bind(this);
        AppEvents.AddEventListener("albums-loading", this.$options.onAlbumsLoadingH);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("theme-changed", this.$options.onThemeChangedH);
        AppEvents.RemoveEventListener("app-status-update", this.$options.onAppStatusUpdateH);
        AppEvents.RemoveEventListener("auth-status-changed", this.$options.onAuthStatusChangedH);
        AppEvents.RemoveEventListener("auth-status-loading", this.$options.onAuthStatusLoadingH);
        AppEvents.RemoveEventListener("tags-loading", this.$options.onTagsLoadingH);
        AppEvents.RemoveEventListener("albums-loading", this.$options.onAlbumsLoadingH);
    },
});
</script>
