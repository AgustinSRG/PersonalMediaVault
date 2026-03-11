<template>
    <div class="player-pending-checker">
        <div v-if="error" class="player-task-info">
            <div class="player-task-info-row">
                <span>{{
                    $t("Error: Could not load the media. This may be a network error or maybe the media resource is corrupted.")
                }}</span>
            </div>
            <div v-if="errorMessage" class="player-task-info-row">
                <span>{{ errorMessage }}</span>
            </div>
            <div class="player-task-info-row">
                <button type="button" class="btn btn-primary" @click="refreshMedia">
                    <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                </button>
            </div>
        </div>
        <div v-else-if="status === 'loading'" class="player-lds-ring">
            <div></div>
            <div></div>
            <div></div>
            <div></div>
        </div>
        <div v-else-if="status === 'not-ready'" class="player-task-info">
            <div class="player-task-info-row">
                <span v-if="encodingError">{{
                    $t("The encoding process failed. Check the error trace below and report it as a bug.")
                }}</span>
                <span v-else>{{
                    $t("It seems the media is not ready yet. This means the media is still being uploaded or it is corrupted.")
                }}</span>
            </div>
            <div v-if="encodingError" class="player-task-info-row">
                <textarea
                    ref="encodingErrorTextArea"
                    class="form-control form-textarea player-task-encoding-error"
                    rows="10"
                    :value="encodingError"
                    readonly
                ></textarea>
            </div>
            <div class="player-task-info-row">
                <button type="button" class="btn btn-primary" @click="refreshMedia">
                    <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                </button>
            </div>
        </div>
        <div v-else-if="status === 'task' && stageNumber < 0" class="player-task-info">
            <div class="player-task-info-row">
                <span>{{ $t("The media is still pending to be encoded. The task will start as soon as possible.") }}</span>
            </div>
            <div class="player-task-info-row">
                <button type="button" class="btn btn-primary" @click="refreshMedia">
                    <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                </button>
            </div>
        </div>
        <div v-else-if="status === 'task' && stageNumber >= 0" class="player-task-info">
            <div class="player-task-info-row">
                <span>{{ $t("The media is being encoded.") }}</span>
            </div>
            <div class="player-task-info-row">
                <span>{{ $t("Stage") }} ({{ stageNumber + 1 }} / 7):&nbsp;</span>

                <span v-if="stage === 'PREPARE'">{{ $t("Preparing task environment") }}...</span>
                <span v-if="stage === 'COPY'">{{ $t("Copying assets to be encoded") }}...</span>
                <span v-if="stage === 'PROBE'">{{ $t("Extracting metadata") }}...</span>
                <span v-if="stage === 'ENCODE'">{{ $t("Encoding media assets") }}...</span>
                <span v-if="stage === 'ENCRYPT'">{{ $t("Encrypting and storing in the vault") }}...</span>
                <span v-if="stage === 'UPDATE'">{{ $t("Updating metadata") }}...</span>
                <span v-if="stage === 'FINISH'">{{ $t("Cleaning up") }}...</span>
            </div>

            <div v-if="progress > 0" class="player-task-info-row">
                <span
                    >{{ $t("Stage progress") }}: {{ cssProgress }} / {{ $t("Remaining time (estimated)") }}:
                    {{ renderedEstimatedRemainingTime }}</span
                >
            </div>
            <div v-if="progress > 0" class="player-task-info-row">
                <div class="player-task-progress-bar">
                    <div class="player-task-progress-bar-current" :style="{ width: cssProgress }"></div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { MediaData, TaskStatus } from "@/api/models";
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { renderTimeSeconds } from "@/utils/time";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { computed, nextTick, onMounted, ref, useTemplateRef, watch } from "vue";
import { apiMediaGetMedia } from "@/api/api-media";
import { apiTasksGetTask } from "@/api/api-tasks";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { loadCurrentMedia } from "@/control/media";

// Known stages in order
const KNOWN_STAGES_IN_ORDER = ["PREPARE", "COPY", "PROBE", "ENCODE", "ENCRYPT", "UPDATE", "FINISH"];

/**
 * Gets the stage number from the stage name
 * @param s The stage name
 * @returns The stage number
 */
function getStageNumber(s: string): number {
    const i = KNOWN_STAGES_IN_ORDER.indexOf(s);
    return i >= 0 ? i : 0;
}

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
    /**
     * Media ID
     */
    mid: Number,

    /**
     * Task ID
     */
    tid: Number,

    /**
     * Resolution index
     */
    res: Number,

    /**
     * Is error
     */
    error: Boolean,

    /**
     * Error message
     */
    errorMessage: String,

    /**
     * Encoding error message
     */
    encodingError: String,

    /**
     * True if the media can be auto-reloaded
     */
    canAutoReload: Boolean,
});

