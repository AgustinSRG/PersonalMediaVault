<template>
    <div class="player-editor-sub-content">
        <!--- Resolutions -->

        <div class="form-group" v-if="canWrite && (type === 2 || type === 1)">
            <label v-if="type === 2"
                >{{ $t("Extra resolutions for videos. These resolutions can be used for slow connections or small screens") }}:</label
            >
            <label v-if="type === 1"
                >{{ $t("Extra resolutions for images. These resolutions can be used for slow connections or small screens") }}:</label
            >
        </div>

        <div class="form-group" v-if="canWrite && (type === 2 || type === 1)">
            <label v-if="type === 1">{{ $t("Original resolution") }}: {{ width }}x{{ height }}</label>
            <label v-if="type === 2"> {{ $t("Original resolution") }}: {{ width }}x{{ height }}, {{ fps }} fps </label>
        </div>

        <div v-if="canWrite && (type === 2 || type === 1)" class="table-responsive">
            <table class="table">
                <thead>
                    <tr>
                        <th class="text-left">{{ $t("Name") }}</th>
                        <th class="text-left">{{ $t("Properties") }}</th>
                        <th class="text-right"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="res in resolutions" :key="res.name">
                        <td class="bold">{{ res.name }}</td>
                        <td v-if="type === 1">
                            {{ renderResolutionProperties(res.width, res.height, width, height) }}
                        </td>
                        <td v-if="type === 2">{{ res.width }}x{{ res.height }}, {{ res.fps }} fps</td>
                        <td class="text-right">
                            <button
                                v-if="!res.enabled"
                                type="button"
                                class="btn btn-primary btn-xs"
                                :disabled="busy"
                                @click="addResolution(res)"
                            >
                                <i class="fas fa-plus"></i> {{ $t("Encode") }}
                            </button>
                            <button
                                v-if="res.enabled"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy"
                                @click="deleteResolution(res)"
                            >
                                <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <ResolutionConfirmationModal
            ref="resolutionConfirmationModal"
            v-model:display="displayResolutionConfirmation"
        ></ResolutionConfirmationModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { Request } from "@asanrom/request-browser";
import { defineComponent } from "vue";

import ResolutionConfirmationModal from "@/components/modals/ResolutionConfirmationModal.vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO } from "@/api/models";
import { PagesController } from "@/control/pages";
import { apiMediaAddResolution, apiMediaRemoveResolution } from "@/api/api-media-edit";

