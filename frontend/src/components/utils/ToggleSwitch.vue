<template>
    <div class="switch-button-container" tabindex="0" @keydown="onKeyDown">
        <div class="switch-button-control">
            <div class="switch-button" :class="{ enabled: val, disabled: disabled }" @click="toggle">
                <div class="button"></div>
            </div>
        </div>
        <div class="switch-button-label">
            <slot></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
// Properties
const props = defineProps({
    disabled: Boolean,
});

// Toggle switch value
const val = defineModel<boolean>("val");

/**
 * Toggles the value of the switch
 */
const toggle = () => {
    if (props.disabled) {
        return;
    }
    val.value = !val.value;
};

/**
 * Handler for the 'keydown' event
 */
const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === " " || e.key === "Enter") {
        toggle();
    }
    e.stopPropagation();
};
</script>
