<template>
    <div class="upload-list-item-status-bar">
        <div
            v-if="loaded"
            class="upload-list-item-status-bar-current"
            :class="{ error: usage > 90 }"
            :style="{ width: percentUsage(usage) }"
        ></div>
        <div v-if="loaded || loadingTimedOut" class="upload-list-item-status-bar-text" :title="renderDiskUsage(loaded, usage, free, total)">
            {{ renderDiskUsage(loaded, usage, free, total) }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { apiDiskUsage } from "@/api/api-about";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { useTimeout } from "@/composables/use-timeout";
import { emitAppEvent, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED, EVENT_NAME_UPLOAD_LIST_ENTRY_READY } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { renderSize } from "@/utils/size";
import { makeNamedApiRequest } from "@asanrom/request-browser";
import { onMounted, ref } from "vue";

// Translation function
const { $t } = useI18n();

/**
 * Renders the usage as a percentage
 * @param usage The dick usage
 * @returns The usage as a percentage
 */
const percentUsage = (usage: number): string => {
    return Math.min(100, Math.max(0, Math.round(usage))) + "%";
};

/**
 * Renders the disk usage in order to display it to the user
 * @param loaded True if loaded
 * @param usage The disk usage
 * @param free The free space
 * @param total The total space
 */
const renderDiskUsage = (loaded: boolean, usage: number, free: number, total: number): string => {
    if (!loaded) {
        return $t("Loading disk usage") + "...";
    }

    return $t("Disk usage") + ": " + percentUsage(usage) + " (" + renderSize(total - free) + " / " + renderSize(total) + ")";
};

// Disk usage
const usage = ref(0);

// Free disk space
const free = ref(0);

// Total disk space
const total = ref(0);

// Interval (in milliseconds) for automatically reloading the disk usage
const RELOAD_INTERVAL_MS = 60 * 1000;

// True of loaded for the first time
const loaded = ref(false);

// True if the loading process took too long, so the loader should be displayed
const loadingTimedOut = ref(false);

// Timeout for displaying the loading status
const loadingTimeout = useTimeout();

// Unique request ID for loading
const loadRequestId = useRequestId();

// True if loading
const loading = ref(false);

/**
 * Loads the disk usage
 * calling the corresponding API
 */
const load = () => {
    if (AuthController.Locked) {
        return;
    }

    loading.value = true;

    clearNamedTimeout(loadRequestId);

    makeNamedApiRequest(loadRequestId, apiDiskUsage())
        .onSuccess((res) => {
            loading.value = false;
            loadingTimeout.clear();

            loaded.value = true;

            usage.value = res.usage || 0;
            free.value = res.free || 0;
            total.value = res.total || 0;

            setNamedTimeout(loadRequestId, RELOAD_INTERVAL_MS, load);
        })
        .onRequestError((err, handleErr) => {
            loading.value = false;
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    setNamedTimeout(loadRequestId, 1500, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            loading.value = false;
            console.error(err);
            setNamedTimeout(loadRequestId, 1500, load);
        });
};

/// Delay to display the loader (milliseconds)
const LOADING_DISPLAY_DELAY = 333;

/// On mounter, load the data for the first time
/// and set the timeout to show the loader
onMounted(() => {
    loadingTimeout.set(() => {
        loadingTimedOut.value = true;
    }, LOADING_DISPLAY_DELAY);

    load();
});

// Load the data again when the auth status changes
onApplicationEvent(EVENT_NAME_AUTH_CHANGED, load);

// If a new file is uploaded, reload the disk usage immediately,
// unless it is already loading
onApplicationEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_READY, () => {
    if (loading.value) {
        return;
    }

    load();
});
</script>
