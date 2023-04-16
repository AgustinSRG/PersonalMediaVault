<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @click="close" @keydown="keyDownHandle">
    <div v-if="display" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Size Statistics") }}
        </div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="loading" class="modal-body">
        <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
      </div>
      <div v-if="!loading" class="modal-body no-padding table-responsive">
        <table class="table table-text-overflow">
          <thead>
            <tr>
              <th class="text-left">{{ $t("Asset") }}</th>
              <th class="text-left">{{ $t("Size") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>METADATA</td>
              <td>{{ renderSize(metaSize) }}</td>
            </tr>
            <tr v-for="a in assets" :key="a.key">
              <td>{{ a.name }}</td>
              <td>{{ renderSize(a.size) }}</td>
            </tr>
            <tr>
              <td class="bold">{{ $t("Total") }}</td>
              <td class="bold">{{ renderSize(total) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import { Timeouts } from "@/utils/timeout";
import { Request } from "@/utils/request";
import { AuthController } from "@/control/auth";
import { MediaAPI } from "@/api/api-media";
import { AppEvents } from "@/control/app-events";

export default defineComponent({
  name: "SizeStatsModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
    mid: Number,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      loading: false,
      metaSize: 0,
      assets: [],
      total: 0,
    };
  },
  methods: {
    load: function () {
      Timeouts.Abort("media-size-stats-load");
      Request.Abort("media-size-stats-load");

      if (!this.display) {
        return;
      }

      this.loading = true;

      if (AuthController.Locked) {
        return; // Vault is locked
      }

      Request.Pending(
        "media-size-stats-load",
        MediaAPI.GetMediaSizeStats(this.mid)
      )
        .onSuccess((result) => {
          this.loading = false;
          this.metaSize = result.meta_size;
          this.assets = result.assets;

          let total = 0;

          total += result.meta_size;

          for (let asset of result.assets) {
            total += asset.size;
          }

          this.total = total;
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized", false);
            })
            .add(404, "*", () => {
              this.close();
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("media-size-stats-load", 1500, this.load.bind(this));
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("media-size-stats-load", 1500, this.load.bind(this));
        });
    },

    renderSize: function (bytes: number): string {
      if (bytes > 1024 * 1024 * 1024) {
        let gb = bytes / (1024 * 1024 * 1024);
        gb = Math.floor(gb * 100) / 100;
        return gb + " GB";
      } else if (bytes > 1024 * 1024) {
        let mb = bytes / (1024 * 1024);
        mb = Math.floor(mb * 100) / 100;
        return mb + " MB";
      } else if (bytes > 1024) {
        let kb = bytes / 1024;
        kb = Math.floor(kb * 100) / 100;
        return kb + " KB";
      } else {
        return bytes + " Bytes";
      }
    },

    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
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
    if (this.display) {
      this.$options.focusTrap.activate();
      nextTick(() => {
        this.$el.focus();
      });
      this.load();
    }
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
        this.load();
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },

    mid: function () {
      this.load();
    },
  },
});
</script>
