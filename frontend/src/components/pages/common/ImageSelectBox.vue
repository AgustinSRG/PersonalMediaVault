<template>
    <div class="form-group">
        <input ref="hiddenFileInput" type="file" class="file-hidden" name="image-select" @change="onImageFileChanged" />
        <div
            class="upload-box auto-focus"
            :class="{ dragging: imageDragging }"
            tabindex="0"
            @click="clickToSelect"
            @dragover="dragOver"
            @dragenter="dragEnter"
            @dragstart="dragEnter"
            @dragend="dragLeave"
            @dragleave="dragLeave"
            @drop="onDrop"
            @keydown="clickOnEnter"
        >
            <div v-if="(loading && !vectorLoaded) || loadingImageThumbnail" class="upload-box-hint loading-delayed">
                {{ $t("Loading image") }}...
            </div>
            <div v-else-if="imageError" class="upload-box-hint">{{ imageError }}</div>
            <div v-else-if="imageUrl" class="upload-box-image-container">
                <img class="upload-box-image" :src="imageUrl" />
            </div>
            <div v-else class="upload-box-hint">
                {{ $t("Drop an image here or click to open the file selection dialog.") }}
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { apiAdvancedSearch } from "@/api/api-search";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { LOAD_RETRY_DELAY } from "@/constants";
import { emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/global-state/app-events";
import { getAssetURL } from "@/utils/api";
import { clickOnEnter } from "@/utils/events";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue";

// Translation
const { $t } = useI18n();

// Props
defineProps({
    /**
     * Loading image?
     */
    loading: Boolean,

    /**
     * True if the vector has been loaded from the image
     */
    vectorLoaded: Boolean,

    /**
     * Error of the image
     */
    imageError: String,
});

// Emits
const emit = defineEmits<{
    /**
     * Emitted when the image changes to start the search process.
     */
    (e: "start-search"): void;
}>();

// Image file model
const imageFile = defineModel<File | null>("imageFile");

// Image media ID
const imageMediaId = defineModel<number>("imageMediaId");

// Current image URL (for image search)
const imageUrl = ref<string | null>(null);

// True if the imageUrl is revokable
const imageUrlIsRevokable = ref(false);

// If of the image where the URL comes from
const imageUrlFromId = ref(-1);

// True if an image is being dragged into the box
const imageDragging = ref(false);

/**
 * Event handler for 'dragover' on the image upload box
 * @param e The drag event
 */
const dragOver = (e: DragEvent) => {
    e.preventDefault();
};

/**
 * Event handler for 'dragenter' on the image upload box
 * @param e The drag event
 */
const dragEnter = (e: DragEvent) => {
    e.preventDefault();
    imageDragging.value = true;
};

/**
 * Event handler for 'dragleave' on the image upload box
 * @param e The drag event
 */
const dragLeave = (e: DragEvent) => {
    e.preventDefault();
    imageDragging.value = false;
};

/**
 * Event handler for 'drop' on the image upload box
 * @param e The drag event
 */
const onDrop = (e: DragEvent) => {
    e.preventDefault();
    imageDragging.value = false;
    const data = e.dataTransfer.files;
    if (data && data.length > 0) {
        setImageFile(data[0]);
    }
};

// Hidden file input element
const hiddenFileInput = useTemplateRef("hiddenFileInput");

/**
 * User clicked on the upload box.
 * The file input must be triggered.
 */
const clickToSelect = () => {
    if (hiddenFileInput.value) {
        hiddenFileInput.value.value = null;
        hiddenFileInput.value.click();
    }
};

/**
 * Event handler for 'change' on the file input
 * @param e The event
 */
const onImageFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        setImageFile(data[0]);
    }
};

/**
 * Clears image URL
 */
const clearImageUrl = () => {
    if (imageUrl.value && imageUrlIsRevokable.value) {
        URL.revokeObjectURL(imageUrl.value);
    }

    imageUrl.value = null;
    imageUrlFromId.value = -1;
};

// Ensure the image URL is revoked before the component unmounts
onBeforeUnmount(clearImageUrl);

/**
 * Sets the image file and creates an URL to visualize it
 * @param image The image file
 */
const setImageFile = (image: File) => {
    clearThumbnailLoad();

    if (imageFile.value === image) {
        return;
    }

    clearImageUrl();

    if (!image) {
        return;
    }

    imageFile.value = image;
    imageUrl.value = URL.createObjectURL(image);
    imageUrlIsRevokable.value = true;

    imageMediaId.value = -1;
    imageUrlFromId.value = -1;

    emit("start-search");
};

// Loading image thumbnail?
const loadingImageThumbnail = ref(false);

// Request ID to load image thumbnail
const loadingImageThumbnailRequestId = useRequestId();

/**
 * Loads thumbnail image
 */
const loadImageThumbnail = () => {
    clearImageUrl();
    clearNamedTimeout(loadingImageThumbnailRequestId);

    loadingImageThumbnail.value = true;

    const id = imageMediaId.value;

    if (id < 0) {
        abortNamedApiRequest(loadingImageThumbnailRequestId);
        loadingImageThumbnail.value = false;
        return;
    }

    makeNamedApiRequest(loadingImageThumbnailRequestId, apiAdvancedSearch("allof", [], "asc", "" + (id - 1), 1))
        .onSuccess((result) => {
            if (result.items.length > 0 && result.items[0].thumbnail) {
                imageUrlFromId.value = id;
                imageUrl.value = getAssetURL(result.items[0].thumbnail);
                imageUrlIsRevokable.value = false;
            }

            loadingImageThumbnail.value = false;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadingImageThumbnailRequestId, LOAD_RETRY_DELAY, loadImageThumbnail);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadingImageThumbnailRequestId, LOAD_RETRY_DELAY, loadImageThumbnail);
        });
};

// Retry loading thumbnail on unauthorized
onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    if (loadingImageThumbnail.value) {
        loadImageThumbnail();
    }
});

/**
 * Clears the thumbnail loading
 */
const clearThumbnailLoad = () => {
    clearNamedTimeout(loadingImageThumbnailRequestId);
    abortNamedApiRequest(loadingImageThumbnailRequestId);
};

watch(imageMediaId, () => {
    clearThumbnailLoad();

    if (imageMediaId.value !== imageUrlFromId.value) {
        if (imageMediaId.value >= 0) {
            imageFile.value = null;
            loadImageThumbnail();
        } else if (!imageFile.value) {
            clearImageUrl();
        }
    }
});

onMounted(() => {
    if (imageMediaId.value >= 0) {
        loadImageThumbnail();
    }
    if (imageFile.value) {
        imageUrl.value = URL.createObjectURL(imageFile.value);
        imageUrlIsRevokable.value = true;
    }
});
</script>
