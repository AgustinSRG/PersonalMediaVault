<template>
    <div
        ref="container"
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
        <SideBar v-model:display="displaySidebar" @skip-to-content="skipToMainContent"></SideBar>
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

<script setup lang="ts">
import { defineAsyncComponent, nextTick, ref, useTemplateRef } from "vue";

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
import { useI18n } from "@/composables/use-i18n";
import { onApplicationEvent } from "@/composables/on-app-event";

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

// Ref to the container element
const container = useTemplateRef("container");

// Translation function
const { $t } = useI18n();

// Current theme
const theme = ref(getTheme());

// Is vault locked?
const locked = ref(AuthController.Locked);

// Is auth status being loaded?
const loadingAuth = ref(AuthController.Loading);

// Is there an error loading the auth status?
const loadingAuthError = ref(false);

// Current layout
const layout = ref(AppStatus.CurrentLayout);

// Current focus
const focus = ref(AppStatus.CurrentFocus);

// Display the logout modal?
const displayLogout = ref(false);

// Display the vault settings dropdown?
const displayVaultSettings = ref(false);

// Display the account settings dropdown?
const displayAccountSettings = ref(false);

// Display the theme dropdown
const displayTheme = ref(false);

// Display the language dropdown?
const displayLang = ref(false);

// Display the modal to change username?
const displayUsernameModal = ref(false);

// Display modal to change password?
const displayPasswordModal = ref(false);

// Display advanced settings modal?
const displayAdvancedSettings = ref(false);

// Display batch operation modal?
const displayBatchOperation = ref(false);

// Display the invite options modal?
const displayInvite = ref(false);

// Display security settings modal?
const displaySecuritySettings = ref(false);

// Display account admin modal?
const displayAccountAdmin = ref(false);

// Display task list modal?
const displayTaskList = ref(false);

// Display clear browser data modal?
const displayClearBrowserData = ref(false);

// Display sidebar?
const displaySidebar = ref(window.innerWidth >= 1000);

// Display help modal?
const displayHelpModal = ref(false);

// Display about modal?
const displayAboutModal = ref(false);

// Display keyboard help modal?
const displayKeyboardHelpModal = ref(false);

// Display upload modal?
const displayUpload = ref(false);

// Is a new version available?
const newVersionAvailable = ref(false);

// Has the user dismissed the warning of a new version
const newVersionDismissed = ref(false);

onApplicationEvent(EVENT_NAME_THEME_CHANGED, (t: ColorThemeName) => {
    theme.value = t;
});

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    layout.value = AppStatus.CurrentLayout;
    focus.value = AppStatus.CurrentFocus;
});

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, (l: boolean) => {
    locked.value = l;
    loadingAuthError.value = false;

    if (locked.value) {
        // Close all modals
        displayLogout.value = false;
        displayVaultSettings.value = false;
        displayAccountSettings.value = false;
        displayTheme.value = false;
        displayLang.value = false;
        displayUsernameModal.value = false;
        displayPasswordModal.value = false;
        displayAdvancedSettings.value = false;
        displayBatchOperation.value = false;
        displayInvite.value = false;
        displaySecuritySettings.value = false;

        displayAccountAdmin.value = false;
    }
});

onApplicationEvent(EVENT_NAME_AUTH_LOADING, (l: boolean) => {
    loadingAuth.value = l;
});

onApplicationEvent(EVENT_NAME_AUTH_ERROR, () => {
    loadingAuthError.value = true;
});

onApplicationEvent(EVENT_NAME_APP_NEW_VERSION, () => {
    newVersionAvailable.value = true;
});

/**
 * Opens logout modal
 */
const logout = () => {
    displayLogout.value = true;
};

/**
 * Opens vault settings dropdown
 */
const showVaultSettings = () => {
    displayVaultSettings.value = !displayVaultSettings.value;
    displayAccountSettings.value = false;
    displayHelpModal.value = false;
    displayTheme.value = false;
    displayLang.value = false;
};

/**
 * Opens account settings dropdown
 */
const showAccountSettings = () => {
    displayAccountSettings.value = !displayAccountSettings.value;
    displayVaultSettings.value = false;
    displayHelpModal.value = false;
    displayTheme.value = false;
    displayLang.value = false;
};

/**
 * Opens the help dropdown
 */
const showHelp = () => {
    displayHelpModal.value = !displayHelpModal.value;
    displayVaultSettings.value = false;
    displayAccountSettings.value = false;
    displayTheme.value = false;
    displayLang.value = false;
};

/**
 * Handler for events on the setting dropdowns.
 * Opens the corresponding modal or performs the corresponding action.
 * @param o Selected option
 */
const onGoSettings = (o: string) => {
    switch (o) {
        case "theme":
            displayTheme.value = true;
            break;
        case "lang":
            displayLang.value = true;
            break;
        case "username":
            displayUsernameModal.value = true;
            break;
        case "password":
            displayPasswordModal.value = true;
            break;
        case "advanced":
            displayAdvancedSettings.value = true;
            break;
        case "batch":
            displayBatchOperation.value = true;
            break;
        case "tasks":
            displayTaskList.value = true;
            break;
        case "clear-browser-data":
            displayClearBrowserData.value = true;
            break;
        case "admin":
            displayAccountAdmin.value = true;
            break;
        case "invite":
            displayInvite.value = true;
            break;
        case "security":
            displaySecuritySettings.value = true;
            break;
        case "logout":
            AuthController.Logout();
            break;
    }
};

/**
 * Handler for events on the help dropdown.
 * Opens the corresponding modal or performs the corresponding action.
 * @param o Selected option
 */
const onGoHelp = (o: string) => {
    switch (o) {
        case "about":
            displayAboutModal.value = true;
            break;
        case "keyboard":
            displayKeyboardHelpModal.value = true;
            break;
    }
};

/**
 * Focuses sidebar element
 */
const focusSidebar = () => {
    nextTick(() => {
        const sideBar = container.value?.querySelector(".side-bar") as HTMLElement;
        if (sideBar) {
            sideBar.focus();
        }
    });
};

/**
 * Toggles sidebar visibility
 */
const toggleSidebar = () => {
    displaySidebar.value = !displaySidebar.value;

    if (displaySidebar.value) {
        focusSidebar();
    }
};

/**
 * Opens the sidebar
 */
const openSideBar = () => {
    displaySidebar.value = true;

    focusSidebar();
};

/**
 * Hides the sidebar
 */
const hideSidebar = () => {
    displaySidebar.value = false;
};

/**
 * Skips focus to main content
 * @param event The click event
 */
const skipToMainContent = (event?: Event) => {
    if (event) {
        event.preventDefault();
    }
    let skipTo: HTMLElement;
    switch (AppStatus.CurrentLayout) {
        case "media":
        case "media-split":
        case "album":
            skipTo = container.value?.querySelector(".player-container");
            break;
        default:
            skipTo = container.value?.querySelector(".page-content");
            if (skipTo && !isTouchDevice()) {
                const autoFocused = skipTo.querySelector(".auto-focus") as HTMLElement;
                if (autoFocused) {
                    skipTo = autoFocused;
                }
            }
    }
    if (skipTo) {
        skipTo.focus();
    }
};

/**
 * Dismissed warning for new version
 */
const dismissNewVersion = () => {
    newVersionDismissed.value = true;
};

/**
 * Reloads the page clearing the cache,
 * forcing to use the newest version.
 */
const hardReload = () => {
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
};
</script>
