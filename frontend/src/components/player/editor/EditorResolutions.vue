<template>
    <div class="player-editor-sub-content">
        <!--- Resolutions -->

        <div class="form-group">
            <label v-if="type === 2"
                >{{ $t("Extra resolutions for videos. These resolutions can be used for slow connections or small screens") }}.</label
            >
            <label v-if="type === 1"
                >{{ $t("Extra resolutions for images. These resolutions can be used for slow connections or small screens") }}.</label
            >
        </div>

        <div class="form-group">
            <label v-if="type === 1">{{ $t("Original resolution") }}: {{ width }}x{{ height }}</label>
            <label v-if="type === 2"> {{ $t("Original resolution") }}: {{ width }}x{{ height }}, {{ fps }} fps </label>
        </div>

        <div v-if="(type === 2 || type === 1) && resolutions.length > 0" class="table-responsive">
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
                        <td v-if="canWrite" class="text-right">
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
                        <td v-else class="text-right">
                            <i v-if="res.enabled" class="fas fa-check"></i>
                            <i v-else class="fas fa-times"></i>
                            <span>&nbsp;{{ res.enabled ? $t("Available") : $t("Not available") }}</span>
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

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { EVENT_NAME_MEDIA_UPDATE } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, onMounted, ref } from "vue";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import type { MediaData, MediaResolution, MediaType, NamedResolution } from "@/api/models";
import { MEDIA_TYPE_IMAGE, MEDIA_TYPE_VIDEO } from "@/api/models";
import { apiMediaAddResolution, apiMediaRemoveResolution } from "@/api/api-media-edit";
import { STANDARD_VIDEO_RESOLUTIONS } from "@/utils/resolutions";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useRequestId } from "@/composables/use-request-id";
import { showSnackBarRight } from "@/control/snack-bar";
import { modifyCurrentMediaData } from "@/control/media";

const ResolutionConfirmationModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ResolutionConfirmationModal.vue"),
});

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Emits
const emit = defineEmits<{
    /**
     * Media changed
     */
    (e: "changed"): void;
}>();

// Map of standard resolutions
const standardResolutionsMap = new Map<string, string>();

// Initialize map of standard resolutions
STANDARD_VIDEO_RESOLUTIONS.forEach((sr) => {
    standardResolutionsMap.set(`${sr.width}x${sr.height}-${sr.fps}`, sr.name);
    standardResolutionsMap.set(`${sr.height}x${sr.width}-${sr.fps}`, sr.name);

    if (sr.fps === 30) {
        standardResolutionsMap.set(`${sr.width}x${sr.height}`, sr.name);
        standardResolutionsMap.set(`${sr.height}x${sr.width}`, sr.name);
    }
});

// Media type
const type = ref<MediaType>(0);

// Media original resolution
const width = ref(0);
const height = ref(0);
const fps = ref(0);

// Available media resolutions
const resolutions = ref<NamedResolution[]>([]);

/**
 * Updates media metadata
 */
const updateMediaData = (mediaData: MediaData | null) => {
    if (!mediaData) {
        return;
    }

    type.value = mediaData.type;

    width.value = mediaData.width;
    height.value = mediaData.height;
    fps.value = mediaData.fps;

    updateResolutions(mediaData.resolutions || []);
};

onMounted(updateMediaData);
onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, updateMediaData);

/**
 * Updates resolutions
 * @param newResolutions The resolutions from the media metadata
 */
