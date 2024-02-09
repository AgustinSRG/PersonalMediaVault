<template>
    <div class="resizable-widget-container">
        <ResizableWidget
            :title="$t('Extended description')"
            v-model:display="displayStatus"
            :contextOpen="contextOpen"
            :position-key="'ext-desc-widget-pos'"
            @clicked="propagateClick"
            :busy="busy"
            :action-buttons="actionButtons"
            @action-btn="clickActionButton"
        >
            <div class="extended-description-body" tabindex="-1">
                <LoadingOverlay v-if="loading"></LoadingOverlay>
                <div v-if="!loading && editing" class="extended-description-edit">
                    <textarea
                        :disabled="busy || !canWrite"
                        class="form-control form-textarea auto-focus"
                        v-model="contentToChange"
                        :placeholder="$t('Input your description here') + '...'"
                    ></textarea>
                </div>
                <div
                    v-if="!loading && !editing"
                    class="extended-description-view"
                    :style="{ '--base-font-size': baseFontSize + 'px' }"
                    v-html="renderContent(content)"
                ></div>
            </div>
        </ResizableWidget>
    </div>
</template>

<script lang="ts">
import { useVModel } from "@/utils/v-model";
import { defineComponent } from "vue";

import ResizableWidget, { ActionButton } from "@/components/player/ResizableWidget.vue";
import { nextTick } from "vue";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { AppStatus } from "@/control/app-status";
import { AppEvents } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest, RequestErrorHandler, makeApiRequest } from "@asanrom/request-browser";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { getAssetURL } from "@/utils/api";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { apiMediaSetExtendedDescription } from "@/api/api-media-edit";
import { escapeHTML } from "@/utils/html";

import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import { getExtendedDescriptionSize, setExtendedDescriptionSize } from "@/control/player-preferences";

export default defineComponent({
    components: {
        ResizableWidget,
        LoadingOverlay,
    },
    name: "ExtendedDescriptionWidget",
    emits: ["update:display", "tags-update", "clicked"],
    props: {
        display: Boolean,
        contextOpen: Boolean,
        currentTime: Number,
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

            editing: false,

            content: "",
            contentToChange: "",

            contentStoredId: -1,
            contentStored: "",

            loading: true,
            busy: false,
            canWrite: AuthController.CanWrite,

            baseFontSize: getExtendedDescriptionSize(),
        };
    },
    computed: {
        actionButtons: function (): ActionButton[] {
            const buttons: ActionButton[] = [];

            if (this.loading) {
                return [];
            }

            if (!this.editing) {
                if (this.baseFontSize !== 18) {
                    buttons.push({
                        id: "size-reset",
                        name: this.$t("Reset font size"),
                        icon: "fas fa-magnifying-glass",
                        key: ["r", "R"],
                    });
                }

                buttons.push({
                    id: "size-minus",
                    name: this.$t("Smaller font size"),
                    icon: "fas fa-magnifying-glass-minus",
                    key: "-",
                });

                buttons.push({
                    id: "size-plus",
                    name: this.$t("Bigger font size"),
                    icon: "fas fa-magnifying-glass-plus",
                    key: "+",
                });
            }

            if (this.canWrite) {
                if (this.editing) {
                    buttons.push({
                        id: "save",
                        name: this.$t("Save changes"),
                        icon: "fas fa-check",
                    });
                } else {
                    buttons.push({
                        id: "edit",
                        name: this.$t("Edit"),
                        icon: "fas fa-pencil-alt",
                    });
                }
            }

            return buttons;
        },
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

        clickActionButton: function (id: string) {
            switch (id) {
                case "save":
                    this.saveChanges();
                    break;
                case "edit":
                    this.startEdit();
                    break;
                case "size-reset":
                    this.baseFontSize = 18;
                    this.saveBaseFontSize();
                    break;
                case "size-plus":
                    this.baseFontSize = Math.min(128, this.baseFontSize + 1);
                    this.saveBaseFontSize();
                    break;
                case "size-minus":
                    this.baseFontSize = Math.max(1, this.baseFontSize - 1);
                    this.saveBaseFontSize();
                    break;
            }
        },

        saveBaseFontSize: function () {
            setExtendedDescriptionSize(this.baseFontSize);
        },

        propagateClick: function () {
            this.$emit("clicked");
        },

        close: function () {
            this.displayStatus = false;
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

        updateMediaData: function () {
            this.mid = AppStatus.CurrentMedia;
            this.load();
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
                    MediaController.Load();
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
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));

        if (this.display) {
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
                this.load();
            }
        },
    },
});
</script>
