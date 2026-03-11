<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            v-model:display="display"
            :title="$t('Time slices')"
            :context-open="contextOpen"
            :position-key="'time-slices-helper-pos'"
            @clicked="propagateClick"
        >
            <div ref="container" class="table-responsive time-slices-table">
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

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { clone } from "@/utils/objects";
import { makeApiRequest } from "@asanrom/request-browser";
import { renderTimeSeconds } from "@/utils/time";
import { parseTimeSeconds } from "@/utils/time-slices";
import { defineAsyncComponent, onBeforeUnmount, onMounted, ref, shallowRef, useTemplateRef, watch } from "vue";
import ResizableWidget from "@/components/widgets/common/ResizableWidget.vue";
import { nextTick } from "vue";
import { apiMediaChangeTimeSlices } from "@/api/api-media-edit";
import type { MediaTimeSlice } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { showSnackBar } from "@/control/snack-bar";
import { getCurrentMediaData, loadCurrentMedia, modifyCurrentMediaData } from "@/control/media";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Display model
const display = defineModel<boolean>("display");

// Props
const props = defineProps({
    /**
     * A context menu is opened
     */
    contextOpen: Boolean,

    /**
     * The current time of the video (seconds)
     */
    currentTime: Number,
});

// Emits
const emit = defineEmits<{
    /**
     * The user updated the time slices
     */
    (e: "update-time-slices"): void;

    /**
     * The user clicked the widget
     */
    (e: "clicked"): void;
}>();

// Array of time slices
const timeSlicesArray = ref<EditorTimeSlice[]>([]);

/**
 * Updates the media data
 */
const updateMediaData = () => {
    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        return;
    }

    reset();

    timeSlicesArray.value = (mediaData.time_slices || []).map((s) => {
        return {
            time: s.time,
            timeStr: renderTimeSeconds(s.time),
            name: s.name,
            deleted: false,
        };
    });

    saveState.value.callback = null;

    saveState.value = {
        saving: false,
        pendingSave: false,
        mid: mediaData.id,
        timeSlices: clone(mediaData.time_slices),
        callback: onSaved,
    };
};

onMounted(updateMediaData);
onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, updateMediaData);

/**
 * Save request state for time slices
 */
interface SaveRequestState {
    /**
     * True if saving
     */
    saving: boolean;

    /**
     * True if a save process is pending
     */
    pendingSave: boolean;

    /**
     * The media ID
     */
    mid: number;

    /**
     * List of time slices
     */
    timeSlices: {
        time: number;
        name: string;
    }[];

    /**
     * Callback
     */
    callback: () => void;
}

/**
 * Called after the time slices are saved
 */
const onSaved = () => {
    showSnackBar($t("Successfully changed time slices"));
    emit("update-time-slices");
};

// Save state
const saveState = shallowRef<SaveRequestState>({
    saving: false,
    pendingSave: false,
    mid: -1,
    timeSlices: [] as MediaTimeSlice[],
    callback: onSaved,
});

onBeforeUnmount(() => {
    saveState.value.callback = null;
});

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

/**
 * Saves the time slices
 * @param ss The save state
 */
const saveTimeSlices = (ss: SaveRequestState) => {
    if (ss.mid < 0) {
        return;
    }

    if (ss.saving) {
        ss.pendingSave = true;
        return;
    }

    ss.saving = true;

    makeApiRequest(apiMediaChangeTimeSlices(ss.mid, ss.timeSlices))
        .onSuccess(() => {
            ss.saving = false;

            modifyCurrentMediaData(ss.mid, (metadata) => {
                metadata.time_slices = clone(ss.timeSlices);
            });

            if (ss.pendingSave) {
                ss.pendingSave = false;
                saveTimeSlices(ss);
            } else if (ss.callback) {
                ss.callback();
            }
        })
        .onCancel(() => {
            ss.saving = false;

            if (ss.pendingSave) {
                ss.pendingSave = false;
                saveTimeSlices(ss);
            }
        })
        .onRequestError((err, handleErr) => {
            ss.saving = false;
            handleErr(err, {
                unauthorized,
                badRequest,
                accessDenied,
                notFound: () => {
                    notFound();
                    loadCurrentMedia();
                },
                serverError,
                networkError,
            });
            if (ss.pendingSave) {
                ss.pendingSave = false;
                saveTimeSlices(ss);
            }
        })
        .onUnexpectedError((err) => {
            console.error(err);
            ss.saving = false;
            setError(err.message);
            if (ss.pendingSave) {
                ss.pendingSave = false;
                saveTimeSlices(ss);
            }
        });
};

/**
 * Requests the time slices to be saved
 */
