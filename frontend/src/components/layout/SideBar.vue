<template>
    <div
        ref="container"
        class="side-bar"
        :class="{ hidden: !display }"
        tabindex="-1"
        :role="isInitialLayout ? '' : 'dialog'"
        @click="stopPropagationEvent"
        @keydown="keyDownHandle"
    >
        <div class="side-bar-header">
            <div class="top-bar-logo-td">
                <button type="button" class="top-bar-button" :title="$t('Main menu')" @click="close">
                    <i class="fas fa-bars"></i>
                </button>
                <img class="top-bar-logo-img" src="/img/icons/favicon.png" :alt="appLogo" />
                <span :title="appTitle" class="top-bar-title">{{ appLogo }}</span>
            </div>
        </div>
        <div class="side-bar-body" tabindex="-1">
            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'home' }"
                :title="$t('Home')"
                :href="getPageURL('home')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('home', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-home"></i></div>
                <div class="side-bar-option-text">{{ $t("Home") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'albums' }"
                :title="$t('Albums')"
                :href="getPageURL('albums', cachedAlbumsSearchParams)"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('albums', $event, cachedAlbumsSearchParams)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-list"></i></div>
                <div class="side-bar-option-text">{{ $t("Albums") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'media' }"
                :title="$t('Media')"
                :href="getPageURL('media')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('media', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-photo-film"></i></div>
                <div class="side-bar-option-text">{{ $t("Media") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'search' }"
                :title="$t('Find media')"
                :href="getPageURL('search')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('search', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-search"></i></div>
                <div class="side-bar-option-text">{{ $t("Find media") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'random' }"
                :title="$t('Random')"
                :href="getPageURL('random')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('random', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-shuffle"></i></div>
                <div class="side-bar-option-text">{{ $t("Random") }}</div>
            </a>

            <a
                v-if="canWrite"
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'upload' }"
                :title="$t('Upload')"
                :href="getPageURL('upload')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('upload', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-upload"></i></div>
                <div class="side-bar-option-text">{{ $t("Upload") }}</div>
            </a>

            <div v-if="albumsFavorite.length > 0" class="side-bar-separator"></div>

            <a
                v-for="a in albumsFavorite"
                :key="a.id"
                class="side-bar-option"
                :class="{ selected: album == a.id }"
                :title="a.name"
                :href="getAlbumURL(a.id)"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToAlbum(a.id, $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-star"></i></div>
                <div class="side-bar-option-text">{{ a.name }}</div>
            </a>

            <div class="side-bar-separator"></div>

            <a
                v-for="a in albumsRest"
                :key="a.id"
                class="side-bar-option"
                :class="{ selected: album == a.id }"
                :title="a.name"
                :href="getAlbumURL(a.id)"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToAlbum(a.id, $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-list-ol"></i></div>
                <div class="side-bar-option-text">{{ a.name }}</div>
            </a>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { AlbumListItemMinExt } from "@/control/albums";
import { AlbumsController } from "@/control/albums";
import { getAlbumFavoriteList, getAlbumsOrderMap } from "@/control/app-preferences";
import type { AppStatusPage } from "@/control/app-status";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { getFrontendUrl } from "@/utils/api";
import { computed, nextTick, ref, useTemplateRef, watch } from "vue";
import {
    EVENT_NAME_ALBUM_SIDEBAR_TOP,
    EVENT_NAME_ALBUMS_LIST_UPDATE,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_FAVORITE_ALBUMS_UPDATED,
} from "@/control/app-events";
import { useI18n } from "@/composables/use-i18n";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { stopPropagationEvent } from "@/utils/events";

// Translation
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Events
const emit = defineEmits<{
    /**
     * The user wants to skip the focus to the main content.
     */
    (e: "skip-to-content"): void;
}>();

// Custom title
const customTitle = ref(AuthController.Title);

// Custom logo
const customLogo = ref(AuthController.Logo);

// Application title
const appTitle = computed(() => customTitle.value || $t("Personal Media Vault"));

// Application logo
const appLogo = computed(() => customLogo.value || "PMV");

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    customTitle.value = AuthController.Title;
    customLogo.value = AuthController.Logo;
});

// User permissions
const { canWrite } = useUserPermissions();

// Current page
const page = ref(AppStatus.CurrentPage);

// Current album
const album = ref(AppStatus.CurrentAlbum);

// Current layout
const layout = ref(AppStatus.CurrentLayout);

// Is initial layout?
const isInitialLayout = computed(() => layout.value === "initial");

// Current search
const search = ref(AppStatus.CurrentSearch);

// Cached album search params
const cachedAlbumsSearchParams = ref(AppStatus.CurrentPage === "albums" ? AppStatus.SearchParams : "");

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    if (AppStatus.CurrentLayout !== "initial") {
        display.value = false;
    } else if (layout.value !== "initial") {
        display.value = true;
    }

    layout.value = AppStatus.CurrentLayout;

    page.value = AppStatus.CurrentPage;
    album.value = AppStatus.CurrentAlbum;

    search.value = AppStatus.CurrentSearch;

    if (AppStatus.CurrentPage === "albums") {
        cachedAlbumsSearchParams.value = AppStatus.SearchParams;
    }
});

// Albums
const albumsFavorite = ref<AlbumListItemMinExt[]>();
const albumsRest = ref<AlbumListItemMinExt[]>();

// Max number of albums in the sidebar
const MAX_ALBUMS_LIST_LENGTH_SIDEBAR = 10;

// Complete list of albums
let albums: AlbumListItemMinExt[] = [];

/**
 * Updates the albums lists
 */
const updateAlbums = () => {
    const albumsOrderMap = getAlbumsOrderMap();
    albums = AlbumsController.GetAlbumsListMin().sort((a, b) => {
        const lruA = albumsOrderMap[a.id + ""] || 0;
        const lruB = albumsOrderMap[b.id + ""] || 0;
        if (lruA > lruB) {
            return -1;
        } else if (lruA < lruB) {
            return 1;
        } else if (a.nameLowerCase < b.nameLowerCase) {
            return -1;
        } else if (a.nameLowerCase > b.nameLowerCase) {
            return 1;
        } else {
            return 0;
        }
    });
    const favIdList = getAlbumFavoriteList();
    const fav = [];
    const rest = [];
    albums.forEach((album) => {
        if (favIdList.includes(album.id + "")) {
            fav.push(album);
        } else {
            rest.push(album);
        }
    });
    albumsFavorite.value = fav;
    albumsRest.value = rest.slice(0, MAX_ALBUMS_LIST_LENGTH_SIDEBAR);
};

updateAlbums();
onApplicationEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, updateAlbums);
onApplicationEvent(EVENT_NAME_FAVORITE_ALBUMS_UPDATED, updateAlbums);

