<template>
    <div ref="container" class="album-selector-container" tabindex="0" :class="{ expanded: expanded }" @keydown="onKeyDown">
        <div class="album-selector" :class="{ expanded: expanded, disabled: disabled }" @click="toggleExpand">
            <div class="album-selected-name">{{ albumName }}</div>
            <div class="album-selector-chevron">
                <div class="chevron"></div>
            </div>
        </div>

        <div v-if="expanded" class="album-selector-suggestions-container">
            <div class="name-filter-input-container">
                <input
                    ref="nameFilterInput"
                    v-model="filter"
                    type="text"
                    class="form-control form-control-full-width name-filter-input"
                    autocomplete="off"
                    :placeholder="$t('Filter by name') + '...'"
                    @input="onFilterChanged"
                    @keydown="onInputKeyDown"
                />
            </div>
            <div ref="suggestionsContainer" class="album-selector-suggestions" @scroll.passive="onSuggestionsScroll">
                <a
                    v-for="s in suggestions"
                    :key="s.id"
                    target="_blank"
                    rel="noopener noreferrer"
                    :href="getSuggestionURL(s)"
                    class="album-selector-suggestion"
                    tabindex="0"
                    :title="s.name"
                    @click="clickSuggestion(s, $event)"
                    @keydown="onSuggestionKeyDown"
                >
                    <span>{{ s.name }}</span>
                </a>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onApplicationEvent } from "@/composables/on-app-event";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { useI18n } from "@/composables/use-i18n";
import { useTimeout } from "@/composables/use-timeout";
import { AlbumsController } from "@/control/albums";
import { EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/app-events";
import { getFrontendUrl } from "@/utils/api";
import { BigListScroller } from "@/utils/big-list-scroller";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { nextTick, ref, useTemplateRef, watch } from "vue";

// Translation function
const { $t } = useI18n();

/**
 * Filtered album item
 */
interface AlbumItemFiltered {
    // ID
    id: number;

    // Name
    name: string;

    // Name (lowercase)
    nameLowerCase: string;

    // True if starts
    starts: boolean;

    // True if contains
    contains: boolean;
}

const props = defineProps({
    /**
     * True if the control is disabled
     */
    disabled: Boolean,
});

// Album ID model
const album = defineModel<number>("album", {
    default: -1,
});

/**
 * Gets the name of a selected album
 * @param id Album ID
 * @returns Album name
 */
const getAlbumName = (id: number): string => {
    if (id < 0) {
        return "--";
    }

    const album = AlbumsController.AlbumsMap.get(id);

    if (!album) {
        return "";
    }

    return album.name;
};

// Album name
const albumName = ref(getAlbumName(album.value));

watch(album, () => {
    albumName.value = getAlbumName(album.value);
});

// Full list of albums
let albums = AlbumsController.GetAlbumsListMin();

// When the albums list changes, update everything
onApplicationEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, () => {
    albums = AlbumsController.GetAlbumsListMin();

    albumName.value = getAlbumName(album.value);

    updateSuggestions();
});

// Album title filter
const filter = ref("");

// Album suggestions
const suggestions = ref<AlbumItemFiltered[]>([]);

// Max number of items that will fit in the visible section of the scroller
const LIST_SCROLLER_ITEMS_FIT = 9;

// Scroller for the suggestions
const bigListScroller = new BigListScroller(BigListScroller.GetWindowSize(LIST_SCROLLER_ITEMS_FIT), {
    get: () => {
        return suggestions.value;
    },
    set: (list) => {
        suggestions.value = list;
    },
});

/**
 * Event handler for 'scroll' on the suggestions container
 * @param e The scroll event
 */
const onSuggestionsScroll = (e: Event) => {
    bigListScroller.checkElementScroll(e.target as HTMLElement);
};

// Container element for suggestions
const suggestionsContainer = useTemplateRef("suggestionsContainer");

/**
 * Updates the album suggestions based on the filter
 * typed by the user
 */
