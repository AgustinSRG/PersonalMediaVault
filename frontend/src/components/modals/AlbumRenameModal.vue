<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Rename album") }}
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
                        class="form-control form-control-full-width auto-focus auto-select"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Rename album") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { AlbumsController } from "@/control/albums";
import { EVENT_NAME_CURRENT_ALBUM_UPDATED } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import { apiAlbumsRenameAlbum } from "@/api/api-albums";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose } = useModal(display, container);

// Current album ID
const currentAlbum = ref(AlbumsController.CurrentAlbum);

// Old name
const oldName = ref(AlbumsController.CurrentAlbumData?.name || "");

// New name
const name = ref(oldName.value);

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, () => {
    currentAlbum.value = AlbumsController.CurrentAlbum;
    oldName.value = AlbumsController.CurrentAlbumData?.name || "";
    name.value = oldName.value;
});

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, badRequest, accessDenied, serverError, networkError } = useCommonRequestErrors();

watch(display, () => {
    if (display.value) {
        name.value = oldName.value;
        error.value = "";
    }
});

/**
 * Handler for the 'submit' event
 * @param e The event
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

    if (name.value === oldName.value) {
        forceClose();
        return;
    }

    busy.value = true;
    error.value = "";

    const albumId = currentAlbum.value;

    makeApiRequest(apiAlbumsRenameAlbum(albumId, name.value))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Album renamed") + ": " + name.value);

            forceClose();

            AlbumsController.OnChangedAlbum(albumId);
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
                notFound: () => {
                    forceClose();
                    AlbumsController.OnChangedAlbum(albumId);
                },
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