/**
 * Moves an album to the first position when accessed
 * @param albumId The album ID
 */
const putAlbumFirst = (albumId: number) => {
    for (let i = 0; i < albumsFavorite.value.length; i++) {
        if (albumsFavorite.value[i].id === albumId) {
            const albumEntry = albumsFavorite.value.splice(i, 1)[0];
            albumsFavorite.value.unshift(albumEntry);
            return;
        }
    }
    for (let i = 0; i < albumsRest.value.length; i++) {
        if (albumsRest.value[i].id === albumId) {
            const albumEntry = albumsRest.value.splice(i, 1)[0];
            albumsRest.value.unshift(albumEntry);
            return;
        }
    }
    for (let i = 0; i < albums.length; i++) {
        if (albums[i].id === albumId) {
            const albumEntry = albums[i];
            albumsRest.value.unshift(albumEntry);
            if (albumsRest.value.length > MAX_ALBUMS_LIST_LENGTH_SIDEBAR) {
                albumsRest.value.pop();
            }
            return;
        }
    }
};

onApplicationEvent(EVENT_NAME_ALBUM_SIDEBAR_TOP, putAlbumFirst);

// Ref to the container element
const container = useTemplateRef("container");

// Auto focus container when the menu is displayed
watch(display, () => {
    if (display.value && !isInitialLayout.value) {
        nextTick(() => {
            container.value?.focus();
        });
    }
});

/**
 * Closes the sidebar
 */
const close = () => {
    display.value = false;
};

/**
 * Call when the sidebar loses focus
 */
const lostFocus = () => {
    if (!isInitialLayout.value) {
        close();
    }
};

// Focus trap
useFocusTrap(container, display, lostFocus);

/**
 * Gets the URL for a page link
 * @param page The page
 * @param searchParams The search params
 * @returns The URL
 */
const getPageURL = (page: AppStatusPage, searchParams?: string): string => {
    return getFrontendUrl({
        page: page,
        sp: searchParams || null,
    });
};

/**
 * Navigates to a page
 * @param p The page
 * @param e The click event
 * @param searchParams The search params
 */
const goToPage = (p: AppStatusPage, e: Event, searchParams?: string) => {
    if (e) {
        e.preventDefault();
    }
    AppStatus.GoToPageConditionalSplit(p, searchParams);
    nextTick(() => {
        emit("skip-to-content");
    });
    if (window.innerWidth < 1000) {
        close();
    }
};

/**
 * Gets the URL of an album
 * @param albumId The album ID
 * @returns The URL
 */
const getAlbumURL = (albumId: number): string => {
    return getFrontendUrl({
        album: albumId,
    });
};

/**
 * Navigates to an album
 * @param albumId The album ID
 * @param e The click event
 */
const goToAlbum = (albumId: number, e?: Event) => {
    e?.preventDefault();

    AppStatus.ClickOnAlbum(albumId);

    nextTick(() => {
        emit("skip-to-content");
    });

    if (window.innerWidth < 1000) {
        close();
    }
};

/**
 * Event handler for 'keydown'
 * @param e The keyboard event
 */
const keyDownHandle = (e: KeyboardEvent) => {
    if (!isInitialLayout.value && e.key === "Escape") {
        e.stopPropagation();
        close();
    }
};
</script>
