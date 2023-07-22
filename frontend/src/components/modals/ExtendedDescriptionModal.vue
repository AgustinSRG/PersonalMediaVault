<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" @close="onClose" :lock-close="busy">
        <div
            v-if="display"
            class="modal-dialog modal-height-100-wf"
            :class="{ 'modal-lg': modalSize === 'lg', 'modal-xl': modalSize === 'xl' }"
            role="document"
        >
            <div class="modal-header">
                <div class="modal-title with-2-buttons">{{ title || $t("Extended description") }}</div>
                <button v-if="modalSize === 'lg'" type="button" class="modal-close-btn" :title="$t('Expand')" @click="switchModalSize">
                    <i class="fas fa-expand"></i>
                </button>
                <button v-if="modalSize === 'xl'" type="button" class="modal-close-btn" :title="$t('Compress')" @click="switchModalSize">
                    <i class="fas fa-compress"></i>
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
import { AuthController } from "@/control/auth";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { MediaController } from "@/control/media";

import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import { Timeouts } from "@/utils/timeout";
import { GetAssetURL, Request } from "@/utils/request";
import { MediaAPI } from "@/api/api-media";
import { escapeHTML } from "@/utils/html";
import { PlayerPreferences } from "@/control/player-preferences";

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

            loading: true,
            busy: false,
            canWrite: AuthController.CanWrite,

            modalSize: "xl",

            changed: false,
        };
    },
    methods: {
        load: function () {
            Timeouts.Abort("media-ext-desc-load");
            Request.Abort("media-ext-desc-load");

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
                this.autoFocus();
                return;
            }

            this.loading = true;

            Request.Pending("media-ext-desc-load", {
                method: "GET",
                url: GetAssetURL(descFilePath),
            })
                .onSuccess((extendedDescText) => {
                    this.content = extendedDescText;
                    this.contentToChange = extendedDescText;
                    this.loading = false;
                    this.editing = this.canWrite && !this.content;
                    this.autoFocus();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized", false);
                        })
                        .add(404, "*", () => {
                            this.content = "";
                            this.contentToChange = "";
                            this.loading = false;
                            this.editing = !!this.canWrite;
                            this.autoFocus();
                        })
                        .add("*", "*", () => {
                            // Retry
                            Timeouts.Set("media-ext-desc-load", 1500, this.load.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    Timeouts.Set("media-ext-desc-load", 1500, this.load.bind(this));
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

        startEdit: function () {
            this.editing = true;
        },

        cancelEdit: function () {
            this.contentToChange = this.content;
            this.editing = false;
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
            if (this.busy) {
                return;
            }

            this.busy = true;

            Request.Do(MediaAPI.SetExtendedDescription(this.mid, this.contentToChange))
                .onSuccess(() => {
                    this.busy = false;
                    AppEvents.Emit("snack", this.$t("Successfully saved extended description"));
                    this.content = this.contentToChange;
                    this.editing = false;
                    this.changed = true;
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized");
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                });
        },

        updateModalSize: function () {
            this.modalSize = PlayerPreferences.ExtendedDescriptionSize;
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
            PlayerPreferences.SetExtendedDescriptionSize(this.modalSize);
        },
    },
    mounted: function () {
        this.$options.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this.$options.authUpdateH);

        this.$options.mediaUpdateH = this.updateMediaData.bind(this);

        AppEvents.AddEventListener("current-media-update", this.$options.mediaUpdateH);

        if (this.display) {
            this.updateModalSize();
            this.load();
        }
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("auth-status-changed", this.$options.authUpdateH);
        AppEvents.RemoveEventListener("current-media-update", this.$options.mediaUpdateH);
    },
    watch: {
        display: function () {
            if (this.display) {
                this.updateModalSize();
                this.load();
            }
        },
    },
});
</script>
