<template>
    <div ref="container" class="modal-container modal-container-login" tabindex="-1" role="dialog">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("The media vault is locked") }}
                </div>
            </div>
            <div class="modal-body">
                <div v-if="!tfaRequired" class="horizontal-filter-menu two-child no-border">
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Credentials')"
                        :class="{ selected: !isCode }"
                        @click="changeToCredentials"
                        ><i class="fas fa-key"></i> {{ $t("Credentials") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Invite code')"
                        :class="{ selected: isCode }"
                        @click="changeToCode"
                        ><i class="fas fa-user-check"></i> {{ $t("Invite code") }}</a
                    >
                </div>
                <div v-if="!tfaRequired && !isCode" class="div-pt">
                    <div class="form-group">
                        <label>{{ $t("Username") }}:</label>
                        <input
                            v-model="username"
                            type="text"
                            name="username"
                            :disabled="busy"
                            maxlength="255"
                            class="form-control form-control-full-width auto-focus auto-select"
                        />
                    </div>
                    <div class="form-group">
                        <label>{{ $t("Password") }}:</label>
                        <PasswordInput
                            v-model:val="password"
                            :name="'password'"
                            :disabled="busy"
                            @tab-skip="passwordTabSkip"
                        ></PasswordInput>
                    </div>
                    <div class="form-group">
                        <label>{{ $t("Session duration") }}:</label>
                        <select v-model="duration" name="session-duration" class="form-control form-control-full-width form-select">
                            <option :value="'day'">1 {{ $t("day") }}</option>
                            <option :value="'week'">1 {{ $t("week") }} | 7 {{ $t("days") }}</option>
                            <option :value="'month'">1 {{ $t("month") }} | 30 {{ $t("days") }}</option>
                            <option :value="'year'">1 {{ $t("year") }} | 365 {{ $t("days") }}</option>
                        </select>
                    </div>
                </div>
                <div v-else-if="!tfaRequired && isCode" class="div-pt">
                    <div class="invite-code-label">{{ $t("Input your invite code below") }}</div>
                    <SixDigitCodeInput v-model:val="code" :disabled="busy"></SixDigitCodeInput>
                </div>
                <div v-else-if="tfaRequired" class="div-pt">
                    <div class="tfa-label">{{ $t("Input your current one-time code for two factor authentication") }}</div>

                    <div class="tfa-cancel-label">
                        {{ $t("Don't have the code?") }}
                        <a href="javascript:;" @click="cancelTfa">{{ $t("Cancel and try with other method") }}</a>
                    </div>
                    <SixDigitCodeInput v-model:val="tfaCode" :disabled="busy"></SixDigitCodeInput>
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button v-if="!busy && mustWait <= 0" type="submit" class="modal-footer-btn">
                    <i class="fas fa-unlock"></i> {{ $t("Unlock vault") }}
                </button>
                <button v-if="!busy && mustWait === 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait 1 second to try again") }}
                </button>
                <button v-if="!busy && mustWait > 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ stringMultiReplace($t("You must wait $TIME seconds to try again"), { $TIME: mustWait + "" }) }}
                </button>
                <button v-if="busy" type="button" disabled class="modal-footer-btn">
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Unlocking vault") }}...
                </button>
            </div>
        </form>
    </div>
</template>

<script setup lang="ts">
import type { SessionDuration } from "@/api/api-auth";
import { apiAuthLogin } from "@/api/api-auth";
import { apiInvitesLogin } from "@/api/api-invites";
import { AuthController } from "@/control/auth";
import { makeApiRequest } from "@asanrom/request-browser";
import { nextTick, onMounted, ref, useTemplateRef } from "vue";
import PasswordInput from "@/components/utils/PasswordInput.vue";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import { stringMultiReplace } from "@/utils/string-multi-replace";
import { useInterval } from "@/composables/use-interval";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Username
const username = ref("");

// Password
const password = ref("");

// Session duration
const duration = ref<SessionDuration>("day");

