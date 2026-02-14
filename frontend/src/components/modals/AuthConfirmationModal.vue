<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <form ref="form" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Authentication confirmation") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body">
                <div v-if="tfa" class="tfa-label">{{ $t("Input your current one-time code for two factor authentication") }}</div>
                <SixDigitCodeInput v-if="tfa" v-model:val="code"></SixDigitCodeInput>

                <div v-if="!tfa" class="form-group">
                    <label>{{ $t("To confirm the operation, type your account password") }}:</label>
                    <input type="text" class="hidden-input" name="username" :value="username" />
                    <PasswordInput v-model:val="password" :name="'password'" :auto-focus="true" @tab-skip="passwordTabSkip"></PasswordInput>
                </div>

                <div v-if="error" class="form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button v-if="mustWait === 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait 1 second to try again") }}
                </button>
                <button v-else-if="mustWait > 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ stringMultiReplace($t("You must wait $TIME seconds to try again"), { $TIME: mustWait + "" }) }}
                </button>
                <button v-else type="submit" class="modal-footer-btn" :disabled="tfa ? !code : !password">
                    <i class="fas fa-check"></i> {{ $t("Confirm") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { onMounted, ref, useTemplateRef } from "vue";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import PasswordInput from "../utils/PasswordInput.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { AuthController } from "@/control/auth";
import { stringMultiReplace } from "@/utils/string-multi-replace";
import { EVENT_NAME_AUTH_CHANGED } from "@/control/app-events";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useInterval } from "@/composables/use-interval";
import { onApplicationEvent } from "@/composables/on-app-event";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * Error message to display
     */
    error: String,

    /**
     * True if TFA is required
     */
    tfa: Boolean,

    /**
     * Cooldown (milliseconds)
     */
    cooldown: Number,
});

// Events
const emit = defineEmits<{
    /**
     * Confirmation event
     */
    (e: "confirm", confirmation: ProvidedAuthConfirmation): void;
}>();

// The current timestamp
const now = ref(Date.now());

// Number of seconds the user must wait
const mustWait = ref(0);

// A timer to update the 'now' ref
const timer = useInterval();

// Interval delay (milliseconds)
const TIMER_INTERVAL_DELAY = 200;

onMounted(() => {
    timer.set(() => {
        now.value = Date.now();
        if (now.value < props.cooldown) {
            mustWait.value = Math.max(1, Math.round((props.cooldown - now.value) / 1000));
        } else {
            mustWait.value = 0;
        }
    }, TIMER_INTERVAL_DELAY);
});

// Username
const username = ref(AuthController.Username);

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    username.value = AuthController.Username;
});

// TFA code
const code = ref("");

// Password
const password = ref("");

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    if (e) {
        e.preventDefault();
    }

    const providedAuthConfirmation: ProvidedAuthConfirmation = {
        password: password.value,
        tfaCode: code.value,
    };

    emit("confirm", providedAuthConfirmation);

    display.value = false;
};

// Form container
const form = useTemplateRef("form");

/**
 * Handler for tab focus skip on password input
 * @param e The keyboard event
 */
const passwordTabSkip = (e: KeyboardEvent) => {
    const nextElement = form.value?.querySelector(".modal-footer-btn") as HTMLElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();
    }
};
</script>
