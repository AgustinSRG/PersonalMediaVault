<template>
    <div ref="container" class="player-editor-sub-content">
        <!--- Time slices -->

        <div class="form-group">
            <label>{{ $t("Time slices") }}. {{ $t("You can split the time in slices and name them to display at the timeline.") }}</label>
        </div>

        <div class="form-group mt-1">
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

        <div v-if="canWrite" class="form-group">
            <button v-if="dirty || !saved || busy" type="button" class="btn btn-primary" :disabled="busy || !dirty" @click="saveChanges">
                <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Change time slices") }}
            </button>
            <button v-else type="button" disabled class="btn btn-primary">
                <i class="fas fa-check"></i> {{ $t("Saved time slices") }}
            </button>
        </div>

        <SaveChangesAskModal
            v-if="displayExitConfirmation"
            v-model:display="displayExitConfirmation"
            @yes="onExitSaveChanges"
            @no="onExitDiscardChanges"
        ></SaveChangesAskModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { MediaController } from "@/control/media";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import { parseTimeSlices, renderTimeSlices } from "@/utils/time-slices";
import { clone } from "@/utils/objects";
import { apiMediaChangeTimeSlices } from "@/api/api-media-edit";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useRequestId } from "@/composables/use-request-id";
import { useExitPreventer } from "@/composables/use-exit-preventer";
import { showSnackBarRight } from "@/control/snack-bar";

const SaveChangesAskModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SaveChangesAskModal.vue"),
});

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Emits
const emit = defineEmits<{
    /**
     * Media changed
     */
    (e: "changed"): void;
}>();

// Original time slices
const originalTimeSlices = ref(renderTimeSlices(MediaController.MediaData?.time_slices || []));

// New time slices
const timeSlices = ref(originalTimeSlices.value);

// Dirty? (unsaved changes)
const dirty = ref(false);

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, () => {
    if (!MediaController.MediaData) {
        return;
    }

    originalTimeSlices.value = renderTimeSlices(MediaController.MediaData.time_slices);
    timeSlices.value = originalTimeSlices.value;
    dirty.value = false;
});

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    nextTick(() => {
        const elem = container.value?.querySelector(".auto-focus") as HTMLElement;
        if (elem) {
            elem.focus();
        }
    });
};

onMounted(autoFocus);

/**
 * Indicates changes were made
 */
const markDirty = () => {
    dirty.value = true;
};

// Busy status
const busy = ref(false);

// True if saved
const saved = ref(false);

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

// Save request ID
const saveRequestId = useRequestId();

/**
 * Saves the changes
 */
const saveChanges = () => {
    if (busy.value) {
        return;
    }

    busy.value = true;

    const mediaId = AppStatus.CurrentMedia;

    const slices = parseTimeSlices(timeSlices.value);

    makeNamedApiRequest(saveRequestId, apiMediaChangeTimeSlices(mediaId, slices))
        .onSuccess(() => {
            showSnackBarRight($t("Successfully changed time slices"));

            busy.value = false;
            saved.value = true;
            dirty.value = false;
            originalTimeSlices.value = renderTimeSlices(slices);
            timeSlices.value = originalTimeSlices.value;

            if (MediaController.MediaData && MediaController.MediaData.id === mediaId) {
                MediaController.MediaData.time_slices = clone(slices);
            }

            emit("changed");

            onSave();
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                badRequest,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            setError(err.message);
            console.error(err);
            busy.value = false;
        });
};

// Exit preventer
const { displayExitConfirmation, onSave, onExitSaveChanges, onExitDiscardChanges } = useExitPreventer(dirty, saveChanges);
</script>
