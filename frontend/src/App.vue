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
/* Add here any global / shared styles */

.bold {
  font-weight: bold;
}

.text-right {
  text-align: right;
}

.text-center {
  text-align: center;
}

.text-left {
  text-align: left;
}

.one-line {
  white-space: nowrap;
}

.btn {
  display: inline-block;

  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: white;
  background: transparent;
  border-radius: 100vw;
  padding: 12px 24px;
  white-space: nowrap;
}

.light-theme .btn {
  color: black;
}

.dark-theme .btn {
  color: white;
}

.btn-sm {
  font-size: 16px;
}

.btn-xs {
  font-size: 12px;
}

.btn-mr {
  margin-right: 0.5rem;
  margin-bottom: 0.5rem;
}

.btn i {
  margin-right: 12px;
}

.btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.btn:hover {
  opacity: 1;
}

.btn:disabled:hover {
  opacity: 0.7;
}

.light-theme .btn-primary {
  border: solid 1px rgba(0, 0, 0, 0.1);
}

.light-theme .btn-primary:not(:disabled):hover {
  background: rgba(0, 0, 0, 0.1);
}

.dark-theme .btn-primary {
  border: solid 1px rgba(255, 255, 255, 0.1);
}

.dark-theme .btn-primary:not(:disabled):hover {
  background: rgba(255, 255, 255, 0.1);
}

.btn-danger {
  border: solid 1px rgba(255, 0, 0, 0.1);
}

.btn-danger:not(:disabled):hover {
  background: rgba(255, 0, 0, 0.1);
}

.form-group {
  padding-bottom: 0.75rem;
}

.form-group label {
  display: block;
  width: 100%;
  padding-bottom: 0.3rem;
}

.form-control {
  height: calc(1.5em + 0.75rem + 2px);
  padding: 0.375rem 0.75rem;
  font-size: 1rem;
  font-weight: 400;
  line-height: 1.5;
  background-clip: padding-box;

  border-radius: 0.25rem;
  transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
}

.light-theme .form-control {
  border: 1px solid rgba(0, 0, 0, 0.1);
  color: black;
  background: white;
}

.dark-theme .form-control {
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  background: hsl(0, 0%, 7%);
}

.light-theme .form-control:focus {
  color: black;
  background: white;
  border: 1px solid rgba(0, 0, 0, 0.1);
  outline: 0;
  box-shadow: 0 0 0 0.2rem rgba(0, 0, 0, 0.1);
}

.dark-theme .form-control:focus {
  color: #fff;
  background: hsl(0, 0%, 7%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  outline: 0;
  box-shadow: 0 0 0 0.2rem rgba(255, 255, 255, 0.1);
}

.form-control-full-width {
  width: 100%;
}

.form-error {
  color: red;
}

/* Custom scroll bar */

*::-webkit-scrollbar {
  width: 16px;
  height: 16px;
}

*::-webkit-scrollbar-track {
  background: transparent;
}

*::-webkit-scrollbar-thumb {
  background: #757575;
  border-radius: 8px;
  border: 4px solid transparent;
  background-clip: content-box;
}

a, a:visited {
  color: inherit;
}
</style>
