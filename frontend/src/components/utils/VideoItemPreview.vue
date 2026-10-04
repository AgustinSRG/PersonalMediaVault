<template>
    <div class="video-item-preview-container">
        <div v-if="previewReady && previewUrl" class="video-item-preview-inner">
            <video
                ref="videoElement"
                :class="{ 'video-preview-ready': canPlay, hidden: hidden }"
                crossorigin="use-credentials"
                playsinline
                muted
                loop
                autoplay
                :src="previewUrl"
                @canplay="onCanPlay"
            ></video>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRequestId } from "@/composables/use-request-id";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { apiMediaGetMediaPreview } from "@/api/api-media";
import { LOAD_RETRY_DELAY } from "@/constants";
import { getAssetURL } from "@/utils/api";

// Props
const props = defineProps({
    /**
     * Media ID
     */
    mid: {
        type: Number,
        required: true,
    },

    /**
     * True if hidden
     */
    hidden: Boolean,
});

// Preview video URL
const previewUrl = ref("");

// True if the preview is ready
const previewReady = ref(false);

// Request ID to load the preview URL
const loadPreviewUrlRequestId = useRequestId();

// Loads the preview URL
const load = () => {
    clearNamedTimeout(loadPreviewUrlRequestId);

    previewReady.value = false;

    makeNamedApiRequest(loadPreviewUrlRequestId, apiMediaGetMediaPreview(props.mid))
        .onSuccess((res) => {
            previewUrl.value = getAssetURL(res.url);
            previewReady.value = true;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: function (): void {
                    throw new Error("Function not implemented.");
                },
                notFound: () => {
                    previewUrl.value = "";
                },
                temporalError: () => {
                    setNamedTimeout(loadPreviewUrlRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            setNamedTimeout(loadPreviewUrlRequestId, LOAD_RETRY_DELAY, load);
        });
};

// Load on mounted and if the media ID changes
onMounted(load);
watch([() => props.mid], load);

// True if the video is ready for play
const canPlay = ref(false);

/**
 * CanPlay event handler
 */
const onCanPlay = () => {
    canPlay.value = true;
};

// Reset canPlay when the URL changes
watch(previewUrl, () => {
    canPlay.value = false;
});
</script>
