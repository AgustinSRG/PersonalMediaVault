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
/* Font awesome */

@import './assets/font-awesome.css';

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
  font-size: 16px;
  color: var(--theme-fg-color);
  background: transparent;
  border-radius: 100vw;
  padding: 12px 24px;
  white-space: nowrap;
}

.btn-sm {
  font-size: 12px;
}

.btn-xs {
  font-size: 10px;
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

.btn-primary {
  border: solid 1px var(--theme-border-color);
}

.btn-primary:not(:disabled):hover {
  background: var(--hover-color);
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

  max-width: 100%;

  border-radius: 0.25rem;
  transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;

  border: 1px solid var(--theme-border-color);
  color: var(--theme-fg-color);
  background: var(--input-bg-color);
}

.form-control:focus {
  border: 1px solid var(--theme-border-color);
  color: var(--theme-fg-color);
  background: var(--input-bg-color);
  outline: 0;
  box-shadow: 0 0 0 0.2rem var(--theme-border-color);
}

.form-control-full-width {
  width: 100%;
}

.form-control.form-textarea {
  resize: vertical;
  height: auto;
}

.form-error {
  color: red;
}

a,
a:visited {
  color: inherit;
}

.td-shrink {
  width: 1px;
}

.border-top {
  border-top: solid 1px var(--theme-border-color);
  padding-top: 1rem;
}

.file-hidden {
  display: none;
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

/* Modals */

.modal-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;

  background: var(--modal-overlay-bg-color);

  display: flex;
  flex-direction: column;

  padding: 0.5rem;

  transition: opacity 0.2s;
  opacity: 1;

  overflow: auto;
}

.modal-container.hidden {
  transition: opacity 0.2s, visibility 0.2s;
  pointer-events: none;
  opacity: 0;
  visibility: hidden;
}

.modal-container.no-transition,
.modal-container.no-transition.hidden {
  transition: none;
}

.modal-container:focus {
  outline: none;
}

.modal-dialog {
  display: flex;
  margin: auto;
  flex-direction: column;
  background: var(--modal-bg-color);
  box-shadow: var(--modal-shadow);
}

.modal-sm {
  width: 300px;
}

@media (max-width: 300px) {
  .modal-sm {
    width: calc(100% - 1rem);
  }
}

.modal-md {
  width: 500px;
}

@media (max-width: 600px) {
  .modal-md {
    width: calc(100% - 1rem);
  }
}

.modal-lg {
  width: 800px;
}

@media (max-width: 900px) {
  .modal-lg {
    width: calc(100% - 1rem);
  }
}

.modal-xl {
  width: 1140px;
}

@media (max-width: 1240px) {
  .modal-xl {
    width: calc(100% - 1rem);
  }
}

.modal-header {
  display: flex;
  flex-direction: row;
  align-items: center;
  border-bottom: solid 1px var(--theme-border-color);
}

.modal-title {
  width: calc(100% - 48px);
  padding: 1rem;
  font-size: 24px;
  font-weight: bold;
}

.modal-title.no-close {
  width: 100%;
}

.modal-close-btn {
  display: block;
  width: 48px;
  height: 48px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: var(--theme-btn-color);
  background: transparent;
}

.modal-close-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.modal-close-btn:not(:disabled):hover {
  color: var(--theme-btn-hover-color);
}

.modal-body,
.modal-footer {
  padding: 1rem;
}

.modal-footer.no-padding {
  padding: 0;
}

.modal-footer {
  border-top: solid 1px var(--theme-border-color);
}

.modal-footer-btn {
  display: block;
  width: 100%;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  color: var(--theme-btn-color);
  background: transparent;
  text-align: left;
  padding: 1rem;
  white-space: nowrap;
  font-weight: bold;
}

.modal-footer-btn i {
  margin-right: 0.5rem;
}

.modal-footer-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.modal-footer-btn:not(:disabled):hover {
  color: var(--theme-btn-hover-color);
}

.modal-menu {
  width: 100%;
  border-spacing: 0; /* Removes the cell spacing via CSS */
  border-collapse: collapse; /* Optional - if you don't want to have double border where cells touch */
}

.modal-body.with-menu {
  padding: 0;
}

.modal-menu-item {
  cursor: pointer;
}

.modal-menu-item-title {
  padding-top: 1rem;
  padding-right: 1rem;
  padding-bottom: 1rem;
  font-weight: bold;
}

.modal-menu-item-icon {
  padding: 1rem;
  text-align: center;
  width: 2rem;
}

.modal-menu-item:hover {
  background: var(--hover-color);
}

/* Tables */

.table-responsive {
  display: block;
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.table {
  width: 100%;
  margin-bottom: 1rem;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 0.75rem;
  vertical-align: top;
  border-top: 1px solid var(--theme-border-color);
}

.table thead th {
  vertical-align: bottom;
  border-bottom: 2px solid var(--theme-border-color);
}

.table tbody + tbody {
  border-top: 2px solid var(--theme-border-color);
}

.table.table-vmiddle td {
  vertical-align: middle;
}

.table-btn {
  display: inline-block;
  width: 32px;
  height: 32px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  background: transparent;
  color: var(--theme-btn-color);
}

.table-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.table-btn:not(:disabled):hover {
  color: var(--theme-btn-hover-color);
}
</style>
