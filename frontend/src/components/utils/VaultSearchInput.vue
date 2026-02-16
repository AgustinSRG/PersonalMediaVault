<template>
    <form
        ref="container"
        class="top-bar-search-input-container"
        :class="{ focused: focused, 'has-search': hasSearch }"
        tabindex="-1"
        @submit="submitSearch"
    >
        <input
            ref="inputElement"
            v-model="search"
            type="text"
            class="top-bar-search-input"
            name="pmv-search-input"
            spellcheck="false"
            autocorrect="off"
            autocomplete="off"
            autocapitalize="none"
            :placeholder="$t('Search')"
            @keydown="onKeyDown"
            @input="onSearchInput"
            @focus="focusSearch"
        />
        <button v-if="hasSearch" type="button" class="top-bar-search-clear-btn" :title="$t('Clear search')" @click="clearSearch">
            <i class="fas fa-times"></i>
        </button>
        <button type="submit" class="top-bar-button top-bar-search-button" :title="$t('Search')" @focus="blurSearch">
            <i class="fas fa-search"></i>
        </button>
        <div
            ref="suggestionsElement"
            class="top-bar-search-suggestions"
            :class="{ hidden: suggestions.length === 0 }"
            @scroll.passive="onSuggestionsScroll"
        >
            <a
                v-for="s in suggestions"
                :key="s.key"
                target="_blank"
                rel="noopener noreferrer"
                :href="getSuggestionUrl(s)"
                class="top-bar-search-suggestion"
                tabindex="0"
                :title="s.name"
                @click="clickSearch(s, $event)"
                @keydown="onSuggestionKeyDown"
            >
                <i v-if="s.type === 'tag'" class="fas fa-tag"></i>
                <i v-else-if="s.type === 'album'" class="fas fa-list-ol"></i>
                <i v-else-if="s.type === 'search'" class="fas fa-search"></i>
                <i v-else-if="s.type === 'media' && s.mediaType === 1" class="fas fa-image"></i>
                <i v-else-if="s.type === 'media' && s.mediaType === 2" class="fas fa-video"></i>
                <i v-else-if="s.type === 'media' && s.mediaType === 3" class="fas fa-headphones"></i>
                <i v-else class="fas fa-ban"></i>
                <span>{{ s.name }}</span>
            </a>
        </div>
    </form>
</template>

<script setup lang="ts">
import { onApplicationEvent } from "@/composables/on-app-event";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { useI18n } from "@/composables/use-i18n";
import { useTimeout } from "@/composables/use-timeout";
import { AlbumsController } from "@/control/albums";
import { EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { TagsController } from "@/control/tags";
import { getFrontendUrl } from "@/utils/api";
import { BigListScroller } from "@/utils/big-list-scroller";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { parseTagName } from "@/utils/tags";
import { computed, nextTick, ref, useTemplateRef } from "vue";

/**
 * Search bar suggestion
 */
interface SearchBarSuggestion {
    key: string;

    id: string;

    name: string;
    nameLower: string;

    type: "tag" | "album" | "media" | "search";
    mediaType?: number;

    starts: boolean;
    contains: boolean;
}

// Translation function
const { $t } = useI18n();

// Search filter
const search = ref(AppStatus.CurrentSearch);

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    search.value = AppStatus.CurrentSearch;
});

// True if it has a search
const hasSearch = computed(() => !!search.value);

// Ref to the container element
const container = useTemplateRef("container");

// Ref to the input element
const inputElement = useTemplateRef("inputElement");

// Is focused?
const focused = defineModel<boolean>("focused");

// Timeout to blur the element
const blurTimeout = useTimeout();

/**
 * Call when the element gains focus
 */
const focusSearch = () => {
    blurTimeout.clear();

    focused.value = true;

    inputElement.value?.select();

    updateSuggestions();

    AlbumsController.Refresh();
    TagsController.Refresh();
};

/**
 * Called when the element loses focus
 */
const blurSearchInstantly = () => {
    blurTimeout.clear();
    focused.value = false;
};

// Delay to blur (milliseconds)
const BLUR_DELAY = 100;

/**
 * Called when the element loses focus after a delay
 */
const blurSearch = () => {
    blurTimeout.set(blurSearchInstantly, BLUR_DELAY);
};

/**
 * Removes the focus from the input element
 */
const blurSearchInputElement = () => {
    inputElement.value?.blur();
};

