<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div class="modal-dialog modal-xl modal-height-100" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Tasks") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body no-padding">
                <div class="table-responsive">
                    <table class="table">
                        <thead>
                            <tr>
                                <th class="text-left" colspan="2">
                                    {{ $t("List of active tasks") }}
                                </th>
                                <th class="text-left">{{ $t("Type") }}</th>
                                <th class="text-left">{{ $t("Media") }}</th>
                                <th class="text-left">{{ $t("Status") }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-if="tasks.length === 0">
                                <td colspan="5">
                                    {{ $t("There are no active tasks in the vault at the moment.") }}
                                </td>
                            </tr>
                            <tr v-for="t in tasks" :key="t.id">
                                <td class="one-line td-shrink">
                                    <i class="fas fa-circle task-status" :class="{ 'task-running': t.running }"></i>
                                </td>
                                <td class="task-progress-bar-td one-line td-shrink">
                                    <div v-if="t.running" class="task-progress-bar-container">
                                        <div
                                            class="task-progress-bar-current"
                                            :style="{
                                                width: getGlobalProgress(t.stage, t.stage_progress),
                                            }"
                                        ></div>
                                    </div>
                                </td>
                                <td class="bold one-line td-shrink">
                                    {{ renderType(t.type) }}
                                </td>
                                <td class="bold one-line td-shrink">
                                    <a
                                        :href="getMediaURL(t.media_id)"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        @click="goToMedia(t.media_id, $event)"
                                        >{{ t.media_id }}</a
                                    >
                                </td>
                                <td class="one-line">
                                    {{ renderStatus(t.running, t.stage, t.stage_progress, t.stage_start, t.time_now) }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { getFrontendUrl } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { renderTimeSeconds } from "@/utils/time";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { onMounted, ref, useTemplateRef, watch } from "vue";
import { apiTasksGetTasks } from "@/api/api-tasks";
import type { TaskStatus } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Tasks list
const tasks = ref<TaskStatus[]>([]);

/**
 * Sets the task list
 * @param newTasks The new task list
 */
const setTasks = (newTasks: TaskStatus[]) => {
    tasks.value = newTasks.sort((a, b) => {
        if (a.running && !b.running) {
            return -1;
        } else if (!a.running && b.running) {
            return 1;
        } else if (a.type < b.type) {
            return -1;
        } else if (a.type > b.type) {
            return 1;
        } else if (a.id < b.id) {
            return -1;
        } else {
            return 1;
        }
    });
};

// Loading status
const loading = ref(true);

// Load request ID
const loadRequestId = useRequestId();

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the data
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (!display.value) {
        return;
    }

    loading.value = true;

    loadTaskList();
};

// Update period (milliseconds)
const UPDATE_PERIOD = 500;

/**
 * Loads the list of tasks
 */
const loadTaskList = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (!display.value) {
        return;
    }

    makeNamedApiRequest(loadRequestId, apiTasksGetTasks())
        .onSuccess((tasks) => {
            setTasks(tasks);

            loading.value = false;

            setNamedTimeout(loadRequestId, UPDATE_PERIOD, loadTaskList);
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadTaskList);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadTaskList);
        });
};

onMounted(load);
watch(display, load);

/**
 * gets progress as percentage
 * @param p The progress (0-1)
 * @returns The progress as percentage
 */
const getProgressPercent = (p: number) => {
    return Math.floor(p * 100) / 100 + "%";
};

/**
 * Gets global task progress
 * @param stage The current stage
 * @param p The stage progress
 */
const getGlobalProgress = (stage: string, p: number) => {
    let baseProgress = 0;
    let progressCap = 100;
    switch (stage) {
        case "PREPARE":
            baseProgress = 0;
            progressCap = 0;
            break;
        case "COPY":
            baseProgress = 0;
            progressCap = 10;
            break;
        case "PROBE":
            baseProgress = 10;
            progressCap = 0;
            break;
        case "ENCODE":
            baseProgress = 10;
            progressCap = 60;
            break;
        case "ENCRYPT":
            baseProgress = 70;
            progressCap = 25;
            break;
        case "UPDATE":
            baseProgress = 95;
            progressCap = 5;
            break;
        case "FINISH":
            baseProgress = 100;
            progressCap = 0;
            break;
    }

    const realP = baseProgress + (p * progressCap) / 100;

    return Math.floor(realP * 100) / 100 + "%";
};

/**
 * Renders the tasks type
 * @param t The task type as a number
 * @returns The rendered type
 */
const renderType = (t: number) => {
    switch (t) {
        case 0:
            return $t("Encode media");
        case 1:
            return $t("Encode extra resolution");
        case 2:
            return $t("Generate video previews");
        default:
            return "???";
    }
};

/**
 * Renders task status
 * @param running Running status
 * @param stage Current stage
 * @param p Stage progress
 * @param ts Task las updated timestamp
 * @param now Current timestamp
 */
const renderStatus = (running: boolean, stage: string, p: number, ts: number, now: number) => {
    if (running) {
        let stageNumber = 0;

        switch (stage) {
            case "PREPARE":
                stageNumber = 0;
                break;
            case "COPY":
                stageNumber = 1;
                break;
            case "PROBE":
                stageNumber = 2;
                break;
            case "ENCODE":
                stageNumber = 3;
                break;
            case "ENCRYPT":
                stageNumber = 4;
                break;
            case "UPDATE":
                stageNumber = 5;
                break;
            case "FINISH":
                stageNumber = 6;
                break;
        }

        const progressPercent = getProgressPercent(p);

        const estimatedRemainingTime = (((now - ts) / p) * 100 - (now - ts)) / 1000;

        let txt =
            getGlobalProgress(stage, p) +
            " | " +
            $t("Stage") +
            ": " +
            (stageNumber + 1) +
            " / 7 | " +
            $t("Stage progress") +
            ": " +
            progressPercent;

        if (estimatedRemainingTime > 0) {
            txt += " | " + $t("Remaining time (estimated)") + ": " + renderTimeSeconds(estimatedRemainingTime);
        }

        return txt;
    } else {
        return $t("Task is in queue");
    }
};

/**
 * Call when the user click on media link
 * @param mid The media ID
 * @param e The click event
 */
const goToMedia = (mid: number, e?: Event) => {
    if (e) {
        e.preventDefault();
    }
    AppStatus.ClickOnMedia(mid, true);
    close();
};

/**
 * Gets the URL for a media link
 * @param mid The media ID
 * @returns The URL
 */
const getMediaURL = (mid: number): string => {
    return getFrontendUrl({
        media: mid,
    });
};
</script>
