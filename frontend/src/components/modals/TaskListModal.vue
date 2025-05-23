<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <div v-if="display" class="modal-dialog modal-xl modal-height-100" role="document">
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

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { generateURIQuery } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { renderTimeSeconds } from "@/utils/time";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiTasksGetTasks } from "@/api/api-tasks";
import { TaskStatus } from "@/api/models";

export default defineComponent({
    name: "TaskListModal",
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            updateRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            tasks: [] as TaskStatus[],

            loading: true,

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
                this.load();
            }
        },
    },
    mounted: function () {
        this.load();

        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
        clearNamedTimeout(this.updateRequestId);
        abortNamedApiRequest(this.updateRequestId);
    },
    methods: {
        setTasks: function (tasks: TaskStatus[]) {
            this.tasks = tasks.sort((a, b) => {
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
        },

        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);
            clearNamedTimeout(this.updateRequestId);
            abortNamedApiRequest(this.updateRequestId);

            if (!this.display) {
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.loadRequestId, apiTasksGetTasks())
                .onSuccess((tasks) => {
                    this.setTasks(tasks);
                    this.loading = false;
                    setNamedTimeout(this.updateRequestId, 500, this.updateTasks.bind(this));
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        updateTasks: function () {
            clearNamedTimeout(this.updateRequestId);
            abortNamedApiRequest(this.updateRequestId);

            if (!this.display) {
                return;
            }

            makeNamedApiRequest(this.updateRequestId, apiTasksGetTasks())
                .onSuccess((tasks) => {
                    this.setTasks(tasks);
                    setNamedTimeout(this.updateRequestId, 500, this.updateTasks.bind(this));
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.updateRequestId, 1500, this.updateTasks.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.updateRequestId, 1500, this.updateTasks.bind(this));
                });
        },

        getProgressPercent: function (p: number) {
            return Math.floor(p * 100) / 100 + "%";
        },

        getGlobalProgress: function (stage: string, p: number) {
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
        },

        close: function () {
            this.closeSignal++;
        },

        renderType: function (t: number) {
            switch (t) {
                case 0:
                    return this.$t("Encode media");
                case 1:
                    return this.$t("Encode extra resolution");
                case 2:
                    return this.$t("Generate video previews");
                default:
                    return "???";
            }
        },

        renderStatus: function (running: boolean, stage: string, p: number, ts: number, now: number) {
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

                const progressPercent = this.getProgressPercent(p);

                const estimatedRemainingTime = (((now - ts) / p) * 100 - (now - ts)) / 1000;

                let txt =
                    this.getGlobalProgress(stage, p) +
                    " | " +
                    this.$t("Stage") +
                    ": " +
                    (stageNumber + 1) +
                    " / 7 | " +
                    this.$t("Stage progress") +
                    ": " +
                    progressPercent;

                if (estimatedRemainingTime > 0) {
                    txt += " | " + this.$t("Remaining time (estimated)") + ": " + renderTimeSeconds(estimatedRemainingTime);
                }

                return txt;
            } else {
                return this.$t("Task is in queue");
            }
        },

        goToMedia: function (mid, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnMedia(mid, true);
            this.close();
        },

        getMediaURL: function (mid: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    media: mid + "",
                })
            );
        },
    },
});
</script>
