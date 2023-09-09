<template>
    <MainLayout></MainLayout>
</template>

<script lang="ts">

import MainLayout from "./components/layout/MainLayout.vue";
import { AlbumsController } from "./control/albums";
import { AppEvents } from "./control/app-events";
import { AppStatus } from "./control/app-status";
import { MediaController } from "./control/media";
import { UploadEntryMin } from "./control/upload";
import { GetAssetURL } from "./utils/request";
import { AuthController } from "./control/auth";
import { defineComponent } from "vue";

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
                        ? [{ src: GetAssetURL(MediaController.MediaData.thumbnail), sizes: "250x250", type: "image/jpeg" }]
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

        onUploadFinished: function (i, m: UploadEntryMin) {
            if (m.status === "ready") {
                AppEvents.Emit("snack", this.$t("Successfully uploaded") + ": " + m.name);
            } else if (m.status === "error") {
                AppEvents.Emit("snack", this.$t("Error uploading file") + ": " + m.name);
            }
        },

        onLoadedLocale: function (locale: string) {
            this.$updateLocale(locale);
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this.updateAppStatus();
        this._handles.updateH = this.updateAppStatus.bind(this);

        AppEvents.AddEventListener("app-status-update", this._handles.updateH);
        AppEvents.AddEventListener("current-album-update", this._handles.updateH);
        AppEvents.AddEventListener("current-media-update", this._handles.updateH);

        this._handles.uploadDoneH = this.onUploadFinished.bind(this);
        AppEvents.AddEventListener("upload-list-update", this._handles.uploadDoneH);

        this._handles.onLoadedLocaleH = this.onLoadedLocale.bind(this);
        AppEvents.AddEventListener("loaded-locale", this._handles.onLoadedLocaleH);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("app-status-update", this._handles.updateH);
        AppEvents.RemoveEventListener("current-album-update", this._handles.updateH);
        AppEvents.RemoveEventListener("current-media-update", this._handles.updateH);
        AppEvents.RemoveEventListener("upload-list-update", this._handles.uploadDoneH);
        AppEvents.RemoveEventListener("loaded-locale", this._handles.onLoadedLocaleH);
    },
})
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
