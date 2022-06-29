<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <div
      class="modal-dialog modal-sm"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Albums") }}</div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body with-menu limited-size">
        <div class="albums-modal-filter">
          <input
            type="text"
            autocomplete="off"
            @input="updateAlbums"
            v-model="filter"
            class="form-control form-control-full-width"
            :placeholder="$t('Filter by name') + '...'"
          />
        </div>
        <table class="modal-menu">
          <tr v-if="albums.length === 0">
            <td colspan="2" class="albums-menu-empty">
              {{ $t("No albums found") }}
            </td>
          </tr>
          <tr
            v-for="a in albums"
            :key="a.id"
            class="modal-menu-item"
            tabindex="0"
            @click="clickOnAlbum(a)"
          >
            <td class="modal-menu-item-icon">
              <i v-if="a.added" class="far fa-square-check"></i>
              <i v-else class="far fa-square"></i>
            </td>
            <td class="modal-menu-item-title">
              {{ a.name }}
            </td>
          </tr>
        </table>
      </div>

      <div class="modal-footer">
        <button type="button" @click="createAlbum" class="modal-footer-btn">
          <i class="fas fa-plus"></i> {{ $t("Create album") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AmbumsAPI } from "@/api/api-albums";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { Request } from "@/utils/request";
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "AlbumListModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      albums: [],
      filter: "",
      mid: AppStatus.CurrentMedia,
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    createAlbum: function () {
      this.$emit("album-create");
    },

    clickOnAlbum: function (album) {
      if (album.added) {
        // Remove
        Request.Do(AmbumsAPI.RemoveMediaFromAlbum(album.id, this.mid))
          .onSuccess(() => {
            album.added = false;
            AppEvents.Emit("snack", this.$t("Successfully removed from album"));
            AlbumsController.OnChangedAlbum(album.id);
          })
          .onRequestError((err) => {
            Request.ErrorHandler()
              .add(401, "*", () => {
                AppEvents.Emit("unauthorized");
              })
              .handle(err);
          })
          .onUnexpectedError((err) => {
            console.error(err);
          });
      } else {
        // Add
        Request.Do(AmbumsAPI.AddMediaToAlbum(album.id, this.mid))
          .onSuccess(() => {
            album.added = true;
            AppEvents.Emit("snack", this.$t("Successfully added to album"));
            AlbumsController.OnChangedAlbum(album.id);
          })
          .onRequestError((err) => {
            Request.ErrorHandler()
              .add(401, "*", () => {
                AppEvents.Emit("unauthorized");
              })
              .handle(err);
          })
          .onUnexpectedError((err) => {
            console.error(err);
          });
      }
    },

    onUpdateStatus: function () {
      const changed = this.mid !== AppStatus.CurrentMedia;
      this.mid = AppStatus.CurrentMedia;
      if (changed) {
        this.updateAlbums();
      }
    },

    updateAlbums: function () {
      var mid = AppStatus.CurrentMedia;
      var filter = (this.filter + "").toLowerCase();
      this.albums = AlbumsController.GetAlbumsListCopy()
        .filter((a) => {
          return !filter || a.nameLowerCase.indexOf(filter) >= 0;
        })
        .map((a: any) => {
          a.added = mid >= 0 && a.list.indexOf(mid) >= 0;
          return a;
        })
        .sort((a, b) => {
          if (a.nameLowerCase < b.nameLowerCase) {
            return -1;
          } else if (a.nameLowerCase > b.nameLowerCase) {
            return 1;
          } else {
            return 1;
          }
        });
    },
  },
  mounted: function () {
    this.$options.albumsUpdateH = this.updateAlbums.bind(this);
    AppEvents.AddEventListener("albums-update", this.$options.albumsUpdateH);

    this.$options.statusH = this.onUpdateStatus.bind(this);
    AppEvents.AddEventListener("app-status-update", this.$options.statusH);

    this.updateAlbums();
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("albums-update", this.$options.albumsUpdateH);
    AppEvents.RemoveEventListener("app-status-update", this.$options.statusH);
  },
});
</script>

<style>
.modal-body.with-menu.limited-size {
  max-height: 300px;
  overflow-y: auto;
}

.albums-menu-empty {
  padding: 1rem;
  text-align: center;
}

.albums-modal-filter {
  padding: 0.5rem 0.25rem;
}
</style>
