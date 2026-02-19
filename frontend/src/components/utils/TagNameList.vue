<template>
    <div ref="container">
        <div class="form-group media-tags">
            <div v-for="tag in tags" :key="tag" class="media-tag">
                <div class="media-tag-name">{{ tag }}</div>
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
                    class="form-control tag-to-add auto-focus"
                    :placeholder="$t('Add tags') + '...'"
                    @input="onTagFilterChanged"
                    @keydown="onTagInputKeyDown"
                />
            </div>
            <div class="media-tags-adder">
                <button type="button" :disabled="!tagFilter" class="btn btn-primary btn-xs" @click="addTag">
                    <i class="fas fa-plus"></i> {{ $t("Add tag") }}
                </button>
            </div>
        </div>
        <div v-if="tagSuggestions.length > 0" class="form-group">
            <button
                v-for="mt in tagSuggestions"
                :key="mt.id"
                type="button"
                class="btn btn-primary btn-sm btn-tag-mini btn-add-tag"
                @click="addTagSuggestion(mt)"
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
import type { MatchingTag } from "@/control/tags";
import { parseTagName } from "@/utils/tags";
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

// Tag name list (model)
const tags = defineModel<string[]>("tags");

// Tag suggestions
const { tagFilter, onTagFilterChanged, tagSuggestions, updateTagSuggestions } = useTagSuggestions(
    "edit",
    (_id, name) => !tags.value.includes(name),
);

// Container element
const container = useTemplateRef("container");

// Input where the tag name is typed
const tagFilterInput = useTemplateRef("tagFilterInput");

/**
 * Adds a tag to the list
 * @param tag The tag name
 */
const addTagByName = (tag: string) => {
    tag = parseTagName(tag);

    if (tags.value.includes(tag)) {
        tags.value = tags.value.filter((t) => tag !== t);
    }

    tags.value.push(tag);
};

/**
 * Adds the tag typed by the user
 */
const addTag = () => {
    addTagByName(tagFilter.value);

    tagFilter.value = "";

    emit("changed");

    updateTagSuggestions();

    tagFilterInput.value?.focus();
    tagFilterInput.value?.select();
};

/**
 * Adds a tag suggestion
 * @param tag The tag
 */
const addTagSuggestion = (tag: MatchingTag) => {
    addTagByName(tag.name);

    emit("changed");

    updateTagSuggestions();

    tagFilterInput.value?.focus();
    tagFilterInput.value?.select();
};

/**
 * Removes a tag from the list
 * @param tag The tag name
 */
const removeTag = (tag: string) => {
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
const onTagInputKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && !event.shiftKey) {
        if (tagFilter.value) {
            updateTagSuggestions();
            event.preventDefault();

            if (tagSuggestions.value.length > 0 && tagSuggestions.value[0].name !== tagFilter.value) {
                tagFilter.value = tagSuggestions.value[0].name;

                updateTagSuggestions();
            }
        } else {
            event.preventDefault();

            emit("tab-focus-skip");
        }
    } else if (event.key === "Enter") {
        if (tagFilter.value) {
            addTag();
        }
    } else if (event.key === "ArrowDown") {
        const suggestionElem = container.value?.querySelector(".btn-add-tag") as HTMLElement;

        if (suggestionElem) {
            event.preventDefault();

            suggestionElem.focus();
        }
    }
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
