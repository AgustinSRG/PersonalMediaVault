<template>
    <div class="player-editor-sub-content">
        <!--- Resolutions -->

        <div v-if="canWrite && (type === 2 || type === 1)" class="form-group">
            <label v-if="type === 2"
                >{{ $t("Extra resolutions for videos. These resolutions can be used for slow connections or small screens") }}.</label
            >
            <label v-if="type === 1"
                >{{ $t("Extra resolutions for images. These resolutions can be used for slow connections or small screens") }}.</label
            >
        </div>

        <div v-if="canWrite && (type === 2 || type === 1)" class="form-group">
            <label v-if="type === 1">{{ $t("Original resolution") }}: {{ width }}x{{ height }}</label>
            <label v-if="type === 2"> {{ $t("Original resolution") }}: {{ width }}x{{ height }}, {{ fps }} fps </label>
        </div>

        <div v-if="canWrite && (type === 2 || type === 1) && resolutions.length > 0" class="table-responsive">
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
                                <LoadingIcon icon="fas fa-plus" :loading="busy && busyTarget === res.name"></LoadingIcon> {{ $t("Encode") }}
                            </button>
                            <button
                                v-if="res.enabled"
                                type="button"
                                class="btn btn-danger btn-xs"
                                :disabled="busy"
                                @click="deleteResolution(res)"
                            >
                                <LoadingIcon icon="fas fa-trash-alt" :loading="busy && busyTarget === res.name"></LoadingIcon>
                                {{ $t("Delete") }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <ResolutionConfirmationModal
            v-model:display="displayResolutionConfirmation"
            :type="type"
            :resolution="resolutionToConfirm"
            :deleting="resolutionToConfirmDeleting"
            @confirm="onResolutionConfirm"
        ></ResolutionConfirmationModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import ResolutionConfirmationModal from "@/components/modals/ResolutionConfirmationModal.vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO, MediaResolution, NamedResolution } from "@/api/models";
import { PagesController } from "@/control/pages";
import { apiMediaAddResolution, apiMediaRemoveResolution } from "@/api/api-media-edit";
import { STANDARD_VIDEO_RESOLUTIONS } from "@/utils/resolutions";

export default defineComponent({
    name: "EditorResolutions",
    components: {
        LoadingIcon,
        ResolutionConfirmationModal,
    },
    emits: ["changed"],
    setup() {
        const standardResolutions = STANDARD_VIDEO_RESOLUTIONS;

        const standardResolutionsMap = new Map<string, string>();

        standardResolutions.forEach((sr) => {
            standardResolutionsMap.set(`${sr.width}x${sr.height}-${sr.fps}`, sr.name);

            if (sr.fps === 30) {
                standardResolutionsMap.set(`${sr.width}x${sr.height}`, sr.name);
            }
        });

        return {
            standardResolutions: standardResolutions,
            standardResolutionsMap: standardResolutionsMap,
            requestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            type: 0,

            width: 0,
            height: 0,
            fps: 0,

            resolutions: [] as NamedResolution[],

            busy: false,
            busyTarget: "",

            canWrite: AuthController.CanWrite,

            displayResolutionConfirmation: false,

            resolutionToConfirm: null as NamedResolution,
            resolutionToConfirmDeleting: false,
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

            this.width = MediaController.MediaData.width;
            this.height = MediaController.MediaData.height;
            this.fps = MediaController.MediaData.fps;

            this.updateResolutions(MediaController.MediaData.resolutions || []);
        },

        updateResolutions: function (resolutions: MediaResolution[]) {
            const totalPixels = this.width * this.height;
            const totalFps = this.fps;

            const resolutionsLeft = resolutions.slice();

            this.resolutions = this.standardResolutions
                .filter((r) => {
                    if (this.type === MEDIA_TYPE_IMAGE) {
                        return r.fps === 30 && r.width * r.height < totalPixels;
                    } else if (this.type === MEDIA_TYPE_VIDEO) {
                        const pixels = r.width * r.height;
                        return pixels < totalPixels || (pixels === totalPixels && r.fps < totalFps);
                    } else {
                        return false;
                    }
                })
                .map((r) => {
                    let enabled = false;
                    let fps = r.fps;
                    for (let i = 0; i < resolutionsLeft.length; i++) {
                        const res = resolutionsLeft[i];
                        if (res.width === r.width && res.height === r.height && (this.type === MEDIA_TYPE_IMAGE || res.fps === r.fps)) {
                            enabled = true;
                            fps = res.fps;
                            resolutionsLeft.splice(i, 1);
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
                })
                .concat(
                    resolutionsLeft.map((rl) => {
                        const customName =
                            this.type === MEDIA_TYPE_IMAGE ? `${rl.width}x${rl.height}` : `${rl.width}x${rl.height}-${rl.fps}`;
                        return {
                            enabled: true,
                            name: this.standardResolutionsMap.get(customName) || customName,
                            width: rl.width,
                            height: rl.height,
                            fps: rl.fps,
                        };
                    }),
                )
                .sort((a, b) => {
                    const aPixels = a.width * a.height;
                    const bPixels = b.width * b.height;

                    if (aPixels > bPixels) {
                        return -1;
                    } else if (aPixels < bPixels) {
                        return 1;
                    } else if (a.fps > b.fps) {
                        return -1;
                    } else {
                        return 1;
                    }
                });
        },

        addResolution: function (r: NamedResolution) {
            this.resolutionToConfirm = r;
            this.resolutionToConfirmDeleting = false;
            this.displayResolutionConfirmation = true;
        },

        deleteResolution: function (r: NamedResolution) {
            this.resolutionToConfirm = r;
            this.resolutionToConfirmDeleting = true;
            this.displayResolutionConfirmation = true;
        },

        onResolutionConfirm: function () {
            const r = this.resolutionToConfirm;

            if (!r) {
                return;
            }

            if (this.busy) {
                return;
            }

            this.busy = true;
            this.busyTarget = r.name;

            const mediaId = AppStatus.CurrentMedia;

            if (!this.resolutionToConfirmDeleting) {
                // Add resolution
                makeNamedApiRequest(this.requestId, apiMediaAddResolution(mediaId, r.width, r.height, r.fps))
                    .onSuccess((result) => {
                        PagesController.ShowSnackBarRight(this.$t("Added resolution") + ": " + r.name);
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
                                PagesController.ShowSnackBarRight(this.$t("Error") + ": " + this.$t("Access denied"));
                                AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            },
                            duplicatedResolution: () => {
                                // Already added
                                PagesController.ShowSnackBarRight(this.$t("Added resolution") + ": " + r.name);
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
            } else {
                // Remove resolution
                makeNamedApiRequest(this.requestId, apiMediaRemoveResolution(mediaId, r.width, r.height, r.fps))
                    .onSuccess(() => {
                        PagesController.ShowSnackBarRight(this.$t("Removed resolution") + ": " + r.name);
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
            }
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
});
</script>
