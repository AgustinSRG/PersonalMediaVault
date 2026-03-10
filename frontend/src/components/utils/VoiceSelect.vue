<template>
    <div ref="container" class="big-selector-container" tabindex="0" :class="{ expanded: expanded }" @keydown="onKeyDown">
        <div class="big-selector" :class="{ expanded: expanded, disabled: disabled }" @click="toggleExpand">
            <div class="big-selector-name">{{ voice ? voiceName || voice : $t("Default voice") }}</div>
            <div class="big-selector-chevron">
                <div class="chevron"></div>
            </div>
        </div>

        <div v-if="expanded" class="big-selector-suggestions-container">
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
            <div ref="suggestionsContainer" class="big-selector-suggestions" @scroll.passive="onSuggestionsScroll">
                <div
                    v-for="s in suggestions"
                    :key="s.id"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="big-selector-suggestion"
                    tabindex="0"
                    :title="s.name"
                    @click="clickSuggestion(s, $event)"
                    @keydown="onSuggestionKeyDown"
                >
                    <span>{{ s.id ? s.name || s.id : $t("Default voice") }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useFocusTrap } from "@/composables/use-focus-trap";
import { useI18n } from "@/composables/use-i18n";
import { useTimeout } from "@/composables/use-timeout";
import { BigListScroller } from "@/utils/big-list-scroller";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { isSpeechSynthesisAvailable } from "@/utils/voice-synthesis";
import { nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue";

// Speech synthesis available?
const speechSynthesisAvailable = isSpeechSynthesisAvailable();

// Translation function
const { $t } = useI18n();

/**
 * Filtered voice item
 */
interface VoiceItemFiltered {
    // ID
    id: string;

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

// Voice ID model
const voice = defineModel<string>("voice", {
    default: "",
});

// Album name
const voiceName = ref(voice.value);

// Map of voices to voice names
const voicesMap = new Map<string, string>();

/**
 * Updates voices
 */
const updateVoices = () => {
    const voices = speechSynthesis.getVoices();

    voicesMap.clear();

    for (const voice of voices) {
        voicesMap.set(voice.voiceURI, voice.name + " (" + voice.lang + ")");
    }

    const chosenVoice = voicesMap.get(voice.value);
    voiceName.value = chosenVoice || voice.value;
};

if (speechSynthesisAvailable) {
    onMounted(updateVoices);

    speechSynthesis.addEventListener("voiceschanged", updateVoices);

    onBeforeUnmount(() => {
        speechSynthesis.removeEventListener("voiceschanged", updateVoices);
    });
}

watch(voice, () => {
    const chosenVoice = voicesMap.get(voice.value);
    voiceName.value = chosenVoice || voice.value;
});

// Voice name filter
const filter = ref("");

// Voices suggestions
const suggestions = ref<VoiceItemFiltered[]>([]);

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

    const voiceFilter = normalizeString(filter.value).trim().toLowerCase();
    const voiceFilterWords = filterToWords(voiceFilter);

    const suggestions: VoiceItemFiltered[] = Array.from(voicesMap.entries())
        .map(([id, name]) => {
            const i = voiceFilter ? matchSearchFilter(name, voiceFilter, voiceFilterWords) : 0;
            return {
                id: id,
                name: name,
                nameLowerCase: name.toLowerCase(),
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
            id: "",
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

// Main container of the component
const container = useTemplateRef("container");

// Focus trap
useFocusTrap(container, expanded, close);

/**
 * Event handler for 'click on a suggestion'
 * @param s The suggestion item
 * @param e The click event
 */
const clickSuggestion = (s: VoiceItemFiltered, e: Event) => {
    e.preventDefault();
    e.stopPropagation();

    voice.value = s.id;
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
