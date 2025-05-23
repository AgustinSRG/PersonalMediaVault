<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            v-model:display="displayStatus"
            :title="$t('Time slices')"
            :context-open="contextOpen"
            :position-key="'time-slices-helper-pos'"
            @clicked="propagateClick"
        >
            <div class="table-responsive time-slices-table">
                <table class="table">
                    <thead>
                        <tr>
                            <th class="text-left one-line">
                                {{ $t("Timestamp") }}
                            </th>
                            <th class="text-left one-line">
                                {{ $t("Slice name") }}
                            </th>
                            <th class="td-shrink"></th>
                            <th class="td-shrink"></th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(slice, i) in timeSlicesArray" :key="i">
                            <td v-if="slice.deleted" colspan="4">
                                <a href="javascript:;" @click="undoDelete(i)">{{ $t("Undo delete") }}</a>
                            </td>
                            <td v-if="!slice.deleted" class="one-line">
                                <input
                                    v-if="sliceEditIndex === i"
                                    v-model="sliceEditTimestamp"
                                    type="text"
                                    class="form-control form-control-full-width"
                                    placeholder="00:00:00"
                                />
                                <span v-else>{{ slice.timeStr }}</span>
                            </td>
                            <td v-if="!slice.deleted">
                                <input
                                    v-if="sliceEditIndex === i"
                                    v-model="sliceEditName"
                                    type="text"
                                    class="form-control form-control-full-width auto-focus"
                                    @keydown="keyDownEdit"
                                />
                                <span v-else>{{ slice.name }}</span>
                            </td>
                            <td v-if="!slice.deleted" class="td-shrink text-right one-line">
                                <button
                                    v-if="sliceEditIndex < 0"
                                    type="button"
                                    class="time-slices-edit-helper-btn mr-1"
                                    :title="$t('Edit')"
                                    @click="editSlice(i)"
                                >
                                    <i class="fas fa-pencil-alt"></i>
                                </button>
                                <button
                                    v-if="sliceEditIndex === i"
                                    type="button"
                                    class="time-slices-edit-helper-btn mr-1"
                                    :title="$t('Save')"
                                    @click="finishEdit"
                                >
                                    <i class="fas fa-check"></i>
                                </button>
                            </td>
                            <td v-if="!slice.deleted" class="td-shrink text-right one-line">
                                <button
                                    v-if="sliceEditIndex < 0"
                                    type="button"
                                    class="time-slices-edit-helper-btn"
                                    :title="$t('Delete')"
                                    @click="deleteSlice(i)"
                                >
                                    <i class="fas fa-trash-alt"></i>
                                </button>
                                <button
                                    v-if="sliceEditIndex === i"
                                    type="button"
                                    class="time-slices-edit-helper-btn"
                                    :title="$t('Cancel')"
                                    @click="cancelEdit"
                                >
                                    <i class="fas fa-times"></i>
                                </button>
                            </td>
                        </tr>
                        <tr v-if="timeSlicesArray.length === 0">
                            <td colspan="3">{{ $t("There are no time slices yet for this media.") }}</td>
                        </tr>
                        <tr v-if="canWrite && sliceEditIndex < 0">
                            <td>
                                <input
                                    v-model="sliceAddTimestamp"
                                    type="text"
                                    class="form-control form-control-full-width"
                                    :placeholder="timeSlicesArray.length > 0 ? '00:01:00' : '00:00:00'"
                                />
                            </td>
                            <td colspan="2">
                                <input
                                    v-model="sliceAddName"
                                    type="text"
                                    class="form-control form-control-full-width auto-focus"
                                    :placeholder="timeSlicesArray.length > 0 ? $t('Rest of the video') : $t('Opening')"
                                    @keydown="keyDownAdd"
                                />
                            </td>
                            <td class="td-shrink text-right">
                                <button type="button" class="time-slices-edit-helper-btn" :title="$t('Add')" @click="addSlice">
                                    <i class="fas fa-plus"></i>
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </ResizableWidget>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { clone } from "@/utils/objects";
import { makeApiRequest } from "@asanrom/request-browser";
import { renderTimeSeconds } from "@/utils/time";
import { parseTimeSeconds } from "@/utils/time-slices";
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

