<template>
    <div>
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
            <div v-if="loading && !vectorLoaded" class="upload-box-hint loading-delayed">{{ $t("Loading image") }}...</div>
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
import { useI18n } from "@/composables/use-i18n";
import { clickOnEnter } from "@/utils/events";
import { onBeforeUnmount, onMounted, ref, useTemplateRef } from "vue";

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

// Current image URL (for image search)
const imageUrl = ref<string | null>(null);

// Ensure the image URL is revoked before the component unmounts
onBeforeUnmount(() => {
    if (imageUrl.value) {
        URL.revokeObjectURL(imageUrl.value);
    }
});

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
 * Sets the image file and creates an URL to visualize it
 * @param image The image file
 */
const setImageFile = (image: File) => {
    if (imageFile.value === image) {
        return;
    }

    if (imageUrl.value) {
        URL.revokeObjectURL(imageUrl.value);
        imageUrl.value = null;
    }

    if (!image) {
        return;
    }

    imageFile.value = image;
    imageUrl.value = URL.createObjectURL(image);

    emit("start-search");
};

onMounted(() => {
    if (imageFile.value) {
        imageUrl.value = URL.createObjectURL(imageFile.value);
    }
});
</script>
