<template>
    <div
        class="modal-container modal-container-corner no-transition"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
        :aria-hidden="!display"
        @mousedown="close"
        @touchstart="close"
        @keydown="keyDownHandle"
    >
        <div
            v-if="display"
            class="modal-dialog modal-md"
            role="document"
            @click="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
        >
            <div class="modal-header-corner">
                <div class="modal-header-corner-title">{{ $t("Select a theme for the app") }}</div>
            </div>
            <div class="modal-body with-menu limited-height">
                <table class="modal-menu">
                    <tr class="modal-menu-item" tabindex="0" @click="changeTheme('dark')" @keydown="clickOnEnter">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: theme !== 'dark' }"></i>
                        </td>
                        <td class="modal-menu-item-title">{{ $t("Dark Theme") }}</td>
                    </tr>
                    <tr class="modal-menu-item" tabindex="0" @click="changeTheme('light')" @keydown="clickOnEnter">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: theme !== 'light' }"></i>
                        </td>
                        <td class="modal-menu-item-title">{{ $t("Light Theme") }}</td>
                    </tr>
                </table>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "ThemeDropdown",
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
            theme: AppPreferences.Theme,
        };
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        changeTheme: function (t: string) {
            AppPreferences.SetTheme(t);
        },

        themeUpdated: function () {
            this.theme = AppPreferences.Theme;
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
        this.$options.themeHandler = this.themeUpdated.bind(this);
        AppEvents.AddEventListener("theme-changed", this.$options.themeHandler);
        this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "top-bar-button-dropdown");

        if (this.display) {
            this.$options.focusTrap.activate();
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("theme-changed", this.$options.themeHandler);
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