import ResizableWidget from "@/components/player/ResizableWidget.vue";
import { nextTick } from "vue";
import { PagesController } from "@/control/pages";
import { apiMediaChangeTimeSlices } from "@/api/api-media-edit";
import { MediaTimeSlice } from "@/api/models";

interface SaveRequestState {
    saving: boolean;
    pendingSave: boolean;
    mid: number;
    timeSlices: {
        time: number;
        name: string;
    }[];
    callback: () => void;
}

function saveTimeSlices(state: SaveRequestState, $t: (msg: string) => string) {
    if (state.mid < 0) {
        return;
    }

    if (state.saving) {
        state.pendingSave = true;
        return;
    }

    state.saving = true;

    makeApiRequest(apiMediaChangeTimeSlices(state.mid, state.timeSlices))
        .onSuccess(() => {
            state.saving = false;

            if (MediaController.MediaData && MediaController.MediaData.id === state.mid) {
                MediaController.MediaData.time_slices = clone(state.timeSlices);
            }

            if (state.pendingSave) {
                state.pendingSave = false;
                saveTimeSlices(state, $t);
            } else if (state.callback) {
                state.callback();
            }
        })
        .onCancel(() => {
            state.saving = false;

            if (state.pendingSave) {
                state.pendingSave = false;
                saveTimeSlices(state, $t);
            }
        })
        .onRequestError((err, handleErr) => {
            state.saving = false;
            handleErr(err, {
                unauthorized: () => {
                    AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                },
                badRequest: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Bad request"));
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                },
                notFound: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Not found"));
                    MediaController.Load();
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
            if (state.pendingSave) {
                state.pendingSave = false;
                saveTimeSlices(state, $t);
            }
        })
        .onUnexpectedError((err) => {
            console.error(err);
            state.saving = false;
            if (state.pendingSave) {
                state.pendingSave = false;
                saveTimeSlices(state, $t);
            }
        });
}

/**
 * Time slice being modified
 */
export type EditorTimeSlice = {
    // Start time
    time: number;

    // Start time (string)
    timeStr: string;

    // Slice name
    name: string;

    // True if deleted
    deleted: boolean;
};

export default defineComponent({
    name: "TimeSlicesEditHelper",
    components: {
        ResizableWidget,
    },
    props: {
        display: Boolean,
        contextOpen: Boolean,
        currentTime: Number,
    },
    emits: ["update:display", "update-time-slices", "clicked"],
    setup(props) {
        return {
            saveState: {
                saving: false,
                pendingSave: false,
                mid: -1,
                timeSlices: [] as MediaTimeSlice[],
                callback: null as (() => void) | null,
            },
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            timeSlicesArray: [] as EditorTimeSlice[],

            sliceAddTimestamp: "",
            sliceAddName: "",

            sliceEditIndex: -1,
            sliceEditTimestamp: "",
            sliceEditName: "",

            canWrite: AuthController.CanWrite,
        };
    },

    watch: {
        currentTime: function () {
            this.sliceAddTimestamp = renderTimeSeconds(this.currentTime || 0);
        },

        display: function () {
            if (this.display) {
                nextTick(() => {
                    const elem = this.$el.querySelector(".auto-focus");

                    if (elem) {
                        elem.focus();
                        elem.select();
                    }
                });
            }
        },
    },

    mounted: function () {
        this.saveState.callback = this.onSaved.bind(this);

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.sliceAddTimestamp = renderTimeSeconds(this.currentTime || 0);

        this.updateMediaData();

        if (this.display) {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");

                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        }
    },

    beforeUnmount: function () {
        this.saveState.callback = null;
    },

    methods: {
        propagateClick: function () {
            this.$emit("clicked");
        },

        reset: function () {
            this.sliceEditIndex = -1;
        },

        close: function () {
            this.displayStatus = false;
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.reset();

            this.timeSlicesArray = (MediaController.MediaData.time_slices || []).map((s) => {
                return {
                    time: s.time,
                    timeStr: renderTimeSeconds(s.time),
                    name: s.name,
                    deleted: false,
                };
            });

            this.saveState = {
                saving: false,
                pendingSave: false,
                mid: MediaController.MediaData.id,
                timeSlices: clone(MediaController.MediaData.time_slices),
                callback: this.onSaved.bind(this),
            };
        },

        save: function () {
            this.saveState.timeSlices = this.timeSlicesArray
                .filter((s) => {
                    return !s.deleted;
                })
                .map((s) => {
                    return {
                        time: s.time,
                        name: s.name,
                    };
                });
            saveTimeSlices(this.saveState, this.$t);
        },

        onSaved: function () {
            PagesController.ShowSnackBar(this.$t("Successfully changed time slices"));
            this.$emit("update-time-slices");
        },

        addSlice: function () {
            const timeSeconds = parseTimeSeconds(this.sliceAddTimestamp);

            if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
                return;
            }

            if (!this.sliceAddName) {
                return;
            }

            let foundSpace = false;
            const newSlice: EditorTimeSlice = {
                time: timeSeconds,
                timeStr: renderTimeSeconds(timeSeconds),
                name: this.sliceAddName,
                deleted: false,
            };

            for (let j = 0; j < this.timeSlicesArray.length; j++) {
                if (this.timeSlicesArray[j].time > timeSeconds) {
                    foundSpace = true;
                    this.timeSlicesArray.splice(j, 0, newSlice);
                    break;
                }
            }

            if (!foundSpace) {
                this.timeSlicesArray.push(newSlice);
            }

            this.sliceAddName = "";

            this.save();
        },

        deleteSlice: function (i: number) {
            this.timeSlicesArray[i].deleted = true;

            for (let j = 0; j < this.timeSlicesArray.length; j++) {
                if (j === i) {
                    continue;
                }
                if (this.timeSlicesArray[j].deleted) {
                    this.timeSlicesArray.splice(i, 1);
                }
            }

            this.save();
        },

        undoDelete: function (i: number) {
            this.timeSlicesArray[i].deleted = false;
            this.save();
        },

        editSlice: function (i: number) {
            this.sliceEditIndex = i;
            this.sliceEditTimestamp = this.timeSlicesArray[i].timeStr;
            this.sliceEditName = this.timeSlicesArray[i].name;
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");

                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        },

        cancelEdit: function () {
            this.sliceEditIndex = -1;
        },

        finishEdit: function () {
            if (this.sliceEditIndex < 0) {
                return;
            }

            const timeSeconds = parseTimeSeconds(this.sliceEditTimestamp);

            if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
                return;
            }

            this.timeSlicesArray.splice(this.sliceEditIndex, 1);

            let foundSpace = false;

            const newSlice: EditorTimeSlice = {
                time: timeSeconds,
                timeStr: renderTimeSeconds(timeSeconds),
                name: this.sliceEditName,
                deleted: false,
            };

            for (let j = 0; j < this.timeSlicesArray.length; j++) {
                if (this.timeSlicesArray[j].time > timeSeconds) {
                    foundSpace = true;
                    this.timeSlicesArray.splice(j, 0, newSlice);
                    break;
                }
            }

            if (!foundSpace) {
                this.timeSlicesArray.push(newSlice);
            }

            this.sliceEditIndex = -1;

            this.save();
        },

        keyDownAdd: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                this.addSlice();
            }
        },

        keyDownEdit: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                this.finishEdit();
            }
        },
    },
});
</script>
