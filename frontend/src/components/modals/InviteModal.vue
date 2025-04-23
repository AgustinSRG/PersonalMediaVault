<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <div v-if="display" class="modal-dialog modal-lg" role="document">
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
                        <label>{{ $t("Session duration") }}:</label> {{ renderDuration(duration) }}
                    </div>
                    <div v-if="expirationRemaining > 0" class="invite-code-warning">
                        {{ $t("Remaining time until expiration") }}:
                        {{ renderTime(getExpirationRemaining(expirationRemaining, lasReceivedTimestamp, now)) }}
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
                                    <b>{{ $t("Created") }}:</b> {{ renderTimestamp(s.timestamp) }}, <b>{{ $t("Expires") }}:</b>
                                    {{ renderTimestamp(s.expiration) }}
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
import LoadingIcon from "@/components/utils/LoadingIcon.vue";

export default defineComponent({
    name: "InviteModal",
    components: {
        LoadingIcon,
        InviteCloseSessionConfirmationModal,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            loadStatusRequestId: getUniqueStringId(),
            loadSessionsRequestId: getUniqueStringId(),
            updateStatusTimer: null as ReturnType<typeof setInterval> | null,
            updateNowTimer: null as ReturnType<typeof setInterval> | null,
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
            busyClosing: false,

            loadingSessions: true,
            loadingSessionsSilent: true,
            sessions: [] as InviteSession[],

            sessionToClose: 0,
            displayCloseConfirmationModal: false,

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.load();
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
            if (this.busyClosing) {
                return;
            }

            const sessionIndex = this.sessionToClose;

            this.busyClosing = true;

            makeApiRequest(apiInvitesDeleteSession(sessionIndex))
                .onSuccess(() => {
                    this.busyClosing = false;
                    PagesController.ShowSnackBar(this.$t("Session closed"));
                    for (let i = 0; i < this.sessions.length; i++) {
                        if (this.sessions[i].index === sessionIndex) {
                            this.sessions.splice(i, 1);
                            break;
                        }
                    }
                    this.loadSessions();
                })
                .onRequestError((err, handleErr) => {
                    this.busyClosing = false;
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
                    this.busyClosing = false;
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
});
</script>
