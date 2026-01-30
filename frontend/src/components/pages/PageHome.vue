<template>
    <div
        class="page-inner scrollbar-stable"
        :class="{ hidden: !display, 'page-editing': editing, 'is-min': min }"
        :style="{
            '--moving-group-width': movingGroupData.width + 'px',
            '--moving-group-height': movingGroupData.height + 'px',
        }"
        @scroll.passive="onScroll"
    >
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
                    <button type="button" class="btn btn-primary btn-mr" @click="showAddRowPrepend">
                        <i class="fas fa-plus"></i> {{ $t("Add new row") }}
                    </button>
                </div>

                <HomePageRow
                    v-for="(g, i) in groups"
                    :key="g.id"
                    :group="g"
                    :row-size="actualRowSize"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :editing="editing"
                    :moving-over="movingGroup && movingGroupData.movingOver === i + 1"
                    :moving-self="movingGroup && movingGroupData.startPosition === i"
                    :current-media="currentMedia"
                    :tag-version="tagVersion"
                    :is-current-group="currentGroup == g.id"
                    :load-tick="loadTick"
                    :is-mobile-size="isMobileSize"
                    @request-rename="showRenameRow"
                    @request-move="showMoveRow"
                    @request-delete="showDeleteRow"
                    @start-moving="onStartMoving"
                    @loaded-current="scrollToCurrentRow"
                    @must-reload="load"
                    @updated-prev-next="onPrevNextUpdated"
                ></HomePageRow>

                <div
                    v-if="movingGroup && groups.length > 0 && movingGroupData.movingOver > groups.length"
                    class="home-page-moving-groups-extra-padding"
                ></div>

                <HomePageRow
                    v-if="movingGroup && movingGroupData.group"
                    :moving-left="movingGroupData.x + 'px'"
                    :moving-top="movingGroupData.y + 'px'"
                    :group="movingGroupData.group"
                    :row-size="actualRowSize"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :editing="editing"
                    :moving="true"
                    :current-media="currentMedia"
                    :tag-version="tagVersion"
                    :moving-initial-elements="initialMovingElements"
                    :moving-initial-scroll="initialMovingScroll"
                ></HomePageRow>

                <div v-if="editing && groups.length < maxGroupsCount" class="home-add-row-form">
                    <button type="button" class="btn btn-primary btn-mr" @click="showAddRow">
                        <i class="fas fa-plus"></i> {{ $t("Add new row") }}
                    </button>
                </div>
            </div>
        </div>

        <HomePageCreateRowModal
            v-if="displayRowAdd"
            v-model:display="displayRowAdd"
            :prepend="displayRowAddPrepend"
            @new-row="onRowAdded"
        ></HomePageCreateRowModal>

        <HomePageRenameRowModal
            v-if="displayRowRename"
            v-model:display="displayRowRename"
            :selected-row="selectedRow"
            :selected-row-type="selectedRowType"
            :selected-row-name="selectedRowName"
            @renamed="onRowRenamed"
            @must-reload="load"
        ></HomePageRenameRowModal>

        <HomePageMoveRowModal
            v-if="displayRowMove"
            v-model:display="displayRowMove"
            :selected-row="selectedRow"
            :selected-row-type="selectedRowType"
            :selected-row-name="selectedRowName"
            :selected-row-position="selectedRowPosition"
            :max-position="groups.length"
            @moved="onRowMoved"
            @must-reload="load"
        ></HomePageMoveRowModal>

        <HomePageDeleteRowModal
            v-if="displayRowDelete"
            v-model:display="displayRowDelete"
            :selected-row="selectedRow"
            :selected-row-type="selectedRowType"
            :selected-row-name="selectedRowName"
            @row-deleted="onRowDeleted"
            @must-reload="load"
        ></HomePageDeleteRowModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import {
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    PagesController,
} from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import { EVENT_NAME_ALBUMS_CHANGED } from "@/control/albums";
import type { HomePageElement, HomePageGroup } from "@/api/api-home";
import { apiHomeGetGroups, apiHomeGroupMove } from "@/api/api-home";
import HomePageRow from "../layout/HomePageRow.vue";
import {
    doHomePageSilentSaveAction,
    EVENT_NAME_HOME_SCROLL_CHANGED,
    getHomePageBackStatePage,
    HomePageGroupTypes,
    type HomePageGroupStartMovingData,
} from "@/utils/home";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { EVENT_NAME_TAGS_UPDATE, TagsController } from "@/control/tags";

const HomePageCreateRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageCreateRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const HomePageRenameRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageRenameRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const HomePageMoveRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageMoveRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const HomePageDeleteRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageDeleteRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const WINDOW_MOBILE_SIZE = 600;

export default defineComponent({
    name: "PageHome",
    components: {
        LoadingOverlay,
        HomePageRow,
        HomePageCreateRowModal,
        HomePageRenameRowModal,
        HomePageMoveRowModal,
        HomePageDeleteRowModal,
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

            dragCheckInterval: null as ReturnType<typeof setInterval> | null,

            initialMovingElements: null as HomePageElement[] | null,
            initialMovingScroll: 0,

            scrollingToCurrent: false,
        };
    },
    data: function () {
        return {
            groups: [] as HomePageGroup[],

            loading: false,
            firstLoaded: false,

            loadTick: 0,

            windowWidth: 0,

            isMobileSize: document.documentElement.clientWidth <= WINDOW_MOBILE_SIZE,

            canWrite: AuthController.CanWrite,

            actualRowSize: this.rowSize || 1,

            displayRowAdd: false,
            displayRowAddPrepend: false,

            selectedRow: -1,
            selectedRowType: HomePageGroupTypes.Custom,
            selectedRowName: "",
            selectedRowPosition: 0,

            displayRowRename: false,
            displayRowMove: false,
            displayRowDelete: false,

            movingGroup: false,

            mouseX: 0,
            mouseY: 0,

            movingGroupData: {
                group: null as HomePageGroup,

                startX: 0,
                startY: 0,

                offsetX: 0,
                offsetY: 0,

                width: 0,
                height: 0,

                x: 0,
                y: 0,

                startPosition: -1,
                movingOver: -1,
            },

            currentMedia: AppStatus.CurrentMedia,
            currentGroup: AppStatus.CurrentHomePageGroup,

            tagVersion: TagsController.TagsVersion,

            currentGroupNext: -1,
            currentGroupPrev: -1,

            currentGroupFirst: -1,
            currentGroupLast: -1,

            storedScroll: 0,
            shouldRestoreStoreScroll: false,
        };
    },
    watch: {
        display: function () {
            this.load();
            if (this.display) {
                this.autoFocus();
            }
        },
        editing: function () {
            this.movingGroup = false;

            this.storeCurrentScroll();

            this.load();
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
        min: function () {
            this.updateActualRowSize();
        },
    },
    mounted: function () {
        this.$addKeyboardHandler(this.handleGlobalKey.bind(this), 20);

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, () => {
            this.canWrite = AuthController.CanWrite;
            this.load();
        });

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, this.load.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_MEDIA_DELETE, this.load.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_CHANGED, this.load.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_NAV_NEXT, this.nextMedia.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_NAV_PREV, this.prevMedia.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onAppStatusChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.updateTagData();

        this.load();

        if (this.display) {
            this.autoFocus();
        }

        this.updateWindowWidth();

        this.windowResizeObserver = new ResizeObserver(this.updateWindowWidth.bind(this));
        this.windowResizeObserver.observe(this.$el);

        this.$listenOnDocumentEvent("mousemove", this.onDocumentMouseMove.bind(this));

        this.$listenOnDocumentEvent("mouseup", this.onDocumentMouseUp.bind(this));
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
        PagesController.OnPageUnload();
        this.windowResizeObserver.disconnect();
        if (this.dragCheckInterval) {
            clearInterval(this.dragCheckInterval);
            this.dragCheckInterval = null;
        }
    },
    methods: {
        scrollToTop: function () {
            let scroll = 0;

            if (this.shouldRestoreStoreScroll) {
                scroll = this.storedScroll * (this.$el.scrollHeight || 1);
                this.shouldRestoreStoreScroll = false;
            }

            this.$el.scrollTop = scroll;
        },

        storeCurrentScroll: function () {
            this.storedScroll = (this.$el.scrollTop || 0) / (this.$el.scrollHeight || 1);
            this.shouldRestoreStoreScroll = true;
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
                    this.scrollToCurrentRow();
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
            this.isMobileSize = document.documentElement.clientWidth <= WINDOW_MOBILE_SIZE;
            this.updateActualRowSize();
            this.onScroll();
        },

        updateActualRowSize: function () {
            const preferRowSize = (this.min ? this.rowSizeMin : this.rowSize) || 1;

            let itemsWidth = this.windowWidth / preferRowSize;

            itemsWidth = Math.min(itemsWidth, Math.min(this.windowWidth, this.maxItemsSize || 0));

            itemsWidth = Math.max(1, Math.max(itemsWidth, this.minItemsSize || 0));

            this.actualRowSize = Math.ceil(this.windowWidth / itemsWidth);

            if (this.currentMedia) {
                this.scrollToCurrentRow();
            }
        },

        showAddRow: function () {
            this.displayRowAdd = true;
            this.displayRowAddPrepend = false;
        },

        showAddRowPrepend: function () {
            this.displayRowAdd = true;
            this.displayRowAddPrepend = true;
        },

        onRowAdded: function (row: HomePageGroup, prepend: boolean) {
            if (prepend) {
                this.groups.unshift(row);
            } else {
                this.groups.push(row);
            }
        },

        showRenameRow: function (group: HomePageGroup) {
            this.displayRowRename = true;
            this.selectedRow = group.id;
            this.selectedRowName = group.name;
            this.selectedRowType = group.type;
        },

        onRowRenamed: function (id: number, newName: string) {
            for (const g of this.groups) {
                if (g.id === id) {
                    g.name = newName;
                    break;
                }
            }
        },

        showMoveRow: function (group: HomePageGroup) {
            let startPosition = 0;
            for (let i = 0; i < this.groups.length; i++) {
                if (this.groups[i].id === group.id) {
                    startPosition = i;
                    break;
                }
            }

            this.displayRowMove = true;
            this.selectedRow = group.id;
            this.selectedRowPosition = startPosition;
            this.selectedRowName = group.name;
            this.selectedRowType = group.type;
        },

        onRowMoved: function (id: number, position: number) {
            position = Math.max(0, Math.min(position, this.groups.length));
            for (let i = 0; i < this.groups.length; i++) {
                if (this.groups[i].id === id) {
                    this.groups.splice(position, 0, this.groups.splice(i, 1)[0]);
                    return;
                }
            }
        },

        showDeleteRow: function (group: HomePageGroup) {
            this.displayRowDelete = true;
            this.selectedRow = group.id;
            this.selectedRowName = group.name;
            this.selectedRowType = group.type;
        },

        onRowDeleted: function (id: number) {
            for (let i = 0; i < this.groups.length; i++) {
                if (this.groups[i].id === id) {
                    this.groups.splice(i, 1);
                    return;
                }
            }
        },

        onStartMoving: function (group: HomePageGroup, moveData: HomePageGroupStartMovingData) {
            let startPosition = -1;
            for (let i = 0; i < this.groups.length; i++) {
                if (this.groups[i].id === group.id) {
                    startPosition = i;
                    break;
                }
            }

            this.mouseX = moveData.startX;
            this.mouseY = moveData.startY;

            this.movingGroup = true;
            this.movingGroupData.group = group;
            this.movingGroupData.startPosition = startPosition;

            this.movingGroupData.startX = moveData.startX;
            this.movingGroupData.startY = moveData.startY;

            this.movingGroupData.offsetX = moveData.offsetX;
            this.movingGroupData.offsetY = moveData.offsetY;

            this.movingGroupData.width = moveData.width;
            this.movingGroupData.height = moveData.height;

            this.movingGroupData.x = moveData.startX - moveData.offsetX;
            this.movingGroupData.y = moveData.startY - moveData.offsetY;

            this.initialMovingElements = moveData.initialElements;
            this.initialMovingScroll = moveData.initialScroll;

            this.updateMovingOver();

            if (this.dragCheckInterval) {
                clearInterval(this.dragCheckInterval);
                this.dragCheckInterval = null;
            }
            this.dragCheckInterval = setInterval(this.onDragCheck.bind(this), 40);
        },

        updateMovingOver: function () {
            const topAddButtonForm = this.$el.querySelector(".home-add-row-form");
            const container = this.$el;

            if (!container || !topAddButtonForm) {
                this.movingGroupData.movingOver = -1;
                return;
            }

            const offsetTop = container.getBoundingClientRect().top + topAddButtonForm.getBoundingClientRect().height;
            const scrollTop = container.scrollTop || 0;

            const y = this.movingGroupData.y + Math.round(this.movingGroupData.height / 2);

            const height = this.movingGroupData.height;
            const expectedIndex = Math.round((y - offsetTop + scrollTop) / height);

            this.movingGroupData.movingOver = Math.min(this.groups.length + 1, Math.max(1, 1 + expectedIndex));
        },

        onDragCheck: function () {
            const con = this.$el;

            if (!con) {
                return;
            }

            const conBounds = con.getBoundingClientRect();

            if (this.mouseX >= conBounds.left - this.movingGroupData.height) {
                // Auto scroll

                const relTop = (this.mouseY - conBounds.top) / (conBounds.height || 1);
                const scrollStep = Math.floor(conBounds.height / 20);

                if (relTop <= 0.1) {
                    con.scrollTop = Math.max(0, con.scrollTop - scrollStep);
                } else if (relTop >= 0.9) {
                    con.scrollTop = Math.min(con.scrollHeight - conBounds.height, con.scrollTop + scrollStep);
                }
            }

            this.updateMovingOver();
        },

        onDocumentMouseMove: function (event: MouseEvent) {
            if (!this.movingGroup) {
                return;
            }

            if (typeof event.pageX !== "number" || typeof event.pageY !== "number") {
                return;
            }

            if (isNaN(event.pageX) || isNaN(event.pageY)) {
                return;
            }

            this.mouseX = event.pageX;
            this.mouseY = event.pageY;

            this.movingGroupData.x = event.pageX - this.movingGroupData.offsetX;
            this.movingGroupData.y = event.pageY - this.movingGroupData.offsetY;
        },

        onDocumentMouseUp: function (event: MouseEvent) {
            if (!this.movingGroup) {
                return;
            }

            if (this.dragCheckInterval) {
                clearInterval(this.dragCheckInterval);
                this.dragCheckInterval = null;
            }

            event.stopPropagation();

            this.movingGroup = false;

            if (!this.movingGroupData.group) {
                return;
            }

            const groupId = this.movingGroupData.group.id;

            let position =
                this.movingGroupData.movingOver > this.movingGroupData.startPosition + 1
                    ? this.movingGroupData.movingOver - 1
                    : this.movingGroupData.movingOver;

            position = Math.max(0, Math.min(this.groups.length, position - 1));

            const startPosition = this.movingGroupData.startPosition;

            if (startPosition === -1 || position === startPosition) {
                return;
            }

            this.doSilentMove(groupId, position);
            this.groups.splice(position, 0, this.groups.splice(startPosition, 1)[0]);
        },

        doSilentMove: function (rowId: number, position: number) {
            doHomePageSilentSaveAction((callback) => {
                makeApiRequest(apiHomeGroupMove(rowId, position))
                    .onSuccess(() => {
                        callback();
                    })
                    .onCancel(() => {
                        callback();
                    })
                    .onRequestError((err, handleErr) => {
                        callback();
                        handleErr(err, {
                            unauthorized: () => {
                                AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            },
                            accessDenied: () => {
                                AuthController.CheckAuthStatus();
                            },
                            notFound: () => {
                                this.load();
                            },
                            temporalError: () => {
                                this.load();
                            },
                        });
                    })
                    .onUnexpectedError((err) => {
                        callback();
                        console.error(err);
                        this.load();
                    });
            });
        },

        onAppStatusChanged: function () {
            const changed = this.currentMedia !== AppStatus.CurrentMedia || this.currentGroup !== AppStatus.CurrentHomePageGroup;

            this.currentMedia = AppStatus.CurrentMedia;
            this.currentGroup = AppStatus.CurrentHomePageGroup;

            if (changed) {
                this.currentGroupPrev = -1;
                this.currentGroupNext = -1;
                this.currentGroupFirst = -1;
                this.currentGroupLast = -1;
                PagesController.OnHomeGroupLoad(this.currentGroupPrev >= 0, this.currentGroupNext >= 0);

                this.scrollToCurrentRow();
            }
        },

        scrollToCurrentRow: function () {
            if (this.scrollingToCurrent) {
                return;
            }
            this.scrollingToCurrent = true;

            const backState = getHomePageBackStatePage();

            if (backState !== null && this.currentGroup === -1) {
                nextTick(() => {
                    this.scrollingToCurrent = false;
                    const currentElem = this.$el.querySelector(".home-page-row-" + backState);
                    if (currentElem) {
                        currentElem.scrollIntoView();
                    }
                });
                return;
            }

            nextTick(() => {
                this.scrollingToCurrent = false;
                const currentElem = this.$el.querySelector(".home-page-row.current");
                if (currentElem) {
                    currentElem.scrollIntoView();
                }
            });
        },

        updateTagData: function () {
            this.tagVersion = TagsController.TagsVersion;
        },

        onScroll: function () {
            AppEvents.Emit(EVENT_NAME_HOME_SCROLL_CHANGED);
        },

        onPrevNextUpdated: function (
            currentGroupPrev: number,
            currentGroupNext: number,
            currentGroupFirst: number,
            currentGroupLast: number,
        ) {
            this.currentGroupPrev = currentGroupPrev;
            this.currentGroupNext = currentGroupNext;
            this.currentGroupFirst = currentGroupFirst;
            this.currentGroupLast = currentGroupLast;

            PagesController.OnHomeGroupLoad(this.currentGroupPrev >= 0, this.currentGroupNext >= 0);
        },

        goToMedia: function (id: number) {
            AppStatus.ClickOnMedia(id, true, this.currentGroup);
        },

        nextMedia: function () {
            if (this.currentGroupNext >= 0) {
                this.goToMedia(this.currentGroupNext);
            }
        },

        prevMedia: function () {
            if (this.currentGroupPrev >= 0) {
                this.goToMedia(this.currentGroupPrev);
            }
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPageVisible() || !this.display || !event.key || event.ctrlKey || this.editing) {
                return false;
            }

            if (event.shiftKey && event.key === "PageUp") {
                const activeElement = document.activeElement as HTMLElement;

                if (activeElement && activeElement.classList.contains("home-page-row") && activeElement.previousElementSibling) {
                    const sibling = activeElement.previousElementSibling as HTMLElement;
                    sibling.focus();
                } else if (
                    activeElement &&
                    activeElement.parentElement &&
                    activeElement.parentElement.parentElement &&
                    activeElement.parentElement.parentElement.parentElement &&
                    activeElement.parentElement.parentElement.parentElement.parentElement
                ) {
                    const rowElement = activeElement.parentElement.parentElement.parentElement.parentElement;

                    if (rowElement.classList.contains("home-page-row") && rowElement.previousElementSibling) {
                        const sibling = rowElement.previousElementSibling as HTMLElement;
                        sibling.focus();
                    }
                }

                return true;
            }

            if (event.shiftKey && event.key === "PageDown") {
                const activeElement = document.activeElement as HTMLElement;

                if (activeElement && activeElement.classList.contains("home-page-row") && activeElement.nextElementSibling) {
                    const sibling = activeElement.nextElementSibling as HTMLElement;
                    sibling.focus();
                } else if (
                    activeElement &&
                    activeElement.parentElement &&
                    activeElement.parentElement.parentElement &&
                    activeElement.parentElement.parentElement.parentElement &&
                    activeElement.parentElement.parentElement.parentElement.parentElement
                ) {
                    const rowElement = activeElement.parentElement.parentElement.parentElement.parentElement;

                    if (rowElement.classList.contains("home-page-row") && rowElement.nextElementSibling) {
                        const sibling = rowElement.nextElementSibling as HTMLElement;
                        sibling.focus();
                    }
                }

                return true;
            }

            if (event.key === "Home") {
                if (this.currentGroupFirst > 0) {
                    this.goToMedia(this.currentGroupFirst);
                }
                return true;
            }

            if (event.key === "End") {
                if (this.currentGroupLast > 0) {
                    this.goToMedia(this.currentGroupLast);
                }
                return true;
            }

            if (event.key === "PageUp" || event.key === "ArrowLeft") {
                this.prevMedia();
                return true;
            }

            if (event.key === "PageDown" || event.key === "ArrowRight") {
                this.nextMedia();
                return true;
            }

            return false;
        },
    },
});
</script>
