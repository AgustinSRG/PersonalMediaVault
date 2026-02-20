<template>
    <div
        ref="container"
        :class="{ 'page-inner': !inModal, 'scrollbar-stable': !inModal, 'page-in-modal': !!inModal }"
        tabindex="-1"
        @scroll.passive="onPageScroll"
    >
        <form class="media-search-form" @submit="onSubmit">
            <div class="form-group search-option-container">
                <button type="button" class="page-option" :class="{ current: mode === 'basic' }" @click="setMode('basic')">
                    {{ $t("Basic search") }}
                </button>

                <button type="button" class="page-option" :class="{ current: mode === 'adv' }" @click="setMode('adv')">
                    {{ $t("Advanced search") }}
                </button>

                <button
                    v-if="semanticSearchAvailable"
                    type="button"
                    class="page-option"
                    :class="{ current: mode === 'semantic' }"
                    @click="setMode('semantic')"
                >
                    {{ $t("Semantic search") }}
                </button>

                <button
                    v-if="semanticSearchAvailable"
                    type="button"
                    class="page-option"
                    :class="{ current: mode === 'image' }"
                    @click="setMode('image')"
                >
                    {{ $t("Search by image") }}
                </button>
            </div>

            <div v-if="mode === 'basic' || mode === 'adv'">
                <div class="form-group">
                    <label>{{ $t("Title must contain") }}:</label>
                    <input
                        v-model="textSearch"
                        type="text"
                        name="title-search"
                        autocomplete="off"
                        maxlength="255"
                        class="form-control form-control-full-width"
                        :class="{ 'auto-focus': !!textSearch }"
                        @input="markDirty"
                    />
                </div>

                <div v-if="mode === 'adv'" class="form-group">
                    <label>{{ $t("Tags") }}:</label>
                    <select v-model="tagMode" class="form-control form-select form-control-full-width" @change="markDirty">
                        <option :value="'all'">
                            {{ $t("Media must contain ALL of the selected tags") }}
                        </option>
                        <option :value="'any'">
                            {{ $t("Media must contain ANY of the selected tags") }}
                        </option>
                        <option :value="'none'">
                            {{ $t("Media must contain NONE of the selected tags") }}
                        </option>
                        <option :value="'untagged'">
                            {{ $t("Media must be untagged") }}
                        </option>
                    </select>
                </div>

                <TagIdList
                    v-if="tagMode !== 'untagged'"
                    v-model:tags="tags"
                    :label="$t('Tags')"
                    @changed="markDirty"
                    @tab-focus-skip="skipTagSuggestions"
                ></TagIdList>

                <div v-if="mode === 'adv'">
                    <div class="form-group">
                        <label>{{ $t("Media type") }}:</label>
                        <select
                            v-model="type"
                            class="form-control form-select form-control-full-width tags-focus-skip"
                            @change="markDirty"
                            @keydown="onTagsSkipKeyDown"
                        >
                            <option :value="0">{{ $t("Any media") }}</option>
                            <option :value="1">{{ $t("Images") }}</option>
                            <option :value="2">{{ $t("Videos") }}</option>
                            <option :value="3">{{ $t("Audios") }}</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Album") }}:</label>
                        <AlbumSelect v-model:album="albumSearch" @update:album="markDirty"></AlbumSelect>
                    </div>

                    <div class="form-group">
                        <label>{{ $t("Order") }}:</label>
                        <select v-model="order" class="form-control form-select form-control-full-width" @change="markDirty">
                            <option :value="''">{{ $t("Default order") }}</option>
                            <option :value="'desc'">{{ $t("Show most recent") }}</option>
                            <option :value="'asc'">{{ $t("Show oldest") }}</option>
                        </select>
                    </div>
                </div>
            </div>

            <div v-else-if="mode === 'semantic'">
                <div class="form-group">
                    <label>{{ $t("Semantic search mode") }}:</label>
                    <select v-model="onlyImages" class="form-control form-select form-control-full-width" @change="onOnlyImagesChanged">
                        <option :value="false">{{ $t("Search by title and image content") }}</option>
                        <option :value="true">{{ $t("Search only image content (ignore titles)") }}</option>
                    </select>
                </div>

                <div class="form-group">
                    <label>{{ $t("Type what you are looking for") }}:</label>
                    <input
                        v-model="textSearch"
                        type="text"
                        name="semantic-search"
                        autocomplete="off"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                        @input="markDirty"
                    />
                </div>
            </div>

            <ImageSelectBox
                v-else-if="mode === 'image'"
                v-model:image-file="imageFile"
                :loading="loading"
                :vector-loaded="vectorLoaded"
                :image-error="imageError"
                @start-search="startSearch"
            ></ImageSelectBox>
        </form>

        <div class="search-results tags-focus-skip" tabindex="-1" @keydown="onTagsSkipKeyDown">
            <div v-if="!loading && started && fullListLength === 0" class="search-results-msg-display">
                <div class="search-results-msg-icon"><i class="fas fa-search"></i></div>
                <div class="search-results-msg-text">
                    {{ $t("Could not find any result") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="onSubmit()">
                        <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                    </button>
                </div>
            </div>

            <div v-if="pageItems.length > 0" class="search-results-final-display">
                <div v-for="i in rowPaddingPreserveCols" :key="'pad-prev-' + i" class="search-result-item">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>

                <MediaItem
                    v-for="(item, i) in pageItems"
                    :key="i"
                    :item="item"
                    :current="currentMedia == item.id"
                    :last-current="lastCurrentMedia === item.id"
                    :display-titles="displayTitles"
                    @go="goToMedia(item)"
                ></MediaItem>

                <div v-for="i in lastRowPadding" :key="'pad-last-' + i" class="search-result-item"></div>
            </div>

            <div v-if="!finished" class="search-continue-mark">
                <button type="button" class="btn btn-primary btn-mr" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{ Math.round(progress) + "%" }})
                </button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { MediaType } from "@/api/models";
