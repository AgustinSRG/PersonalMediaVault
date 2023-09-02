<template>
    <div class="player-pending-checker">
        <div v-if="error" class="player-task-info">
            <div class="player-task-info-row">
                <span>{{
                    $t("Error: Could not load the media. This may be a network error or maybe the media resource is corrupted.")
                }}</span>
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
                <span>{{
                    $t("It seems the media is not ready yet. This means the media is still being uploaded or it is corrupted.")
                }}</span>
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

            <div class="player-task-info-row" v-if="progress > 0">
                <span
                    >{{ $t("Stage progress") }}: {{ cssProgress(progress) }} / {{ $t("Remaining time (estimated)") }}:
                    {{ renderTime(estimatedRemainingTime) }}</span
                >
            </div>
            <div class="player-task-info-row" v-if="progress > 0">
                <div class="player-task-progress-bar">
                    <div class="player-task-progress-bar-current" :style="{ width: cssProgress(progress) }"></div>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { MediaAPI } from "@/api/api-media";
import { TasksAPI } from "@/api/api-tasks";
import { TaskStatus } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { MediaController, MediaData } from "@/control/media";
import { Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";

export default defineComponent({
    name: "PlayerEncodingPending",
    props: {
        mid: Number,
        tid: Number,
        res: Number,
        error: Boolean,
    },
    data: function () {
        return {
            status: "loading",
            progress: 0,
            stage: "",
            stageNumber: -1,
            stageProgress: 0,
            startTime: 0,
            estimatedRemainingTime: 0,

            pendingId: "",
        };
    },

    methods: {
        start: function () {
            this.checkTask();
        },

        stop: function () {
            Timeouts.Abort(this.pendingId);
            Request.Abort(this.pendingId);
            this.status = "loading";
            this.progress = 0;
            this.stage = "";
            this.stageNumber = -1;
            this.startTime = 0;
            this.estimatedRemainingTime = 0;
        },

        checkTask: function () {
            Timeouts.Abort(this.pendingId);
            Request.Abort(this.pendingId);

            if (this.error) {
                return;
            }

            if (this.tid <= 0) {
                this.status = "not-ready";
                Timeouts.Set(this.pendingId, 1000, this.refreshMedia.bind(this));
                return;
            }

            Request.Pending(this.pendingId, TasksAPI.GetTask(this.tid))
                .onSuccess((task: TaskStatus) => {
                    this.status = "task";
                    if (task.running) {
                        this.progress = task.stage_progress;
                        this.startTime = task.stage_start;
                        this.stage = task.stage;

                        this.estimatedRemainingTime =
                            (((task.time_now - task.stage_start) / task.stage_progress) * 100 - (task.time_now - task.stage_start)) / 1000;

                        switch (this.stage) {
                        case "PREPARE":
                            this.stageNumber = 0;
                            break;
                        case "COPY":
                            this.stageNumber = 1;
                            break;
                        case "PROBE":
                            this.stageNumber = 2;
                            break;
                        case "ENCODE":
                            this.stageNumber = 3;
                            break;
                        case "ENCRYPT":
                            this.stageNumber = 4;
                            break;
                        case "UPDATE":
                            this.stageNumber = 5;
                            break;
                        case "FINISH":
                            this.stageNumber = 6;
                            break;
                        default:
                            this.stageNumber = 0;
                        }

                        this.stageProgress = (this.stageNumber * 100) / 6;

                        Timeouts.Set(this.pendingId, 500, this.checkTask.bind(this));
                    } else {
                        this.stageNumber = -1;
                        this.stage = "QUEUE";
                        this.progress = 0;
                        Timeouts.Set(this.pendingId, 1500, this.checkTask.bind(this));
                    }
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized", false);
                        })
                        .add(404, "*", () => {
                            this.status = "loading";
                            this.checkMediaStatus();
                        })
                        .add("*", "*", () => {
                            // Retry
                            Timeouts.Set(this.pendingId, 1500, this.checkTask.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    Timeouts.Set(this.pendingId, 1500, this.checkTask.bind(this));
                });
        },

        checkMediaStatus: function () {
            Timeouts.Abort(this.pendingId);
            Request.Abort(this.pendingId);

            Request.Pending(this.pendingId, MediaAPI.GetMedia(this.mid))
                .onSuccess((media: MediaData) => {
                    if (this.res >= 0) {
                        if (media.resolutions[this.res] && media.resolutions[this.res].ready) {
                            this.refreshMedia();
                        } else {
                            this.status = "not-ready";
                        }
                    } else {
                        if (media.encoded) {
                            this.refreshMedia();
                        } else {
                            this.status = "not-ready";
                        }
                    }
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized", false);
                        })
                        .add(404, "*", () => {
                            this.refreshMedia();
                        })
                        .add("*", "*", () => {
                            // Retry
                            Timeouts.Set(this.pendingId, 1500, this.checkMediaStatus.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    Timeouts.Set(this.pendingId, 1500, this.checkMediaStatus.bind(this));
                });
        },

        refreshMedia: function () {
            MediaController.Load();
        },

        renderTime: function (s: number): string {
            return renderTimeSeconds(s);
        },

        cssProgress: function (p: number) {
            return Math.round(p) + "%";
        },
    },

    mounted: function () {
        this.pendingId = MediaController.GetPendingId();
        this.start();
    },

    beforeUnmount: function () {
        this.stop();
    },

    watch: {
        mid: function () {
            this.stop();
            this.start();
        },

        res: function () {
            this.stop();
            this.start();
        },

        tid: function () {
            this.stop();
            this.start();
        },
    },
});
</script>
