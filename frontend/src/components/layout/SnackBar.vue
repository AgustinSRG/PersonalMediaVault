<template>
    <div class="snackbar" :class="{ hidden: !shown, center: position === 'center', right: position === 'right' }">
        <div class="snackbar-box" @mouseenter="hide" @click="hide">{{ message }}</div>
    </div>
</template>

<script setup lang="ts">
import { onApplicationEvent } from "@/composables/on-app-event";
import { useTimeout } from "@/composables/use-timeout";
import { EVENT_NAME_SNACK_BAR } from "@/control/app-events";
import type { SnackBarPosition } from "@/control/pages";
import { ref } from "vue";

// Timeout to hide
const timeout = useTimeout();

// Snackbar shown?
const shown = ref(false);

// Message
const message = ref("");

// Delay to hide the snackbar (milliseconds)
const HIDE_DELAY = 3000;

// Position
const position = ref<SnackBarPosition>("left");

/**
 * Shows the snackbar
 * @param msg The message
 * @param pos The position
 */
const show = (msg: string, pos?: SnackBarPosition) => {
    timeout.clear();

    shown.value = true;
    message.value = msg;
    position.value = pos || "left";

    timeout.set(() => {
        shown.value = false;
    }, HIDE_DELAY);
};

onApplicationEvent(EVENT_NAME_SNACK_BAR, show);

/**
 * Hides the snackbar
 */
const hide = () => {
    timeout.clear();

    shown.value = false;
};
</script>