import { type MediaListItem } from "@/api/models";
import { AlbumsController } from "@/control/albums";
import {
    emitAppEvent,
    EVENT_NAME_ADVANCED_SEARCH_GO_TOP,
    EVENT_NAME_ADVANCED_SEARCH_SCROLL,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { TagsController } from "@/control/tags";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import type { PropType } from "vue";
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, useTemplateRef, watch } from "vue";
import { BigListScroller } from "@/utils/big-list-scroller";
import { PagesController } from "@/control/pages";
import { apiAlbumsGetAlbum } from "@/api/api-albums";
import { apiAdvancedSearch } from "@/api/api-search";
import { isTouchDevice } from "@/utils/touch";
import AlbumSelect from "../utils/AlbumSelect.vue";
import MediaItem from "../utils/MediaItem.vue";
import TagIdList from "../utils/TagIdList.vue";
import type { SearchMode } from "@/control/app-preferences";
import {
    getPreferredSearchMode,
    getSemanticSearchOnlyImages,
    setPreferredSearchMode,
    setSemanticSearchOnlyImages,
} from "@/control/app-preferences";
import { apiSemanticSearch, apiSemanticSearchEncodeImage, apiSemanticSearchEncodeText } from "@/api/api-semantic-search";
import { useRequestId } from "@/composables/use-request-id";
import { useI18n } from "@/composables/use-i18n";
import { usePageLastRowPadding } from "@/composables/use-page-last-row-padding";
import { useInterval } from "@/composables/use-interval";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";

const ImageSelectBox = defineAsyncComponent({
    loader: () => import("./common/ImageSelectBox.vue"),
});

// Ref to the container element
const container = useTemplateRef("container");

// Ref to the top container
const topContainer = computed<HTMLElement>(() => {
    if (props.inModal) {
        if (
            container.value &&
            container.value.parentElement &&
            container.value.parentElement.parentElement &&
            container.value.parentElement.parentElement.parentElement
        ) {
            return container.value.parentElement.parentElement.parentElement;
        } else {
            return null;
        }
    } else {
        return container.value;
    }
});

// Translation
const { $t } = useI18n();

const props = defineProps({
    /**
     * True if the page is being displayed in a modal
     */
    inModal: Boolean,

    /**
     * Optional set of elements to remove from the list
     */
    removeMediaFromList: Object as PropType<Set<number>>,

    /**
     * Optional album ID to filter out
     */
    noAlbum: Number,

    /**
     * Size of the page
     */
    pageSize: {
        type: Number,
        required: true,
    },

    /**
     * True to display titles
     */
    displayTitles: Boolean,

    /**
     * True if using miniature mode
     */
    min: Boolean,

    /**
     * Preferred row size
     */
    rowSize: {
        type: Number,
        required: true,
    },

    /**
     * Preferred row size for miniature mode
     */
    rowSizeMin: {
        type: Number,
        required: true,
    },

    /**
     * Min size for items
     */
    minItemsSize: {
        type: Number,
        required: true,
    },

    /**
     * Max size for items
     */
    maxItemsSize: {
        type: Number,
        required: true,
    },
});