const updateResolutions = (newResolutions: MediaResolution[]) => {
    const totalPixels = width.value * height.value;
    const totalFps = fps.value;

    const resolutionsLeft = newResolutions.slice();

    resolutions.value = STANDARD_VIDEO_RESOLUTIONS.filter((r) => {
        if (type.value === MEDIA_TYPE_IMAGE) {
            return r.fps === 30 && r.width * r.height < totalPixels;
        } else if (type.value === MEDIA_TYPE_VIDEO) {
            const pixels = r.width * r.height;
            return pixels < totalPixels || (pixels === totalPixels && r.fps < totalFps);
        } else {
            return false;
        }
    })
        .map((r) => {
            if ((r.width >= r.height && width.value >= height.value) || (r.width <= r.height && width.value <= height.value)) {
                return r;
            } else {
                // Rotate the resolution for vertical videos or images
                return {
                    name: r.name,
                    width: r.height,
                    height: r.width,
                    fps: r.fps,
                };
            }
        })
        .map((r) => {
            let enabled = false;
            let fps = r.fps;
            for (let i = 0; i < resolutionsLeft.length; i++) {
                const res = resolutionsLeft[i];
                if (res.width === r.width && res.height === r.height && (type.value === MEDIA_TYPE_IMAGE || res.fps === r.fps)) {
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
                const customName = type.value === MEDIA_TYPE_IMAGE ? `${rl.width}x${rl.height}` : `${rl.width}x${rl.height}-${rl.fps}`;
                return {
                    enabled: true,
                    name: standardResolutionsMap.get(customName) || customName,
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
};

/**
 * Renders resolution properties
 * @param resWidth The resolution width
 * @param resHeight The resolution height
 * @param originalWidth The original width
 * @param originalHeight The original height
 * @returns The resolution properties
 */
const renderResolutionProperties = (resWidth: number, resHeight: number, originalWidth: number, originalHeight: number): string => {
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
};

// Display resolution confirmation?
const displayResolutionConfirmation = ref(false);

// Selected resolution for confirmation
const resolutionToConfirm = ref<NamedResolution | null>(null);

// Deleting the resolution?
const resolutionToConfirmDeleting = ref(false);

/**
 * Asks the user to add a resolution
 * @param r The resolution
 */
const addResolution = (r: NamedResolution) => {
    resolutionToConfirm.value = r;
    resolutionToConfirmDeleting.value = false;
    displayResolutionConfirmation.value = true;
};

/**
 * Asks the user to delete a resolution
 * @param r The resolution
 */
const deleteResolution = (r: NamedResolution) => {
    resolutionToConfirm.value = r;
    resolutionToConfirmDeleting.value = true;
    displayResolutionConfirmation.value = true;
};

// Busy status
const busy = ref(false);

// Target for the busy request
const busyTarget = ref("");

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

// Save request ID
const saveRequestId = useRequestId();

/**
 * The user confirmed to add or delete a resolution
 */
const onResolutionConfirm = () => {
    if (!resolutionToConfirmDeleting.value) {
        performAddResolution();
    } else {
        performDeleteResolution();
    }
};

/**
 * Performs the request to add a resolution
 */
const performAddResolution = () => {
    const r = resolutionToConfirm.value;

    if (!r) {
        return;
    }

    if (busy.value) {
        return;
    }

    busy.value = true;
    busyTarget.value = r.name;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(saveRequestId, apiMediaAddResolution(mediaId, r.width, r.height, r.fps))
        .onSuccess((result) => {
            showSnackBarRight($t("Added resolution") + ": " + r.name);
            busy.value = false;
            r.enabled = true;
            r.fps = result.fps;

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.resolutions = resolutions.value
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
            });

            emit("changed");
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                duplicatedResolution: () => {
                    // Already added
                    showSnackBarRight($t("Added resolution") + ": " + r.name);
                    busy.value = false;
                    r.enabled = true;

                    modifyCurrentMediaData(mediaId, (metadata) => {
                        metadata.resolutions = resolutions.value
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
                    });

                    emit("changed");
                },
                badRequest,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            setError(err.message);
            console.error(err);
        });
};

/**
 * Performs request to delete a resolution
 */
const performDeleteResolution = () => {
    const r = resolutionToConfirm.value;

    if (!r) {
        return;
    }

    if (busy.value) {
        return;
    }

    busy.value = true;
    busyTarget.value = r.name;

    const mediaId = AppStatus.CurrentMedia;

    makeNamedApiRequest(saveRequestId, apiMediaRemoveResolution(mediaId, r.width, r.height, r.fps))
        .onSuccess(() => {
            showSnackBarRight($t("Removed resolution") + ": " + r.name);
            busy.value = false;
            r.enabled = false;

            modifyCurrentMediaData(mediaId, (metadata) => {
                metadata.resolutions = resolutions.value
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
            });

            emit("changed");
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                badRequest,
                accessDenied,
                notFound,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            setError(err.message);
            console.error(err);
        });
};
</script>
