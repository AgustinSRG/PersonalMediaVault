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
        <a v-if="!locked" href="javascript:;" class="skip-to-main-content" @click="skipToMainContent">{{ $t("Skip to main content") }}</a>
        <SideBar v-model:display="displaySidebar" :initial-layout="layout === 'initial'" @skip-to-content="skipToMainContent"></SideBar>
        <TopBar
            @logout="logout"
            @vault-settings="showVaultSettings"
            @account-settings="showAccountSettings"
            @menu="toggleSidebar"
            @menu-focus="openSideBar"
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

        <VaultSettingsDropdown
            v-if="displayVaultSettings"
            v-model:display="displayVaultSettings"
            @goto="onGoSettings"
        ></VaultSettingsDropdown>
        <AccountSettingsDropdown
            v-if="displayAccountSettings"
            v-model:display="displayAccountSettings"
            @goto="onGoSettings"
        ></AccountSettingsDropdown>

        <LanguageDropdown v-if="displayLang" v-model:display="displayLang"></LanguageDropdown>
        <ThemeDropdown v-if="displayTheme" v-model:display="displayTheme"></ThemeDropdown>
        <ChangeUsernameModal v-if="displayUsernameModal" v-model:display="displayUsernameModal"></ChangeUsernameModal>
        <ChangePasswordModal v-if="displayPasswordModal" v-model:display="displayPasswordModal"></ChangePasswordModal>
        <AdvancedSettingsModal v-if="displayAdvancedSettings" v-model:display="displayAdvancedSettings"></AdvancedSettingsModal>
        <BatchOperationModal v-if="displayBatchOperation" v-model:display="displayBatchOperation"></BatchOperationModal>

        <InviteModal v-if="displayInvite" v-model:display="displayInvite"></InviteModal>

        <AccountSecuritySettingsModal
            v-if="displaySecuritySettings"
            v-model:display="displaySecuritySettings"
        ></AccountSecuritySettingsModal>

        <AccountsAdminModal v-if="displayAccountAdmin" v-model:display="displayAccountAdmin"></AccountsAdminModal>

        <TaskListModal v-if="displayTaskList" v-model:display="displayTaskList"></TaskListModal>

        <ClearBrowserDataModal v-if="displayClearBrowserData" v-model:display="displayClearBrowserData"></ClearBrowserDataModal>

        <HelpHubDropdown v-if="displayHelpModal" v-model:display="displayHelpModal" @goto="onGoHelp"></HelpHubDropdown>

        <AboutModal v-if="displayAboutModal" v-model:display="displayAboutModal"></AboutModal>

        <KeyboardGuideModal v-if="displayKeyboardHelpModal" v-model:display="displayKeyboardHelpModal"></KeyboardGuideModal>

        <LogoutModal v-if="displayLogout" v-model:display="displayLogout"></LogoutModal>

        <LoadingOverlay v-if="locked || loadingAuth" :fixed="true" :issues="loadingAuthError"></LoadingOverlay>

        <LoginModal v-if="locked && !loadingAuth"></LoginModal>

        <SnackBar></SnackBar>

        <div v-if="newVersionAvailable && !newVersionDismissed" class="new-version-notice">
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

import { AuthController } from "../../control/auth";
import type { ColorThemeName } from "@/control/app-preferences";
import { getTheme } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";
import { isTouchDevice } from "@/utils/touch";
import {
    EVENT_NAME_THEME_CHANGED,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_AUTH_LOADING,
    EVENT_NAME_AUTH_ERROR,
    EVENT_NAME_APP_NEW_VERSION,
} from "@/control/app-events";

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
    loader: () => import("@/components/layout/LoginModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const LogoutModal = defineAsyncComponent({
    loader: () => import("@/components/modals/LogoutModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const VaultSettingsDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/VaultSettingsDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountSettingsDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/AccountSettingsDropdown.vue"),
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

const ClearBrowserDataModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ClearBrowserDataModal.vue"),
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

const InviteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/InviteModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountSecuritySettingsModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountSecuritySettingsModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

export default defineComponent({
    name: "MainLayout",
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
        VaultSettingsDropdown,
        AccountSettingsDropdown,
        LanguageDropdown,
        ThemeDropdown,
        ChangeUsernameModal,
        ChangePasswordModal,
        AdvancedSettingsModal,
        AccountsAdminModal,
        TaskListModal,
        ClearBrowserDataModal,
        HelpHubDropdown,
        AboutModal,
        KeyboardGuideModal,
        BatchOperationModal,
        InviteModal,
        AccountSecuritySettingsModal,
        SnackBar,
    },
    data: function () {
        return {
            theme: getTheme(),

            locked: AuthController.Locked,
            loadingAuth: AuthController.Loading,
            loadingAuthError: false,

            layout: AppStatus.CurrentLayout,
            focus: AppStatus.CurrentFocus,

            displayLogout: false,
            displayVaultSettings: false,
            displayAccountSettings: false,
            displayTheme: false,
            displayLang: false,
            displayUsernameModal: false,
            displayPasswordModal: false,
            displayAdvancedSettings: false,
            displayBatchOperation: false,
            displayInvite: false,
            displaySecuritySettings: false,

            displayAccountAdmin: false,

            displayTaskList: false,

            displayClearBrowserData: false,

            displaySidebar: window.innerWidth >= 1000,

            displayHelpModal: false,

            displayAboutModal: false,
            displayKeyboardHelpModal: false,

            newVersionAvailable: false,
            newVersionDismissed: false,

            displayUpload: false,
        };
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_THEME_CHANGED, this.onThemeChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onAppStatusUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.onAuthStatusChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_LOADING, this.onAuthStatusLoading.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_ERROR, this.onAuthLoadingError.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_APP_NEW_VERSION, this.onNewAppVersion.bind(this));
    },
    methods: {
        logout: function () {
            this.displayLogout = true;
        },

        showVaultSettings: function () {
            this.displayVaultSettings = !this.displayVaultSettings;
            this.displayAccountSettings = false;
            this.displayHelpModal = false;
            this.displayTheme = false;
            this.displayLang = false;
        },

        showAccountSettings: function () {
            this.displayAccountSettings = !this.displayAccountSettings;
            this.displayVaultSettings = false;
            this.displayHelpModal = false;
            this.displayTheme = false;
            this.displayLang = false;
        },

        showHelp: function () {
            this.displayHelpModal = !this.displayHelpModal;
            this.displayVaultSettings = false;
            this.displayAccountSettings = false;
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
                case "clear-browser-data":
                    this.displayClearBrowserData = true;
                    break;
                case "admin":
                    this.displayAccountAdmin = true;
                    break;
                case "invite":
                    this.displayInvite = true;
                    break;
                case "security":
                    this.displaySecuritySettings = true;
                    break;
                case "logout":
                    AuthController.Logout();
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

        skipToMainContent: function (event: Event) {
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
                    if (skipTo && !isTouchDevice()) {
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
                this.displayVaultSettings = false;
                this.displayAccountSettings = false;
                this.displayTheme = false;
                this.displayLang = false;
                this.displayUsernameModal = false;
                this.displayPasswordModal = false;
                this.displayAdvancedSettings = false;
                this.displayBatchOperation = false;
                this.displayInvite = false;
                this.displaySecuritySettings = false;

                this.displayAccountAdmin = false;
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
});
</script>
