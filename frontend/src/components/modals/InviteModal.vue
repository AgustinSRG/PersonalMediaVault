<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <div class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Invite") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body">
                <div v-if="hasCode" class="invite-code-modal-body">
                    <div class="invite-code-display">
                        <div v-for="(c, i) in code.split('')" :key="i" class="invite-code-char-container">
                            <div class="invite-code-char">{{ c }}</div>
                        </div>
                    </div>
                    <div class="invite-code-info">
                        {{ $t("Use this code to log into the vault in read-only mode") }}
                    </div>
                    <div class="invite-code-info">
                        <label>{{ $t("Session duration") }}:</label> {{ renderedDurationDays }}
                    </div>
                    <div v-if="expirationRemaining > 0" class="invite-code-warning">
                        {{ $t("Remaining time until expiration") }}:
                        {{ renderedExpirationRemaining }}
                    </div>
                    <div v-if="expirationRemaining <= 0" class="invite-code-error">
                        {{ $t("The invite code has expired") }}
                    </div>
                    <div class="invite-code-btn-container">
                        <button type="button" :disabled="busy" class="btn btn-danger" @click="clearCode">
                            <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Clear invite code") }}
                        </button>
                    </div>
                </div>
                <div v-else class="invite-code-modal-body">
                    <div class="invite-code-info">{{ $t("You can invite other users or devices by creating an invite code") }}</div>
                    <div class="invite-code-info">
                        {{ $t("Invite codes are single use, and allow to login into the vault in read-only mode") }}
                    </div>
                    <div class="invite-code-btn-container">
                        <button type="button" :disabled="busy" class="btn btn-primary" @click="generateCode">
                            <LoadingIcon icon="fas fa-plus" :loading="busy"></LoadingIcon> {{ $t("Create invite code") }}
                        </button>
                    </div>
                    <div class="invite-code-select-container">
                        <label>{{ $t("Session duration") }}:</label>
                        <select v-model="durationGenerate" class="form-control form-select">
                            <option :value="'day'">1 {{ $t("day") }}</option>
                            <option :value="'week'">1 {{ $t("week") }} | 7 {{ $t("days") }}</option>
                            <option :value="'month'">1 {{ $t("month") }} | 30 {{ $t("days") }}</option>
                            <option :value="'year'">1 {{ $t("year") }} | 365 {{ $t("days") }}</option>
                        </select>
                    </div>
                </div>
            </div>
            <div v-if="loadingSessions" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loadingSessions" class="modal-body no-padding border-top">
                <div class="table-responsive">
                    <table class="table">
                        <thead>
                            <tr>
                                <th colspan="2" class="text-center">{{ $t("List of invited sessions") }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-if="sessions.length === 0">
                                <td colspan="2">
                                    {{ $t("There are no active invited sessions.") }}
                                </td>
                            </tr>
                            <tr v-for="s in sessions" :key="s.index">
                                <td>
                                    <b>{{ $t("Created") }}:</b> {{ renderDateAndTime(s.timestamp, locale) }}, <b>{{ $t("Expires") }}:</b>
                                    {{ renderDateAndTime(s.expiration, locale) }}
                                </td>
                                <td class="text-right td-shrink">
                                    <button
                                        type="button"
                                        :disabled="busyClosing"
                                        class="btn btn-danger btn-xs"
                                        @click="askCloseSession(s.index)"
                                    >
                                        <LoadingIcon
                                            icon="fas fa-trash-alt"
                                            :loading="busyClosing && sessionToClose === s.index"
                                        ></LoadingIcon>
                                        {{ $t("Close") }}
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <InviteCloseSessionConfirmationModal
                v-if="displayCloseConfirmationModal"
                v-model:display="displayCloseConfirmationModal"
                @confirm="closeSession"
            ></InviteCloseSessionConfirmationModal>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { computed, onMounted, ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import type { InviteSession } from "@/api/api-invites";
import {
    apiInvitesClearCode,
    apiInvitesDeleteSession,
    apiInvitesGenerateCode,
    apiInvitesGetSessions,
    apiInvitesGetStatus,
} from "@/api/api-invites";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { renderDateAndTime, renderTimeSeconds } from "@/utils/time";
import InviteCloseSessionConfirmationModal from "@/components/modals/InviteCloseSessionConfirmationModal.vue";
import type { SessionDuration } from "@/api/api-auth";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";
import { useInterval } from "@/composables/use-interval";
import { onApplicationEvent } from "@/composables/on-app-event";

// Translation function
const { $t, locale } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose } = useModal(display, container);

// True if a code exists
const hasCode = ref(false);

// The code
const code = ref("");

// Session duration (milliseconds)
const duration = ref(0);

// Rendered session duration (days)
const renderedDurationDays = computed(() => {
    const days = Math.max(1, Math.round(duration.value / (24 * 60 * 60 * 1000)));

    if (days === 1) {
        return "1" + " " + $t("day");
    } else {
        return days + " " + $t("days");
    }
});

// Remaining time until expiration (milliseconds)
const expirationRemaining = ref(0);

// Last time the status was updated
const lastReceivedTimestamp = ref(0);

// Current timestamp
const now = ref(Date.now());

// Interval to update the current timestamp (milliseconds)
const NOW_UPDATE_INTERVAL = 500;

// Timer to update the current time
const updateNowTimer = useInterval();

updateNowTimer.set(() => {
    now.value = Date.now();
}, NOW_UPDATE_INTERVAL);

// Rendered remaining time until expiration, updated based on the current timestamp
const renderedExpirationRemaining = computed(() =>
    renderTimeSeconds(Math.round(Math.max(0, expirationRemaining.value - (now.value - lastReceivedTimestamp.value)) / 1000)),
);

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the data
 */
const load = () => {
    if (!display.value) {
        return;
    }

    loading.value = true;
    loadingSessions.value = true;

    loadStatus();
    loadSessions();
};

onMounted(load);
onApplicationEvent(EVENT_NAME_AUTH_CHANGED, load);
watch(display, load);

// Loading status
const loading = ref(true);

// Loading status after the first time the status is loaded
// Silent: No loader
const loadingSilent = ref(true);

// Request ID for load request
const loadStatusRequestId = useRequestId();

/**
 * Loads the invite code status
 */
const loadStatus = () => {
    clearNamedTimeout(loadStatusRequestId);

    loadingSilent.value = true;

    const firstTime = loading.value;

    makeNamedApiRequest(loadStatusRequestId, apiInvitesGetStatus())
        .onSuccess((response) => {
            hasCode.value = response.has_code;
            code.value = response.code;
            expirationRemaining.value = response.expiration_remaining;
            lastReceivedTimestamp.value = Date.now();
            now.value = Date.now();
            duration.value = response.duration;

            loading.value = false;
            loadingSilent.value = false;

            if (firstTime) {
                focus();
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    forceClose();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadStatusRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadStatusRequestId, LOAD_RETRY_DELAY, load);
        });
};

// Loading status for sessions
const loadingSessions = ref(true);

// Loading status for sessions (after first time)
// Silent: No loader
const loadingSessionsSilent = ref(true);

// List of active invited sessions
const sessions = ref<InviteSession[]>([]);

// Request ID for session loading
const loadSessionsRequestId = useRequestId();

/**
 * Loads the list of active sessions
 */
const loadSessions = () => {
    clearNamedTimeout(loadSessionsRequestId);

    loadingSessionsSilent.value = true;

    makeNamedApiRequest(loadSessionsRequestId, apiInvitesGetSessions())
        .onSuccess((response) => {
            sessions.value = response;

            loadingSessions.value = false;
            loadingSessionsSilent.value = false;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    forceClose();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadSessionsRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadSessionsRequestId, LOAD_RETRY_DELAY, load);
        });
};

