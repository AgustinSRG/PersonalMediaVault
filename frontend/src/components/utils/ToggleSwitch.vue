<template>
    <div class="switch-button-container" tabindex="0" disabled="disabled" @keydown="keyToggle">
        <div class="switch-button-control">
            <div class="switch-button" :class="{ enabled: val }" @click="toggle">
                <div class="button"></div>
            </div>
        </div>
        <div class="switch-button-label">
            <slot></slot>
        </div>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "ToggleSwitch",
    emits: ["update:val"],
    props: {
        val: Boolean,
        disabled: Boolean,
    },
    setup(props) {
        return {
            valState: useVModel(props, "val"),
        };
    },
    data: function () {
        return {};
    },
    methods: {
        toggle: function () {
            if (this.disabled) {
                return;
            }
            this.valState = !this.valState;
        },
        keyToggle: function (e) {
            if (e.key === " " || e.key === "Enter") {
                this.toggle();
            }
            e.stopPropagation();
        },
    },
    mounted: function () {},
    beforeUnmount: function () {},
});
</script>
