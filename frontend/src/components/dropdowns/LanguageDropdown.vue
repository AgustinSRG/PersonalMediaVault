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
                <div class="modal-header-corner-title">{{ $t("Choose your language") }}</div>
            </div>
            <div class="modal-body with-menu limited-height">
                <table class="modal-menu">
                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="changeLocale('en')">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: lang !== 'en' }"></i>
                        </td>
                        <td class="modal-menu-item-title">English ({{ $t("Default") }})</td>
                    </tr>
                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="changeLocale('es')">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: lang !== 'es' }"></i>
                        </td>
                        <td class="modal-menu-item-title">Español (Internacional)</td>
                    </tr>
                </table>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { getLanguage, setLanguage } from "@/control/app-preferences";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "LanguageDropdown",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            lang: getLanguage(),
        };
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        changeLocale: function (l: string) {
            this.lang = l;
            setLanguage(l);
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
});
</script>
