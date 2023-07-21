<template>
  <div class="
        modal-container modal-container-corner modal-container-help
        no-transition
      " :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @mousedown="close" @touchstart="close" @keydown="keyDownHandle">
    <div v-if="display" class="modal-dialog modal-sm" role="document" @click="stopPropagationEvent" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent">
      <div class="modal-header-corner">
        <div class="modal-header-corner-title">{{ $t("Help") }}</div>
      </div>
      <div class="modal-body with-menu">
        <table class="modal-menu">
          <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('about')">
            <td class="modal-menu-item-icon"><i class="fas fa-info"></i></td>
            <td class="modal-menu-item-title">
              {{ $t("About PMV") }}
            </td>
          </tr>

          <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('keyboard')">
            <td class="modal-menu-item-icon">
              <i class="fas fa-keyboard"></i>
            </td>
            <td class="modal-menu-item-title">
              {{ $t("Keyboard shortcuts") }}
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "HelpHubDropdown",
    emits: ["update:display", "goto"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {};
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        clickOnOption: function (o: string) {
            this.$emit("goto", o);
            this.close();
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },

        keyDownHandle: function (e) {
            e.stopPropagation();
            if (e.key === "Escape") {
                this.close();
            }
        },
    },
    mounted: function () {
        this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "top-bar-button-dropdown");
        if (this.display) {
            this.$options.focusTrap.activate();
            nextTick(() => {
                this.$el.focus();
            });
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
            } else {
                if (this.$options.focusTrap) {
                    this.$options.focusTrap.deactivate();
                }
            }
        },
    },
});
</script>

