<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-lg" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Account security settings") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-else class="modal-body">
                <div class="form-group">
                    <label
                        >{{ $t("Two factor authentication") }}:
                        <span class="tfa-status" :class="{ enabled: tfa }">{{
                            tfa ? $t("Enabled") + " (" + tfaMethod + ")" : $t("Disabled")
                        }}</span></label
                    >
                </div>

                <div class="form-group">
                    <button v-if="tfa" type="button" class="btn btn-danger btn-sm auto-focus" @click="disableTfa">
                        <i class="fas fa-trash-alt"></i> {{ $t("Disable two factor authentication") }}
                    </button>
                    <button v-else type="button" class="btn btn-primary btn-sm auto-focus" @click="enableTfa">
                        <i class="fas fa-check"></i> {{ $t("Enable two factor authentication") }}
                    </button>
                </div>

                <div class="form-group">
                    <table class="table no-border">
                        <tbody>
                            <tr>
                                <td class="text-right td-shrink no-padding">
                                    <ToggleSwitch v-model:val="authConfirmation" @update:val="markDirty"></ToggleSwitch>
                                </td>
                                <td>
                                    {{
                                        $t(
                                            "Enable auth confirmation. (when enabled, you will be asked for a confirmation when performing dangerous actions)",
                                        )
                                    }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="form-group">
                    <table class="table no-border">
                        <tbody>
                            <tr>
                                <td class="text-right td-shrink no-padding">
                                    <ToggleSwitch v-model:val="authConfirmationPreferTfa" @update:val="markDirty"></ToggleSwitch>
                                </td>
                                <td>
                                    {{ $t("Prefer two factor authentication (if enabled) for auth confirmation.") }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="form-group">
                    <label
                        >{{
                            $t(
                                "Auth confirmation period, in seconds. (After a successful auth confirmation, you won't be asked for it again for a period of time)",
                            )
                        }}:</label
                    >
                    <input
                        v-model.number="authConfirmationPeriodSeconds"
                        type="number"
                        autocomplete="off"
                        :disabled="busy"
                        min="0"
                        max="2147483647"
                        class="form-control form-control-full-width"
                        @change="markDirty"
                    />
                </div>

                <div class="form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy || !dirty">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Save changes") }}
                </button>
            </div>
        </form>

        <AccountTfaEnableModal v-if="displayTfaEnableModal" v-model:display="displayTfaEnableModal" @done="load"></AccountTfaEnableModal>
        <AccountTfaDisableModal
            v-if="displayTfaDisableModal"
            v-model:display="displayTfaDisableModal"
            @done="load"
        ></AccountTfaDisableModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="performRequest"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, onMounted, ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { apiAccountGetSecuritySettings, apiAccountSetSecuritySettings } from "@/api/api-account";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import { useModal } from "@/composables/use-modal";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";

const AccountTfaEnableModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountTfaEnableModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountTfaDisableModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountTfaDisableModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose, focus } = useModal(display, container);

// User has TFA enabled?
const tfa = ref(false);

// TFA method
const tfaMethod = ref("TOTP");

// User has auth confirmation enabled?
const authConfirmation = ref(false);

// Does the user prefer TFA over password for auth confirmation?
const authConfirmationPreferTfa = ref(true);

// Auth confirmation period (seconds)
const authConfirmationPeriodSeconds = ref(120);

// Loading status
const loading = ref(true);

// True if changes were made
const dirty = ref(false);

/**
 * Indicates changes were made
 * by the user
 */
const markDirty = () => {
    dirty.value = true;
};

// Load request ID
const loadRequestId = useRequestId();

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the data
 */
const load = () => {
    loading.value = true;

    clearNamedTimeout(loadRequestId);

    makeNamedApiRequest(loadRequestId, apiAccountGetSecuritySettings())
        .onSuccess((response) => {
            tfa.value = response.tfa;
            tfaMethod.value = ((response.tfaMethod || "").split(":")[0] || "").toUpperCase();

            authConfirmation.value = response.authConfirmation;
            authConfirmationPreferTfa.value = response.authConfirmationMethod !== "pw";
            authConfirmationPeriodSeconds.value = response.authConfirmationPeriodSeconds || 0;

            loading.value = false;

            dirty.value = false;

            error.value = "";

            focus();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    close();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
        });
};

onMounted(() => {
    if (display.value) {
        load();
    }
});

// Display the TFA enable modal
const displayTfaEnableModal = ref(false);

// Display the TFA disable modal
const displayTfaDisableModal = ref(false);

watch(display, () => {
    if (display.value) {
        displayTfaEnableModal.value = false;
        displayTfaDisableModal.value = false;

        load();
    }
});

/**
 * Displays the modal to enable TFA
 */
const enableTfa = () => {
    displayTfaEnableModal.value = true;
};

/**
 * Displays the modal to disable tfa
 */
const disableTfa = () => {
    displayTfaDisableModal.value = true;
};

// Auth confirmation
const {
    displayAuthConfirmation,
    authConfirmationCooldown,
    authConfirmationTfa,
    authConfirmationError,
    requiredAuthConfirmationPassword,
    invalidPassword,
    requiredAuthConfirmationTfa,
    invalidTfaCode,
    cooldown,
} = useAuthConfirmation();

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
 * @param confirmation The auth confirmation
 */
const performRequest = (confirmation: ProvidedAuthConfirmation) => {
    if (busy.value) {
        return;
    }

    resetErrors();

    busy.value = true;

    makeApiRequest(
        apiAccountSetSecuritySettings(
            authConfirmation.value,
            authConfirmationPreferTfa.value ? "tfa" : "pw",
            authConfirmationPeriodSeconds.value,
            confirmation,
        ),
    )
        .onSuccess(() => {
            busy.value = false;

            dirty.value = false;

            PagesController.ShowSnackBar($t("Saved security settings"));

            forceClose();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidSettings: () => {
                    error.value = $t("Invalid security settings");
                },
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied,
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

    performRequest({});
};
</script>
