<template>
    <div ref="container">
        <div class="form-group media-tags">
            <label v-if="label">{{ label }}:</label>
            <div v-for="tag in tags" :key="tag" class="media-tag">
                <div class="media-tag-name">{{ getTagName(tag) }}</div>
                <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTag(tag)">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="media-tags-finder">
                <input
                    ref="tagFilterInput"
                    v-model="tagFilter"
                    type="text"
                    autocomplete="off"
                    maxlength="255"
                    class="form-control auto-focus tags-input-search"
                    :placeholder="$t('Search for tags') + '...'"
                    @input="onTagFilterChanged"
                    @keydown="onTagAddKeyDown"
                />
            </div>
        </div>
        <div v-if="tagSuggestions.length > 0" class="button-list-row form-group">
            <button
                v-for="mt in tagSuggestions"
                :key="mt.id"
                type="button"
                class="btn btn-primary btn-sm btn-add-tag"
                @click="addTag(mt)"
                @keydown="onSuggestionKeydown"
            >
                <i class="fas fa-plus"></i> {{ mt.name }}
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { useTagSuggestions } from "@/composables/use-tag-suggestions";
import { useTagNames } from "@/composables/use-tags-names";
import { setLastUsedTag } from "@/control/app-preferences";
import type { MatchingTag } from "@/control/tags";
import { useTemplateRef } from "vue";

const emit = defineEmits<{
    /**
     * The user pressed tag to skip to the next element
     */
    (e: "tab-focus-skip"): void;

    /**
     * The tag list was changed
     */
    (e: "changed"): void;
}>();

// Translation function
const { $t } = useI18n();

// Tag names
const { getTagName } = useTagNames();

defineProps({
    /**
     * Label to display above the input
     */
    label: String,
});

// Tag ID list (model)
const tags = defineModel<number[]>("tags");

// Tag suggestions
const { tagFilter, onTagFilterChanged, tagSuggestions, updateTagSuggestions } = useTagSuggestions(
    "search",
    (id) => !tags.value.includes(id),
);

// Container element
const container = useTemplateRef("container");

// Input where the tag name is typed
const tagFilterInput = useTemplateRef("tagFilterInput");

/**
 * Adds a tag to the list
 * @param tag The tag
 */
const addTag = (tag: MatchingTag) => {
    if (tags.value.includes(tag.id)) {
        return;
    }

    tags.value.push(tag.id);

    emit("changed");

    setLastUsedTag(tag.id, "search");

    updateTagSuggestions();

    tagFilterInput.value?.focus();
    tagFilterInput.value?.select();
};

/**
 * Removes a tag from the list
 * @param tag The tag ID
 */
const removeTag = (tag: number) => {
    tags.value = tags.value.filter((t) => tag !== t);

    emit("changed");

    updateTagSuggestions();

    tagFilterInput.value?.focus();
    tagFilterInput.value?.select();
};

/**
 * Event handler for 'keydown' on the input element
 * @param event The keyboard event
 */
const onTagAddKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Enter") {
        event.preventDefault();

        updateTagSuggestions();

        if (tagSuggestions.value.length > 0) {
            addTag(tagSuggestions.value[0]);

            tagFilter.value = "";

            updateTagSuggestions();
        }
    } else if (event.key === "Tab" && tagFilter.value && !event.shiftKey) {
        updateTagSuggestions();
        event.preventDefault();

        if (tagSuggestions.value.length > 0 && tagSuggestions.value[0].name !== tagFilter.value) {
            tagFilter.value = tagSuggestions.value[0].name;

            updateTagSuggestions();
        }
    } else if (event.key === "Tab" && !event.shiftKey) {
        event.preventDefault();

        emit("tab-focus-skip");
    } else if (event.key === "ArrowDown") {
        const suggestionElem = container.value?.querySelector(".btn-add-tag") as HTMLElement;

        if (suggestionElem) {
            event.preventDefault();

            suggestionElem.focus();
        }
    }

    event.stopPropagation();
};

/**
 * Event handler for 'keydown' on a tag suggestion button
 * @param e The keyboard event
 */
const onSuggestionKeydown = (e: KeyboardEvent) => {
    if (e.key === "ArrowRight" || e.key === "ArrowDown") {
        e.preventDefault();
        e.stopPropagation();

        const nextElem = (e.target as HTMLElement).nextSibling as HTMLElement;

        if (nextElem && nextElem.focus) {
            nextElem.focus();
        }
    } else if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
        e.preventDefault();
        e.stopPropagation();

        const prevElem = (e.target as HTMLElement).previousSibling as HTMLElement;

        if (prevElem && prevElem.focus) {
            prevElem.focus();
        } else {
            tagFilterInput.value?.focus();
            tagFilterInput.value?.select();
        }
    }
};
</script>