// Emits
const emit = defineEmits<{
    /**
     * User selected a media element
     */
    (e: "select-media", m: MediaListItem, callback: () => void): void;
}>();

// Page scroll
const pageScroll = defineModel<number>("pageScroll");

onMounted(() => {
    pageScroll.value = 0;
});

// Map (media ID -> index in the list)
const mediaIndexMap = new Map<number, number>();

// Initial window size for scroller
const INITIAL_WINDOW_SIZE = 50;

// Page items
const pageItems = ref<MediaListItem[]>([]);

// Position of the scroll window
const windowPosition = ref(0);

// List scroller
const listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
    get: () => {
        return pageItems.value;
    },
    set: (l) => {
        pageItems.value = l;
    },
    onChange: () => {
        windowPosition.value = listScroller.windowPosition;
    },
});

/**
 * Event handler for scroll on the page
 * @param e
 */
const onPageScroll = (e: Event) => {
    if (props.inModal) {
        return;
    }

    pageScroll.value = (e.target as HTMLElement).scrollTop || 0;

    listScroller.checkElementScroll(e.target as HTMLElement);
};

onApplicationEvent(EVENT_NAME_ADVANCED_SEARCH_SCROLL, (e: Event) => {
    listScroller.checkElementScroll(e.target as HTMLElement);
});

// Loading status
const loading = ref(false);

// Search parameters
const searchParams = ref(AppStatus.SearchParams);

/**
 * Page search parameters
 */
type PageSearchParameters = {
    textSearch: string;
    tags: number[];
};

/**
 * Pages the search parameters
 * @returns The parsed search parameters
 */
const parsePageSearchParameters = (): PageSearchParameters => {
    if (props.inModal || !AppStatus.SearchParams) {
        return {
            textSearch: "",
            tags: [],
        };
    }

    const parts = AppStatus.SearchParams.split("~");

    const tags = parts[0]
        .split("-")
        .filter((p) => !!p)
        .map((p) => parseInt(p, 10))
        .filter((p) => !isNaN(p) && p >= 0);

    const tagSet = new Set<number>();
    const filteredTags: number[] = [];

    for (const tag of tags) {
        if (tagSet.has(tag)) {
            continue;
        }

        tagSet.add(tag);
        filteredTags.push(tag);
    }

    const textSearch = parts.slice(1).join("~").substring(0, 128).trim();

    return {
        tags: filteredTags,
        textSearch,
    };
};

/**
 * Updates the search parameters
 */
const updateSearchParams = () => {
    if (props.inModal) {
        return;
    }

    let newSearchParams = "";

    if (tags.value.length > 0 || textSearch.value.length > 0) {
        newSearchParams = tags.value.join("-") + (textSearch.value ? "~" + textSearch.value : "");
    }

    searchParams.value = newSearchParams;

    if (AppStatus.SearchParams !== newSearchParams) {
        AppStatus.ChangeSearchParams(newSearchParams);
    }
};

// Page order
const order = ref("" as "" | "desc" | "asc");

// Text search
const textSearch = ref("");

// Media type filter
const type = ref<MediaType>(0);

// Current media
const currentMedia = ref(AppStatus.CurrentMedia);

// Last current media
const lastCurrentMedia = ref(AppStatus.CurrentMedia);

// Current page
const page = ref(0);

// Total number of pages
const totalPages = ref(0);

// Search progress
const progress = ref(0);

// Reference to continue loading more results
const continueRef = ref<number | null>(null);

// Length of the full list
const fullListLength = ref(0);

// True if started searching
const started = ref(false);

// True if finished searching
const finished = ref(true);

// Search mode
const mode = ref(getPreferredSearchMode(AuthController.SemanticSearchAvailable));

// Tags to filter by
const tags = ref<number[]>([]);

// Mode for the tags filter
const tagMode = ref("all");

/**
 * Gets the tags list for the API
 * @returns The tags list for the API
 */
const getTagList = (): string[] => {
    if (tagMode.value === "untagged") {
        return [];
    }
    if (tagMode.value === "any" && tags.value.length > 16) {
        return [];
    }
    return tags.value
        .map((tag) => {
            return TagsController.GetTagName(tag, TagsController.TagsVersion);
        })
        .slice(0, 16);
};

