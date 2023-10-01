<template>
    <div class="player-editor-sub-content">
        <!--- Time slices -->

        <div class="form-group">
            <label>{{ $t("Time slices") }}. {{ $t("You can split the time in slices and name them to display at the timeline.") }}</label>
        </div>

        <div class="horizontal-filter-menu two-child">
            <a href="javascript:;" @click="setTextMode(true)" class="horizontal-filter-menu-item" :class="{ selected: textMode }">
                {{ $t("Text view") }}</a
            >
            <a href="javascript:;" @click="setTextMode(false)" class="horizontal-filter-menu-item" :class="{ selected: !textMode }">
                {{ $t("Table view") }}</a
            >
        </div>

        <div class="form-group table-responsive" v-if="!textMode">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("Timestamp") }}</th>
                        <th class="text-left">{{ $t("Slice name") }}</th>
                        <th class="td-shrink"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(s, i) in timeSlicesArray" :key="i">
                        <td>
                            <input
                                v-if="canWrite"
                                type="text"
                                :disabled="busy"
                                v-model="s.timeStr"
                                class="form-control form-control-full-width auto-focus"
                                :placeholder="i == 0 ? '00:01:00' : '00:00:00'"
                                @change="updateSliceTime(s)"
                            />
                            <span v-if="!canWrite">{{ s.timeStr }}</span>
                        </td>
                        <td>
                            <input
                                v-if="canWrite"
                                type="text"
                                :disabled="busy"
                                v-model="s.name"
                                class="form-control form-control-full-width"
                                :placeholder="i !== 0 ? $t('Rest of the video') : $t('Opening')"
                                @input="markDirty"
                            />
                            <span v-if="!canWrite">{{ s.name }}</span>
                        </td>
                        <td class="td-shrink">
                            <button v-if="canWrite" type="button" class="btn btn-danger btn-xs" :disabled="busy" @click="deleteSlice(i)">
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                    <tr v-if="timeSlicesArray.length === 0">
                        <td colspan="3">{{ $t("There are no time slices yet for this media.") }}</td>
                    </tr>
                    <tr v-if="canWrite">
                        <td>
                            <input
                                type="text"
                                :disabled="busy"
                                v-model="sliceAddTimestamp"
                                class="form-control form-control-full-width auto-focus"
                                :placeholder="timeSlicesArray.length > 0 ? '00:01:00' : '00:00:00'"
                            />
                        </td>
                        <td>
                            <input
                                type="text"
                                :disabled="busy"
                                v-model="sliceAddName"
                                class="form-control form-control-full-width"
                                :placeholder="timeSlicesArray.length > 0 ? $t('Rest of the video') : $t('Opening')"
                            />
                        </td>
                        <td class="td-shrink">
                            <button
                                type="button"
                                class="btn btn-primary btn-xs"
                                :disabled="busy || !sliceAddName || !sliceAddTimestamp"
                                @click="addSlice"
                            >
                                <i class="fas fa-plus"></i> {{ $t("Add") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group mt-1" v-if="textMode">
            <textarea
                v-model="timeSlices"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea auto-focus"
                :placeholder="'00:00:00 ' + $t('Opening') + '\n00:01:00 ' + $t('Rest of the video')"
                rows="10"
                :disabled="busy"
                @input="markDirty"
            ></textarea>
        </div>

        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-primary" :disabled="busy || !dirty" @click="changeTimeSlices">
                <i class="fas fa-pencil-alt"></i> {{ $t("Change time slices") }}
            </button>
        </div>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { MediaController } from "@/control/media";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { parseTimeSeconds, parseTimeSlices, renderTimeSlices } from "@/utils/time-slices";
import { EditMediaAPI } from "@/api/api-media-edit";
import { clone } from "@/utils/objects";
import { renderTimeSeconds } from "@/utils/time";

export default defineComponent({
    components: {},
    name: "EditorTimeSlices",
    emits: ["changed"],
    data: function () {
        return {
            type: 0,

            textMode: true,

            timeSlices: "",
            originalTimeSlices: "",

            timeSlicesArray: [],

            sliceAddTimestamp: "",
            sliceAddName: "",

            busy: false,

            dirty: false,

            canWrite: AuthController.CanWrite,
        };
    },

    methods: {
        autoFocus: function () {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        deleteSlice: function (i: number) {
            this.timeSlicesArray.splice(i, 1);
            this.dirty = true;
        },

        updateSliceTime: function (slice: { time: number; timeStr: string; name: string }) {
            const timeSeconds = parseTimeSeconds(slice.timeStr);

            if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
                slice.timeStr = renderTimeSeconds(slice.time);
            }

            slice.time = timeSeconds;
            this.dirty = true;
        },

        addSlice: function () {
            const timeSeconds = parseTimeSeconds(this.sliceAddTimestamp);

            if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
                return;
            }

            if (!this.sliceAddName) {
                return;
            }

            this.timeSlicesArray.push({
                time: timeSeconds,
                timeStr: renderTimeSeconds(timeSeconds),
                name: this.sliceAddName,
            });

            this.sliceAddTimestamp = "";
            this.sliceAddName = "";
            this.dirty = true;
            this.autoFocus();
        },

        markDirty: function () {
            this.dirty = true;
        },

        setTextMode: function (m: boolean) {
            this.textMode = m;

            if (this.textMode) {
                this.timeSlices = renderTimeSlices(this.timeSlicesArray);
            } else {
                this.timeSlicesArray = parseTimeSlices(this.timeSlices).map((s) => {
                    return {
                        time: s.time,
                        timeStr: renderTimeSeconds(s.time),
                        name: s.name,
                    };
                });
            }

            this.autoFocus();
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.originalTimeSlices = renderTimeSlices(MediaController.MediaData.time_slices);
            this.timeSlices = this.originalTimeSlices;
            this.timeSlicesArray = parseTimeSlices(this.timeSlices).map((s) => {
                return {
                    time: s.time,
                    timeStr: renderTimeSeconds(s.time),
                    name: s.name,
                };
            });
            this.dirty = false;
        },

        changeTimeSlices: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            const slices = this.textMode ? parseTimeSlices(this.timeSlices) : clone(this.timeSlicesArray);

            Request.Pending("media-editor-busy-time-slices", EditMediaAPI.ChangeTimeSlices(mediaId, slices))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed time slices"));
                    this.busy = false;
                    this.dirty = false;
                    this.originalTimeSlices = renderTimeSlices(slices);
                    this.timeSlices = this.originalTimeSlices;
                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(400, "*", () => {
                            AppEvents.Emit("snack", this.$t("Bad request"));
                        })
                        .add(401, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                            AppEvents.Emit("unauthorized");
                        })
                        .add(403, "*", () => {
                            AppEvents.Emit("snack", this.$t("Access denied"));
                        })
                        .add(404, "*", () => {
                            AppEvents.Emit("snack", this.$t("Not found"));
                        })
                        .add(500, "*", () => {
                            AppEvents.Emit("snack", this.$t("Internal server error"));
                        })
                        .add("*", "*", () => {
                            AppEvents.Emit("snack", this.$t("Could not connect to the server"));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    AppEvents.Emit("snack", err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },

    mounted: function () {
        this._handles = Object.create(null);
        this.updateMediaData();

        this._handles.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this._handles.mediaUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);

        this.autoFocus();
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-media-update", this._handles.mediaUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        Request.Abort("media-editor-busy-time-slices");
    },
});
</script>
