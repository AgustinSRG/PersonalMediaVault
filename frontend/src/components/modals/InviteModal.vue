<template>
    <ModalDialogContainer
        :closeSignal="closeSignal"
        :forceCloseSignal="forceCloseSignal"
        v-model:display="displayStatus"
        :lock-close="busy"
    >
        <div v-if="display" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Invite") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body" v-if="loading">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div class="modal-body" v-if="!loading">
                <div v-if="hasCode" class="invite-code-modal-body">
                    <div class="invite-code-display">
                        <div v-for="(c, i) in code.split('')" class="invite-code-char-container" :key="i">
                            <div class="invite-code-char">{{ c }}</div>
                        </div>
                    </div>
                    <div class="invite-code-info">
                        {{ $t("Use this code to log into the vault in read-only mode") }}
                    </div>
                    <div class="invite-code-info">
                        <label>{{ $t("Session duration") }}:</label> {{ renderDuration(duration) }}
                    </div>
                    <div class="invite-code-warning" v-if="expirationRemaining > 0">
                        {{ $t("Remaining time until expiration") }}:
                        {{ renderTime(getExpirationRemaining(expirationRemaining, lasReceivedTimestamp, now)) }}
                    </div>
                    <div class="invite-code-error" v-if="expirationRemaining <= 0">
                        {{ $t("The invite code has expired") }}
                    </div>
                    <div class="invite-code-btn-container">
                        <button type="button" :disabled="busy" @click="clearCode" class="btn btn-danger">
                            <i class="fas fa-trash-alt"></i> {{ $t("Clear invite code") }}
                        </button>
                    </div>
                </div>
                <div v-else class="invite-code-modal-body">
                    <div class="invite-code-info">{{ $t("You can invite other users or devices by creating an invite code") }}</div>
                    <div class="invite-code-info">
                        {{ $t("Invite codes are single use, and allow to login into the vault in read-only mode") }}
                    </div>
                    <div class="invite-code-btn-container">
                        <button type="button" :disabled="busy" @click="generateCode" class="btn btn-primary">
                            <i class="fas fa-plus"></i> {{ $t("Create invite code") }}
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
            <div class="modal-body" v-if="loadingSessions">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div class="modal-body no-padding border-top" v-if="!loadingSessions">
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
                                    <b>{{ $t("Created") }}:</b> {{ renderTimestamp(s.timestamp) }}, <b>{{ $t("Expires") }}:</b>
                                    {{ renderTimestamp(s.expiration) }}
                                </td>
                                <td class="text-right td-shrink">
                                    <button type="button" :disabled="busy" @click="askCloseSession(s.index)" class="btn btn-danger btn-xs">
                                        <i class="fas fa-trash-alt"></i> {{ $t("Close") }}
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

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { PagesController } from "@/control/pages";
import {
    InviteSession,
    apiInvitesClearCode,
    apiInvitesDeleteSession,
    apiInvitesGenerateCode,
    apiInvitesGetSessions,
    apiInvitesGetStatus,
} from "@/api/api-invites";
import { getUniqueStringId } from "@/utils/unique-id";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { renderDateAndTime, renderTimeSeconds } from "@/utils/time";

import InviteCloseSessionConfirmationModal from "@/components/modals/InviteCloseSessionConfirmationModal.vue";
import { SessionDuration } from "@/api/api-auth";

