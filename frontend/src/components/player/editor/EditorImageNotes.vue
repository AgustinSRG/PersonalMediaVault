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

        <div class="form-group" v-if="canWrite">
            <button type="button" class="btn btn-primary" :disabled="busy || !dirty" @click="changeImageNotes">
                <i class="fas fa-pencil-alt"></i> {{ $t("Change image notes") }}
            </button>
        </div>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { Request } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { NOTES_TEXT_SEPARATOR, imageNotesToText, textToImageNotes } from "@/utils/notes-format";
import { EVENT_NAME_IMAGE_NOTES_UPDATE, ImageNotesController } from "@/control/img-notes";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaSetNotes } from "@/api/api-media-edit";

export default defineComponent({
    components: {},
    name: "EditorImageNotes",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            imageNotesSeparator: NOTES_TEXT_SEPARATOR,

            imageNotes: "",

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

            Request.Pending(this.requestId, apiMediaSetNotes(mediaId, notes))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Successfully changed image notes"));
                    this.busy = false;
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
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    PagesController.ShowSnackBar(err.message);
                    console.error(err);
                    this.busy = false;
                });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },
    },

    mounted: function () {
        this.updateImageNotes();

        this.$listenOnAppEvent(EVENT_NAME_IMAGE_NOTES_UPDATE, this.updateImageNotes.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.autoFocus();
    },

    beforeUnmount: function () {
        Request.Abort(this.requestId);
    },
});
</script>