// TFA required?
const tfaRequired = ref(false);

// TFA code
const tfaCode = ref("");

// Using invite code?
const isCode = ref(false);

// Invite code
const code = ref("");

// The current timestamp
const now = ref(Date.now());

// Cooldown (milliseconds)
const cooldown = ref(0);

// Number of seconds the user must wait
const mustWait = ref(0);

// Timer to update the wait time
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

/**
 * Changes mode to invite code
 */
const changeToCode = () => {
    isCode.value = true;
    autoFocus();
};

/**
 * Changes mode to credentials
 */
const changeToCredentials = () => {
    isCode.value = false;
    autoFocus();
};

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, serverError, networkError } = useCommonRequestErrors();

// Cooldown (milliseconds) for an invalid attempt
const INVALID_ATTEMPT_COOLDOWN = 5000;

/**
 * Login with credentials (username + password + tfa (optional))
 */
const loginWithCredentials = () => {
    if (busy.value) {
        return;
    }

    busy.value = true;
    error.value = "";

    makeApiRequest(apiAuthLogin(username.value, password.value, duration.value, tfaRequired.value ? tfaCode.value : undefined))
        .onSuccess((response) => {
            busy.value = false;

            username.value = "";
            password.value = "";
            tfaCode.value = "";

            AuthController.SetSession(response.session_id, response.vault_fingerprint);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                invalidCredentials: () => {
                    error.value = $t("Invalid username or password");
                    tfaRequired.value = false;
                    autoFocus();
                },
                tfaRequired: () => {
                    tfaRequired.value = true;
                    tfaCode.value = "";
                    autoFocus();
                },
                invalidTfaCode: () => {
                    error.value = $t("Invalid one-time code");
                    tfaCode.value = "";
                    cooldown.value = Date.now() + INVALID_ATTEMPT_COOLDOWN;
                    autoFocus();
                },
                wrongCredentials: () => {
                    error.value = $t("Invalid username or password");
                    cooldown.value = Date.now() + INVALID_ATTEMPT_COOLDOWN;
                    tfaRequired.value = false;
                    autoFocus();
                },
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
 * Log with invite code
 */
const loginWithCode = () => {
    if (busy.value) {
        return;
    }

    busy.value = true;
    error.value = "";

    makeApiRequest(apiInvitesLogin(code.value))
        .onSuccess((response) => {
            busy.value = false;

            code.value = "";

            AuthController.SetSession(response.session_id, response.vault_fingerprint);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                invalidCode: () => {
                    error.value = $t("Invalid invite code");
                    autoFocus();
                },
                wrongCode: () => {
                    error.value = $t("Invalid invite code");
                    cooldown.value = Date.now() + INVALID_ATTEMPT_COOLDOWN;
                    autoFocus();
                },
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
 * Handler for the 'submit' event
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    if (isCode.value) {
        loginWithCode();
    } else {
        loginWithCredentials();
    }
};

/**
 * Cancels TFA attempt
 */
const cancelTfa = () => {
    if (busy.value) {
        return;
    }

    tfaRequired.value = false;
};

// Container element
const container = useTemplateRef("container");

/**
 * Focuses the element
 */
const focus = () => {
    const focusElement = container.value?.querySelector(`.auto-focus`) as HTMLElement;
    if (focusElement) {
        focusElement.focus();

        if (focusElement.classList.contains("auto-select") && typeof (focusElement as HTMLInputElement).select === "function") {
            (focusElement as HTMLInputElement).select();
        }
    } else {
        container.value?.focus();
    }
};

/**
 * Focuses the element on the next tick
 */
const autoFocus = () => {
    nextTick(focus);
};

onMounted(autoFocus);

/**
 * Handler for tab focus skip on password input
 * @param e The keyboard event
 */
const passwordTabSkip = (e: KeyboardEvent) => {
    const nextElement = container.value?.querySelector('[name="session-duration"]') as HTMLElement;

    if (nextElement) {
        e.preventDefault();
        nextElement.focus();
    }
};
</script>