// Focus trap
useFocusTrap(container, focused, blurSearch);

// Search suggestions
const suggestions = ref<SearchBarSuggestion[]>([]);

// Max number of items that will fit in the visible section of the scroller
const LIST_SCROLLER_ITEMS_FIT = 9;

// Big list scroller for suggestions
const bigListScroller = new BigListScroller(BigListScroller.GetWindowSize(LIST_SCROLLER_ITEMS_FIT), {
    get: () => {
        return suggestions.value;
    },
    set: (list) => {
        suggestions.value = list;
    },
});

/**
 * Scroll event handler for the suggestions container element
 * @param e The scroll event
 */
const onSuggestionsScroll = (e: Event) => {
    bigListScroller.checkElementScroll(e.target as HTMLElement);
};

// Suggestions element
const suggestionsElement = useTemplateRef("suggestionsElement");

// Timeout to update suggestions with delay
const updateSuggestionsTimeout = useTimeout();

/**
 * Updates the search suggestions
 */
const updateSuggestions = () => {
    updateSuggestionsTimeout.clear();

    const tagFilter = normalizeString(parseTagName(search.value));
    const albumFilter = normalizeString(search.value).trim().toLowerCase();
    const albumFilterWords = filterToWords(albumFilter);

    let newSuggestions: SearchBarSuggestion[] = [];

    newSuggestions = newSuggestions.concat(
        Array.from(TagsController.Tags.entries())
            .map((a): SearchBarSuggestion => {
                const i = tagFilter ? normalizeString(a[1]).indexOf(tagFilter) : 0;
                return {
                    key: "tag:" + a[0],
                    id: a[0] + "",
                    name: a[1],
                    nameLower: a[1],
                    starts: i === 0,
                    contains: i >= 0,
                    type: "tag",
                };
            })
            .filter((a) => {
                return a.starts || a.contains;
            }),
    );

    newSuggestions = newSuggestions.concat(
        AlbumsController.GetAlbumsList()
            .map((a): SearchBarSuggestion => {
                const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
                return {
                    key: "album:" + a.id,
                    id: a.id + "",
                    name: a.name,
                    nameLower: a.name.toLowerCase(),
                    starts: i === 0,
                    contains: i >= 0,
                    type: "album",
                };
            })
            .filter((a) => {
                return a.starts || a.contains;
            }),
    );

    if (AlbumsController.CurrentAlbumData) {
        newSuggestions = newSuggestions.concat(
            AlbumsController.CurrentAlbumData.list
                .map((a): SearchBarSuggestion => {
                    const i = albumFilter ? matchSearchFilter(a.title, albumFilter, albumFilterWords) : 0;
                    return {
                        key: "media:" + a.id,
                        id: a.id + "",
                        name: a.title,
                        nameLower: a.title.toLowerCase(),
                        starts: i === 0,
                        contains: i >= 0,
                        type: "media",
                        mediaType: a.type,
                    };
                })
                .filter((a) => {
                    return a.starts || a.contains;
                }),
        );
    }

    newSuggestions = newSuggestions.sort((a, b) => {
        if (a.starts && !b.starts) {
            return -1;
        } else if (b.starts && !a.starts) {
            return 1;
        } else if (a.nameLower < b.nameLower) {
            return -1;
        } else {
            return 1;
        }
    });

    if (search.value) {
        newSuggestions.push({
            key: "search:text",
            id: "search:text",
            name: search.value,
            nameLower: search.value.toLowerCase(),
            starts: true,
            contains: true,
            type: "search",
        });
    }

    bigListScroller.reset();
    bigListScroller.addElements(newSuggestions);

    nextTick(() => {
        if (suggestionsElement.value) {
            suggestionsElement.value.scrollTop = 0;
        }
    });
};

// Delay to update suggestions (milliseconds)
const SUGGESTIONS_UPDATE_DELAY = 200;

/**
 * Called when the user changes the search input
 */
const onSearchInput = () => {
    if (updateSuggestionsTimeout.isSet()) {
        return;
    }

    updateSuggestionsTimeout.set(updateSuggestions, SUGGESTIONS_UPDATE_DELAY);
};

/**
 * Submits the search form
 * @param event The submit event
 */
