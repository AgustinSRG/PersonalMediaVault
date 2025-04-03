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
                <div class="modal-header-corner-title">{{ $t("Choose your language") }}</div>
            </div>
            <div class="modal-body with-menu limited-height">
                <table class="modal-menu">
                    <tbody>
                        <tr
                            v-for="l in languages"
                            :key="l.id"
                            class="modal-menu-item"
                            tabindex="0"
                            @keydown="clickOnEnter"
                            @click="changeLocale(l.id)"
                        >
                            <td class="modal-menu-item-icon">
                                <i class="fas fa-check" :class="{ unchecked: lang !== l.id }"></i>
                            </td>
                            <td class="modal-menu-item-title">{{ l.name }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { getLanguage, setLanguage } from "@/i18n";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import { AVAILABLE_LANGUAGES } from "@/i18n";

export default defineComponent({
    name: "LanguageDropdown",
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
            languages: AVAILABLE_LANGUAGES.map((l) => {
                return {
                    id: l.id,
                    name: l.name,
                };
            }).sort((a, b) => {
                if (a.name < b.name) {
                    return -1;
                } else {
                    return 1;
                }
            }),
            lang: getLanguage(),
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
});
</script>
