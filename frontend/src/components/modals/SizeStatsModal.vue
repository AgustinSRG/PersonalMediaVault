<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div class="modal-dialog modal-sm" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Size Statistics") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body no-padding table-responsive">
                <table class="table table-text-overflow no-margin">
                    <tbody>
                        <tr>
                            <td>{{ $t("Metadata") }}</td>
                            <td class="text-right">{{ renderSize(metaSize) }}</td>
                        </tr>

                        <tr v-if="originalSize > 0">
                            <td>{{ $t("Original") }}</td>
                            <td class="text-right">{{ renderSize(originalSize) }}</td>
                        </tr>

                        <tr v-if="resizedSize > 0">
                            <td>{{ $t("Resized media") }}</td>
                            <td class="text-right">{{ renderSize(resizedSize) }}</td>
                        </tr>

                        <tr v-if="thumbnailSize > 0">
                            <td>{{ $t("Thumbnail") }}</td>
                            <td class="text-right">{{ renderSize(thumbnailSize) }}</td>
                        </tr>

                        <tr v-if="videoPreviewsSize > 0">
                            <td>{{ $t("Video previews") }}</td>
                            <td class="text-right">{{ renderSize(videoPreviewsSize) }}</td>
                        </tr>

                        <tr v-if="subtitlesSize > 0">
                            <td>{{ $t("Subtitles") }}</td>
                            <td class="text-right">{{ renderSize(subtitlesSize) }}</td>
                        </tr>

                        <tr v-if="audioTracksSize > 0">
                            <td>{{ $t("Audio tracks") }}</td>
                            <td class="text-right">{{ renderSize(audioTracksSize) }}</td>
                        </tr>

                        <tr v-if="imageNotesSize > 0">
                            <td>{{ $t("Image notes") }}</td>
                            <td class="text-right">{{ renderSize(imageNotesSize) }}</td>
                        </tr>

                        <tr v-if="descriptionSize > 0">
                            <td>{{ $t("Description") }}</td>
                            <td class="text-right">{{ renderSize(descriptionSize) }}</td>
                        </tr>

                        <tr v-if="attachmentsSize > 0">
                            <td>{{ $t("Attachments") }}</td>
                            <td class="text-right">{{ renderSize(attachmentsSize) }}</td>
                        </tr>

                        <tr v-if="otherSize > 0">
                            <td>{{ $t("Other assets") }}</td>
                            <td class="text-right">{{ renderSize(otherSize) }}</td>
                        </tr>

                        <tr>
                            <td class="bold">{{ $t("Total") }}</td>
                            <td class="bold text-right">{{ renderSize(total) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { onMounted, ref, useTemplateRef, watch } from "vue";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { AuthController } from "@/control/auth";
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { apiMediaGetMediaSizeStats } from "@/api/api-media";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";
import { renderSize } from "@/utils/size";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * ID of the media
     */
    mid: {
        type: Number,
        required: true,
    },
});

// Sizes for every kind of asset in the media
const metaSize = ref(0);
const originalSize = ref(0);
const thumbnailSize = ref(0);
const resizedSize = ref(0);
const videoPreviewsSize = ref(0);
const subtitlesSize = ref(0);
const audioTracksSize = ref(0);
const imageNotesSize = ref(0);
const descriptionSize = ref(0);
const attachmentsSize = ref(0);
const otherSize = ref(0);

// Total media size
const total = ref(0);

// Loading status
const loading = ref(false);

// Load request ID
const loadRequestId = useRequestId();

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the data
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (!display.value) {
        return;
    }

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiMediaGetMediaSizeStats(props.mid))
        .onSuccess((result) => {
            loading.value = false;

            metaSize.value = result.meta_size;

            originalSize.value = 0;
            resizedSize.value = 0;
            thumbnailSize.value = 0;
            subtitlesSize.value = 0;
            audioTracksSize.value = 0;
            imageNotesSize.value = 0;
            descriptionSize.value = 0;
            attachmentsSize.value = 0;
            videoPreviewsSize.value = 0;
            otherSize.value = 0;

            total.value = 0;

            // Compute size of assets and total size

            total.value += result.meta_size;

            for (const asset of result.assets) {
                const assetName = asset.name + "";

                if (assetName === "ORIGINAL") {
                    originalSize.value += asset.size;
                } else if (assetName === "THUMBNAIL") {
                    thumbnailSize.value += asset.size;
                } else if (assetName === "VIDEO_PREVIEWS") {
                    videoPreviewsSize.value += asset.size;
                } else if (assetName === "IMG_NOTES") {
                    imageNotesSize.value += asset.size;
                } else if (assetName === "DESCRIPTION") {
                    descriptionSize.value += asset.size;
                } else if (assetName.startsWith("RESIZED_")) {
                    resizedSize.value += asset.size;
                } else if (assetName.startsWith("SUBTITLES_")) {
                    subtitlesSize.value += asset.size;
                } else if (assetName.startsWith("AUDIO_TRACK_")) {
                    audioTracksSize.value += asset.size;
                } else if (assetName.startsWith("ATTACHMENT:")) {
                    attachmentsSize.value += asset.size;
                } else {
                    otherSize.value += asset.size;
                }

                total.value += asset.size;
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    close();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
        });
};

onMounted(load);
watch(display, load);
watch(() => props.mid, load);
</script>
