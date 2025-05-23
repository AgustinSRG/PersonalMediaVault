<template>
    <div
        class="modal-container modal-container-corner modal-container-help no-transition"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
        @mousedown="close"
        @touchstart="close"
        @keydown="keyDownHandle"
    >
        <div
            v-if="display"
            class="modal-dialog modal-sm"
            role="document"
            @click="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
        >
            <div class="modal-header-corner">
                <div class="modal-header-corner-title">{{ $t("Help") }}</div>
            </div>
            <div class="modal-body with-menu">
                <table class="modal-menu">
                    <tbody>
                        <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('about')">
                            <td class="modal-menu-item-icon"><i class="fas fa-question"></i></td>
                            <td class="modal-menu-item-title">
                                {{ $t("About PMV") }}
                            </td>
                        </tr>

                        <a class="tr-link modal-menu-item" :href="docsURL" target="_blank" rel="noopener noreferrer" @click="close">
                            <td class="modal-menu-item-icon"><i class="fas fa-book"></i></td>
                            <td class="modal-menu-item-title">
                                {{ $t("Documentation") }}
                            </td>
                        </a>

                        <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('keyboard')">
                            <td class="modal-menu-item-icon">
                                <i class="fas fa-keyboard"></i>
                            </td>
                            <td class="modal-menu-item-title">
                                {{ $t("Keyboard shortcuts") }}
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
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "HelpHubDropdown",
    props: {
        display: Boolean,
    },
    emits: ["update:display", "goto"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            docsURL: import.meta.env.VITE__DOCS_URL || "#",
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

        clickOnOption: function (o: string) {
            this.$emit("goto", o);
            this.close();
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