const save = () => {
    saveState.value.timeSlices = timeSlicesArray.value
        .filter((s) => {
            return !s.deleted;
        })
        .map((s) => {
            return {
                time: s.time,
                name: s.name,
            };
        });
    saveTimeSlices(saveState.value);
};

/**
 * Propagates the click event to the parent element
 */
const propagateClick = () => {
    emit("clicked");
};

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

// Timestamp for the time slice to add
const sliceAddTimestamp = ref(renderTimeSeconds(props.currentTime || 0));

watch(
    () => props.currentTime,
    () => {
        sliceAddTimestamp.value = renderTimeSeconds(props.currentTime || 0);
    },
);

// Name for the time slice to add
const sliceAddName = ref("");

/**
 * Adds a time slice
 */
const addSlice = () => {
    const timeSeconds = parseTimeSeconds(sliceAddTimestamp.value);

    if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
        return;
    }

    if (!sliceAddName.value) {
        return;
    }

    let foundSpace = false;

    const newSlice: EditorTimeSlice = {
        time: timeSeconds,
        timeStr: renderTimeSeconds(timeSeconds),
        name: sliceAddName.value,
        deleted: false,
    };

    for (let j = 0; j < timeSlicesArray.value.length; j++) {
        if (timeSlicesArray.value[j].time > timeSeconds) {
            foundSpace = true;
            timeSlicesArray.value.splice(j, 0, newSlice);
            break;
        }
    }

    if (!foundSpace) {
        timeSlicesArray.value.push(newSlice);
    }

    sliceAddName.value = "";

    save();
};

/**
 * Handler for 'keydown' on add input
 * @param event The keyboard event
 */
const keyDownAdd = (event: KeyboardEvent) => {
    if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        addSlice();
    }
};

/**
 * Deletes a time slice
 * @param i The time slice index
 */
const deleteSlice = (i: number) => {
    timeSlicesArray.value[i].deleted = true;

    for (let j = 0; j < timeSlicesArray.value.length; j++) {
        if (j === i) {
            continue;
        }
        if (timeSlicesArray.value[j].deleted) {
            timeSlicesArray.value.splice(i, 1);
        }
    }

    save();
};

/**
 * Undoes a deletion
 * @param i The index of the deleted time slice
 */
const undoDelete = (i: number) => {
    timeSlicesArray.value[i].deleted = false;

    save();
};

// Index of the slice being edited
const sliceEditIndex = ref(-1);

// Timestamp of the slice being edited
const sliceEditTimestamp = ref("");

// Name of the slice being edited
const sliceEditName = ref("");

/**
 * Resets the edit state
 */
const reset = () => {
    sliceEditIndex.value = -1;
};

/**
 * Start editing an slice
 * @param i The index of the slice
 */
const editSlice = (i: number) => {
    sliceEditIndex.value = i;
    sliceEditTimestamp.value = timeSlicesArray.value[i].timeStr;
    sliceEditName.value = timeSlicesArray.value[i].name;

    autoFocus();
};

/**
 * Cancels editing
 */
const cancelEdit = () => {
    sliceEditIndex.value = -1;
};

/**
 * Finish editing a time slice
 */
const finishEdit = () => {
    if (sliceEditIndex.value < 0) {
        return;
    }

    const timeSeconds = parseTimeSeconds(sliceEditTimestamp.value);

    if (isNaN(timeSeconds) || !isFinite(timeSeconds) || timeSeconds < 0) {
        return;
    }

    timeSlicesArray.value.splice(sliceEditIndex.value, 1);

    let foundSpace = false;

    const newSlice: EditorTimeSlice = {
        time: timeSeconds,
        timeStr: renderTimeSeconds(timeSeconds),
        name: sliceEditName.value,
        deleted: false,
    };

    for (let j = 0; j < timeSlicesArray.value.length; j++) {
        if (timeSlicesArray.value[j].time > timeSeconds) {
            foundSpace = true;
            timeSlicesArray.value.splice(j, 0, newSlice);
            break;
        }
    }

    if (!foundSpace) {
        timeSlicesArray.value.push(newSlice);
    }

    sliceEditIndex.value = -1;

    save();
};

/**
 * Event handler for 'keydown' on the slice edit input
 * @param event The keyboard event
 */
const keyDownEdit = (event: KeyboardEvent) => {
    if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        finishEdit();
    }
};

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    nextTick(() => {
        const elem = container.value?.querySelector(".auto-focus") as HTMLInputElement;

        if (elem) {
            elem.focus();
            if (elem.select) {
                elem.select();
            }
        }
    });
};

onMounted(() => {
    if (display.value) {
        autoFocus();
    }
});

watch(display, () => {
    if (display.value) {
        autoFocus();
    }
});
</script>