/**
 * Gets the tag mode for the API
 * @returns The API tag mode
 */
const getTagMode = (): "allof" | "anyof" | "noneof" => {
    switch (tagMode.value) {
        case "any":
            if (tags.value.length > 16) {
                return "allof";
            }
            return "anyof";
        case "none":
            return "noneof";
        default:
            return "allof";
    }
};

// Album to search into
const albumSearch = ref(-1);

// True if semantic search is available
const semanticSearchAvailable = ref(AuthController.SemanticSearchAvailable);

// Vector (for semantic search)
const vector = ref([]);

// True if the vector has been loaded
const vectorLoaded = ref(false);

// True to use only image vectors
const onlyImages = ref(getSemanticSearchOnlyImages());

// The current uploaded image file to search by image
const imageFile = shallowRef<File | null>(null);

// Error of the image (for "Search by Image")
const imageError = ref("");

// Load request ID
const loadRequestId = useRequestId();

/**
 * Resets the search status
 */
const resetSearch = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);
    listScroller.reset();
    fullListLength.value = 0;
    mediaIndexMap.clear();
    page.value = 0;
    totalPages.value = 0;
    progress.value = 0;
    continueRef.value = null;
    loading.value = false;
    finished.value = true;
    started.value = false;
    startSearch();
};

onApplicationEvent(EVENT_NAME_MEDIA_DELETE, resetSearch);
onApplicationEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, resetSearch);
watch(() => props.pageSize, resetSearch);

// ID for the dirty timeout
const dirtyTimeoutId = useRequestId();

// Delay to start loading after changes are made (milliseconds)
const DIRTY_LOAD_DELAY = 330;

/**
 * Indicates changes were made on the search parameters
 */
const markDirty = () => {
    setNamedTimeout(dirtyTimeoutId, DIRTY_LOAD_DELAY, startSearch);
};

/**
 * Starts the search process
 */
const startSearch = () => {
    clearNamedTimeout(dirtyTimeoutId);

    updateSearchParams();

    loading.value = true;
    listScroller.reset();
    fullListLength.value = 0;
    mediaIndexMap.clear();
    page.value = 0;
    continueRef.value = null;
    totalPages.value = 0;
    progress.value = 0;
    started.value = true;
    finished.value = false;
    vectorLoaded.value = false;
    imageError.value = "";

    continueSearch();
};

onMounted(() => {
    // Parse initial search parameters
    const parsedParams = parsePageSearchParameters();
    textSearch.value = parsedParams.textSearch;
    if (textSearch.value != "" && mode.value === "image") {
        mode.value = "semantic";
    }
    tags.value = parsedParams.tags;

    updateSearchParams();

    // Start searching
    startSearch();
});

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    const changed = currentMedia.value !== AppStatus.CurrentMedia;
    currentMedia.value = AppStatus.CurrentMedia;

    if (AppStatus.CurrentMedia >= 0) {
        lastCurrentMedia.value = AppStatus.CurrentMedia;
    }

    if (!props.inModal) {
        if (changed) {
            nextTick(() => {
                if (!checkContainerHeight()) {
                    scrollToCurrentMedia();
                }
            });
        }
        onCurrentMediaChanged();
    }

    if (!props.inModal && searchParams.value !== AppStatus.SearchParams) {
        const parsedParams = parsePageSearchParameters();
        textSearch.value = parsedParams.textSearch;
        tags.value = parsedParams.tags;
        updateSearchParams();
        startSearch();
        autoFocus();
    }
});

onApplicationEvent(EVENT_NAME_AUTH_CHANGED, () => {
    semanticSearchAvailable.value = AuthController.SemanticSearchAvailable;

    if ((mode.value === "semantic" || mode.value === "image") && !semanticSearchAvailable.value) {
        setMode("basic");
    }

    startSearch();
});

/**
 * Cancels the search process
 */
const cancel = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);
    loading.value = false;
    finished.value = true;
};

/**
 * Continues the search process
 */
const continueSearch = () => {
    if (mode.value === "semantic" && textSearch.value) {
        loadSemantic();
    } else if (mode.value === "image" && imageFile.value) {
        loadSemanticImage();
    } else if (albumSearch.value >= 0 && mode.value === "adv") {
        loadAlbumSearch();
    } else {
        load();
    }
};

