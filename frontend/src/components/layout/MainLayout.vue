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
        <a v-if="!locked" href="javascript:;" @click="skipToMainContent" class="skip-to-main-content">{{ $t("Skip to main content") }}</a>
        <SideBar v-model:display="displaySidebar" :initialLayout="layout === 'initial'" @skip-to-content="skipToMainContent"></SideBar>
        <TopBar
            @logout="logout"
            @settings="showSettings"
            @menu="toggleSidebar"
            @menu-focus="openSideBar"
            @search-open="openSearchModal"
            @help="showHelp"
        ></TopBar>
        <PlayerContainer
            v-if="layout === 'media-split' || layout === 'media' || layout === 'album'"
            :display-upload="displayUpload"
        ></PlayerContainer>
        <PageContent v-if="layout === 'initial' || layout === 'media-split'" :min="layout === 'media-split'"></PageContent>
        <AlbumContainer v-if="layout === 'album'" v-model:display-upload="displayUpload"></AlbumContainer>

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

        <LoadingOverlay v-if="locked || loadingAuth" :fixed="true" :issues="loadingAuthError"></LoadingOverlay>

        <LoginModal v-if="locked && !loadingAuth" :display="locked && !loadingAuth"></LoginModal>

        <SnackBar></SnackBar>

        <div class="new-version-notice" v-if="newVersionAvailable && !newVersionDismissed">
            <div class="new-version-notice-msg">
                {{ $t("You are using an older version of PersonalMediaVault than the server's") }}.
                {{ $t("Refresh the page in order to use the latest version") }}.
            </div>
            <button type="button" class="modal-close-btn" :title="$t('Refresh')" @click="hardReload">
                <i class="fas fa-sync-alt"></i>
            </button>
            <button type="button" class="modal-close-btn" :title="$t('Close')" @click="dismissNewVersion">
                <i class="fas fa-times"></i>
            </button>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent, defineAsyncComponent, nextTick } from "vue";

import TopBar from "./TopBar.vue";
import BottomBar from "./BottomBar.vue";
import SideBar from "./SideBar.vue";
import SnackBar from "./SnackBar.vue";

import LoadingOverlay from "./LoadingOverlay.vue";
import PlayerContainerLoader from "./PlayerContainerLoader.vue";
import AlbumContainerLoader from "./AlbumContainerLoader.vue";
import PageContentLoader from "./PageContentLoader.vue";

import { AuthController, EVENT_NAME_APP_NEW_VERSION } from "../../control/auth";
import { AppEvents } from "../../control/app-events";
import { ColorThemeName, EVENT_NAME_THEME_CHANGED, getTheme } from "@/control/app-preferences";
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
            theme: getTheme(),

            locked: AuthController.Locked,
            loadingAuth: AuthController.Loading,
            loadingAuthError: false,

            layout: AppStatus.CurrentLayout,
            focus: AppStatus.CurrentFocus,

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

            displaySidebar: window.innerWidth >= 1000,

            displayHelpModal: false,

            displayAboutModal: false,
            displayKeyboardHelpModal: false,

            newVersionAvailable: false,
            newVersionDismissed: false,

            displayUpload: false,
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
            if (this.displaySidebar) {
                nextTick(() => {
                    const sideBar = this.$el.querySelector(".side-bar");
                    if (sideBar) {
                        sideBar.focus();
                    }
                });
            }
        },

        openSideBar: function () {
            this.displaySidebar = true;
            nextTick(() => {
                const sideBar = this.$el.querySelector(".side-bar");
                if (sideBar) {
                    sideBar.focus();
                }
            });
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
                    if (skipTo) {
                        const autoFocused = skipTo.querySelector(".auto-focus");
                        if (autoFocused) {
                            skipTo = autoFocused;
                        }
                    }
            }
            if (skipTo) {
                skipTo.focus();
            }
        },

        onThemeChanged: function (theme: ColorThemeName) {
            this.theme = theme;
        },

        onAppStatusUpdate: function () {
            this.layout = AppStatus.CurrentLayout;
            this.focus = AppStatus.CurrentFocus;
        },

        onAuthStatusChanged: function (locked: boolean) {
            this.locked = locked;
            this.loadingAuthError = false;

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

        onAuthLoadingError: function () {
            this.loadingAuthError = true;
        },

        onNewAppVersion: function () {
            this.newVersionAvailable = true;
        },

        dismissNewVersion: function () {
            this.newVersionDismissed = true;
        },

        hardReload: function () {
            try {
                navigator.serviceWorker.getRegistrations().then((registrations) => {
                    registrations.forEach((registration) => {
                        registration.unregister();
                    });

                    caches.keys().then((allCaches) => {
                        allCaches.forEach((cache) => {
                            caches.delete(cache);
                        });

                        const loc: any = window.location;
                        loc.reload(true);
                    });
                });
            } catch (ex) {
                console.error(ex);
            }
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.onThemeChangedH = this.onThemeChanged.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_THEME_CHANGED, this._handles.onThemeChangedH);

        this._handles.onAppStatusUpdateH = this.onAppStatusUpdate.bind(this);
        AppStatus.AddEventListener(this._handles.onAppStatusUpdateH);

        this._handles.onAuthStatusChangedH = this.onAuthStatusChanged.bind(this);
        AuthController.AddChangeEventListener(this._handles.onAuthStatusChangedH);

        this._handles.onAuthStatusLoadingH = this.onAuthStatusLoading.bind(this);
        AuthController.AddLoadingEventListener(this._handles.onAuthStatusLoadingH);

        this._handles.onAuthLoadingErrorH = this.onAuthLoadingError.bind(this);
        AuthController.AddErrorEventListener(this._handles.onAuthLoadingErrorH);

        this._handles.onNewAppVersionH = this.onNewAppVersion.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_APP_NEW_VERSION, this._handles.onNewAppVersionH);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener(EVENT_NAME_THEME_CHANGED, this._handles.onThemeChangedH);
        AppStatus.RemoveEventListener(this._handles.onAppStatusUpdateH);
        AuthController.RemoveChangeEventListener(this._handles.onAuthStatusChangedH);
        AuthController.RemoveLoadingEventListener(this._handles.onAuthStatusLoadingH);
        AuthController.RemoveErrorEventListener(this._handles.onAuthLoadingErrorH);
        AppEvents.RemoveEventListener(EVENT_NAME_APP_NEW_VERSION, this._handles.onNewAppVersionH);
    },
});
</script>