const submitSearch = (event?: Event) => {
    if (event) {
        event.preventDefault();
    }

    blurSearchInstantly();

    if (!search.value) {
        goSearch();
        return;
    }

    updateSuggestions();

    if (suggestions.value.length > 0) {
        selectSearch(suggestions.value[0]);
    } else {
        goFindMedia();
    }
};

/**
 * Clears the search query
 */
const clearSearch = () => {
    blurSearchInstantly();
    search.value = "";
    goSearch();
};

/**
 * Navigates to the search page
 */
const goSearch = () => {
    AppStatus.GoToSearch(search.value);
    blurSearchInputElement();
};

/**
 * Navigates to the 'Find Media' page
 */
const goFindMedia = () => {
    AppStatus.GoToSearch("");
    AppStatus.GoFindMedia(search.value);
};

/**
 * The user selected a suggestion
 * to navigate into it
 * @param s The suggestion
 */
const selectSearch = (s: SearchBarSuggestion) => {
    if (s.type === "album") {
        search.value = "";
        AppStatus.ClickOnAlbum(Number(s.id));
        blurSearchInputElement();
    } else if (s.type === "media") {
        search.value = "";
        AppStatus.ClickOnMedia(Number(s.id), false);
        blurSearchInputElement();
    } else if (s.type === "tag") {
        search.value = s.name;
        goSearch();
    } else {
        goFindMedia();
    }

    blurSearchInstantly();
};

/**
 * The user clicked a suggestion element
 * @param s The element
 * @param e The click event
 */
const clickSearch = (s: SearchBarSuggestion, e: Event) => {
    e.preventDefault();
    blurSearch();
    selectSearch(s);
};

/**
 * Gets the URL for a suggestion
 * @param s The suggestion
 * @returns The URL
 */
const getSuggestionUrl = (s: SearchBarSuggestion): string => {
    switch (s.type) {
        case "media":
            return getFrontendUrl({
                media: s.id,
            });
        case "album":
            return getFrontendUrl({
                album: s.id,
            });
        case "tag":
            return getFrontendUrl({
                page: "media",
                search: s.name,
            });
        default:
            return getFrontendUrl({ page: "home" });
    }
};

/**
 * Event handler for 'keydown'
 * @param event The keyboard event
 */
const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && search.value && !event.shiftKey) {
        if (suggestions.value.length > 0 && search.value !== suggestions.value[0].name) {
            search.value = suggestions.value[0].name;
            onSearchInput();
            event.preventDefault();
        } else if (suggestions.value.length > 0 && search.value === suggestions.value[0].name) {
            event.preventDefault();
            const suggestionElement = suggestionsElement.value?.querySelector(".top-bar-search-suggestion") as HTMLElement;
            if (suggestionElement) {
                suggestionElement.focus();
            }
        }
    } else if (event.key.toUpperCase() === "F" && event.ctrlKey) {
        blurSearch();
    } else if (event.key === "Escape") {
        event.preventDefault();
        const inputElem = event.target as HTMLElement;
        if (inputElem && inputElem.blur) {
            inputElem.blur();
        }
        blurSearch();
    } else if (event.key === "ArrowDown" && suggestions.value.length > 0) {
        event.preventDefault();
        const suggestionElement = suggestionsElement.value?.querySelector(".top-bar-search-suggestion") as HTMLElement;
        if (suggestionElement) {
            suggestionElement.focus();
        }
    }

    event.stopPropagation();
};

/**
 * Event handler for 'keydown' on a suggestion element
 * @param event The keyboard event
 */
const onSuggestionKeyDown = (event: KeyboardEvent) => {
    const target = event.target as HTMLElement;
    if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        target.click();
    } else if (event.key === "ArrowUp") {
        event.preventDefault();
        event.stopPropagation();
        const sibling = target.previousElementSibling as HTMLElement;
        if (sibling && sibling.focus) {
            sibling.focus();
        } else {
            inputElement.value?.focus();
        }
    } else if (event.key === "ArrowDown") {
        event.preventDefault();
        event.stopPropagation();
        const sibling = target.nextElementSibling as HTMLElement;
        if (sibling && sibling.focus) {
            sibling.focus();
        }
    } else if (event.key === "Escape") {
        event.preventDefault();
        event.stopPropagation();
        inputElement.value?.focus();
    }
};

/**
 * Focuses the input element
 */
const focus = () => {
    focusSearch();
    nextTick(() => {
        inputElement.value.focus();
    });
};

// Expose methods
defineExpose({
    focus,
});
</script>
