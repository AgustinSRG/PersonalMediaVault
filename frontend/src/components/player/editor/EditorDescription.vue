<template>
    <div class="player-editor-sub-content">
        <!--- Description -->

        <div class="form-group">
            <label>{{ $t("Add a description for the media.") }}</label>
        </div>

        <div class="form-group mt-1">
            <textarea
                v-model="description"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea auto-focus"
                :placeholder="getPlaceholder(loading)"
                rows="10"
                :disabled="loading || busy"
                @input="markDirty"
            ></textarea>
        </div>

        <div v-if="canWrite" class="form-group">
            <button
                v-if="!loading && (dirty || busy || !saved)"
                type="button"
                class="btn btn-primary"
                :disabled="busy || !dirty"
                @click="saveChanges"
            >
                <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Change description") }}
            </button>
            <button v-else type="button" disabled class="btn btn-primary">
                <i class="fas fa-check"></i> {{ $t("Saved description") }}
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
import {
    emitAppEvent,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_DESCRIPTION_UPDATE,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest, RequestErrorHandler, makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaSetDescription } from "@/api/api-media-edit";
import SaveChangesAskModal from "@/components/modals/SaveChangesAskModal.vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { ExitPreventer } from "@/control/exit-prevent";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";

export default defineComponent({
    name: "EditorDescription",
    components: {
        SaveChangesAskModal,
        LoadingIcon,
    },
    emits: ["changed"],
    setup() {
        return {
            loadRequestId: getUniqueStringId(),
            requestId: getUniqueStringId(),

            exitCallback: null as () => void,
        };
    },
    data: function () {
        return {
            loading: true,

            description: "",

            busy: false,
            saved: false,

            dirty: false,

            canWrite: AuthController.CanWrite,

            displayExitConfirmation: false,
            exitOnSave: false,
        };
    },

    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, this.updateDescription.bind(this));

        this.load();

        ExitPreventer.SetupExitPrevent(this.checkExitPrevent.bind(this), this.onExit.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);

        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);

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

        getPlaceholder: function (loading: boolean): string {
            if (loading) {
                return this.$t("Loading") + "...";
            } else {
                return this.$t("Example paragraph") + "\n" + this.$t("Example paragraph") + "\n";
            }
        },

        updateMediaData: function () {
            this.load();
        },

        updateDescription: function (source: string) {
            if (source === "editor") {
                return;
            }

            this.load();
        },

        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            this.description = "";

            if (!MediaController.MediaData) {
                return;
            }

            const descFilePath = MediaController.MediaData.description_url;

            if (!descFilePath) {
                this.description = "";
                this.loading = false;

                this.autoFocus();
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.loadRequestId, {
                method: "GET",
                url: getAssetURL(descFilePath),
            })
                .onSuccess((descriptionText) => {
                    this.description = descriptionText;
                    this.loading = false;
                    this.dirty = false;

                    this.autoFocus();
                })
                .onRequestError((err) => {
                    new RequestErrorHandler()
                        .add(401, "*", () => {
                            emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(404, "*", () => {
                            this.description = "";
                            this.loading = false;
                            this.dirty = false;
                            this.autoFocus();
                        })
                        .add("*", "*", () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        markDirty: function () {
            this.dirty = true;
        },

        saveChanges: function () {
            if (this.busy) {
                return;
            }

            if (!MediaController.MediaData) {
                return;
            }

            this.busy = true;

            const mid = MediaController.MediaData.id;

            makeApiRequest(apiMediaSetDescription(mid, this.description))
                .onSuccess((res) => {
                    this.busy = false;

                    this.dirty = false;

                    PagesController.ShowSnackBar(this.$t("Successfully saved description"));

                    if (MediaController.MediaData && MediaController.MediaData.id === mid) {
                        MediaController.MediaData.description_url = res.url || "";
                    }

                    emitAppEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, "editor");

                    this.$emit("changed");

                    if (this.exitOnSave) {
                        this.exitOnSave = false;
                        if (this.exitCallback) {
                            this.exitCallback();
                        }
                    }
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            emitAppEvent(EVENT_NAME_UNAUTHORIZED);
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
                    this.busy = false;
                    console.error(err);
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
                this.saveChanges();
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