export default defineComponent({
    components: {
        InviteCloseSessionConfirmationModal,
    },
    name: "InviteModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            loadStatusRequestId: getUniqueStringId(),
            loadSessionsRequestId: getUniqueStringId(),
            updateStatusTimer: null,
            updateNowTimer: null,
        };
    },
    data: function () {
        return {
            loading: true,
            loadingSilent: true,
            hasCode: false,
            code: "",
            duration: 0,
            expirationRemaining: 0,
            lasReceivedTimestamp: 0,
            now: Date.now(),

            durationGenerate: "day" as SessionDuration,

            busy: false,

            loadingSessions: true,
            loadingSessionsSilent: true,
            sessions: [] as InviteSession[],

            sessionToClose: 0,
            displayCloseConfirmationModal: false,

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    methods: {
        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        load: function () {
            if (!this.display) {
                return;
            }

            this.loading = true;
            this.loadingSessions = true;

            this.loadStatus();
            this.loadSessions();
        },

        loadStatus: function () {
            clearNamedTimeout(this.loadStatusRequestId);

            this.loadingSilent = true;

            makeNamedApiRequest(this.loadStatusRequestId, apiInvitesGetStatus())
                .onSuccess((response) => {
                    this.hasCode = response.has_code;
                    this.code = response.code;
                    this.expirationRemaining = response.expiration_remaining;
                    this.lasReceivedTimestamp = Date.now();
                    this.now = Date.now();
                    this.duration = response.duration;

                    this.loading = false;
                    this.loadingSilent = false;

                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.close();
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadStatusRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadStatusRequestId, 1500, this.load.bind(this));
                });
        },

        loadSessions: function () {
            clearNamedTimeout(this.loadSessionsRequestId);

            this.loadingSessionsSilent = true;

            makeNamedApiRequest(this.loadSessionsRequestId, apiInvitesGetSessions())
                .onSuccess((response) => {
                    this.sessions = response;

                    this.loadingSessions = false;
                    this.loadingSessionsSilent = false;

                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.close();
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadSessionsRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadSessionsRequestId, 1500, this.load.bind(this));
                });
        },

        checkForLoad: function () {
            if (!this.display) {
                return;
            }

            if (!this.loadingSilent) {
                this.loadStatus();
            }

            if (!this.loadingSessionsSilent) {
                this.loadSessions();
            }
        },

        close: function () {
            this.closeSignal++;
        },

        renderTimestamp: function (t: number): string {
            return renderDateAndTime(t, this.$locale.value);
        },

        askCloseSession: function (s: number) {
            this.sessionToClose = s;
            this.displayCloseConfirmationModal = true;
        },

        closeSession: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(apiInvitesDeleteSession(this.sessionToClose))
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Session closed"));
                    this.loadSessions();
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.busy = false;
                    console.error(err);
                    PagesController.ShowSnackBar(this.$t("Error") + ": " + err.message);
                });
        },

        generateCode: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(apiInvitesGenerateCode(this.durationGenerate))
                .onSuccess((response) => {
                    this.busy = false;
                    this.hasCode = response.has_code;
                    this.code = response.code;
                    this.expirationRemaining = response.expiration_remaining;
                    this.lasReceivedTimestamp = Date.now();
                    this.now = Date.now();
                    this.duration = response.duration;

                    clearNamedTimeout(this.loadStatusRequestId);
                    abortNamedApiRequest(this.loadStatusRequestId);
                    this.loadingSilent = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        limitReached: () => {
                            PagesController.ShowSnackBar(
                                this.$t("Error") + ": " + this.$t("You reached the limit of invited sessions you can have"),
                            );
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.busy = false;
                    console.error(err);
                    PagesController.ShowSnackBar(this.$t("Error") + ": " + err.message);
                });
        },

        clearCode: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(apiInvitesClearCode())
                .onSuccess(() => {
                    this.busy = false;
                    this.hasCode = false;

                    clearNamedTimeout(this.loadStatusRequestId);
                    abortNamedApiRequest(this.loadStatusRequestId);
                    this.loadingSilent = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.busy = false;
                    console.error(err);
                    PagesController.ShowSnackBar(this.$t("Error") + ": " + err.message);
                });
        },

        getExpirationRemaining: function (expirationRemaining: number, lasReceivedTimestamp: number, now: number): number {
            return Math.max(0, expirationRemaining - (now - lasReceivedTimestamp));
        },

        renderTime: function (t: number): string {
            return renderTimeSeconds(Math.round(t / 1000));
        },

        updateNow: function () {
            this.now = Date.now();
        },

        renderDuration: function (t: number) {
            const days = Math.max(1, Math.round(t / (24 * 60 * 60 * 1000)));

            if (days === 1) {
                return "1" + " " + this.$t("day");
            } else {
                return days + " " + this.$t("days");
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.load.bind(this));

        this.updateStatusTimer = setInterval(this.checkForLoad.bind(this), 2000);
        this.updateNowTimer = setInterval(this.updateNow.bind(this), 500);

        if (this.display) {
            this.load();
        }
    },
    beforeUnmount: function () {
        clearInterval(this.updateStatusTimer);
        clearInterval(this.updateNowTimer);

        clearNamedTimeout(this.loadStatusRequestId);
        abortNamedApiRequest(this.loadStatusRequestId);

        clearNamedTimeout(this.loadSessionsRequestId);
        abortNamedApiRequest(this.loadSessionsRequestId);
    },
    watch: {
        display: function () {
            if (this.display) {
                this.load();
            }
        },
    },
});
</script>
<style>
.invite-code-modal-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 240px;
}

.invite-code-info,
.invite-code-warning,
.invite-code-error {
    padding: 0.5rem;
    font-size: large;
    text-align: center;
}

.invite-code-error {
    color: red;
}

.invite-code-btn-container,
.invite-code-select-container {
    padding: 0.5rem;
}

.invite-code-select-container {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.invite-code-select-container label {
    padding-bottom: 0.5rem;
}

.invite-code-display {
    padding: 0.5rem;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
}

.invite-code-char-container {
    padding-left: 0.25rem;
    padding-right: 0.25rem;
}

.invite-code-char {
    padding: 1rem;
    border: solid 1px var(--theme-border-color);
    font-size: xx-large;
    font-weight: bold;
}

@media (max-width: 400px) {
    .invite-code-char {
        padding: 0.5rem;
        font-size: large;
    }
}

@media (max-width: 250px) {
    .invite-code-char {
        padding: 0;
        border: none;
        font-size: large;
    }

    .invite-code-display {
        flex-wrap: wrap;
    }
}
</style>
