<template>
    <div
        class="modal-container modal-container-corner no-transition"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
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
                    <tbody>
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
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { ColorThemeName, EVENT_NAME_THEME_CHANGED, getTheme, setTheme } from "@/control/app-preferences";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "ThemeDropdown",
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            theme: getTheme(),
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_THEME_CHANGED, this.themeUpdated.bind(this));
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "top-bar-button-dropdown");

        if (this.display) {
            this.focusTrap.activate();
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        changeTheme: function (t: ColorThemeName) {
            setTheme(t);
        },

        themeUpdated: function (theme: ColorThemeName) {
            this.theme = theme;
        },

        keyDownHandle: function (e: KeyboardEvent) {
            e.stopPropagation();
            if (e.key === "Escape") {
                this.close();
            }
        },
    },
});
</script>