// Interval to periodically update the status
const STATUS_UPDATE_INTERVAL = 2000;

// Timer to update the status
const updateStatusTimer = useInterval();

updateStatusTimer.set(() => {
    if (!display.value) {
        return;
    }

    if (!loadingSilent.value) {
        loadStatus();
    }

    if (!loadingSessionsSilent.value) {
        loadSessions();
    }
}, STATUS_UPDATE_INTERVAL);

// The duration for the next generated code
const durationGenerate = ref<SessionDuration>("day");

// Busy status
const busy = ref(false);

/**
 * Generates invite code
 */
const generateCode = () => {
    if (busy.value) {
        return;
    }

    busy.value = true;

    makeApiRequest(apiInvitesGenerateCode(durationGenerate.value))
        .onSuccess((response) => {
            busy.value = false;

            hasCode.value = response.has_code;
            code.value = response.code;
            expirationRemaining.value = response.expiration_remaining;
            lastReceivedTimestamp.value = Date.now();
            now.value = Date.now();
            duration.value = response.duration;

            // Abort the load status request to avoid
            // an old request re-updating with old data
            clearNamedTimeout(loadStatusRequestId);
            abortNamedApiRequest(loadStatusRequestId);
            loadingSilent.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                limitReached: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("You reached the limit of invited sessions you can have"));
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            console.error(err);

            PagesController.ShowSnackBar($t("Error") + ": " + err.message);
        });
};

/**
 * Clears the current invite code
 */
const clearCode = () => {
    if (busy.value) {
        return;
    }

    busy.value = true;

    makeApiRequest(apiInvitesClearCode())
        .onSuccess(() => {
            busy.value = false;
            hasCode.value = false;

            // Abort the load status request to avoid
            // an old request re-updating with old data
            clearNamedTimeout(loadStatusRequestId);
            abortNamedApiRequest(loadStatusRequestId);
            loadingSilent.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            console.error(err);

            PagesController.ShowSnackBar($t("Error") + ": " + err.message);
        });
};

// Index of the session being closed
const sessionToClose = ref(0);

// True to display the confirmation modal to close a session
const displayCloseConfirmationModal = ref(false);

/**
 * Asks the user to close a session
 * @param s The session to close
 */
const askCloseSession = (s: number) => {
    sessionToClose.value = s;
    displayCloseConfirmationModal.value = true;
};

// Busy status for session being closed
const busyClosing = ref(false);

/**
 * Closes the selected session
 */
const closeSession = () => {
    if (busyClosing.value) {
        return;
    }

    const sessionIndex = sessionToClose.value;
    busyClosing.value = true;

    makeApiRequest(apiInvitesDeleteSession(sessionIndex))
        .onSuccess(() => {
            busyClosing.value = false;

            PagesController.ShowSnackBar($t("Session closed"));

            for (let i = 0; i < sessions.value.length; i++) {
                if (sessions.value[i].index === sessionIndex) {
                    sessions.value.splice(i, 1);
                    break;
                }
            }

            loadSessions();
        })
        .onRequestError((err, handleErr) => {
            busyClosing.value = false;

            handleErr(err, {
                unauthorized: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            busyClosing.value = false;

            console.error(err);

            PagesController.ShowSnackBar($t("Error") + ": " + err.message);
        });
};
</script>
