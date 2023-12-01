<template>
    <MainLayout></MainLayout>
</template>

<script lang="ts">
import MainLayout from "./components/layout/MainLayout.vue";
import { AlbumsController, EVENT_NAME_CURRENT_ALBUM_UPDATED } from "./control/albums";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "./control/app-status";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "./control/media";
import { EVENT_NAME_UPLOAD_LIST_UPDATE, UploadEntryMin } from "./control/upload";
import { getAssetURL } from "@/utils/api";
import { AuthController } from "./control/auth";
import { defineComponent } from "vue";
import { EVENT_NAME_LOADED_LOCALE } from "./i18n";
import { PagesController } from "./control/pages";

export default defineComponent({
    name: "App",
    components: {
        MainLayout,
    },
    data: function () {
        return {};
    },
    methods: {
        updateTitle: function () {
            if (AppStatus.CurrentMedia >= 0 && MediaController.MediaData) {
                if (AppStatus.CurrentAlbum >= 0) {
                    // Media with album list
                    if (AlbumsController.CurrentAlbumData) {
                        document.title =
                            MediaController.MediaData.title + " | " + AlbumsController.CurrentAlbumData.name + " | " + this.getAppTitle();
                    } else {
                        document.title = MediaController.MediaData.title + " | " + this.getAppTitle();
                    }
                } else if (AppStatus.ListSplitMode) {
                    // Media with list
                    document.title = MediaController.MediaData.title + " | " + this.getAppTitle();
                } else {
                    // Media alone
                    document.title = MediaController.MediaData.title + " | " + this.getAppTitle();
                }
            } else if (AppStatus.CurrentAlbum >= 0) {
                if (AlbumsController.CurrentAlbumData) {
                    document.title = AlbumsController.CurrentAlbumData.name + " | " + this.getAppTitle();
                } else {
                    document.title = this.getAppTitle();
                }
            } else {
                switch (AppStatus.CurrentPage) {
                    case "search":
                        document.title = this.$t("Search results") + ": " + AppStatus.CurrentSearch + " | " + this.getAppTitle();
                        break;
                    case "upload":
                        document.title = this.$t("Upload") + " | " + this.getAppTitle();
                        break;
                    case "random":
                        document.title = this.$t("Random") + " | " + this.getAppTitle();
                        break;
                    case "albums":
                        document.title = this.$t("Albums") + " | " + this.getAppTitle();
                        break;
                    default:
                        document.title = this.getAppTitle();
                }
            }
        },

        getAppTitle: function () {
            return AuthController.Title || this.$t("Personal Media Vault");
        },

        updateMediaMetadata: function () {
            if (!window.navigator || !window.navigator.mediaSession) {
                return;
            }
            if (AppStatus.CurrentMedia >= 0 && MediaController.MediaData) {
                window.navigator.mediaSession.metadata = new MediaMetadata({
                    title: MediaController.MediaData.title,
                    album:
                        AppStatus.CurrentAlbum >= 0 && AlbumsController.CurrentAlbumData
                            ? AlbumsController.CurrentAlbumData.name
                            : undefined,
                    artwork: MediaController.MediaData.thumbnail
                        ? [{ src: getAssetURL(MediaController.MediaData.thumbnail), sizes: "250x250", type: "image/jpeg" }]
                        : undefined,
                });
            } else {
                window.navigator.mediaSession.metadata = null;
            }
        },

        updateAppStatus: function () {
            this.updateTitle();
            this.updateMediaMetadata();
        },

        onUploadFinished: function (mode: "push" | "rm" | "update", m: UploadEntryMin) {
            if (mode !== "update") {
                return;
            }
            if (m.status === "ready") {
                PagesController.ShowSnackBar(this.$t("Successfully uploaded") + ": " + m.name);
            } else if (m.status === "error") {
                PagesController.ShowSnackBar(this.$t("Error uploading file") + ": " + m.name);
            }
        },

        onLoadedLocale: function (locale: string) {
            this.$updateLocale(locale);
        },
    },
    mounted: function () {
        this.updateAppStatus();

        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.updateAppStatus.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, this.updateAppStatus.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateAppStatus.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_UPDATE, this.onUploadFinished.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_LOADED_LOCALE, this.onLoadedLocale.bind(this));
    },
});
</script>

<style>
/* Base style */

@import "./style/base.css";

/* Font awesome (minified version) */

@import "./assets/font-awesome.css";

/* Custom scroll bar */

@import "./style/scrollbar.css";

/* Theme colors */

@import "./style/theme-colors.css";

/* Common styles */

@import "./style/common/forms.css";
@import "./style/common/h-menu.css";
@import "./style/common/modals.css";
@import "./style/common/switch.css";
@import "./style/common/tables.css";

/* Layout */

@import "./style/layout/bottom-bar.css";
@import "./style/layout/loader.css";
@import "./style/layout/main-layout.css";
@import "./style/layout/side-bar.css";
@import "./style/layout/snack-bar.css";
@import "./style/layout/top-bar.css";

/* Content */

@import "./style/content/albums.css";
@import "./style/content/media-results.css";
@import "./style/content/media-tags.css";
@import "./style/content/page.css";
@import "./style/content/paginated-menu.css";
@import "./style/content/tasks.css";
@import "./style/content/upload.css";

/* Player style imported in PlayerContainer component (for code-split) */
</style>
