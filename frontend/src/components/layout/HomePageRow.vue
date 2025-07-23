<template>
    <div
        class="home-page-row"
        :class="{
            moving: moving,
            'moving-over': movingOver,
            'moving-self': movingSelf,
            current: isCurrentGroup,
            customizable: group.type == groupTypeCustom,
            'fast-transition': fastTransition,
        }"
        :style="{ '--actual-row-size': rowSize + '', '--row-scroll-index': rowIndex + '', top: movingTop, left: movingLeft }"
        tabindex="-1"
        @keydown="onKeyDown"
    >
        <div class="home-page-row-inner">
            <div
                class="home-page-row-head"
                :draggable="!isTouchDevice && !movePositionModalDisplay && !displayAddElement && !contextMenuShown"
                @dragstart="onDrag"
            >
                <div class="home-page-row-title" :title="getGroupName(group)">{{ getGroupName(group) }}</div>
                <div v-if="editing" class="home-page-row-head-buttons">
                    <button
                        v-if="!loading && group.type == groupTypeCustom"
                        type="button"
                        class="page-header-btn"
                        :title="$t('Add elements')"
                        @click="addElements"
                    >
                        <i class="fas fa-plus"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Rename row')" @click="renameRow">
                        <i class="fas fa-pencil-alt"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Move row')" @click="moveRow">
                        <i class="fas fa-arrows-up-down-left-right"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Delete row')" @click="deleteRow">
                        <i class="fas fa-trash-alt"></i>
                    </button>
                </div>
            </div>

            <div
                v-if="(!editing || group.type !== groupTypeCustom) && !loadDisplay && firstLoaded && elements.length === 0"
                class="home-page-row-content home-page-row-loading"
            >
                <div v-for="f in loadingFiller.slice(0, 1)" :key="f" class="search-result-item hidden">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
                <div v-if="editing && group.type !== groupTypeCustom" class="home-page-row-message">
                    <span>{{ $t("This row cannot be customized") }} </span>
                </div>
                <div v-else class="home-page-row-message home-page-row-message-empty">
                    <span><i class="fas fa-box-open"></i> {{ $t("This row is empty") }}</span>
                </div>
            </div>

            <div v-else-if="!visible" class="home-page-row-content home-page-row-loading">
                <div v-for="f in loadingFiller.slice(0, 1)" :key="f" class="search-result-item hidden">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
            </div>

            <div v-else-if="loadDisplay || !firstLoaded" class="home-page-row-content home-page-row-loading">
                <div v-for="f in loadingFiller" :key="f" class="search-result-item" :class="{ hidden: !loadDisplay }">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
            </div>

            <div
                v-else
                class="home-page-row-content"
                :style="{
                    '--moving-element-width': draggingData.width + 'px',
                }"
                @scroll.passive="onContentScroll"
            >
                <div
                    v-for="(item, i) in elements"
                    :key="i"
                    class="search-result-item"
                    :class="{
                        current: !editing && isCurrent(item, currentMedia, isCurrentGroup),
                        dragging: dragging && draggingData.startPosition === i,
                        'dragging-over': dragging && draggingData.movingOver === i + 1,
                    }"
                >
                    <a
                        class="clickable"
                        :class="{ moveable: editing, ['home-page-row-element-' + i]: true }"
                        :href="getElementURL(item)"
                        target="_blank"
                        rel="noopener noreferrer"
                        @click="goToElement(item, $event)"
                        @focus="focusElementIndex(i)"
                        @contextmenu="showContextMenu(i, $event)"
                        @dragstart="onDragStart(i, $event)"
                    >
                        <div class="search-result-thumb" :title="renderHintTitle(item, tagVersion)">
                            <div v-if="item.media" class="search-result-thumb-inner">
                                <div v-if="!item.media.thumbnail" class="no-thumb">
                                    <i v-if="item.media.type === 1" class="fas fa-image"></i>
                                    <i v-else-if="item.media.type === 2" class="fas fa-video"></i>
                                    <i v-else-if="item.media.type === 3" class="fas fa-headphones"></i>
                                    <i v-else class="fas fa-ban"></i>
                                </div>
                                <ThumbImage v-if="item.media.thumbnail" :src="getThumbnail(item.media.thumbnail)"></ThumbImage>
                                <DurationIndicator
                                    v-if="item.media.type === 2 || item.media.type === 3"
                                    :type="item.media.type"
                                    :duration="item.media.duration"
                                ></DurationIndicator>
                            </div>
                            <div v-else-if="item.album" class="search-result-thumb-inner">
                                <div v-if="!item.album.thumbnail" class="no-thumb">
                                    <i class="fas fa-list-ol"></i>
                                </div>
                                <ThumbImage v-if="item.album.thumbnail" :src="getThumbnail(item.album.thumbnail)"></ThumbImage>
                                <div v-if="item.album.size == 0" class="thumb-bottom-right-tag" :title="$t('Album') + ' - ' + $t('Empty')">
                                    <i class="fas fa-list-ol"></i> {{ $t("Empty") }}
                                </div>
                                <div
                                    v-else-if="item.album.size == 1"
                                    class="thumb-bottom-right-tag"
                                    :title="$t('Album') + ' - 1 ' + $t('item')"
                                >
                                    <i class="fas fa-list-ol"></i> 1 {{ $t("item") }}
                                </div>
                                <div
                                    v-else-if="item.album.size > 1"
                                    class="thumb-bottom-right-tag"
                                    :title="$t('Album') + ' - ' + item.album.size + ' ' + $t('items')"
                                >
                                    <i class="fas fa-list-ol"></i> {{ item.album.size }} {{ $t("items") }}
                                </div>
                            </div>
                            <div v-else class="search-result-thumb-inner">
                                <div class="no-thumb">
                                    <i class="fas fa-ban"></i>
                                </div>
                            </div>

                            <button
                                v-if="editing"
                                type="button"
                                class="home-page-row-context-btn"
                                :title="$t('Options to modify element')"
                                @click="showContextMenu(i, $event)"
                            >
                                <i class="fas fa-bars"></i>
                            </button>
                        </div>
                        <div v-if="displayTitles" class="search-result-title">
                            {{ getTitle(item) }}
                        </div>
                    </a>
                </div>

                <div
                    v-if="dragging && elements.length > 0 && draggingData.movingOver > elements.length"
                    class="home-page-moving-elements-extra-padding"
                ></div>

                <div
                    v-if="editing && group.type === groupTypeCustom && elements.length < limitCustomGroupElements"
                    class="search-result-item add-home-element"
                >
                    <div tabindex="0" class="search-result-thumb add-home-element-btn" :title="$t('Add elements')" @click="addElements">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-plus"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Add elements") }}</div>
                </div>

                <div v-else-if="editing && group.type !== groupTypeCustom" class="home-page-row-message">
                    <span>{{ $t("This row cannot be customized") }}</span>
                </div>

                <div v-if="!editing && rowIndex > 0" class="home-page-row-go-left">
                    <button type="button" class="home-page-row-go-button" :title="$t('Scroll to the left')" @click="goLeft">
                        <i class="fas fa-chevron-left"></i>
                    </button>
                </div>

                <div v-if="!editing && rowIndex < rowSplitCount - 1" class="home-page-row-go-right">
                    <button type="button" class="home-page-row-go-button" :title="$t('Scroll to the right')" @click="goRight">
                        <i class="fas fa-chevron-right"></i>
                    </button>
                </div>
            </div>
        </div>

        <div
            v-if="dragging && draggingElement"
            class="search-result-item home-element-dragging"
            :class="{ current: !editing && isCurrent(draggingElement, currentMedia, isCurrentGroup) }"
            :style="{
                left: draggingData.x + 'px',
                top: draggingData.y + 'px',
                width: draggingData.width + 'px',
                height: draggingData.height + 'px',
                'min-width': draggingData.width + 'px',
                'min-height': draggingData.height + 'px',
            }"
        >
            <a class="clickable" :href="getElementURL(draggingElement)" target="_blank" rel="noopener noreferrer">
                <div class="search-result-thumb" :title="renderHintTitle(draggingElement, tagVersion)">
                    <div v-if="draggingElement.media" class="search-result-thumb-inner">
                        <div v-if="!draggingElement.media.thumbnail" class="no-thumb">
                            <i v-if="draggingElement.media.type === 1" class="fas fa-image"></i>
                            <i v-else-if="draggingElement.media.type === 2" class="fas fa-video"></i>
                            <i v-else-if="draggingElement.media.type === 3" class="fas fa-headphones"></i>
                            <i v-else class="fas fa-ban"></i>
                        </div>
                        <ThumbImage
                            v-if="draggingElement.media.thumbnail"
                            :src="getThumbnail(draggingElement.media.thumbnail)"
                        ></ThumbImage>
                        <DurationIndicator
                            v-if="draggingElement.media.type === 2 || draggingElement.media.type === 3"
                            :type="draggingElement.media.type"
                            :duration="draggingElement.media.duration"
                        ></DurationIndicator>
                    </div>
                    <div v-else-if="draggingElement.album" class="search-result-thumb-inner">
                        <div v-if="!draggingElement.album.thumbnail" class="no-thumb">
                            <i class="fas fa-list-ol"></i>
                        </div>
                        <ThumbImage
                            v-if="draggingElement.album.thumbnail"
                            :src="getThumbnail(draggingElement.album.thumbnail)"
                        ></ThumbImage>
                        <div
                            v-if="draggingElement.album.size == 0"
                            class="thumb-bottom-right-tag"
                            :title="$t('Album') + ' - ' + $t('Empty')"
                        >
                            <i class="fas fa-list-ol"></i> {{ $t("Empty") }}
                        </div>
                        <div
                            v-else-if="draggingElement.album.size == 1"
                            class="thumb-bottom-right-tag"
                            :title="$t('Album') + ' - 1 ' + $t('item')"
                        >
                            <i class="fas fa-list-ol"></i> 1 {{ $t("item") }}
                        </div>
                        <div
                            v-else-if="draggingElement.album.size > 1"
                            class="thumb-bottom-right-tag"
                            :title="$t('Album') + ' - ' + draggingElement.album.size + ' ' + $t('items')"
                        >
                            <i class="fas fa-list-ol"></i> {{ draggingElement.album.size }} {{ $t("items") }}
                        </div>
                    </div>
                    <div v-else class="search-result-thumb-inner">
                        <div class="no-thumb">
                            <i class="fas fa-ban"></i>
                        </div>
                    </div>
                    <button v-if="editing" type="button" class="home-page-row-context-btn">
                        <i class="fas fa-bars"></i>
                    </button>
                </div>
                <div v-if="displayTitles" class="search-result-title">
                    {{ getTitle(draggingElement) }}
                </div>
            </a>
        </div>

        <HomePageElementContextMenu
            v-if="contextMenuShown"
            v-model:shown="contextMenuShown"
            :element-index="contextMenuSelectedIndex"
            :x="contextMenuX"
            :y="contextMenuY"
            @element-remove="onElementRemove"
            @change-pos="onChangePositionRequest"
        ></HomePageElementContextMenu>

        <HomePageRowAddElementModal
            v-if="displayAddElement"
            v-model:display="displayAddElement"
            :group-id="group.id"
            :group-elements="elements"
            @must-reload="load"
            @added-element="load"
        ></HomePageRowAddElementModal>

        <HomePageMoveElementModal
            v-if="movePositionModalDisplay"
            v-model:display="movePositionModalDisplay"
            :selected-position="movePositionModalSelectedIndex"
            :max-position="elements.length - 1"
            @move-element="onMoveElement"
        ></HomePageMoveElementModal>
    </div>
