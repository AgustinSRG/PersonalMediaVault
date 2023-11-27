<template>
    <div
        class="modal-container modal-container-dialog"
        :class="{ hidden: !display, closing: closing }"
        tabindex="-1"
        role="dialog"
        :aria-hidden="!display"
        @keydown="keyDownHandle"
        @animationend="onAnimationEnd"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @click="stopPropagationEvent"
    >
        <div class="modal-out-close-area" @mousedown="onMouseDown" @touchstart="onTouchOutside"></div>
        <slot @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent" @click="stopPropagationEvent"></slot>
    </div>
</template>

<script lang="ts">
import { FocusTrap } from "@/utils/focus-trap";
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

export default defineComponent({
    name: "ModalDialogContainer",
    emits: ["update:display", "key", "close"],
    props: {
        display: Boolean,
        lockClose: Boolean,
        static: Boolean,
        closeSignal: Number,
        forceCloseSignal: Number,
        closeCallback: Function,
    },
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            closing: false,
        };
    },
    methods: {
        close: function (forced?: boolean) {
            if (this.lockClose && forced !== true) {
                return;
            }
            if (this.closeCallback && forced !== true) {
                this.closeCallback(() => {
                    this.$emit("close");
                    this.closing = true;
                });
            } else {
                this.$emit("close");
                this.closing = true;
            }
        },

        onAnimationEnd: function (e: AnimationEvent) {
            e.stopPropagation();
            if (e.animationName === "modal-close-animation") {
                this.closing = false;
                this.displayStatus = false;
            }
        },

        keyDownHandle: function (e) {
            e.stopPropagation();
            if (e.key === "Escape" && this.display && !this.closing) {
                this.close();
            } else {
                this.$emit("key", e);
            }
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        onTouchOutside: function (e: TouchEvent) {
            e.stopPropagation();
            if (!this.static) {
                this.close();
            }
        },

        onMouseDown: function (e: MouseEvent) {
            e.stopPropagation();
            if (e.button === 0 && !this.static) {
                this.close();
            }
        },

        focusLost: function () {
            if (this.display) {
                this.$el.focus();
            }
        },
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.focusLost.bind(this));

        if (this.display) {
            this.focusTrap.activate();
        }
    },
    beforeUnmount: function () {
        if (this.focusTrap) {
            this.focusTrap.destroy();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                this.closing = false;
                if (this.focusTrap) {
                    this.focusTrap.activate();
                }
            } else {
                if (this.focusTrap) {
                    this.focusTrap.deactivate();
                }
            }
        },
        closeSignal: function () {
            if (this.closeSignal > 0) {
                this.close();
            }
        },
        forceCloseSignal: function () {
            if (this.forceCloseSignal > 0) {
                this.close(true);
            }
        },
    },
});
</script>
