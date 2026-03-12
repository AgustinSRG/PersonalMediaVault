<template>
    <MainLayout></MainLayout>
</template>

<script setup lang="ts">
import MainLayout from "./components/layout/MainLayout.vue";
import { AppStatus } from "./control/app-status";
import { type UploadEntryMin } from "./control/upload";
import { getAssetURL } from "@/utils/api";
import { AuthController } from "./control/auth";
import {
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_CURRENT_ALBUM_UPDATED,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_UPLOAD_LIST_ENTRY_READY,
    EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR,
} from "./control/app-events";
import { useI18n } from "./composables/use-i18n";
import { onApplicationEvent } from "./composables/on-app-event";
import { showSnackBar } from "@/control/snack-bar";
import { getCurrentAlbumData } from "./control/album";
import { getCurrentMediaData } from "./control/media";

// Translation function
const { $t } = useI18n();

/**
 * Gets the application title
 * @returns The application title
 */
const getAppTitle = () => {
    return AuthController.Title || $t("Personal Media Vault");
};

/**
 * Updates the document title
 */
const updateTitle = () => {
    const currentMediaData = getCurrentMediaData();

    if (AppStatus.CurrentMedia >= 0 && currentMediaData) {
        if (AppStatus.CurrentAlbum >= 0) {
            // Media with album list
            const currentAlbumData = getCurrentAlbumData();

            if (currentAlbumData) {
                document.title = currentMediaData.title + " | " + currentAlbumData.name + " | " + getAppTitle();
            } else {
                document.title = currentMediaData.title + " | " + getAppTitle();
            }
        } else if (AppStatus.ListSplitMode) {
            // Media with list
            document.title = currentMediaData.title + " | " + getAppTitle();
        } else {
            // Media alone
            document.title = currentMediaData.title + " | " + getAppTitle();
        }
    } else if (AppStatus.CurrentAlbum >= 0) {
        // Empty album

        const currentAlbumData = getCurrentAlbumData();

        if (currentAlbumData) {
            document.title = currentAlbumData.name + " | " + getAppTitle();
        } else {
            document.title = getAppTitle();
        }
    } else {
        const searchPart = AppStatus.CurrentSearch ? " (" + $t("Tag") + ": " + AppStatus.CurrentSearch + ")" : "";
        switch (AppStatus.CurrentPage) {
            case "upload":
                document.title = $t("Upload") + " | " + getAppTitle();
                break;
            case "random":
                document.title = $t("Random") + searchPart + " | " + getAppTitle();
                break;
            case "albums":
                document.title = $t("Albums") + " | " + getAppTitle();
                break;
            case "media":
                document.title = $t("Media") + searchPart + " | " + getAppTitle();
                break;
            case "search":
                document.title = $t("Find media") + " | " + getAppTitle();
                break;
            default:
                document.title = getAppTitle();
        }
    }
};

/**
 * Updates 'mediaSession.metadata' for navigator media features
 */
const updateMediaMetadata = () => {
    if (!window.navigator || !window.navigator.mediaSession) {
        return;
    }

    const currentMediaData = getCurrentMediaData();
    const currentAlbumData = getCurrentAlbumData();

    if (AppStatus.CurrentMedia >= 0 && currentMediaData) {
        window.navigator.mediaSession.metadata = new MediaMetadata({
            title: currentMediaData.title,
            album: AppStatus.CurrentAlbum >= 0 && currentAlbumData ? currentAlbumData.name : undefined,
            artwork: currentMediaData.thumbnail
                ? [{ src: getAssetURL(currentMediaData.thumbnail), sizes: "250x250", type: "image/jpeg" }]
                : undefined,
        });
    } else {
        window.navigator.mediaSession.metadata = null;
    }
};

/**
 * Updates application status
 */
const updateAppStatus = () => {
    updateTitle();
    updateMediaMetadata();
};

updateAppStatus();
onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, updateAppStatus);
onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, updateAppStatus);
onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, updateAppStatus);

// Notify when upload is ready
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_READY, (m: UploadEntryMin) => {
    showSnackBar($t("Successfully uploaded") + ": " + m.name);
});

// Notify when upload fails
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR, (m: UploadEntryMin) => {
    showSnackBar($t("Error uploading file") + ": " + m.name);
});
</script>

<style>
/* Base style */

@import "@/style/base.css";

/* Font awesome (minified version) */

@import "./assets/font-awesome.css";

/* Custom scroll bar */

@import "@/style/scrollbar.css";

/* Theme colors */

@import "@/style/theme-colors.css";

/* Common styles */

@import "@/style/common/forms.css";
@import "@/style/common/h-menu.css";
@import "@/style/common/modals.css";
@import "@/style/common/switch.css";
@import "@/style/common/tables.css";

/* Layout */

@import "@/style/layout/bottom-bar.css";
@import "@/style/layout/loader.css";
@import "@/style/layout/main-layout.css";
@import "@/style/layout/side-bar.css";
@import "@/style/layout/snack-bar.css";
@import "@/style/layout/top-bar.css";

/* Content */

@import "@/style/content/albums.css";
@import "@/style/content/media-results.css";
@import "@/style/content/media-tags.css";
@import "@/style/content/page.css";
@import "@/style/content/paginated-menu.css";
@import "@/style/content/tasks.css";
@import "@/style/content/home.css";
@import "@/style/content/upload.css";
@import "@/style/content/batch-operation.css";
@import "@/style/content/invite-modal.css";
@import "@/style/content/page-settings.css";
@import "@/style/content/thumbnail-crop.css";
@import "@/style/content/tfa.css";
@import "@/style/content/big-selector.css";

/* Player styles */

@import "@/style/player/common.css";

@import "@/style/player/loader.css";

@import "@/style/player/player-top-bar.css";
@import "@/style/player/editor.css";

@import "@/style/player/player-config.css";
@import "@/style/player/attachments-list.css";
@import "@/style/player/related-media-list.css";
@import "@/style/player/context-menu.css";
@import "@/style/player/change-preview.css";

@import "@/style/player/empty-player.css";

@import "@/style/player/video-player.css";
@import "@/style/player/audio-player.css";
@import "@/style/player/image-player.css";

@import "@/style/player/timeline.css";

@import "@/style/player/volume.css";
@import "@/style/player/scale.css";

@import "@/style/player/subtitles.css";

@import "@/style/player/image-notes.css";

@import "@/style/player/resizable-widget.css";

@import "@/style/player/time-slices-edit.css";

@import "@/style/player/tags-edit.css";

@import "@/style/player/description.css";
</style>