</template>

<script lang="ts">
import type { HomePageElement, HomePageGroup } from "@/api/api-home";
import { apiHomeGetGroupElements, apiHomeGroupDeleteElement, apiHomeGroupMoveElement, getHomePageElementReference } from "@/api/api-home";
import { AppStatus } from "@/control/app-status";
import { TagsController } from "@/control/tags";
import { generateURIQuery, getAssetURL } from "@/utils/api";
import type { HomePageGroupStartMovingData } from "@/utils/home";
import { doHomePageSilentSaveAction, EVENT_NAME_HOME_SCROLL_CHANGED, getDefaultGroupName, HomePageGroupTypes } from "@/utils/home";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { renderDateAndTime } from "@/utils/time";
import { isTouchDevice } from "@/utils/touch";
import { getUniqueStringId } from "@/utils/unique-id";
import { abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import type { PropType } from "vue";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import DurationIndicator from "@/components/utils/DurationIndicator.vue";
import ThumbImage from "@/components/utils/ThumbImage.vue";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { apiSearch } from "@/api/api-search";
import { AppEvents } from "@/control/app-events";
import { apiAlbumsGetAlbums } from "@/api/api-albums";

const HomePageRowAddElementModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageRowAddElementModal.vue"),
});

const HomePageElementContextMenu = defineAsyncComponent({
    loader: () => import("@/components/layout/HomePageElementContextMenu.vue"),
});

const HomePageMoveElementModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageMoveElementModal.vue"),
});

export default defineComponent({
    name: "HomePageRow",
    components: {
        DurationIndicator,
        ThumbImage,
        HomePageRowAddElementModal,
        HomePageElementContextMenu,
        HomePageMoveElementModal,
    },
    props: {
        rowSize: Number,

        pageSize: Number,

        displayTitles: Boolean,

        editing: Boolean,

        moving: Boolean,
        movingInitialElements: Array as PropType<HomePageElement[]>,
        movingInitialScroll: Number,

        movingOver: Boolean,
        movingSelf: Boolean,
        movingLeft: String,
        movingTop: String,

        group: Object as PropType<HomePageGroup>,

        loadTick: Number,

        tagVersion: Number,
        currentMedia: Number,

        isCurrentGroup: Boolean,
    },
    emits: ["request-rename", "request-move", "request-delete", "start-moving", "loaded-current", "must-reload", "updated-prev-next"],
    setup() {
        return {
            limitCustomGroupElements: 256,
            groupTypeCustom: HomePageGroupTypes.Custom,

            loadRequestId: getUniqueStringId(),

            isTouchDevice: isTouchDevice(),

            dragCheckInterval: null as ReturnType<typeof setInterval> | null,
            mouseMoveHandler: null,
            mouseUpHandler: null,
        };
    },
    data: function () {
        return {
            visible: false,

            loadTriggered: false,

            rowIndex: 0,
            rowSplitCount: 1,

            loading: true,
            firstLoaded: false,

            loadDisplay: false,

            displayAddElement: false,

            fastTransition: false,

            contextMenuShown: false,
            contextMenuSelectedIndex: -1,
            contextMenuX: 0,
            contextMenuY: 0,

            movePositionModalDisplay: false,
            movePositionModalSelectedIndex: -1,

            elements: [] as HomePageElement[],

            loadingFiller: Array(this.pageSize)
                .fill(0)
                .map((_v, i) => i),

            dragging: false,
            draggingElement: null as HomePageElement,
            draggingData: {
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

            mouseX: 0,
            mouseY: 0,
        };
    },
    watch: {
        pageSize: function () {
            this.loadingFiller = Array(this.pageSize)
                .fill(0)
                .map((_v, i) => i);
            this.checkLoad(true);
        },
        loadTick: function () {
            this.checkLoad(true);
        },
        rowSize: function () {
            this.rowIndex = 0;
            this.fastTransition = true;
            nextTick(() => {
                this.fastTransition = false;
            });
            this.updateRowSplits();
            this.updateCurrentMedia();
        },
        currentMedia: function () {
            this.updateCurrentMedia();
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_HOME_SCROLL_CHANGED, () => {
            this.contextMenuShown = false;
            this.checkLoad();
        });
        this.checkLoad(true);
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);

        if (this.dragCheckInterval) {
            clearInterval(this.dragCheckInterval);
            this.dragCheckInterval = null;
        }

        if (this.mouseMoveHandler) {
            document.removeEventListener("mousemove", this.mouseMoveHandler);
            this.mouseMoveHandler = null;
        }

        if (this.mouseUpHandler) {
            document.removeEventListener("mouseup", this.mouseUpHandler);
            this.mouseUpHandler = null;
        }
    },
    methods: {
        unload: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);
            this.loadTriggered = false;
            this.loading = true;
            this.firstLoaded = false;
            this.elements = [];
        },

        checkLoad: function (forced?: boolean) {
            const container = this.$el.parentElement?.parentElement?.parentElement as HTMLElement;

            if (!container) {
                return;
            }

            const containerBounds = container.getBoundingClientRect();
            const rowBounds = (this.$el as HTMLElement).getBoundingClientRect();

            const graceSize = rowBounds.height / 2;

            const visibleTop = rowBounds.top >= containerBounds.top - graceSize && rowBounds.top <= containerBounds.bottom + graceSize;
            const visibleBottom =
                rowBounds.bottom >= containerBounds.top - graceSize && rowBounds.bottom <= containerBounds.bottom + graceSize;

            const visible = visibleTop || visibleBottom;

            this.visible = visible;

            if (this.loadTriggered) {
                if (forced) {
                    this.load();
                } else {
                    const outOfReach =
                        !visible &&
                        Math.abs(containerBounds.top + containerBounds.height / 2 - (rowBounds.top + rowBounds.height / 2)) >
                            containerBounds.height * 2;

                    if (outOfReach) {
                        this.unload();
                    }
                }
                return;
            }

            if (visible) {
                this.loadTriggered = true;
                this.load();
            }
        },

        load: function () {
            this.loading = true;

            if (this.moving && this.movingInitialElements) {
                this.onElementsLoaded(this.movingInitialElements);

                if (this.movingInitialScroll) {
                    this.applyInitialMovingScroll();
                }
                return;
            }

            switch (this.group.type) {
                case HomePageGroupTypes.Custom:
                    this.loadCustomElements();
                    break;
                case HomePageGroupTypes.RecentMedia:
                    this.loadRecentMedia();
                    break;
                case HomePageGroupTypes.RecentAlbums:
                    this.loadRecentAlbums();
                    break;
                default:
                    this.onElementsLoaded([]);
            }
        },

        onElementsLoaded(elements: HomePageElement[]) {
            this.elements = elements;
            this.firstLoaded = true;
            this.loading = false;
            this.loadDisplay = false;

            this.rowIndex = 0;

            this.updateRowSplits();
            this.updateCurrentMedia();

            if (this.isCurrentGroup) {
                this.$emit("loaded-current");
            }
        },

        updateRowSplits() {
            this.rowSplitCount = Math.ceil(this.elements.length / (this.rowSize || 1)) || 1;

            this.rowIndex = Math.min(this.rowIndex, this.rowSplitCount - 1);

            if (!this.editing) {
                nextTick(() => {
                    const scrollContainer = this.$el.querySelector(".home-page-row-content") as HTMLElement;
                    if (scrollContainer) {
                        scrollContainer.scrollLeft = 0;
                    }
                });
            }
        },

        onContentScroll: function () {
            if (this.contextMenuShown) {
                this.contextMenuShown = false;
            }
            if (!this.editing) {
                const scrollContainer = this.$el.querySelector(".home-page-row-content") as HTMLElement;
                if (scrollContainer) {
                    scrollContainer.scrollLeft = 0;
                }
            }
        },

        updateCurrentMedia() {
            if (!this.isCurrentGroup || this.currentMedia === -1) {
                return;
            }

            let currentMediaPos = -1;

            for (let i = 0; i < this.elements.length; i++) {
                if (this.elements[i].media && this.elements[i].media.id === this.currentMedia) {
                    currentMediaPos = i;
                    break;
                }
            }

            let firstElement = -1;
            let lastElement = -1;

            let prevElement = -1;
            let nextElement = -1;

            if (currentMediaPos >= 0) {
                this.rowIndex = Math.floor(currentMediaPos / (this.rowSize || 1));

                for (let i = 0; i < this.elements.length; i++) {
                    if (firstElement === -1 && this.elements[i].media) {
                        firstElement = this.elements[i].media.id;
                    }

                    if (this.elements[i].media) {
                        lastElement = this.elements[i].media.id;
                    }

                    if (i < currentMediaPos && this.elements[i].media) {
                        prevElement = this.elements[i].media.id;
                    } else if (nextElement === -1 && i > currentMediaPos && this.elements[i].media) {
                        nextElement = this.elements[i].media.id;
                    }
                }
            }

            this.$emit("updated-prev-next", prevElement, nextElement, firstElement, lastElement);
        },

        loadCustomElements: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            setNamedTimeout(this.loadRequestId, 330, () => {
                this.loadDisplay = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiHomeGetGroupElements(this.group.id))
                .onSuccess((result) => {
                    TagsController.OnMediaListReceived(result.filter((r) => !!r.media).map((r) => r.media));
                    clearNamedTimeout(this.loadRequestId);

                    this.onElementsLoaded(result);
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        notFound: () => {
                            this.$emit("must-reload");
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

        loadRecentMedia: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            setNamedTimeout(this.loadRequestId, 330, () => {
                this.loadDisplay = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiSearch("", "desc", 0, this.pageSize))
                .onSuccess((result) => {
                    TagsController.OnMediaListReceived(result.page_items);
                    clearNamedTimeout(this.loadRequestId);

                    this.onElementsLoaded(
                        result.page_items.map((r) => {
                            return {
                                media: r,
                            };
                        }),
                    );
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

        loadRecentAlbums: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            setNamedTimeout(this.loadRequestId, 330, () => {
                this.loadDisplay = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiAlbumsGetAlbums())
                .onSuccess((result) => {
                    clearNamedTimeout(this.loadRequestId);

                    this.onElementsLoaded(
                        result
                            .sort((a, b) => {
                                if (a.lm > b.lm) {
                                    return -1;
                                } else if (b.lm > a.lm) {
                                    return 1;
                                } else if (a.id < b.id) {
                                    return 1;
                                } else {
                                    return -1;
                                }
                            })
                            .slice(0, this.pageSize || this.rowSize)
                            .map((a) => {
                                return { album: a };
                            }),
                    );
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

        getGroupName(group: HomePageGroup): string {
            if (group.name) {
                return group.name;
            }

            return this.getDefaultGroupName(group.type, this.$t);
        },

        getDefaultGroupName: getDefaultGroupName,

        renameRow: function () {
            this.$emit("request-rename", this.group);
        },

        moveRow: function () {
            this.$emit("request-move", this.group);
        },

        deleteRow: function () {
            this.$emit("request-delete", this.group);
        },

        onDrag: function (event: DragEvent) {
            event.preventDefault();

            if (!this.editing) {
                return;
            }

            if (isTouchDevice()) {
                return;
            }

            const startX = event.pageX;
            const startY = event.pageY;

            const bounds = (this.$el as HTMLElement).getBoundingClientRect();

            const scrollContainer = this.$el.querySelector(".home-page-row-content") as HTMLElement;
            const initialScroll = scrollContainer ? scrollContainer.scrollLeft || 0 : 0;

            const data: HomePageGroupStartMovingData = {
                startX,
                startY,
                offsetX: startX - bounds.left,
                offsetY: startY - bounds.top,
                width: bounds.width,
                height: bounds.height,
                initialElements: this.loading ? null : this.elements.slice(),
                initialScroll,
            };

            this.$emit("start-moving", this.group, data);
        },

        goToElement: function (element: HomePageElement, e?: Event) {
            if (e) {
                e.preventDefault();
            }

            if (this.editing) {
                return;
            }

            if (element.media) {
                AppStatus.ClickOnMedia(element.media.id, true, this.group.id);
            } else if (element.album) {
                AppStatus.ClickOnAlbum(element.album.id);
            }
        },

        getElementURL: function (element: HomePageElement): string {
            if (element.media) {
                return (
                    window.location.protocol +
                    "//" +
                    window.location.host +
                    window.location.pathname +
                    generateURIQuery({
                        media: element.media.id + "",
                    })
                );
            } else if (element.album) {
                return (
                    window.location.protocol +
                    "//" +
                    window.location.host +
                    window.location.pathname +
                    generateURIQuery({
                        album: element.album.id + "",
                    })
                );
            } else {
                return "#";
            }
        },

        renderHintTitle(item: HomePageElement, tagVersion: number): string {
            if (item.media) {
                const parts = [item.media.title || this.$t("Untitled")];

                if (item.media.tags.length > 0) {
                    const tagNames = [];

                    for (const tag of item.media.tags) {
                        tagNames.push(TagsController.GetTagName(tag, tagVersion));
                    }

                    parts.push(this.$t("Tags") + ": " + tagNames.join(", "));
                }

                return parts.join("\n");
            } else if (item.album) {
                return (
                    (item.album.name || this.$t("Untitled album")) +
                    (item.album.lm ? "\n" + this.$t("Last modified") + ": " + renderDateAndTime(item.album.lm, this.$locale.value) : "")
                );
            } else {
                return "";
            }
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
        },

        getTitle: function (element: HomePageElement): string {
            if (element.media) {
                return element.media.title || this.$t("Untitled");
            } else if (element.album) {
                return element.album.name || this.$t("Untitled");
            } else {
                return "-";
            }
        },

        isCurrent(item: HomePageElement, currentMedia: number, isCurrentGroup: boolean): boolean {
            return isCurrentGroup && item.media && item.media.id === currentMedia;
        },

        addElements: function () {
            this.displayAddElement = true;
        },

        applyInitialMovingScroll: function () {
            nextTick(() => {
                const scrollContainer = this.$el.querySelector(".home-page-row-content") as HTMLElement;
                if (scrollContainer) {
                    scrollContainer.scrollLeft = this.movingInitialScroll || 0;
                }
            });
        },

        goLeft: function () {
            this.rowIndex = Math.max(0, this.rowIndex - 1);
        },

        goRight: function () {
            this.rowIndex = Math.min(this.rowIndex + 1, this.rowSplitCount - 1);
        },

        focusElementIndex: function (i: number) {
            this.rowIndex = Math.floor(i / (this.rowSize || 1));
        },

        moveFocusToFirstRowElement: function () {
            const firstRowElementIndex = this.rowIndex * (this.rowSize || 1);
            const firstRowElement = this.$el.querySelector(".home-page-row-element-" + firstRowElementIndex);
            if (firstRowElement) {
                firstRowElement.focus();
            }
        },

        onKeyDown: function (event: KeyboardEvent) {
            if (event.key === "ArrowLeft" || event.key === "PageUp") {
                event.preventDefault();
                this.goLeft();
                this.moveFocusToFirstRowElement();
            } else if (event.key === "ArrowRight" || event.key === "PageDown") {
                event.preventDefault();
                this.goRight();
                this.moveFocusToFirstRowElement();
            } else if (event.key === "Home") {
                event.preventDefault();
                this.rowIndex = 0;
                this.moveFocusToFirstRowElement();
            } else if (event.key === "End") {
                event.preventDefault();
                this.rowIndex = this.rowSplitCount - 1;
                const firstRowElement = this.$el.querySelector(".home-page-row-element-" + (this.elements.length - 1));
                if (firstRowElement) {
                    firstRowElement.focus();
                }
            }
        },

        showContextMenu: function (i: number, event: MouseEvent) {
            if (!this.editing) {
                return;
            }

            event.preventDefault();

            this.contextMenuSelectedIndex = i;
            this.contextMenuX = event.pageX;
            this.contextMenuY = event.pageY;
            this.contextMenuShown = true;
        },

        onElementRemove: function (i: number) {
            const element = this.elements[i];

            if (!element) {
                return;
            }

            this.elements.splice(i, 1);

            this.doSilentDelete(element);
        },

        doSilentDelete: function (element: HomePageElement) {
            doHomePageSilentSaveAction((callback) => {
                makeApiRequest(apiHomeGroupDeleteElement(this.group.id, getHomePageElementReference(element)))
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
                            notCustomGroup: () => {
                                this.$emit("must-reload");
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

        onChangePositionRequest: function (i: number) {
            this.movePositionModalSelectedIndex = i;
            this.movePositionModalDisplay = true;
        },

        onMoveElement: function (oldPos: number, newPos: number) {
            const element = this.elements[oldPos];

            if (!element) {
                return;
            }

            this.elements.splice(newPos, 0, this.elements.splice(oldPos, 1)[0]);

            this.doSilentMove(element, newPos);
        },

        doSilentMove: function (element: HomePageElement, position: number) {
            doHomePageSilentSaveAction((callback) => {
                makeApiRequest(apiHomeGroupMoveElement(this.group.id, getHomePageElementReference(element), position))
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
                            notCustomGroup: () => {
                                this.$emit("must-reload");
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

        onDragStart: function (i: number, event: DragEvent) {
            if (!this.editing) {
                return;
            }

            event.preventDefault();
            event.stopPropagation();

            const element = this.elements[i];

            if (!element) {
                return;
            }

            this.mouseX = event.pageX;
            this.mouseY = event.pageY;

            const elClickable = this.$el.querySelector(".home-page-row-element-" + i) as HTMLElement;

            if (!elClickable) {
                return;
            }

            const target = elClickable.parentElement as HTMLElement;

            if (!target) {
                return;
            }

            const targetBounds = target.getBoundingClientRect();

            const startX = event.pageX;
            const startY = event.pageY;

            const offsetX = startX - targetBounds.left;
            const offsetY = startY - targetBounds.top;

            this.dragging = true;
            this.draggingElement = element;

            this.draggingData.startX = startX;
            this.draggingData.startY = startY;

            this.draggingData.offsetX = offsetX;
            this.draggingData.offsetY = offsetY;

            this.draggingData.width = targetBounds.width;
            this.draggingData.height = targetBounds.height;

            this.draggingData.x = startX - offsetX;
            this.draggingData.y = startY - offsetY;

            this.draggingData.startPosition = i;

            if (this.dragCheckInterval) {
                clearInterval(this.dragCheckInterval);
                this.dragCheckInterval = null;
            }

            if (this.mouseMoveHandler) {
                document.removeEventListener("mousemove", this.mouseMoveHandler);
                this.mouseMoveHandler = null;
            }

            if (this.mouseUpHandler) {
                document.removeEventListener("mouseup", this.mouseUpHandler);
                this.mouseUpHandler = null;
            }

            this.mouseMoveHandler = this.onDocumentMouseMove.bind(this);
            document.addEventListener("mousemove", this.mouseMoveHandler);

            this.mouseUpHandler = this.onDocumentMouseUp.bind(this);
            document.addEventListener("mouseup", this.mouseUpHandler);

            this.dragCheckInterval = setInterval(this.onDragCheck.bind(this), 40);

            this.updateMovingOver();
        },

        updateMovingOver: function () {
            const container = this.$el.querySelector(".home-page-row-content") as HTMLElement;

            if (!container) {
                this.draggingData.movingOver = -1;
                return;
            }

            const containerBounds = container.getBoundingClientRect();

            const offsetLeft = containerBounds.left;
            const scrollLeft = container.scrollLeft || 0;

            const x = this.draggingData.x + Math.round(this.draggingData.width / 2);
            const y = this.draggingData.y + Math.round(this.draggingData.height / 2);

            if (
                y + Math.round(this.draggingData.height / 2) < containerBounds.top ||
                y - Math.round(this.draggingData.height / 2) > containerBounds.top + containerBounds.height
            ) {
                this.draggingData.movingOver = -1;
                return;
            }

            const width = this.draggingData.width;
            const expectedIndex = Math.round((x - offsetLeft + scrollLeft) / width);

            this.draggingData.movingOver = Math.min(this.elements.length + 1, Math.max(1, 1 + expectedIndex));
        },

        onDragCheck: function () {
            const con = this.$el.querySelector(".home-page-row-content");

            if (!con) {
                return;
            }

            const conBounds = con.getBoundingClientRect();

            if (this.mouseX >= conBounds.left - this.draggingData.width) {
                // Auto scroll

                const relTop = (this.mouseX - conBounds.left) / (conBounds.width || 1);
                const scrollStep = Math.ceil(conBounds.width / 25);

                if (relTop <= 0.1) {
                    con.scrollLeft = Math.max(0, con.scrollLeft - scrollStep);
                } else if (relTop >= 0.9) {
                    con.scrollLeft = Math.min(con.scrollWidth - conBounds.width, con.scrollLeft + scrollStep);
                }
            }

            this.updateMovingOver();
        },

        onDocumentMouseMove: function (event: MouseEvent) {
            if (!this.dragging) {
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

            this.draggingData.x = event.pageX - this.draggingData.offsetX;
            this.draggingData.y = event.pageY - this.draggingData.offsetY;
        },

        onDocumentMouseUp: function (event: MouseEvent) {
            if (!this.dragging) {
                return;
            }

            if (this.dragCheckInterval) {
                clearInterval(this.dragCheckInterval);
                this.dragCheckInterval = null;
            }

            if (this.mouseMoveHandler) {
                document.removeEventListener("mousemove", this.mouseMoveHandler);
                this.mouseMoveHandler = null;
            }

            if (this.mouseUpHandler) {
                document.removeEventListener("mouseup", this.mouseUpHandler);
                this.mouseUpHandler = null;
            }

            event.stopPropagation();

            this.dragging = false;

            if (!this.draggingElement) {
                return;
            }

            const draggingElement = this.draggingElement;

            let position =
                this.draggingData.movingOver > this.draggingData.startPosition + 1
                    ? this.draggingData.movingOver - 1
                    : this.draggingData.movingOver;

            position = Math.max(0, Math.min(this.elements.length, position - 1));

            const startPosition = this.draggingData.startPosition;

            if (startPosition === -1 || position === startPosition) {
                return;
            }

            this.doSilentMove(draggingElement, position);
            this.elements.splice(position, 0, this.elements.splice(startPosition, 1)[0]);
        },
    },
});
</script>
