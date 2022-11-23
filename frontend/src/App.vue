<template>
  <MainLayout></MainLayout>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import MainLayout from "./components/layout/MainLayout.vue";
import { AlbumsController } from "./control/albums";
import { AppEvents } from "./control/app-events";
import { AppStatus } from "./control/app-status";
import { MediaController } from "./control/media";

@Options({
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
              MediaController.MediaData.title +
              " | " +
              AlbumsController.CurrentAlbumData.name +
              " | " +
              this.$t("Personal Media Vault");
          } else {
            document.title =
              MediaController.MediaData.title +
              " | " +
              this.$t("Personal Media Vault");
          }
        } else if (AppStatus.ListSplitMode) {
          // Media with list
          document.title =
            MediaController.MediaData.title +
            " | " +
            this.$t("Personal Media Vault");
        } else {
          // Media alone
          document.title =
            MediaController.MediaData.title +
            " | " +
            this.$t("Personal Media Vault");
        }
      } else if (AppStatus.CurrentAlbum >= 0) {
        if (AlbumsController.CurrentAlbumData) {
          document.title =
            AlbumsController.CurrentAlbumData.name +
            " | " +
            this.$t("Personal Media Vault");
        } else {
          document.title = this.$t("Personal Media Vault");
        }
      } else {
        switch (AppStatus.CurrentPage) {
          case "search":
            document.title =
              this.$t("Search results") +
              ": " +
              AppStatus.CurrentSearch +
              " | " +
              this.$t("Personal Media Vault");
            break;
          case "upload":
            document.title =
              this.$t("Upload") + " | " + this.$t("Personal Media Vault");
            break;
          case "random":
            document.title =
              this.$t("Random") + " | " + this.$t("Personal Media Vault");
            break;
          case "albums":
            document.title =
              this.$t("Albums") + " | " + this.$t("Personal Media Vault");
            break;
          default:
            document.title = this.$t("Personal Media Vault");
        }
      }
    },
  },
  mounted: function () {
    this.updateTitle();
    this.$options.updateH = this.updateTitle.bind(this);

    AppEvents.AddEventListener("app-status-update", this.$options.updateH);
    AppEvents.AddEventListener("current-album-update", this.$options.updateH);
    AppEvents.AddEventListener("current-media-update", this.$options.updateH);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("app-status-update", this.$options.updateH);
    AppEvents.RemoveEventListener(
      "current-album-update",
      this.$options.updateH
    );
    AppEvents.RemoveEventListener(
      "current-media-update",
      this.$options.updateH
    );
  },
})
export default class App extends Vue {}
</script>

<style>
/* Font awesome (minified version) */

@import "./assets/font-awesome.css";

/* Base style */

@import "./style/base.css";

/* Custom scroll bar */

@import "./style/scrollbar.css";

/* Theme colors */

@import "./style/theme-colors.css";

/* Common styles */

@import "./style/common/forms.css";
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

/* Player */

@import "./style/player/common.css";

@import "./style/player/loader.css";

@import "./style/player/player-top-bar.css";
@import "./style/player/editor.css";

@import "./style/player/player-config.css";
@import "./style/player/context-menu.css";
@import "./style/player/change-preview.css";

@import "./style/player/empty-player.css";

@import "./style/player/video-player.css";
@import "./style/player/audio-player.css";
@import "./style/player/image-player.css";

@import "./style/player/timeline.css";

@import "./style/player/volume.css";
@import "./style/player/scale.css";

@import "./style/player/subtitles.css";
</style>
