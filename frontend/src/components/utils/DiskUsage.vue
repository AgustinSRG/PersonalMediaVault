<template>
    <div class="upload-list-item-status-bar">
        <div v-if="loaded" class="upload-list-item-status-bar-current success" :style="{ width: percentUsage(100 - usage) }"></div>
        <div v-if="loaded" class="upload-list-item-status-bar-current right error" :style="{ width: percentUsage(usage) }"></div>
        <div
            v-if="loaded || loadingTimedOut"
            class="upload-list-item-status-bar-text"
            :title="renderDiskUsage(loaded, usage, available, free, total)"
        >
            {{ renderDiskUsage(loaded, usage, available, free, total) }}
        </div>
    </div>
</template>

<script lang="ts">
import { apiDiskUsage } from "@/api/api-about";
import { AppEvents } from "@/control/app-events";
import { EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_UPLOAD_LIST_UPDATE, UploadEntryMin } from "@/control/upload";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { renderSize } from "@/utils/size";
import { getUniqueStringId } from "@/utils/unique-id";
import { abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { defineComponent } from "vue";

const RELOAD_INTERVAL_MS = 60 * 1000;

export default defineComponent({
    name: "DiskUsage",
    setup: function () {
        return {
            loadRequestId: getUniqueStringId(),
            loadingTimeout: null as ReturnType<typeof setTimeout> | null,
        };
    },
    data: function () {
        return {
            usage: 0,
            available: 0,
            free: 0,
            total: 0,

            loaded: false,
            loadingTimedOut: false,

            loading: false,
        };
    },
    mounted: function () {
        this.loadingTimeout = setTimeout(() => {
            this.loadingTimedOut = true;
        }, 333);

        this.load();

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.load.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_UPLOAD_LIST_UPDATE, this.onUploadDone.bind(this));
    },
    beforeUnmount: function () {
        if (this.loadingTimeout) {
            clearTimeout(this.loadingTimeout);
        }

        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    methods: {
        load: function () {
            this.loading = true;

            clearNamedTimeout(this.loadRequestId);

            makeNamedApiRequest(this.loadRequestId, apiDiskUsage())
                .onSuccess((res) => {
                    this.loading = false;

                    if (this.loadingTimeout) {
                        clearTimeout(this.loadingTimeout);
                        this.loadingTimeout = null;
                    }

                    this.loaded = true;

                    this.usage = res.usage || 0;
                    this.available = res.available || 0;
                    this.free = res.free || 0;
                    this.total = res.total || 0;

                    setNamedTimeout(this.loadRequestId, RELOAD_INTERVAL_MS, this.load.bind(this));
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        percentUsage: function (usage: number) {
            return Math.min(100, Math.max(0, Math.round(usage))) + "%";
        },

        renderDiskUsage: function (loaded: boolean, usage: number, available: number, free: number, total: number): string {
            if (!loaded) {
                return this.$t("Loading disk usage") + "...";
            }

            return (
                this.$t("Disk usage") +
                ": " +
                this.percentUsage(usage) +
                " (" +
                renderSize(available) +
                " " +
                this.$t("available") +
                ", " +
                renderSize(free) +
                " " +
                this.$t("free") +
                ", " +
                renderSize(total) +
                " " +
                this.$t("total size") +
                ")"
            );
        },

        onUploadDone: function (mode: "push" | "rm" | "update" | "clear", entry?: UploadEntryMin) {
            if (this.loading) {
                return;
            }

            if (mode !== "update") {
                return;
            }

            if (entry.status === "ready") {
                this.load();
            }
        },
    },
});
</script>
