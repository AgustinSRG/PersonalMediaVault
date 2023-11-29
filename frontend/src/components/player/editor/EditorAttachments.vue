<template>
    <div class="player-editor-sub-content">
        <!--- Attachments -->

        <div class="form-group">
            <label>{{ $t("Attachments") }}. {{ $t("You can attach arbitrary files for safe keeping in the vault.") }}</label>
        </div>

        <div class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-left">{{ $t("Size") }}</th>
                        <th class="text-right"></th>
                        <th class="text-right td-shrink" v-if="canWrite"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="attachments.length === 0">
                        <td colspan="3">
                            {{ $t("There are no attachments yet for this media.") }}
                        </td>
                        <td class="text-right td-shrink" v-if="canWrite"></td>
                    </tr>
                    <tr v-for="att in attachments" :key="att.id">
                        <td class="bold" v-if="attachmentEdit != att.id">{{ att.name }}</td>
                        <td v-if="attachmentEdit == att.id">
                            <input
                                type="text"
                                maxlength="255"
                                :disabled="busy"
                                class="form-control form-control-full-width"
                                v-model="attachmentEditName"
                            />
                        </td>
                        <td>{{ renderSize(att.size) }}</td>
                        <td class="text-right td-shrink one-line">
                            <button type="button" class="btn btn-primary btn-xs mr-1" :disabled="busy" @click="downloadAttachment(att)">
                                <i class="fas fa-download"></i> {{ $t("Download") }}
                            </button>
                        </td>
                        <td class="text-right td-shrink one-line" v-if="canWrite">
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy"
                                @click="editAttachment(att)"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy"
                                @click="saveEditAttachment"
                            >
                                <i class="fas fa-check"></i> {{ $t("Save") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busy"
                                @click="cancelEditAttachment"
                            >
                                <i class="fas fa-times"></i> {{ $t("Cancel") }}
                            </button>
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy"
                                @click="removeAttachment(att)"
                            >
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="form-group" v-if="canWrite">
            <input type="file" class="file-hidden attachment-file-hidden" @change="attachmentFileChanged" name="attachment-upload" />
            <button
                v-if="!busy || attachmentUploadProgress <= 0"
                type="button"
                class="btn btn-primary"
                :disabled="busy"
                @click="selectAttachmentFile"
            >
                <i class="fas fa-upload"></i> {{ $t("Upload attachment") }}
            </button>
            <button
                v-if="busy && attachmentUploadProgress > 0 && attachmentUploadProgress < 100"
                type="button"
                class="btn btn-primary"
                disabled
            >
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ attachmentUploadProgress }}%)
            </button>
            <button v-if="busy && attachmentUploadProgress >= 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
        </div>

        <AttachmentDeleteModal
            v-model:display="displayAttachmentDelete"
            :attachment-to-delete="attachmentToDelete"
            @confirm="removeAttachmentConfirm"
        ></AttachmentDeleteModal>
    </div>
</template>

<script lang="ts">
import { MediaAttachment } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";

import AttachmentDeleteModal from "@/components/modals/AttachmentDeleteModal.vue";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaRemoveAttachment, apiMediaRenameAttachment, apiMediaUploadAttachment } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        AttachmentDeleteModal,
    },
    name: "EditorAttachments",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            attachments: [],
            attachmentUploadProgress: 0,
            attachmentEdit: -1,
            attachmentEditName: "",

            busy: false,

            canWrite: AuthController.CanWrite,

            displayAttachmentDelete: false,
            attachmentToDelete: null as MediaAttachment,
        };
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.attachments = (MediaController.MediaData.attachments || []).map((a) => {
                return {
                    id: a.id,
                    name: a.name,
                    size: a.size,
                    url: a.url,
                };
            });
            this.attachmentUploadProgress = 0;
            this.attachmentEdit = -1;
            this.attachmentEditName = "";
        },

        // Attachments

        selectAttachmentFile: function () {
            const fileElem = this.$el.querySelector(".attachment-file-hidden");
            if (fileElem) {
                fileElem.value = null;
                fileElem.click();
            }
        },

        attachmentFileChanged: function (e) {
            const data = e.target.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.addAttachment(file);
            }
        },

        addAttachment: function (file: File) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaUploadAttachment(mediaId, file))
                .onSuccess((res) => {
                    PagesController.ShowSnackBar(this.$t("Added attachment") + ": " + res.name);
                    this.busy = false;
                    this.attachmentUploadProgress = 0;
                    this.attachments.push(res);

                    if (MediaController.MediaData) {
                        MediaController.MediaData.attachments = clone(this.attachments);
                    }

                    this.$emit("changed");
                })
                .onUploadProgress((loaded, total) => {
                    if (total) {
                        this.attachmentUploadProgress = Math.floor(((loaded * 100) / total) * 100) / 100;
                    }
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

        removeAttachment: function (att: MediaAttachment) {
            this.attachmentToDelete = att;
            this.displayAttachmentDelete = true;
        },

        removeAttachmentConfirm: function () {
            const att = this.attachmentToDelete;

            if (this.busy || !att) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const id = att.id;

            makeNamedApiRequest(this.requestId, apiMediaRemoveAttachment(mediaId, id))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Removed attachment") + ": " + att.name);
                    this.busy = false;
                    for (let i = 0; i < this.attachments.length; i++) {
                        if (this.attachments[i].id === id) {
                            this.attachments.splice(i, 1);
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.attachments = clone(this.attachments);
                    }
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

        downloadAttachment: function (att: MediaAttachment) {
            const link = document.createElement("a");
            link.target = "_blank";
            link.rel = "noopener noreferrer";
            link.href = getAssetURL(att.url);
            link.click();
        },

        editAttachment: function (att: MediaAttachment) {
            this.attachmentEdit = att.id;
            this.attachmentEditName = att.name;
        },

        cancelEditAttachment: function () {
            this.attachmentEdit = -1;
            this.attachmentEditName = "";
        },

        saveEditAttachment: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;

            const mediaId = AppStatus.CurrentMedia;
            const id = this.attachmentEdit;

            makeNamedApiRequest(this.requestId, apiMediaRenameAttachment(mediaId, id, this.attachmentEditName))
                .onSuccess((res) => {
                    PagesController.ShowSnackBar(this.$t("Renamed attachment") + ": " + res.name);
                    this.busy = false;
                    this.attachmentEdit = -1;
                    this.attachmentEditName = "";
                    for (let i = 0; i < this.attachments.length; i++) {
                        if (this.attachments[i].id === res.id) {
                            this.attachments[i].name = res.name;
                            this.attachments[i].url = res.url;
                            break;
                        }
                    }
                    if (MediaController.MediaData) {
                        MediaController.MediaData.attachments = clone(this.attachments);
                    }
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
                        invalidName: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Invalid attachment name"));
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
            if (!this.canWrite) {
                this.attachmentEdit = -1;
            }
        },

        renderSize: function (bytes: number): string {
            if (bytes > 1024 * 1024 * 1024) {
                let gb = bytes / (1024 * 1024 * 1024);
                gb = Math.floor(gb * 100) / 100;
                return gb + " GB";
            } else if (bytes > 1024 * 1024) {
                let mb = bytes / (1024 * 1024);
                mb = Math.floor(mb * 100) / 100;
                return mb + " MB";
            } else if (bytes > 1024) {
                let kb = bytes / 1024;
                kb = Math.floor(kb * 100) / 100;
                return kb + " KB";
            } else {
                return bytes + " Bytes";
            }
        },
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);
    },
});
</script>
