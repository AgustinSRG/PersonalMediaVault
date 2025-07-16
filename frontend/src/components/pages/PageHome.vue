<template>
    <div class="page-inner" :class="{ hidden: !display, 'page-editing': editing }">
        <div class="search-results auto-focus" tabindex="-1">
            <LoadingOverlay v-if="loading"></LoadingOverlay>

            <div v-else-if="!loading && groups.length == 0 && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ $t("The home page is empty") }}
                </div>
                <div v-if="!editing" class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="load"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
                <div v-else class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary"><i class="fas fa-plus"></i> {{ $t("Add first row") }}</button>
                </div>
            </div>

            <div v-else>
                <div v-if="editing && groups.length < maxGroupsCount" class="home-add-row-form">
                    <button type="button" class="btn btn-primary btn-mr"><i class="fas fa-plus"></i> {{ $t("Add new row") }}</button>
                </div>

                <HomePageRow
                    v-for="g in groups"
                    :key="g.id"
                    :group="g"
                    :row-size="actualRowSize"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :editing="editing"
                ></HomePageRow>

                <div v-if="editing && groups.length < maxGroupsCount" class="home-add-row-form">
                    <button type="button" class="btn btn-primary btn-mr"><i class="fas fa-plus"></i> {{ $t("Add new row") }}</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { EVENT_NAME_MEDIA_DELETE, EVENT_NAME_MEDIA_METADATA_CHANGE, PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import { EVENT_NAME_ALBUMS_CHANGED } from "@/control/albums";
import type { HomePageGroup } from "@/api/api-home";
import { apiHomeGetGroups } from "@/api/api-home";
import HomePageRow from "../utils/HomePageRow.vue";

export default defineComponent({
    name: "PageHome",
    components: {
        LoadingOverlay,
        HomePageRow,
    },
    props: {
        display: Boolean,
        min: Boolean,
        pageSize: Number,
        displayTitles: Boolean,

        rowSize: Number,
        rowSizeMin: Number,
        minItemsSize: Number,
        maxItemsSize: Number,

        editing: Boolean,
    },
    setup() {
        return {
            maxGroupsCount: 1024,

            loadRequestId: getUniqueStringId(),
            windowResizeObserver: null as ResizeObserver,
        };
    },
    data: function () {
        return {
            groups: [] as HomePageGroup[],

            loading: false,
            firstLoaded: false,

            loadTick: 0,

            windowWidth: 0,

            canWrite: AuthController.CanWrite,

            actualRowSize: this.rowSize || 1,
        };
    },
    watch: {
        display: function () {
            this.load();
            if (this.display) {
                this.autoFocus();
            }
        },
        pageSize: function () {
            this.updatePageSize();
        },
        rowSize: function () {
            this.updateActualRowSize();
        },
        rowSizeMin: function () {
            this.updateActualRowSize();
        },
        minItemsSize: function () {
            this.updateActualRowSize();
        },
        maxItemsSize: function () {
            this.updateActualRowSize();
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, () => {
            this.canWrite = AuthController.CanWrite;
            this.load();
        });

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, this.load.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_MEDIA_DELETE, this.load.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_CHANGED, this.load.bind(this));

        this.load();

        if (this.display) {
            this.autoFocus();
        }

        this.updateWindowWidth();

        this.windowResizeObserver = new ResizeObserver(this.updateWindowWidth.bind(this));
        this.windowResizeObserver.observe(this.$el);
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
        PagesController.OnPageUnload();
        this.windowResizeObserver.disconnect();
    },
    methods: {
        scrollToTop: function () {
            this.$el.scrollTop = 0;
        },

        autoFocus: function () {
            nextTick(() => {
                const el = this.$el.querySelector(".auto-focus");
                if (el) {
                    el.focus();
                    if (el.select) {
                        el.select();
                    }
                }
            });
        },

        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            this.scrollToTop();

            setNamedTimeout(this.loadRequestId, 330, () => {
                this.loading = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiHomeGetGroups())
                .onSuccess((groups) => {
                    clearNamedTimeout(this.loadRequestId);
                    this.loading = false;
                    this.firstLoaded = true;
                    this.loadTick++;
                    this.groups = groups;
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            // Retry
                            this.loading = true;
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    this.loading = true;
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        updatePageSize: function () {
            this.load();
        },

        updateWindowWidth: function () {
            this.windowWidth = this.$el.getBoundingClientRect().width;
            this.updateActualRowSize();
        },

        updateActualRowSize: function () {
            const preferRowSize = (this.min ? this.rowSizeMin : this.rowSize) || 1;

            let itemsWidth = this.windowWidth / preferRowSize;

            itemsWidth = Math.min(itemsWidth, Math.min(this.windowWidth, this.maxItemsSize || 0));

            itemsWidth = Math.max(1, Math.max(itemsWidth, this.minItemsSize || 0));

            this.actualRowSize = Math.ceil(this.windowWidth / itemsWidth);
        },

        changeNameFilter: function () {},

        editHomePage: function () {},
    },
});
</script>
