<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" @close="onClose" :lock-close="busy" @key="onKeyPress">
        <div
            v-if="display"
            class="modal-dialog modal-height-100-wf"
            :class="{ 'modal-lg': modalSize === 'lg', 'modal-xl': modalSize === 'xl' }"
            role="document"
        >
            <div class="modal-header">
                <div class="modal-title">{{ title || $t("Extended description") }}</div>
                <button
                    type="button"
                    class="modal-close-btn"
                    :title="modalSize === 'xl' ? $t('Compress') : $t('Expand')"
                    @click="switchModalSize"
                >
                    <i v-if="modalSize === 'lg'" class="fas fa-expand"></i>
                    <i v-if="modalSize === 'xl'" class="fas fa-compress"></i>
                </button>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body no-padding">
                <LoadingOverlay v-if="loading"></LoadingOverlay>
                <div v-if="!loading && editing" class="modal-body-textarea-container">
                    <textarea
                        :disabled="busy"
                        class="form-control form-textarea no-resize auto-focus"
                        v-model="contentToChange"
                        :placeholder="$t('Input your description here') + '...'"
                    ></textarea>
                </div>
                <div v-if="!loading && !editing" class="extended-description-container" v-html="renderContent(content)"></div>
            </div>

            <div class="modal-footer text-right">
                <button
                    v-if="canWrite && !editing"
                    type="button"
                    @click="startEdit"
                    :disabled="busy || loading"
                    class="btn btn-primary btn-mr"
                >
                    <i class="fas fa-pencil-alt"></i> {{ $t("Edit") }}
                </button>
                <button
                    v-if="canWrite && editing"
                    type="button"
                    @click="cancelEdit"
                    :disabled="busy || loading"
                    class="btn btn-primary btn-mr"
                >
                    <i class="fas fa-times"></i> {{ $t("Cancel") }}
                </button>
                <button v-if="canWrite && editing" type="button" @click="saveChanges" :disabled="busy || loading" class="btn btn-primary">
                    <i class="fas fa-check"></i> {{ $t("Save changes") }}
                </button>
                <button v-if="!editing" type="button" @click="close" :disabled="busy" class="btn btn-primary">
                    <i class="fas fa-check"></i> {{ $t("Done") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";

import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { getAssetURL } from "@/utils/api";
import { RequestErrorHandler, abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { escapeHTML } from "@/utils/html";
import { getExtendedDescriptionSize, setExtendedDescriptionSize } from "@/control/player-preferences";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaSetExtendedDescription } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        LoadingOverlay,
    },
    name: "ExtendedDescriptionModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            mid: AppStatus.CurrentMedia,
            title: MediaController.MediaData ? MediaController.MediaData.title : "",

            editing: false,

            content: "",
            contentToChange: "",

            contentStoredId: -1,
            contentStored: "",

            loading: true,
            busy: false,
            canWrite: AuthController.CanWrite,

            modalSize: "xl",

            changed: false,
        };
    },
    methods: {
        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            if (!MediaController.MediaData) {
                return;
            }

            const descFilePath = MediaController.MediaData.ext_desc_url;

            if (!descFilePath) {
                this.content = "";
                this.contentToChange = "";
                this.loading = false;
                this.editing = !!this.canWrite;

                if (this.contentStoredId === MediaController.MediaData.id) {
                    this.contentToChange = this.contentStored;
                } else {
                    this.contentStoredId = -1;
                    this.contentStored = "";
                }

                this.autoFocus();
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.loadRequestId, {
                method: "GET",
                url: getAssetURL(descFilePath),
            })
                .onSuccess((extendedDescText) => {
                    this.content = extendedDescText;
                    this.contentToChange = extendedDescText;
                    this.loading = false;
                    this.editing = this.canWrite && !this.content;

                    if (this.contentStoredId === MediaController.MediaData.id) {
                        this.contentToChange = this.contentStored;
                        this.editing = !!this.canWrite;
                    } else {
                        this.contentStoredId = -1;
                        this.contentStored = "";
                    }

                    this.autoFocus();
                })
                .onRequestError((err) => {
                    new RequestErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED, false);
                        })
                        .add(404, "*", () => {
                            this.content = "";
                            this.contentToChange = "";
                            this.loading = false;
                            this.editing = !!this.canWrite;

                            if (this.contentStoredId === MediaController.MediaData.id) {
                                this.contentToChange = this.contentStored;
                            } else {
                                this.contentStoredId = -1;
                                this.contentStored = "";
                            }
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

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                } else {
                    this.$el.focus();
                }
            });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
            if (!this.canWrite) {
                this.cancelEdit();
            }
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        onClose: function () {
            if (this.changed) {
                MediaController.Load();
            }
        },

        updateMediaData: function () {
            this.mid = AppStatus.CurrentMedia;
            this.title = MediaController.MediaData ? MediaController.MediaData.title : "";
            this.load();
        },

        onKeyPress: function (e: KeyboardEvent) {
            if (this.canWrite && (e.key === "e" || e.key === "E") && !this.editing) {
                e.preventDefault();
                this.startEdit();
            }
        },

        startEdit: function () {
            if (this.editing) {
                return;
            }
            this.editing = true;
            this.autoFocus();
        },

        cancelEdit: function () {
            if (!this.editing) {
                return;
            }
            this.contentToChange = this.content;
            this.editing = false;
            this.autoFocus();
        },

        renderContent: function (text: string): string {
            return text
                .split("\n\n")
                .map((paragraph) => {
                    if (paragraph.startsWith("###")) {
                        return "<h3>" + escapeHTML(paragraph.substring(3)).replace(/\n/g, "<br>") + "</h3>";
                    } else if (paragraph.startsWith("##")) {
                        return "<h2>" + escapeHTML(paragraph.substring(2)).replace(/\n/g, "<br>") + "</h2>";
                    } else if (paragraph.startsWith("#")) {
                        return "<h1>" + escapeHTML(paragraph.substring(1)).replace(/\n/g, "<br>") + "</h1>";
                    } else {
                        return "<p>" + escapeHTML(paragraph).replace(/\n/g, "<br>") + "</p>";
                    }
                })
                .join("");
        },

        saveChanges: function () {
            if (!this.editing) {
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(apiMediaSetExtendedDescription(this.mid, this.contentToChange))
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Successfully saved extended description"));
                    this.content = this.contentToChange;
                    this.editing = false;
                    this.changed = true;
                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.contentStoredId = this.mid;
                            this.contentStored = this.contentToChange;
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
                    this.busy = false;
                    console.error(err);
                });
        },

        updateModalSize: function () {
            this.modalSize = getExtendedDescriptionSize();
            if (!["lg", "xl"].includes(this.modalSize)) {
                this.modalSize = "xl";
            }
        },

        switchModalSize: function () {
            if (this.modalSize === "lg") {
                this.modalSize = "xl";
            } else {
                this.modalSize = "lg";
            }
            setExtendedDescriptionSize(this.modalSize);
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));

        if (this.display) {
            this.updateModalSize();
            this.load();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    watch: {
        display: function () {
            if (this.display) {
                this.contentStoredId = -1;
                this.contentStored = "";
                this.updateModalSize();
                this.load();
            }
        },
    },
});
</script>
