<template>
    <div ref="container" class="player-editor-sub-content">
        <!--- Image notes -->

        <div class="form-group">
            <label
                >{{ $t("Image notes") }}.
                {{ $t("You can add notes to your image, in order to add extra information or translations.") }}</label
            >
        </div>

        <div class="form-group mt-1">
            <textarea
                v-model="imageNotes"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea auto-focus"
                :placeholder="
                    '[50, 100] (240 x 360)\n' +
                    NOTES_TEXT_SEPARATOR +
                    '\n' +
                    $t('Example image note') +
                    '\n' +
                    $t('Another line') +
                    '\n' +
                    NOTES_TEXT_SEPARATOR +
                    '\n'
                "
                rows="10"
                :disabled="busy"
                @input="markDirty"
            ></textarea>
        </div>

        <div v-if="canWrite" class="form-group">
            <button v-if="dirty || busy || !saved" type="button" class="btn btn-primary" :disabled="busy || !dirty" @click="saveChanges">
                <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Change image notes") }}
            </button>
            <button v-else type="button" disabled class="btn btn-primary">
                <i class="fas fa-check"></i> {{ $t("Saved image notes") }}
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
import { emitAppEvent, EVENT_NAME_IMAGE_NOTES_UPDATE } from "@/control/app-events";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import { NOTES_TEXT_SEPARATOR, imageNotesToText, textToImageNotes } from "@/utils/notes-format";
import { ImageNotesController } from "@/control/img-notes";
import { PagesController } from "@/control/pages";
import { apiMediaSetNotes } from "@/api/api-media-edit";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useRequestId } from "@/composables/use-request-id";
import { useExitPreventer } from "@/composables/use-exit-preventer";

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

// Image notes
const imageNotes = ref(imageNotesToText(ImageNotesController.Notes));

onApplicationEvent(EVENT_NAME_IMAGE_NOTES_UPDATE, () => {
    imageNotes.value = imageNotesToText(ImageNotesController.Notes);
});

// Dirty? (unsaved changes)
const dirty = ref(false);

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

    const mediaId = ImageNotesController.MediaId;

    const notes = textToImageNotes(imageNotes.value);

    makeNamedApiRequest(saveRequestId, apiMediaSetNotes(mediaId, notes))
        .onSuccess(() => {
            PagesController.ShowSnackBarRight($t("Successfully changed image notes"));

            busy.value = false;
            saved.value = true;
            dirty.value = false;

            ImageNotesController.Notes = notes;
            emitAppEvent(EVENT_NAME_IMAGE_NOTES_UPDATE);

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
