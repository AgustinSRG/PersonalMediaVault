<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @keydown="keyDownHandle"
  >
    <form
      v-if="display"
      @submit="submit"
      class="modal-dialog modal-lg"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Advanced settings") }}</div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="loading" class="modal-body">
        <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
      </div>
      <div v-if="!loading" class="modal-body">
        <div class="form-group">
          <label
            >{{
              $t("Max number of tasks in parallel (0 for unlimited)")
            }}:</label
          >
          <input
            type="text"
            autocomplete="off"
            v-model.number="maxTasks"
            :disabled="busy"
            min="0"
            class="form-control form-control-full-width"
          />
        </div>
        <div class="form-group">
          <label
            >{{
              $t(
                "Max number threads for each task (0 to use the number of cores)"
              )
            }}:</label
          >
          <input
            type="text"
            autocomplete="off"
            v-model.number="encodingThreads"
            :disabled="busy"
            min="0"
            class="form-control form-control-full-width"
          />
        </div>
        <div class="form-group">
          <label
            >{{
              $t(
                "Extra resolutions for videos. These resolutions can be used for slow connections or small screens"
              )
            }}:</label
          >
          <div class="table-responsive">
            <table class="table">
              <thead>
                <tr>
                  <th class="text-left">{{ $t("Name") }}</th>
                  <th class="text-left">{{ $t("Properties") }}</th>
                  <th class="text-right">{{ $t("Enabled") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="res in resolutions" :key="res.name">
                  <td class="bold">{{ res.name }}</td>
                  <td>{{ res.width }}x{{ res.height }}, {{ res.fps }} fps</td>
                  <td class="text-right">
                    <ToggleSwitch v-model:val="res.enabled"></ToggleSwitch>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="form-group">
          <label
            >{{
              $t(
                "Extra resolutions for images. These resolutions can be used for slow connections or small screens"
              )
            }}:</label
          >
          <div class="table-responsive">
            <table class="table">
              <thead>
                <tr>
                  <th class="text-left">{{ $t("Name") }}</th>
                  <th class="text-left">{{ $t("Properties") }}</th>
                  <th class="text-right">{{ $t("Enabled") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="res in imageResolutions" :key="res.name">
                  <td class="bold">{{ res.name }}</td>
                  <td>{{ res.width }}x{{ res.height }}</td>
                  <td class="text-right">
                    <ToggleSwitch v-model:val="res.enabled"></ToggleSwitch>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer  no-padding">
        <button type="submit" class="modal-footer-btn">
          <i class="fas fa-check"></i> {{ $t("Save changes") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { ConfigAPI, VaultUserConfig } from "@/api/api-config";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  components: {
    ToggleSwitch,
  },
  name: "AdvancedSettingsModal",
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
      maxTasks: 0,
      encodingThreads: 0,
      resolutions: [],
      imageResolutions: [],

      standardResolutions: [
        {
          name: "144p",
          width: 256,
          height: 144,
          fps: 30,
        },
        {
          name: "240p",
          width: 352,
          height: 240,
          fps: 30,
        },
        {
          name: "360p",
          width: 480,
          height: 360,
          fps: 30,
        },
        {
          name: "480p",
          width: 858,
          height: 480,
          fps: 30,
        },
        {
          name: "720p",
          width: 1280,
          height: 720,
          fps: 30,
        },
        {
          name: "720p60",
          width: 1280,
          height: 720,
          fps: 60,
        },
        {
          name: "1080p",
          width: 1920,
          height: 1080,
          fps: 30,
        },
        {
          name: "1080p60",
          width: 1920,
          height: 1080,
          fps: 60,
        },
        {
          name: "2k",
          width: 2048,
          height: 1152,
          fps: 30,
        },
        {
          name: "2k60",
          width: 2048,
          height: 1152,
          fps: 60,
        },
        {
          name: "4k",
          width: 3860,
          height: 2160,
          fps: 30,
        },
        {
          name: "4k60",
          width: 3860,
          height: 2160,
          fps: 60,
        },
      ],

      loading: true,
      busy: false,
      error: "",
    };
  },
  methods: {
    autoFocus: function () {
      if (!this.display) {
        return;
      }
      nextTick(() => {
        const elem = this.$el.querySelector(".auto-focus");
        elem.focus();
      });
    },

    updateResolutions: function (resolutions, imageResolutions) {
      this.resolutions = this.standardResolutions.map((r) => {
        let enabled = false;
        for (let res of resolutions) {
          if (
            res.width === r.width &&
            res.height === r.height &&
            res.fps === r.fps
          ) {
            enabled = true;
            break;
          }
        }
        return {
          enabled: enabled,
          name: r.name,
          width: r.width,
          height: r.height,
          fps: r.fps,
        };
      });

      this.imageResolutions = this.standardResolutions
        .filter((r) => {
          return r.fps === 30;
        })
        .map((r) => {
          let enabled = false;
          for (let res of imageResolutions) {
            if (res.width === r.width && res.height === r.height) {
              enabled = true;
              break;
            }
          }
          return {
            enabled: enabled,
            name: r.name,
            width: r.width,
            height: r.height,
          };
        });
    },

    getResolutions: function () {
      return this.resolutions
        .filter((r) => {
          return r.enabled;
        })
        .map((r) => {
          return {
            width: r.width,
            height: r.height,
            fps: r.fps,
          };
        });
    },

    getImageResolutions: function () {
      return this.imageResolutions
        .filter((r) => {
          return r.enabled;
        })
        .map((r) => {
          return {
            width: r.width,
            height: r.height,
          };
        });
    },

    load: function () {
      Timeouts.Abort("advanced-settings");
      Request.Abort("advanced-settings");

      if (!this.display) {
        return;
      }

      this.loading = true;

      Request.Pending("advanced-settings", ConfigAPI.GetConfig())
        .onSuccess((response: VaultUserConfig) => {
          this.maxTasks = response.max_tasks;
          this.encodingThreads = response.encoding_threads;
          this.updateResolutions(
            response.resolutions,
            response.image_resolutions
          );
          this.loading = false;

          this.autoFocus();
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
              // Retry
              Timeouts.Set("advanced-settings", 1500, this.load.bind(this));
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("advanced-settings", 1500, this.load.bind(this));
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("advanced-settings", 1500, this.load.bind(this));
        });
    },

    submit: function (e) {
      e.preventDefault();

      if (this.busy) {
        return;
      }

      this.busy = true;
      this.error = "";

      Request.Do(
        ConfigAPI.SetConfig({
          max_tasks: this.maxTasks,
          encoding_threads: this.encodingThreads,
          resolutions: this.getResolutions(),
          image_resolutions: this.getImageResolutions(),
        })
      )
        .onSuccess(() => {
          this.busy = false;
          AppEvents.Emit("snack", this.$t("Vault configuration updated!"));
          this.close();
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              this.error = this.$t("Invalid configuration provided");
            })
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.busy = false;
        });
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
    this.load();
  },
  beforeUnmount: function () {
    Timeouts.Abort("advanced-settings");
    Request.Abort("advanced-settings");
    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        this.error = "";
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
  },
});
</script>