const updateSuggestions = () => {
    filterChangedTimeout.clear();

    const albumFilter = normalizeString(filter.value).trim().toLowerCase();
    const albumFilterWords = filterToWords(albumFilter);

    const suggestions: AlbumItemFiltered[] = albums
        .map((a) => {
            const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
            return {
                id: a.id,
                name: a.name,
                nameLowerCase: a.nameLowerCase,
                starts: i === 0,
                contains: i >= 0,
            };
        })
        .filter((a) => {
            return a.starts || a.contains;
        })
        .sort((a, b) => {
            if (a.starts && !b.starts) {
                return -1;
            } else if (b.starts && !a.starts) {
                return 1;
            } else if (a.nameLowerCase < b.nameLowerCase) {
                return -1;
            } else {
                return 1;
            }
        });

    if (!filter.value) {
        suggestions.unshift({
            id: -1,
            name: "--",
            nameLowerCase: "--",
            contains: true,
            starts: true,
        });
    }

    bigListScroller.reset();
    bigListScroller.addElements(suggestions);

    nextTick(() => {
        if (suggestionsContainer.value) {
            // Reset scroll of the suggestions container
            suggestionsContainer.value.scrollTop = 0;
        }
    });
};

// Timeout to update the suggestions
const filterChangedTimeout = useTimeout();

// Delay to update the album suggestions (milliseconds)
const UPDATE_SUGGESTIONS_DELAY = 200;

/**
 * Call when the filter changes
 */
const onFilterChanged = () => {
    if (filterChangedTimeout.isSet()) {
        return;
    }

    filterChangedTimeout.set(updateSuggestions, UPDATE_SUGGESTIONS_DELAY);
};

// Timeout to blur the element and hide the suggestions
const blurTimeout = useTimeout();

// True if the suggestions are visible
const expanded = ref(false);

// Reference to the input element
const nameFilterInput = useTemplateRef("nameFilterInput");

const expand = () => {
    blurTimeout.clear();

    expanded.value = true;

    updateSuggestions();

    AlbumsController.Refresh();

    nextTick(() => {
        nameFilterInput.value?.focus();
        nameFilterInput.value?.select();
    });
};

/**
 * Closes the suggestions instantly
 */
const closeInstantly = () => {
    blurTimeout.clear();

    expanded.value = false;

    filter.value = "";
};

// Delay to blur the element
const BLUR_DELAY = 10;

/**
 * Closes the suggestions
 */
const close = () => {
    blurTimeout.set(closeInstantly, BLUR_DELAY);
};

/**
 * Toggles the expanded status
 */
const toggleExpand = () => {
    if (expanded.value) {
        closeInstantly();
        container.value?.focus();
    } else if (!props.disabled) {
        expand();
    }
};

/**
 * Gets the URL of a suggestion
 * @param s The suggestion item
 * @returns The URL
 */
const getSuggestionURL = (s: AlbumItemFiltered): string => {
    return getFrontendUrl({
        album: s.id,
    });
};

// Main container of the component
const container = useTemplateRef("container");

// Focus trap
useFocusTrap(container, expanded, close);

/**
 * Event handler for 'click on a suggestion'
 * @param s The suggestion item
 * @param e The click event
 */
const clickSuggestion = (s: AlbumItemFiltered, e: Event) => {
    e.preventDefault();
    e.stopPropagation();

    album.value = s.id;
    closeInstantly();
    container.value?.focus();
};

/**
 * Handler for 'keydown' on the container
 * @param event The keyboard event
 */
const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        toggleExpand();
    }
};

/**
 * Handler for 'keydown' on the input text element
 * @param event The keyboard event
 */
const onInputKeyDown = (event: KeyboardEvent) => {
    if (event.key === "ArrowDown" && suggestions.value.length > 0) {
        event.preventDefault();
        const firstSuggestionElement = container.value?.querySelector(".album-selector-suggestion") as HTMLElement;
        if (firstSuggestionElement) {
            firstSuggestionElement.focus();
        }
    } else if (event.key === "Escape") {
        event.preventDefault();
        closeInstantly();
        container.value?.focus();
    }
    event.stopPropagation();
};

/**
 * Handler for 'keydown' on a suggestion element
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
            nameFilterInput.value?.focus();
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
        closeInstantly();
        container.value?.focus();
    }
};
</script>
