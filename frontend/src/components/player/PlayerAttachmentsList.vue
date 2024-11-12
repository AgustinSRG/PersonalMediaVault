<template>
    <div
        class="player-attachments-list"
        :class="{ hidden: !shown }"
        tabindex="-1"
        role="dialog"
        :aria-hidden="!shown"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @mouseenter="enterConfig"
        @mouseleave="leaveConfig"
        @keydown="keyDownHandle"
    >
        <table>
            <a
                v-for="att in attachments || []"
                :key="att.id"
                class="tr-button"
                tabindex="0"
                :href="getAttachmentUrl(att)"
                target="_blank"
                rel="noopener noreferrer"
                @click="clickAttachmentLink"
                @keydown="clickOnEnter"
            >
                <td>
                    <i class="fas fa-paperclip icon-config"></i>
                    <b>{{ att.name }}</b>
                </td>
            </a>
        </table>
    </div>
</template>

<script lang="ts">
import { getAutoNextTime, getImageNotesVisible } from "@/control/player-preferences";
import { defineComponent, nextTick, PropType } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import { MediaAttachment } from "@/api/models";
import { getAssetURL } from "@/utils/api";

export default defineComponent({
    name: "PlayerAttachmentsList",
    emits: ["update:shown", "enter", "leave"],
    props: {
        shown: Boolean,
        attachments: Array as PropType<MediaAttachment[]>,
    },
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            shownState: useVModel(props, "shown"),
        };
    },
    data: function () {
        return {
            autoNext: getAutoNextTime(),
            autoNextOptions: [0, 3, 5, 10, 15, 20, 25, 30],
            hideNotes: !getImageNotesVisible(),
        };
    },
    methods: {
        enterConfig: function () {
            this.$emit("enter");
        },

        leaveConfig: function () {
            this.$emit("leave");
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        focus: function () {
            nextTick(() => {
                this.$el.focus();
            });
        },

        clickOnEnter: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                (event.target as HTMLElement).click();
            }
        },

        clickAttachmentLink: function (event: Event) {
            event.stopPropagation();
            this.close();
        },

        getAttachmentUrl: function (att: MediaAttachment): string {
            return getAssetURL(att.url);
        },

        close: function () {
            this.shownState = false;
        },

        keyDownHandle: function (e: KeyboardEvent) {
            if (e.ctrlKey) {
                return;
            }
            if (e.key === "Escape") {
                this.close();
                e.stopPropagation();
            }
        },
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    watch: {
        shown: function () {
            if (this.shown) {
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