// Delay to retry loading (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the media elements (default)
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (finished.value) {
        return;
    }

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    const pageSize = props.pageSize;
    const orderVal = mode.value === "adv" ? order.value || "desc" : "desc";

    makeNamedApiRequest(loadRequestId, apiAdvancedSearch(getTagMode(), getTagList(), orderVal, continueRef.value, pageSize))
        .onSuccess((result) => {
            const completePageList = listScroller.list;

            filterElements(result.items);

            page.value = result.scanned;
            totalPages.value = result.total_count;
            progress.value = (Math.max(0, result.scanned) / Math.max(1, result.total_count)) * 100;
            continueRef.value = result["continue"];

            if (completePageList.length >= pageSize) {
                // Done for now
                loading.value = false;

                if (page.value >= totalPages.value) {
                    finished.value = true;
                }

                if (!props.inModal) {
                    onCurrentMediaChanged();
                }
            } else if (result.scanned < result.total_count) {
                // Maybe there are more items
                load();
            } else {
                loading.value = false;
                finished.value = true;
                if (!props.inModal) {
                    onCurrentMediaChanged();
                }
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
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

/**
 * Loads the elements of the album
 */
const loadAlbumSearch = () => {
    abortNamedApiRequest(loadRequestId);

    makeNamedApiRequest(loadRequestId, apiAlbumsGetAlbum(albumSearch.value))
        .onSuccess((result) => {
            if (order.value === "asc") {
                filterElements(
                    result.list.sort((a, b) => {
                        if (a.id < b.id) {
                            return -1;
                        } else {
                            return 1;
                        }
                    }),
                );
            } else if (order.value === "desc") {
                filterElements(
                    result.list.sort((a, b) => {
                        if (a.id > b.id) {
                            return -1;
                        } else {
                            return 1;
                        }
                    }),
                );
            } else {
                filterElements(result.list);
            }

            page.value = 1;
            totalPages.value = 1;
            progress.value = 100;
            loading.value = false;
            finished.value = true;
            if (!props.inModal) {
                onCurrentMediaChanged();
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    cancel();
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    filterElements([]);
                    page.value = 1;
                    totalPages.value = 1;
                    progress.value = 100;
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                temporalError: () => {
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadAlbumSearch);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            cancel();
        });
};

/**
 * Performs a semantic search query
 */
const loadSemantic = () => {
    if (!vectorLoaded.value) {
        loadSemanticVector();
        return;
    }

    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (finished.value) {
        return;
    }

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    const pageSize = props.pageSize;

    makeNamedApiRequest(
        loadRequestId,
        apiSemanticSearch({
            vector: vector.value,
            limit: pageSize,
            continuationToken: continueRef.value,
            vectorType: onlyImages.value ? "image" : "any",
        }),
    )
        .onSuccess((result) => {
            const completePageList = listScroller.list;

            filterElements(result.items);

            page.value = result.scanned;
            totalPages.value = result.total_count;
            progress.value = (Math.max(0, result.scanned) / Math.max(1, result.total_count)) * 100;
            continueRef.value = result["continue"];

            if (completePageList.length >= pageSize) {
                // Done for now
                loading.value = false;

                if (page.value >= totalPages.value) {
                    finished.value = true;
                }

                if (!props.inModal) {
                    onCurrentMediaChanged();
                }
            } else if (result.scanned < result.total_count) {
                // Maybe there are more items
                loadSemantic();
            } else {
                loading.value = false;
                finished.value = true;
                if (!props.inModal) {
                    onCurrentMediaChanged();
                }
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                invalidVectorSize: () => {
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemantic);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemantic);
        });
};

/**
 * Loads the semantic vector to perform the semantic search
 */
const loadSemanticVector = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (finished.value) {
        return;
    }

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiSemanticSearchEncodeText(textSearch.value))
        .onSuccess((result) => {
            vector.value = result.vector;
            vectorLoaded.value = true;
            loadSemantic();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                emptyText: () => {
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                notAvailable: () => {
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemanticVector);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemanticVector);
        });
};

/**
 * Performs a semantic search query using an image
 */
const loadSemanticImage = () => {
    if (!vectorLoaded.value) {
        loadSemanticImageVector();
        return;
    }

    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (finished.value) {
        return;
    }

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    const pageSize = props.pageSize;

    makeNamedApiRequest(
        loadRequestId,
        apiSemanticSearch({
            vector: vector.value,
            limit: pageSize,
            continuationToken: continueRef.value,
            vectorType: "image",
        }),
    )
        .onSuccess((result) => {
            const completePageList = listScroller.list;

            filterElements(result.items);

            page.value = result.scanned;
            totalPages.value = result.total_count;
            progress.value = (Math.max(0, result.scanned) / Math.max(1, result.total_count)) * 100;
            continueRef.value = result["continue"];

            if (completePageList.length >= pageSize) {
                // Done for now
                loading.value = false;

                if (page.value >= totalPages.value) {
                    finished.value = true;
                }

                if (!props.inModal) {
                    onCurrentMediaChanged();
                }
            } else if (result.scanned < result.total_count) {
                // Maybe there are more items
                loadSemanticImage();
            } else {
                loading.value = false;
                finished.value = true;
                if (!props.inModal) {
                    onCurrentMediaChanged();
                }
            }
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                invalidVectorSize: () => {
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemanticImage);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemanticImage);
        });
};

/**
 * Loads the vector from the selected image,
 * in order to search by image semantically.
 */
const loadSemanticImageVector = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    if (finished.value) {
        return;
    }

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiSemanticSearchEncodeImage(imageFile.value))
        .onSuccess((result) => {
            vector.value = result.vector;
            vectorLoaded.value = true;

            loadSemanticImage();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                invalidImage: () => {
                    imageError.value = $t("Invalid image file selected.");
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                imageTooLarge: () => {
                    imageError.value = $t("The image you selected is too large. Try using an smaller image.");
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                notAvailable: () => {
                    imageError.value = $t("Image encoding service not available. Try again later.");
                    loading.value = false;
                    finished.value = true;
                    if (!props.inModal) {
                        onCurrentMediaChanged();
                    }
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemanticImageVector);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, loadSemanticImageVector);
        });
};

