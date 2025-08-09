<template>
    <div class="player-editor-sub-content" @drop="onDrop">
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
                        <th v-if="canWrite" class="text-right td-shrink"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="attachments.length === 0">
                        <td colspan="3">
                            {{ $t("There are no attachments yet for this media.") }}
                        </td>
                        <td v-if="canWrite" class="text-right td-shrink"></td>
                    </tr>
                    <tr v-for="att in attachments" :key="att.id">
                        <td v-if="attachmentEdit != att.id" class="bold">
                            <a :href="getAttachmentUrl(att)" target="_blank" rel="noopener noreferrer">{{ att.name }}</a>
                        </td>
                        <td v-if="attachmentEdit == att.id">
                            <input
                                v-model="attachmentEditName"
                                type="text"
                                maxlength="255"
                                :disabled="busyRename || busyDelete"
                                class="form-control form-control-full-width edit-auto-focus"
                                @keydown="editInputKeyEventHandler"
                            />
                        </td>
                        <td class="one-line">{{ renderSize(att.size) }}</td>
                        <td v-if="canWrite" class="text-right td-shrink one-line">
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busyRename || busyDelete"
                                @click="editAttachment(att)"
                            >
                                <i class="fas fa-pencil-alt"></i> {{ $t("Rename") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busyRename || busyDelete"
                                @click="saveEditAttachment"
                            >
                                <LoadingIcon icon="fas fa-check" :loading="busyRename"></LoadingIcon> {{ $t("Save") }}
                            </button>
                            <button
                                v-if="attachmentEdit == att.id"
                                type="button"
                                class="btn btn-primary btn-xs mr-1"
                                :disabled="busyRename || busyDelete"
                                @click="cancelEditAttachment"
                            >
                                <i class="fas fa-times"></i> {{ $t("Cancel") }}
                            </button>
                            <button
                                v-if="attachmentEdit != att.id"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busyRename || busyDelete"
                                @click="removeAttachment(att)"
                            >
                                <LoadingIcon icon="fas fa-trash-alt" :loading="busyDelete && busyDeleteId === att.id"></LoadingIcon>
                                {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="canWrite" class="form-group">
            <input type="file" class="file-hidden attachment-file-hidden" name="attachment-upload" @change="attachmentFileChanged" />
            <button
                v-if="!busyUpload || attachmentUploadProgress <= 0"
                type="button"
                class="btn btn-primary"
                :disabled="busyUpload"
                @click="selectAttachmentFile"
            >
                <i class="fas fa-upload"></i> {{ $t("Upload attachment") }}
            </button>
            <button
                v-if="busyUpload && attachmentUploadProgress > 0 && attachmentUploadProgress < 100"
                type="button"
                class="btn btn-primary"
                disabled
            >
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Uploading") }}... ({{ attachmentUploadProgress }}%)
            </button>
            <button v-if="busyUpload && attachmentUploadProgress >= 100" type="button" class="btn btn-primary" disabled>
                <i class="fa fa-spinner fa-spin"></i> {{ $t("Encrypting") }}...
            </button>
        </div>

        <AttachmentDeleteModal
            v-model:display="displayAttachmentDelete"
            :attachment-to-delete="attachmentToDelete"
            @confirm="removeAttachmentConfirm"
        ></AttachmentDeleteModal>

        <AuthConfirmationModal
            v-if="displayAuthConfirmation"
            v-model:display="displayAuthConfirmation"
            :tfa="authConfirmationTfa"
            :cooldown="authConfirmationCooldown"
            :error="authConfirmationError"
            @confirm="removeAttachmentConfirmInternal"
        ></AuthConfirmationModal>
    </div>
</template>

<script lang="ts">
import type { MediaAttachment } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import AttachmentDeleteModal from "@/components/modals/AttachmentDeleteModal.vue";
import { clone } from "@/utils/objects";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiMediaRemoveAttachment, apiMediaRenameAttachment, apiMediaUploadAttachment } from "@/api/api-media-edit";
import AuthConfirmationModal from "@/components/modals/AuthConfirmationModal.vue";
import type { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "EditorAttachments",
    components: {
        LoadingIcon,
        AttachmentDeleteModal,
        AuthConfirmationModal,
    },
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            attachments: [] as MediaAttachment[],
            attachmentUploadProgress: 0,
            attachmentEdit: -1,
            attachmentEditName: "",

            busyUpload: false,
            busyRename: false,

            busyDelete: false,
            busyDeleteId: -1,

            canWrite: AuthController.CanWrite,

            displayAttachmentDelete: false,
            attachmentToDelete: null as MediaAttachment,

            displayAuthConfirmation: false,
            authConfirmationCooldown: 0,
            authConfirmationTfa: false,
            authConfirmationError: "",
        };
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    beforeUnmount: function () {
        abortNamedApiRequest(this.requestId);
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

        attachmentFileChanged: function (e: InputEvent) {
            const data = (e.target as HTMLInputElement).files;
            if (data && data.length > 0) {
                const file = data[0];
                this.addAttachment(file);
            }
        },

        onDrop: function (e: DragEvent) {
            e.preventDefault();
            const data = e.dataTransfer.files;
            if (data && data.length > 0) {
                const file = data[0];
                this.addAttachment(file);
            }
        },

        addAttachment: function (file: File) {
            if (this.busyUpload) {
                return;
            }

            this.busyUpload = true;

            const mediaId = AppStatus.CurrentMedia;

            makeNamedApiRequest(this.requestId, apiMediaUploadAttachment(mediaId, file))
                .onSuccess((res) => {
                    PagesController.ShowSnackBarRight(this.$t("Added attachment") + ": " + res.name);
                    this.busyUpload = false;
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
                    this.busyUpload = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busyUpload = false;
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
                    this.busyUpload = false;
                });
        },

        removeAttachment: function (att: MediaAttachment) {
            this.attachmentToDelete = att;
            this.displayAttachmentDelete = true;
        },

        removeAttachmentConfirm: function () {
            this.removeAttachmentConfirmInternal({});
        },

        removeAttachmentConfirmInternal: function (confirmation: ProvidedAuthConfirmation) {
            const att = this.attachmentToDelete;

            if (this.busyDelete || !att) {
                return;
            }

            this.busyDelete = true;
            this.busyDeleteId = att.id;

            const mediaId = AppStatus.CurrentMedia;
            const id = att.id;

            makeNamedApiRequest(this.requestId, apiMediaRemoveAttachment(mediaId, id, confirmation))
                .onSuccess(() => {
                    PagesController.ShowSnackBarRight(this.$t("Removed attachment") + ": " + att.name);
                    this.busyDelete = false;
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
                    this.busyDelete = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busyDelete = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        requiredAuthConfirmationPassword: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = "";
                            this.authConfirmationTfa = false;
                        },
                        invalidPassword: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("Invalid password");
                            this.authConfirmationTfa = false;
                            this.authConfirmationCooldown = Date.now() + 5000;
                        },
                        requiredAuthConfirmationTfa: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = "";
                            this.authConfirmationTfa = true;
                        },
                        invalidTfaCode: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("Invalid one-time code");
                            this.authConfirmationTfa = true;
                            this.authConfirmationCooldown = Date.now() + 5000;
                        },
                        cooldown: () => {
                            this.displayAuthConfirmation = true;
                            this.authConfirmationError = this.$t("You must wait 5 seconds to try again");
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
                    this.busyDelete = false;
                });
        },

        getAttachmentUrl: function (att: MediaAttachment): string {
            return getAssetURL(att.url);
        },

        editAttachment: function (att: MediaAttachment) {
            this.attachmentEdit = att.id;
            this.attachmentEditName = att.name;

            nextTick(() => {
                const el = this.$el.querySelector(".edit-auto-focus");
                if (el) {
                    el.focus();
                    el.select();
                }
            });
        },

        cancelEditAttachment: function () {
            this.attachmentEdit = -1;
            this.attachmentEditName = "";
        },

        editInputKeyEventHandler: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();
                this.saveEditAttachment();
            }
        },

        saveEditAttachment: function () {
            if (this.busyRename) {
                return;
            }

            this.busyRename = true;

            const mediaId = AppStatus.CurrentMedia;
            const id = this.attachmentEdit;

            makeNamedApiRequest(this.requestId, apiMediaRenameAttachment(mediaId, id, this.attachmentEditName))
                .onSuccess((res) => {
                    PagesController.ShowSnackBarRight(this.$t("Renamed attachment") + ": " + res.name);
                    this.busyRename = false;
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
                    this.busyRename = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busyRename = false;
                    handleErr(err, {
                        unauthorized: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidName: () => {
                            PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Invalid attachment name"));
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
                    this.busyRename = false;
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
});
</script>
