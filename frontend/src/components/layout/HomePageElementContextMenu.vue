<template>
    <div
        class="album-body-item-options-menu"
        :class="{
            hidden: !shown,
        }"
        :style="{
            top: top,
            left: left,
            right: right,
            bottom: bottom,
            width: width,
            'max-width': maxWidth,
            'max-height': maxHeight,
        }"
        tabindex="-1"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
    >
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="changePosition">
            <i class="fas fa-arrows-up-down-left-right"></i>
            {{ $t("Change position") }}
        </div>
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="removeElement">
            <i class="fas fa-trash-alt"></i> {{ $t("Remove from row") }}
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "HomePageElementContextMenu",
    props: {
        shown: Boolean,

        elementIndex: Number,

        x: Number,
        y: Number,
    },
    emits: ["update:shown", "change-pos", "element-remove"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            shownState: useVModel(props, "shown"),
        };
    },
    data: function () {
        return {
            top: "",
            left: "",
            right: "",
            bottom: "",

            width: "",

            maxWidth: "",
            maxHeight: "",
        };
    },
    watch: {
        x: function () {
            this.computeDimensions();
        },
        y: function () {
            this.computeDimensions();
        },
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
        this.computeDimensions();

        this.$listenOnDocumentEvent("mousedown", this.hide.bind(this));
        this.$listenOnDocumentEvent("touchstart", this.hide.bind(this));

        this.focusTrap = new FocusTrap(this.$el, this.hide.bind(this), "album-body-btn");
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        changePosition: function () {
            this.$emit("change-pos", this.elementIndex);
            this.hide();
        },

        removeElement: function () {
            this.$emit("element-remove", this.elementIndex);
            this.hide();
        },

        hide: function () {
            this.shownState = false;
        },

        computeDimensions: function () {
            const pageWidth = window.innerWidth;
            const pageHeight = window.innerHeight;

            const x = this.x;
            const y = this.y;

            if (y > pageHeight / 2) {
                const bottom = pageHeight - y;

                const maxHeight = pageHeight - bottom;

                this.top = "auto";
                this.bottom = bottom + "px";

                this.maxHeight = maxHeight + "px";
            } else {
                const top = y;

                const maxHeight = pageHeight - top;

                this.top = top + "px";
                this.bottom = "auto";

                this.maxHeight = maxHeight + "px";
            }

            if (x > pageWidth / 2) {
                const right = pageWidth - x;
                const maxWidth = pageWidth - right;

                this.left = "auto";
                this.right = right + "px";

                this.width = "auto";
                this.maxWidth = maxWidth + "px";
            } else {
                const maxWidth = pageWidth - x;

                this.left = x + "px";
                this.right = "auto";

                this.width = "auto";
                this.maxWidth = maxWidth + "px";
            }
        },
    },
});
</script>