/**
 * Filters found elements and adds them to the list if the match the filter
 * @param results The list of results received from the API
 */
const filterElements = (results: MediaListItem[]) => {
    TagsController.OnMediaListReceived(results);

    const filterText = mode.value === "basic" || mode.value === "adv" ? normalizeString(textSearch.value).trim().toLowerCase() : "";
    const filterTextWords = filterToWords(filterText);
    const filterType = type.value;
    const filterTags = tags.value.slice();
    const filterTagMode = tagMode.value;

    let blacklist = new Set();

    if (props.noAlbum >= 0 && AlbumsController.CurrentAlbumData) {
        blacklist = new Set(
            AlbumsController.CurrentAlbumData.list.map((a) => {
                return a.id;
            }),
        );
    } else if (props.removeMediaFromList) {
        blacklist = props.removeMediaFromList;
    }

    const resultsToAdd = [];

    for (const e of results) {
        if (blacklist.has(e.id)) {
            continue;
        }

        if (mediaIndexMap.has(e.id)) {
            continue;
        }

        if (filterText) {
            if (matchSearchFilter(e.title, filterText, filterTextWords) < 0) {
                continue;
            }
        }

        if (mode.value === "adv") {
            if (filterType) {
                if (e.type !== filterType) {
                    continue;
                }
            }
        }

        if (filterTagMode === "all") {
            if (filterTags.length > 0) {
                let passesTags = true;
                for (const tag of filterTags) {
                    if (!e.tags || !e.tags.includes(tag)) {
                        passesTags = false;
                        break;
                    }
                }

                if (!passesTags) {
                    continue;
                }
            }
        } else if (filterTagMode === "none") {
            if (filterTags.length > 0) {
                let passesTags = true;
                for (const tag of filterTags) {
                    if (e.tags && e.tags.includes(tag)) {
                        passesTags = false;
                        break;
                    }
                }

                if (!passesTags) {
                    continue;
                }
            }
        } else if (filterTagMode === "any") {
            if (filterTags.length > 0) {
                let passesTags = false;
                for (const tag of filterTags) {
                    if (!e.tags || e.tags.includes(tag)) {
                        passesTags = true;
                        break;
                    }
                }

                if (!passesTags) {
                    continue;
                }
            }
        } else if (filterTagMode === "untagged") {
            if (e.tags && e.tags.length > 0) {
                continue;
            }
        }

        mediaIndexMap.set(e.id, listScroller.list.length + resultsToAdd.length);

        resultsToAdd.push(e);
    }

    listScroller.addElements(resultsToAdd);
    fullListLength.value = listScroller.list.length;

    nextTick(() => {
        checkContainerHeight();
    });
};

