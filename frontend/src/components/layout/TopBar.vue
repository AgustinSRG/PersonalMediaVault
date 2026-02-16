<template>
    <div class="top-bar" :class="{ 'search-focused': searchFocused }" tabindex="-1">
        <div class="top-bar-logo-td">
            <button type="button" class="top-bar-button top-bar-menu-btn" :title="$t('Main menu')" @click="menu">
                <i class="fas fa-bars"></i>
            </button>
            <img class="top-bar-logo-img" src="/img/icons/favicon.png" :alt="appLogo" />
            <span :title="appTitle" class="top-bar-title">{{ appLogo }}</span>
        </div>
        <div class="top-bar-search-td">
            <div class="top-bar-center-div">
                <VaultSearchInput ref="searchInput" v-model:focused="searchFocused"></VaultSearchInput>
            </div>
        </div>
        <div class="top-bar-user-td">
            <button type="button" class="top-bar-button top-bar-button-small-version" :title="$t('Search')" @click="focusSearch">
                <i class="fas fa-search"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Help')" @click="help">
                <i class="fas fa-question"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Vault settings')" @click="vaultSettings">
                <i class="fas fa-cog"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Account settings')" @click="accountSettings">
                <i class="fas fa-user-cog"></i>
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { AuthController } from "@/control/auth";
import { computed, ref, useTemplateRef } from "vue";
import { EVENT_NAME_AUTH_CHANGED } from "@/control/app-events";
import VaultSearchInput from "../utils/VaultSearchInput.vue";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Events
const emit = defineEmits<{
    /**
     * The user wants to log out
     */
    (e: "logout"): void;

    /**
     * The user wants to open the vault settings menu
     */
    (e: "vault-settings"): void;

    /**
     * The user wants to open the account settings menu
     */
    (e: "account-settings"): void;

    /**
     * The user wants to open the menu
     */
    (e: "menu"): void;

    /**
     * Tne user wants to focus the menu
     */
    (e: "menu-focus"): void;

    /**
     * The user wants to open the help menu
     */
    (e: "help"): void;
}>();

// Is search focused?
const searchFocused = ref(false);

// Custom title
const customTitle = ref(AuthController.Title);

// Custom logo
const customLogo = ref(AuthController.Logo);

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    customTitle.value = AuthController.Title;
    customLogo.value = AuthController.Logo;
});

// Application title
const appTitle = computed(() => customTitle.value || $t("Personal Media Vault"));

// Application logo
const appLogo = computed(() => customLogo.value || "PMV");

/**
 * Opens the menu
 */
const menu = () => {
    emit("menu");
};

/**
 * Logs out the vault
 */
const logout = () => {
    emit("logout");
};

/**
 * Opens vault settings
 */
const vaultSettings = () => {
    emit("vault-settings");
};

/**
 * Opens account settings
 */
const accountSettings = () => {
    emit("account-settings");
};

/**
 * Opens the help menu
 */
const help = () => {
    emit("help");
};

// Ref to the search input component
const searchInput = useTemplateRef("searchInput");

/**
 * Focuses the search input
 */
const focusSearch = () => {
    searchInput.value?.focus();
};

useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !event.key) {
        return false;
    }

    if (event.key.toUpperCase() === "M" && event.ctrlKey) {
        emit("menu-focus");
        return true;
    }

    if (event.key.toUpperCase() === "F" && event.ctrlKey) {
        focusSearch();
        return true;
    }

    if (event.key.toUpperCase() === "S" && event.ctrlKey) {
        if (event.shiftKey) {
            accountSettings();
        } else {
            vaultSettings();
        }

        return true;
    }

    if (event.key.toUpperCase() === "H" && event.ctrlKey && event.shiftKey) {
        help();
        return true;
    }

    if (event.key.toUpperCase() === "Q" && event.ctrlKey) {
        logout();
        return true;
    }

    return false;
});
</script>
