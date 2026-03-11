<template>
    <div ref="container" class="player-editor-sub-content">
        <!--- Description -->

        <div class="form-group">
            <label>{{ $t("Add a description for the media.") }}</label>
        </div>

        <div class="form-group mt-1">
            <textarea
                v-model="description"
                :readonly="!canWrite"
                class="form-control form-control-full-width form-textarea auto-focus"
                :placeholder="placeholder"
                rows="10"
                :disabled="loading || busy"
                @input="markDirty"
            ></textarea>
        </div>

        <div v-if="canWrite" class="form-group">
            <button
                v-if="!loading && (dirty || busy || !saved)"
                type="button"
                class="btn btn-primary"
                :disabled="busy || !dirty"
                @click="saveChanges"
            >
                <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Change description") }}
            </button>
            <button v-else type="button" disabled class="btn btn-primary">
                <i class="fas fa-check"></i> {{ $t("Saved description") }}
            </button>
        </div>

        <SaveChangesAskModal
            v-if="displayExitConfirmation"
            v-model:display="displayExitConfirmation"
            @yes="onExitSaveChanges"
            @no="onExitDiscardChanges"
        ></SaveChangesAskModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </div>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, EVENT_NAME_MEDIA_UPDATE, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest, RequestErrorHandler } from "@asanrom/request-browser";
import { computed, defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import { apiMediaSetDescription } from "@/api/api-media-edit";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { getAssetURL } from "@/utils/api";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useRequestId } from "@/composables/use-request-id";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { useExitPreventer } from "@/composables/use-exit-preventer";
import { showSnackBar } from "@/control/snack-bar";
import { getCurrentMediaData, modifyCurrentMediaData } from "@/control/media";

const SaveChangesAskModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SaveChangesAskModal.vue"),
});

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Ref to the container element
const container = useTemplateRef("container");

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

// Loading status
const loading = ref(false);

// Placeholder for the description textarea
const placeholder = computed<string>(() => {
    if (loading.value) {
        return $t("Loading") + "...";
    } else {
        return $t("Example paragraph") + "\n" + $t("Example paragraph") + "\n";
    }
});

// Description
const description = ref("");

// Dirty? (unsaved changes)
const dirty = ref(false);

// Load request ID
const loadRequestId = useRequestId();

// DElay to retry loading (milliseconds)
const LOAD_RETRY_DELAY = 1500;

const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    description.value = "";

    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        return;
    }

    const descFilePath = mediaData.description_url;

    if (!descFilePath) {
        description.value = "";
        loading.value = false;

        autoFocus();
        return;
    }

    loading.value = true;

    makeNamedApiRequest(loadRequestId, {
        method: "GET",
        url: getAssetURL(descFilePath),
    })
        .onSuccess((descriptionText) => {
            description.value = descriptionText;
            loading.value = false;
            dirty.value = false;

            autoFocus();
        })
        .onRequestError((err) => {
            new RequestErrorHandler()
                .add(401, "*", () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                })
                .add(404, "*", () => {
                    description.value = "";
                    loading.value = false;
                    dirty.value = false;
                    autoFocus();
                })
                .add("*", "*", () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                })
                .handle(err);
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
        });
};

onMounted(load);
onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, load);
onApplicationEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, (source) => {
    if (source === "editor") {
        return;
    }

    load();
});

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    nextTick(() => {
        const elem = container.value?.querySelector(".auto-focus") as HTMLElement;
        if (elem) {
            elem.focus();
        }
    });
};

/**
 * Indicates changes were made
 */
const markDirty = () => {
    dirty.value = true;
};

// Busy status
const busy = ref(false);

// True if saved
const saved = ref(false);

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

// Save request ID
const saveRequestId = useRequestId();

/**
 * Saves the description changes
 */
const saveChanges = () => {
    if (busy.value) {
        return;
    }

    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        return;
    }

    busy.value = true;

    const mid = mediaData.id;

    makeNamedApiRequest(saveRequestId, apiMediaSetDescription(mid, description.value))
        .onSuccess((res) => {
            busy.value = false;

            dirty.value = false;

            showSnackBar($t("Successfully saved description"));

            modifyCurrentMediaData(mid, (metadata) => {
                metadata.description_url = res.url || "";
            });

            emitAppEvent(EVENT_NAME_MEDIA_DESCRIPTION_UPDATE, "editor");

            emit("changed");

            onSave();
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

// Exit preventer
const { displayExitConfirmation, onSave, onExitSaveChanges, onExitDiscardChanges } = useExitPreventer(dirty, saveChanges);
</script>