// Max relative scroll before loading more elements
const MAX_RELATIVE_SCROLL_BEFORE_LOADING_MORE = 0.8;

/**
 * Checks the scroll to load more elements if necessary
 */
const checkContinueSearch = () => {
    if (finished.value || loading.value || fullListLength.value === 0) {
        return;
    }

    if (!listScroller.isAtTheEnd()) {
        return;
    }

    const con = topContainer.value;

    if (!con) {
        return;
    }

    const conBounds = con.getBoundingClientRect();

    const overflowLength = con.scrollHeight - conBounds.height;

    if (overflowLength < 1) {
        continueSearch();
        return;
    }

    const relScroll = con.scrollTop / overflowLength;

    if (relScroll < MAX_RELATIVE_SCROLL_BEFORE_LOADING_MORE) {
        return;
    }

    continueSearch();
};

// Interval to check for continuation for load
const continueCheckInterval = useInterval();

// Delay to check to continue searching
const CHECK_CONTINUE_SEARCH_DELAY = 500;

// Check to load more element periodically
continueCheckInterval.set(checkContinueSearch, CHECK_CONTINUE_SEARCH_DELAY);

/**
 * Checks the container height
 */
const checkContainerHeight = (): boolean => {
    const cont = topContainer.value;

    if (!cont) {
        return false;
    }

    listScroller.checkElementScroll(cont);

    const centerPosition = listScroller.getCenterPosition();

    const el = container.value?.querySelector(".search-result-item") as HTMLElement;

    if (!el) {
        return false;
    }

    const changed = listScroller.checkScrollContainerHeight(cont, el);

    if (changed) {
        if (!scrollToCurrentMedia()) {
            listScroller.moveWindowToElement(centerPosition);
        }
    }

    return changed;
};

// Interval to check the container height
const checkContainerTimer = useInterval();

// DElay to keep checking the container height
const CHECK_CONTAINER_HEIGHT_DELAY = 100;

onMounted(() => {
    checkContainerTimer.set(checkContainerHeight, CHECK_CONTAINER_HEIGHT_DELAY);
});

/**
 * Navigates to the top of the scroll
 */
const goTop = () => {
    if (!props.inModal) {
        listScroller.moveWindowToElement(0);
        nextTick(() => {
            if (container.value) {
                container.value.scrollTop = 0;
            }
        });
    } else {
        listScroller.moveWindowToElement(0);
    }
};

onApplicationEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP, goTop);

/**
 * Scrolls to the current media element
 * @returns true if there is a current media element, false otherwise
 */
const scrollToCurrentMedia = (): boolean => {
    if (!mediaIndexMap.has(currentMedia.value)) {
        return false;
    }
    const index = mediaIndexMap.get(currentMedia.value);

    if (index < listScroller.windowPosition || index >= listScroller.windowPosition + listScroller.windowSize) {
        listScroller.moveWindowToElement(mediaIndexMap.get(currentMedia.value));
    }

    nextTick(() => {
        const currentElem = container.value?.querySelector(".search-result-item.current");
        if (currentElem) {
            currentElem.scrollIntoView();
        }
    });

    return true;
};

/**
 * Scrolls to the last known current media element
 */
const scrollToLastCurrentMedia = () => {
    nextTick(() => {
        if (!mediaIndexMap.has(lastCurrentMedia.value)) {
            return;
        }

        const index = mediaIndexMap.get(lastCurrentMedia.value);

        if (index < listScroller.windowPosition || index >= listScroller.windowPosition + listScroller.windowSize) {
            listScroller.moveWindowToElement(mediaIndexMap.get(lastCurrentMedia.value));
        }

        nextTick(() => {
            const currentElem = container.value?.querySelector(".search-result-item.last-current");
            if (currentElem) {
                currentElem.scrollIntoView();
            }
        });
    });
};

watch(
    () => props.min,
    () => {
        if (!props.min) {
            scrollToLastCurrentMedia();
        }
    },
);

/**
 * Automatically focuses the appropriate element
 */
const autoFocus = () => {
    if (isTouchDevice()) {
        return;
    }
    nextTick(() => {
        const el = container.value?.querySelector(".auto-focus") as HTMLInputElement;
        if (el) {
            el.focus();
            if (el.select) {
                el.select();
            }
        }
    });
};

onMounted(autoFocus);

/**
 * Sets the search mode
 * @param newMode The new search mode
 */
