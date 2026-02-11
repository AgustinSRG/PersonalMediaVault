<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Create new album") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Album name") }}:</label>
                    <input
                        v-model="name"
                        type="text"
                        name="album-name"
                        autocomplete="off"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-plus" :loading="busy"></LoadingIcon> {{ $t("Create album") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { AlbumsController } from "@/control/albums";
import { emitAppEvent, EVENT_NAME_ALBUMS_CHANGED } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef } from "vue";
import { PagesController } from "@/control/pages";
import { apiAlbumsCreateAlbum } from "@/api/api-albums";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose } = useModal(display, container);

// Events
const emit = defineEmits<{
    /**
     * Emitted when the album is created and the modal closed.
     */
    (e: "new-album", id: number, name: string): void;
}>();

// New album name
const name = ref("");

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, badRequest, accessDenied, serverError, networkError } = useCommonRequestErrors();

/**
 * Submits the form
 * @param e The 'submit' event
 */
const submit = (e: Event) => {
    e.preventDefault();

    if (busy.value) {
        return;
    }

    if (!name.value) {
        error.value = $t("Invalid album name provided");
        return;
    }

    if (AlbumsController.FindDuplicatedName(name.value)) {
        error.value = $t("There is already another album with the same name");
        return;
    }

    busy.value = true;
    error.value = "";

    const albumName = name.value;

    makeApiRequest(apiAlbumsCreateAlbum(albumName))
        .onSuccess((response) => {
            PagesController.ShowSnackBar($t("Album created") + ": " + albumName);

            busy.value = false;
            name.value = "";

            forceClose();

            emitAppEvent(EVENT_NAME_ALBUMS_CHANGED);

            AlbumsController.Load();

            emit("new-album", response.album_id, albumName);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;
            handleErr(err, {
                unauthorized,
                invalidName: () => {
                    error.value = $t("Invalid album name provided");
                },
                badRequest,
                accessDenied,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            error.value = err.message;

            console.error(err);
        });
};
</script>
