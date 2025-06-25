<template>
    <div
        class="player-attachments-list"
        :class="{ hidden: !shown }"
        tabindex="-1"
        role="dialog"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @mouseenter="enterConfig"
        @mouseleave="leaveConfig"
        @keydown="keyDownHandle"
    >
        <a
            v-for="att in attachments || []"
            :key="att.id"
            class="player-attachment-link"
            tabindex="0"
            :href="getAttachmentUrl(att)"
            target="_blank"
            rel="noopener noreferrer"
            @click="clickAttachmentLink"
            @keydown="clickOnEnter"
        >
            <div class="attachment-icon-link">
                <i class="fas fa-paperclip"></i>
            </div>
            <div class="attachment-name">
                {{ att.name }}
            </div>
        </a>
    </div>
</template>

<script lang="ts">
import { getAutoNextTime, getImageNotesVisible } from "@/control/player-preferences";
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import type { MediaAttachment } from "@/api/models";
import { getAssetURL } from "@/utils/api";

export default defineComponent({
    name: "PlayerAttachmentsList",
    props: {
        shown: Boolean,
        attachments: Array as PropType<MediaAttachment[]>,
    },
    emits: ["update:shown", "enter", "leave"],
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
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        enterConfig: function () {
            this.$emit("enter");
        },

        leaveConfig: function () {
            this.$emit("leave");
        },

        focus: function () {
            nextTick(() => {
                this.$el.focus();
            });
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
});
</script>