// Pending statuses
type PendingStatus = "loading" | "not-ready" | "task";

// Status
const status = ref<PendingStatus>("loading");

// Progress
const progress = ref(0);

// CSS style progress (as percentage)
const cssProgress = computed(() => Math.round(progress.value) + "%");

// Current stage
const stage = ref("");

// Current stage number
const stageNumber = ref(-1);

// Current stage progress
const stageProgress = ref(0);

// Stage start timestamp (Unix milliseconds)
const startTime = ref(0);

// Estimated remaining time (seconds)
const estimatedRemainingTime = ref(0);

// Rendered estimated remaining time
const renderedEstimatedRemainingTime = computed(() => renderTimeSeconds(estimatedRemainingTime.value));

// Pending refreshing?
const refreshPending = ref(false);

/**
 * Refreshes the media
 */
const refreshMedia = () => {
    loadCurrentMedia();
};

watch(
    () => props.canAutoReload,
    () => {
        if (props.canAutoReload && refreshPending.value) {
            refreshPending.value = true;
            refreshMedia();
        }
    },
);

// Load request ID
const loadRequestId = useRequestId();

/**
 * Loads the data
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    status.value = "loading";
    progress.value = 0;
    stage.value = "";
    stageNumber.value = -1;
    startTime.value = 0;
    estimatedRemainingTime.value = 0;

    onStatusChanged();

    checkTask();
};

onMounted(load);
watch([() => props.mid, () => props.tid, () => props.res], load);

// Delay to automatically refresh the media
const AUTO_RELOAD_DELAY = 1000;

// Delay to refresh the status periodically (milliseconds)
const REFRESH_DELAY = 500;

// Delay to retry after error (milliseconds)
const RETRY_DELAY = 1500;

/**
 * Loads the task status
 */
const checkTask = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (props.error) {
        return;
    }

    if (props.tid <= 0) {
        status.value = "not-ready";
        setNamedTimeout(loadRequestId, AUTO_RELOAD_DELAY, () => {
            if (!props.canAutoReload) {
                return;
            }

            refreshMedia();
        });
        onStatusChanged();
        return;
    }

    makeNamedApiRequest(loadRequestId, apiTasksGetTask(props.tid))
        .onSuccess((task: TaskStatus) => {
            status.value = "task";
            if (task.running) {
                progress.value = task.stage_progress;
                startTime.value = task.stage_start;
                stage.value = task.stage;

                estimatedRemainingTime.value =
                    (((task.time_now - task.stage_start) / task.stage_progress) * 100 - (task.time_now - task.stage_start)) / 1000;

                stageNumber.value = getStageNumber(stage.value);

                stageProgress.value = (stageNumber.value * 100) / 6;

                setNamedTimeout(loadRequestId, REFRESH_DELAY, checkTask);
            } else {
                stageNumber.value = -1;
                stage.value = "QUEUE";
                progress.value = 0;
                setNamedTimeout(loadRequestId, RETRY_DELAY, checkTask);
            }

            onStatusChanged();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    status.value = "loading";
                    checkMediaStatus();
                    onStatusChanged();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, RETRY_DELAY, checkTask);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, RETRY_DELAY, checkTask);
        });
};

/**
 * Checks the current media status if the task is not found
 */
const checkMediaStatus = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    makeNamedApiRequest(loadRequestId, apiMediaGetMedia(props.mid))
        .onSuccess((media: MediaData) => {
            if (props.res >= 0) {
                if (media.resolutions[props.res] && media.resolutions[props.res].ready) {
                    if (props.canAutoReload) {
                        refreshMedia();
                    } else {
                        refreshPending.value = true;
                    }
                } else {
                    status.value = "not-ready";
                    onStatusChanged();
                }
            } else {
                if (media.encoded) {
                    if (props.canAutoReload) {
                        refreshMedia();
                    } else {
                        refreshPending.value = true;
                    }
                } else {
                    status.value = "not-ready";
                    onStatusChanged();
                }
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    if (props.canAutoReload) {
                        refreshMedia();
                    } else {
                        refreshPending.value = true;
                    }
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, RETRY_DELAY, checkMediaStatus);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, RETRY_DELAY, checkMediaStatus);
        });
};

/**
 * Called whenever the status changes
 */
const onStatusChanged = () => {
    if (status.value === "not-ready" && props.encodingError) {
        nextTick(() => {
            scrollErrorToBottom();
        });
    }
};

// Ref to the textarea containing the encoding error
const encodingErrorTextArea = useTemplateRef("encodingErrorTextArea");

/**
 * Scrolls encoding error to bottom
 */
const scrollErrorToBottom = () => {
    const el = encodingErrorTextArea.value;

    if (!el) {
        return;
    }

    const bounds = el.getBoundingClientRect();

    el.scrollTop = el.scrollHeight - bounds.height;
};
</script>
