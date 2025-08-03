<template>
    <div class="player-editor-sub-content">
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
                    imageNotesSeparator +
                    '\n' +
                    $t('Example image note') +
                    '\n' +
                    $t('Another line') +
                    '\n' +
                    imageNotesSeparator +
                    '\n'
                "
                rows="10"
                :disabled="busy"
                @input="markDirty"
            ></textarea>
        </div>

        <div v-if="canWrite" class="form-group">
            <button
                v-if="dirty || busy || !saved"
                type="button"
                class="btn btn-primary"
                :disabled="busy || !dirty"
                @click="changeImageNotes"
            >
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
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { NOTES_TEXT_SEPARATOR, imageNotesToText, textToImageNotes } from "@/utils/notes-format";
import { EVENT_NAME_IMAGE_NOTES_UPDATE, ImageNotesController } from "@/control/img-notes";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaSetNotes } from "@/api/api-media-edit";
import SaveChangesAskModal from "@/components/modals/SaveChangesAskModal.vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { ExitPreventer } from "@/control/exit-prevent";

export default defineComponent({
    name: "EditorImageNotes",
    components: {
        SaveChangesAskModal,
        LoadingIcon,
    },
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),

            exitCallback: null as () => void,
        };
    },
    data: function () {
        return {
            imageNotesSeparator: NOTES_TEXT_SEPARATOR,

            imageNotes: "",

            busy: false,
            saved: false,

            dirty: false,

            canWrite: AuthController.CanWrite,

            displayExitConfirmation: false,
            exitOnSave: false,
        };
    },

    mounted: function () {
        this.updateImageNotes();

        this.$listenOnAppEvent(EVENT_NAME_IMAGE_NOTES_UPDATE, this.updateImageNotes.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.autoFocus();

        ExitPreventer.SetupExitPrevent(this.checkExitPrevent.bind(this), this.onExit.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);

        ExitPreventer.RemoveExitPrevent();
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

        markDirty: function () {
            this.dirty = true;
        },

        updateImageNotes: function () {
            this.imageNotes = imageNotesToText(ImageNotesController.Notes);
        },

        changeImageNotes: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = ImageNotesController.MediaId;

            const notes = textToImageNotes(this.imageNotes);

            makeNamedApiRequest(this.requestId, apiMediaSetNotes(mediaId, notes))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Successfully changed image notes"));
                    this.busy = false;
                    this.saved = true;
                    this.dirty = false;

                    ImageNotesController.Notes = notes;
                    AppEvents.Emit(EVENT_NAME_IMAGE_NOTES_UPDATE);

                    this.$emit("changed");
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBarRight(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        checkExitPrevent: function (): boolean {
            return this.dirty;
        },

        onExit: function (callback: () => void) {
            this.exitCallback = callback;
            this.displayExitConfirmation = true;
        },

        onExitSaveChanges: function () {
            if (this.dirty) {
                this.exitOnSave = true;
                this.changeImageNotes();
            } else {
                if (this.exitCallback) {
                    this.exitCallback();
                }
            }
        },

        onExitDiscardChanges: function () {
            if (this.exitCallback) {
                this.exitCallback();
            }
        },
    },
});
</script>