export default defineComponent({
    components: {
        ResolutionConfirmationModal,
    },
    name: "EditorResolutions",
    emits: ["changed"],
    setup() {
        return {
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            width: 0,
            height: 0,
            fps: 0,

            standardResolutions: [
                {
                    name: "144p",
                    width: 256,
                    height: 144,
                    fps: 30,
                },
                {
                    name: "240p",
                    width: 352,
                    height: 240,
                    fps: 30,
                },
                {
                    name: "360p",
                    width: 480,
                    height: 360,
                    fps: 30,
                },
                {
                    name: "480p",
                    width: 858,
                    height: 480,
                    fps: 30,
                },
                {
                    name: "720p",
                    width: 1280,
                    height: 720,
                    fps: 30,
                },
                {
                    name: "720p60",
                    width: 1280,
                    height: 720,
                    fps: 60,
                },
                {
                    name: "1080p",
                    width: 1920,
                    height: 1080,
                    fps: 30,
                },
                {
                    name: "1080p60",
                    width: 1920,
                    height: 1080,
                    fps: 60,
                },
                {
                    name: "2k",
                    width: 2048,
                    height: 1152,
                    fps: 30,
                },
                {
                    name: "2k60",
                    width: 2048,
                    height: 1152,
                    fps: 60,
                },
                {
                    name: "4k",
                    width: 3860,
                    height: 2160,
                    fps: 30,
                },
                {
                    name: "4k60",
                    width: 3860,
                    height: 2160,
                    fps: 60,
                },
            ],

            resolutions: [],

            busy: false,

            canWrite: AuthController.CanWrite,

            displayResolutionConfirmation: false,
        };
    },

    methods: {
        updateMediaData: function () {
            if (!MediaController.MediaData) {
                return;
            }

            this.type = MediaController.MediaData.type;

            this.width = MediaController.MediaData.width;
            this.height = MediaController.MediaData.height;
            this.fps = MediaController.MediaData.fps;

            this.updateResolutions(MediaController.MediaData.resolutions || []);
        },

        updateResolutions: function (resolutions) {
            this.resolutions = this.standardResolutions
                .filter((r) => {
                    if (this.type === MEDIA_TYPE_IMAGE) {
                        return r.fps === 30;
                    } else if (this.type === MEDIA_TYPE_VIDEO) {
                        return true;
                    } else {
                        return false;
                    }
                })
                .map((r) => {
                    let enabled = false;
                    let fps = r.fps;
                    for (const res of resolutions) {
                        if (res.width === r.width && res.height === r.height && (this.type === MEDIA_TYPE_IMAGE || res.fps === r.fps)) {
                            enabled = true;
                            fps = res.fps;
                            break;
                        }
                    }
                    return {
                        enabled: enabled,
                        name: r.name,
                        width: r.width,
                        height: r.height,
                        fps: fps,
                    };
                });
        },

        addResolution: function (r) {
            this.$refs.resolutionConfirmationModal.show({
                type: this.type,
                deleting: false,
                name: r.name,
                width: r.width,
                height: r.height,
                fps: r.fps,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;

                    Request.Pending(this.requestId, apiMediaAddResolution(mediaId, r.width, r.height, r.fps))
                        .onSuccess((result) => {
                            PagesController.ShowSnackBar(this.$t("Added resolution") + ": " + r.name);
                            this.busy = false;
                            r.enabled = true;
                            r.fps = result.fps;
                            if (MediaController.MediaData) {
                                MediaController.MediaData.resolutions = this.resolutions
                                    .filter((re) => {
                                        return re.enabled;
                                    })
                                    .map((re) => {
                                        return {
                                            width: re.width,
                                            height: re.height,
                                            fps: re.fps,
                                            ready: false,
                                            task: 0,
                                            url: "",
                                        };
                                    });
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
                                duplicatedResolution: () => {
                                    // Already added
                                    PagesController.ShowSnackBar(this.$t("Added resolution") + ": " + r.name);
                                    this.busy = false;
                                    r.enabled = true;
                                    if (MediaController.MediaData) {
                                        MediaController.MediaData.resolutions = this.resolutions
                                            .filter((re) => {
                                                return re.enabled;
                                            })
                                            .map((re) => {
                                                return {
                                                    width: re.width,
                                                    height: re.height,
                                                    fps: re.fps,
                                                    ready: false,
                                                    task: 0,
                                                    url: "",
                                                };
                                            });
                                    }
                                    this.$emit("changed");
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
            });
        },

        deleteResolution: function (r) {
            this.$refs.resolutionConfirmationModal.show({
                type: this.type,
                deleting: true,
                name: r.name,
                width: r.width,
                height: r.height,
                fps: r.fps,
                callback: () => {
                    if (this.busy) {
                        return;
                    }

                    this.busy = true;

                    const mediaId = AppStatus.CurrentMedia;

                    Request.Pending(this.requestId, apiMediaRemoveResolution(mediaId, r.width, r.height, r.fps))
                        .onSuccess(() => {
                            PagesController.ShowSnackBar(this.$t("Removed resolution") + ": " + r.name);
                            this.busy = false;
                            r.enabled = false;
                            if (MediaController.MediaData) {
                                MediaController.MediaData.resolutions = this.resolutions
                                    .filter((re) => {
                                        return re.enabled;
                                    })
                                    .map((re) => {
                                        return {
                                            width: re.width,
                                            height: re.height,
                                            fps: re.fps,
                                            ready: false,
                                            task: 0,
                                            url: "",
                                        };
                                    });
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
            });
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        renderResolutionProperties: function (resWidth: number, resHeight: number, originalWidth: number, originalHeight: number): string {
            let width = originalWidth;
            let height = originalHeight;

            if (width > height) {
                const proportionalHeight = Math.round((height * resWidth) / width);

                if (proportionalHeight > resHeight) {
                    width = Math.round((width * resHeight) / height);
                    height = resHeight;
                } else {
                    width = resWidth;
                    height = proportionalHeight;
                }
            } else {
                const proportionalWidth = Math.round((width * resHeight) / height);

                if (proportionalWidth > resWidth) {
                    height = Math.round((height * resWidth) / width);
                    width = resWidth;
                } else {
                    width = proportionalWidth;
                    height = resHeight;
                }
            }

            return width + "x" + height;
        },
    },

    mounted: function () {
        this.updateMediaData();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMediaData.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
    },

    beforeUnmount: function () {
        Request.Abort(this.requestId);
    },
});
</script>