const setMode = (newMode: SearchMode) => {
    mode.value = newMode;
    setPreferredSearchMode(newMode);

    if (newMode !== "adv") {
        tagMode.value = "all";
    }

    if (newMode === "image") {
        textSearch.value = "";
    }

    startSearch();
    autoFocus();
};

/**
 * Called when the onlyImages value is changed by the user
 */
const onOnlyImagesChanged = () => {
    setSemanticSearchOnlyImages(onlyImages.value);
    startSearch();
};

/**
 * Find the index of the current media in the list
 * @returns the index, or -1 if not found
 */
const findCurrentMediaIndex = (): number => {
    if (mediaIndexMap.has(currentMedia.value)) {
        return mediaIndexMap.get(currentMedia.value);
    } else {
        return -1;
    }
};

/**
 * Call when the current media changes
 */
const onCurrentMediaChanged = () => {
    if (!props.inModal) {
        const completePageList = listScroller.list;
        const i = findCurrentMediaIndex();
        PagesController.OnPageLoad(i, completePageList.length, 0, 1);
    }
};

// Make sure to unload the page context before the component unmounts
onBeforeUnmount(() => {
    if (!props.inModal) {
        PagesController.OnPageUnload();
    }
});

/**
 * Navigates or selects a media element
 * @param m The media element
 */
const goToMedia = (m: MediaListItem) => {
    if (props.inModal) {
        emit("select-media", m, () => {
            const fullList = listScroller.list;
            const centerPosition = listScroller.getCenterPosition();

            const mediaIndex = mediaIndexMap.get(m.id);

            if (mediaIndex !== undefined) {
                fullList.splice(mediaIndex, 1);
                mediaIndexMap.delete(m.id);

                listScroller.moveWindowToElement(centerPosition);

                for (let i = mediaIndex; i < fullList.length; i++) {
                    mediaIndexMap.set(fullList[i].id, i);
                }
            }
        });
    } else {
        AppStatus.ClickOnMedia(m.id, true);
    }
};

/**
 * Navigates to the previous media
 */
const prevMedia = () => {
    const completePageList = listScroller.list;
    const i = findCurrentMediaIndex();
    if (i !== -1 && i > 0) {
        goToMedia(completePageList[i - 1]);
    } else if (i === -1 && completePageList.length > 0) {
        goToMedia(completePageList[0]);
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_PREV, prevMedia);

/**
 * Navigates to the next media element
 */
const nextMedia = () => {
    const completePageList = listScroller.list;
    const i = findCurrentMediaIndex();
    if (i !== -1 && i < completePageList.length - 1) {
        goToMedia(completePageList[i + 1]);
    } else if (i === -1 && completePageList.length > 0) {
        goToMedia(completePageList[0]);
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_NEXT, nextMedia);

/**
 * Event handler for 'submit'
 * @param event The event
 */
const onSubmit = (event?: Event) => {
    if (event) {
        event.preventDefault();
    }

    startSearch();

    const elementToFocus = container.value?.querySelector(".search-results") as HTMLElement;
    if (elementToFocus) {
        elementToFocus.focus();
    }
};

// Padding for the last row
const { lastRowPadding, rowPaddingPreserveCols } = usePageLastRowPadding(
    props,
    topContainer,
    computed(() => pageItems.value.length),
    windowPosition,
);

/**
 * Event handler for 'keydown' on the element after the tags suggestions
 * @param event The keyboard event
 */
const onTagsSkipKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && event.shiftKey) {
        const inputElem = container.value?.querySelector(".tags-input-search") as HTMLElement;
        if (inputElem) {
            event.preventDefault();
            inputElem.focus();
        }
    }
};

/**
 * Called to skip the tag suggestions when tabbing
 */
const skipTagSuggestions = () => {
    const elem = container.value?.querySelector(".tags-focus-skip") as HTMLElement;
    if (elem) {
        elem.focus();
    }
};

// Priority for the global keyboard handler
const KEYBOARD_HANDLER_PRIORITY = 20;

useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !AppStatus.IsPageVisible() || !event.key || event.ctrlKey) {
        return false;
    }

    if (props.inModal) {
        return false;
    }

    if (event.key === "ArrowLeft") {
        prevMedia();
        return true;
    }

    if (event.key === "ArrowRight") {
        nextMedia();
        return true;
    }

    return false;
}, KEYBOARD_HANDLER_PRIORITY);
</script>
