<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
    @keydown="keyDownHandle"
  >
    <div
      class="modal-dialog modal-lg"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Keyboard shortcuts") }}
        </div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body no-padding table-responsive">
        <table class="table table-text-overflow keyboard-help-table">
          <thead>
            <tr>
              <th class="text-left">{{ $t("Action") }}</th>
              <th class="text-left">{{ $t("Shortcut") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{{ $t("Open the menu") }}</td>
              <td><span class="key-shortcut">⌘ + M</span></td>
            </tr>
            <tr>
              <td>{{ $t("Search") }}</td>
              <td><span class="key-shortcut">⌘ + F</span></td>
            </tr>
            <tr>
              <td>{{ $t("Open vault settings") }}</td>
              <td><span class="key-shortcut">⌘ + S</span></td>
            </tr>
            <tr>
              <td>{{ $t("Close vault") }}</td>
              <td><span class="key-shortcut">⌘ + Q</span></td>
            </tr>
            <tr>
              <td>{{ $t("Close current modal or menu") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Escape") }}</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Play") }} / {{ $t("Pause") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Space") }}</span>
                {{ $t("or") }}
                <span class="key-shortcut">K</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Fast forward 5 seconds") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Right arrow") }}</span>
              </td>
            </tr>
            <tr>
              <td>
                {{ $t("Fast forward a single frame (only when paused)") }}
              </td>
              <td>
                <span class="key-shortcut">,</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Rewind 5 seconds") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Left arrow") }}</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Rewind a single frame (only when paused)") }}</td>
              <td>
                <span class="key-shortcut">.</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Go to the beginning of the video / audio") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Home") }}</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Go to the end of the video / audio") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("End") }}</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Increase volume by 5%") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Up arrow") }}</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Decrease volume by 5%") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Down arrow") }}</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Mute") }} / {{ $t("Unmute") }}</td>
              <td>
                <span class="key-shortcut">M</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Toggle loop") }}</td>
              <td>
                <span class="key-shortcut">L</span>
              </td>
            </tr>
            <tr>
              <td>{{ $t("Toggle full screen") }}</td>
              <td>
                <span class="key-shortcut">F</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Increment image scale by 10%") }}</td>
              <td>
                <span class="key-shortcut">⌘ {{ $t("and") }} +</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Decrement image scale by 10%") }}</td>
              <td>
                <span class="key-shortcut">⌘ {{ $t("and") }} -</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Fit image into viewport size") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Space") }}</span>
                {{ $t("or") }}
                <span class="key-shortcut">K</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Hide or show image player controls") }}</td>
              <td>
                <span class="key-shortcut">C</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Scroll image") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Up arrow") }}</span>
                {{ $t("or") }}
                <span class="key-shortcut">{{ $t("Down arrow") }}</span>
                {{ $t("or") }}
                <span class="key-shortcut">{{ $t("Left arrow") }}</span>
                {{ $t("or") }}
                <span class="key-shortcut">{{ $t("Right arrow") }}</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Scroll image to the top") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Home") }}</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Scroll image to the bottom") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("End") }}</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Scroll image to the left") }}</td>
              <td>
                <span class="key-shortcut"
                  >{{ $t("Shift") }} + {{ $t("Home") }}</span
                >
              </td>
            </tr>

            <tr>
              <td>{{ $t("Scroll image to the right") }}</td>
              <td>
                <span class="key-shortcut"
                  >{{ $t("Shift") }} + {{ $t("End") }}</span
                >
              </td>
            </tr>

            <tr>
              <td>{{ $t("Show player configuration") }}</td>
              <td>
                <span class="key-shortcut">S</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Open media information") }}</td>
              <td>
                <span class="key-shortcut">E</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Add current media to an album") }}</td>
              <td>
                <span class="key-shortcut">A</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Go to next") }}</td>
              <td>
                <span class="key-shortcut"
                  >{{ $t("Shift") }} + {{ $t("Right arrow") }}</span
                >
              </td>
            </tr>

            <tr>
              <td>{{ $t("Go to previous") }}</td>
              <td>
                <span class="key-shortcut"
                  >{{ $t("Shift") }} + {{ $t("Left arrow") }}</span
                >
              </td>
            </tr>

            <tr>
              <td>{{ $t("Toggle album loop") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Shift") }} + L</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Toggle album shuffle") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Shift") }} + R</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Add or remove album to favorites") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Shift") }} + F</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Create new album (while in album list)") }}</td>
              <td>
                <span class="key-shortcut">+</span>
              </td>
            </tr>

            <tr>
              <td>
                {{ $t("Close page at the right (in the split view mode)") }}
              </td>
              <td>
                <span class="key-shortcut">Q</span>
              </td>
            </tr>

            <tr>
              <td>{{ $t("Close player (in the split view mode)") }}</td>
              <td>
                <span class="key-shortcut">{{ $t("Backspace") }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "KeyboardGuideModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
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

<style>
.key-shortcut {
  font-weight: bold;
  background: var(--theme-bg-color);
  padding: 0.35rem;
}

.keyboard-help-table td {
  white-space: nowrap;
}
</style>
