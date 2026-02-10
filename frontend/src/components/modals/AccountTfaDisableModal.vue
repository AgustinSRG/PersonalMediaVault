<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Disable two factor authentication") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body">
                <div class="tfa-label">{{ $t("Input your current one-time code for two factor authentication") }}</div>

                <SixDigitCodeInput v-model:val="code" :disabled="busy"></SixDigitCodeInput>

                <div class="form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button v-if="!busy && mustWait === 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait 1 second to try again") }}
                </button>
                <button v-else-if="!busy && mustWait > 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ stringMultiReplace($t("You must wait $TIME seconds to try again"), { $TIME: mustWait + "" }) }}
                </button>
                <button v-else type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Disable two factor authentication") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { makeApiRequest } from "@asanrom/request-browser";
import { onMounted, ref, useTemplateRef } from "vue";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import { apiAccountTfaDisable } from "@/api/api-account";
import { stringMultiReplace } from "@/utils/string-multi-replace";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useInterval } from "@/composables/use-interval";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose, focus } = useModal(display, container);

// Events
const emit = defineEmits<{
    /**
     * Emitted when the request is done successfully.
     */
    (e: "done"): void;
}>();

// The current timestamp
const now = ref(Date.now());

// Cooldown if the request fails
const cooldown = ref(0);

// Number of seconds the user must wait
const mustWait = ref(0);

// A timer to update the 'now' ref
const timer = useInterval();

// Interval delay (milliseconds)
const TIMER_INTERVAL_DELAY = 200;

onMounted(() => {
    timer.set(() => {
        now.value = Date.now();
        if (now.value < cooldown.value) {
            mustWait.value = Math.max(1, Math.round((cooldown.value - now.value) / 1000));
        } else {
            mustWait.value = 0;
        }
    }, TIMER_INTERVAL_DELAY);
});

// TFA code required to disable TFA
const code = ref("");

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Resets the error messages
const resetErrors = () => {
    error.value = "";
};

/**
 * Performs the request
 */
const performRequest = () => {
    if (busy.value) {
        return;
    }

    resetErrors();

    busy.value = true;

    makeApiRequest(apiAccountTfaDisable(code.value))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Two factor authentication disabled"));

            emit("done");

            forceClose();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                tfaNotEnabled: () => {
                    PagesController.ShowSnackBar($t("Two factor authentication is not enabled"));

                    emit("done");

                    forceClose();
                },
                invalidCode: () => {
                    error.value = $t("Invalid one-time code");
                    code.value = "";
                    cooldown.value = Date.now() + 5000;

                    focus();
                },
                accessDenied,
                cooldown: () => {
                    error.value = $t("You must wait 5 seconds to try again");
                },
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            error.value = err.message;

            console.error(err);
        });
};

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    performRequest();
};
</script>
