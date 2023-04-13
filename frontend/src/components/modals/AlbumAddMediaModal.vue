<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @keydown="keyDownHandle"
  >
    <div
      v-if="display"
      class="modal-dialog modal-xl modal-height-100"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title" v-if="!isUpload">
          {{ $t("Search media to add to the album") }}
        </div>
        <div class="modal-title" v-if="isUpload">
          {{ $t("Upload media to add to the album") }}
        </div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body no-padding" v-if="!isUpload">
        <PageAdvancedSearch
          :display="true"
          :inModal="true"
          :noAlbum="aid"
          @select-media="selectMedia"
          @change-to-upload="changeToUpload"
        ></PageAdvancedSearch>
      </div>
      <div class="modal-body no-padding" v-if="isUpload">
        <PageUpload
          :display="true"
          :inModal="true"
          :fixedAlbum="aid"
          @change-to-search="changeToSearch"
          @media-go="close"
        ></PageUpload>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

import PageAdvancedSearch from "@/components/pages/PageAdvancedSearch.vue";
import PageUpload from "@/components/pages/PageUpload.vue";
import { Request } from "@/utils/request";
import { AlbumsAPI } from "@/api/api-albums";
import { AppEvents } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";

export default defineComponent({
  components: {
    PageAdvancedSearch,
    PageUpload,
  },
  name: "AlbumAddMediaModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
    aid: Number,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      busy: false,

      isUpload: false,
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
    },

    changeToUpload: function () {
      this.isUpload = true;
    },

    changeToSearch: function () {
      this.isUpload = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    selectMedia: function (mid, callback) {
      if (this.busy) {
        return;
      }
      const albumId = this.aid;
      this.busy = true;
      // Add
      Request.Do(AlbumsAPI.AddMediaToAlbum(albumId, mid))
        .onSuccess(() => {
          this.busy = false;
          AppEvents.Emit("snack", this.$t("Successfully added to album"));
          AlbumsController.OnChangedAlbum(albumId, true);
          callback();
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.busy = false;
          console.error(err);
        });
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));
  },
  beforeUnmount: function () {
    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        nextTick(() => {
          this.$el.focus();
        });
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },
  },
});
</script>
