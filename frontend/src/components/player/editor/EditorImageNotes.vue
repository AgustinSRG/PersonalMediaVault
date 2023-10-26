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
import { AuthController } from "@/control/auth";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { EditMediaAPI } from "@/api/api-media-edit";
import { NOTES_TEXT_SEPARATOR, imageNotesToText, textToImageNotes } from "@/utils/notes-format";
import { ImageNotesController } from "@/control/img-notes";

export default defineComponent({
    components: {},
    name: "EditorImageNotes",
    emits: ["changed"],
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

            Request.Pending("media-editor-busy-image-notes", EditMediaAPI.SetNotes(mediaId, notes))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully changed image notes"));
                    this.busy = false;
                    this.dirty = false;

                    ImageNotesController.Notes = notes;
                    AppEvents.Emit("img-notes-update");

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
        this.updateImageNotes();

        this._handles.updateImageNotesH = this.updateImageNotes.bind(this);

        AppEvents.AddEventListener("img-notes-update", this._handles.updateImageNotesH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AuthController.AddChangeEventListener(this._handles.authUpdateH);

        this.autoFocus();
    },

    beforeUnmount: function () {
        AppEvents.RemoveEventListener("img-notes-update", this._handles.updateImageNotesH);

        AuthController.RemoveChangeEventListener(this._handles.authUpdateH);

        Request.Abort("media-editor-busy-image-notes");
    },
});
</script>
