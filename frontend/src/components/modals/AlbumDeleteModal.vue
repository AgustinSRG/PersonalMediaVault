<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Delete album") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <table class="table no-margin no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="deleteConfirmation"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the album by accident you would have to recreate it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy || !deleteConfirmation" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Delete album") }}
                </button>
            </div>
        </form>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="performRequest"
        ></AuthConfirmationModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { AlbumsController } from "@/control/albums";
import { EVENT_NAME_CURRENT_ALBUM_UPDATED } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef } from "vue";
import { PagesController } from "@/control/pages";
import { apiAlbumsDeleteAlbum } from "@/api/api-albums";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import AuthConfirmationModal from "./AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";
import { useModal } from "@/composables/use-modal";
import { useI18n } from "@/composables/use-i18n";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useAuthConfirmation } from "@/composables/use-auth-confirmation";
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

// Current album name
const currentAlbumName = ref(AlbumsController.CurrentAlbumData?.name || "");

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, () => {
    currentAlbum.value = AlbumsController.CurrentAlbum;
    currentAlbumName.value = AlbumsController.CurrentAlbumData?.name || "";
});

// Deletion confirmation
const deleteConfirmation = ref(false);

// Auth confirmation
const {
    displayAuthConfirmation,
    authConfirmationCooldown,
    authConfirmationTfa,
    authConfirmationError,
    requiredAuthConfirmationPassword,
    invalidPassword,
    requiredAuthConfirmationTfa,
    invalidTfaCode,
    cooldown,
} = useAuthConfirmation();

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

/**
 * Performs the request
 * @param confirmation The auth confirmation
 */
const performRequest = (confirmation: ProvidedAuthConfirmation) => {
    if (busy.value) {
        return;
    }

    busy.value = true;
    error.value = "";

    const albumId = currentAlbum.value;
    const albumName = currentAlbumName.value;

    makeApiRequest(apiAlbumsDeleteAlbum(albumId, confirmation))
        .onSuccess(() => {
            PagesController.ShowSnackBar($t("Album deleted") + ": " + albumName);

            busy.value = false;
            deleteConfirmation.value = false;

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
                requiredAuthConfirmationPassword,
                invalidPassword,
                requiredAuthConfirmationTfa,
                invalidTfaCode,
                cooldown,
                accessDenied,
                notFound: () => {
                    // Already deleted
                    PagesController.ShowSnackBar($t("Album deleted") + ": " + albumName);

                    deleteConfirmation.value = false;

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

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    performRequest({});
};
</script>
