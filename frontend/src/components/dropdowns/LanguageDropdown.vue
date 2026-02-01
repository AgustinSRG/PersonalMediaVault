<template>
    <DropdownContainer v-model:display="display" :focus-trap-exception-class="'top-bar-button-dropdown'">
        <div class="modal-header-corner">
            <div class="modal-header-corner-title">{{ $t("Choose your language") }}</div>
        </div>
        <div class="modal-body with-menu limited-height">
            <table class="modal-menu">
                <tbody>
                    <tr
                        v-for="l in languages"
                        :key="l.id"
                        class="modal-menu-item"
                        tabindex="0"
                        @keydown="clickOnEnter"
                        @click="changeLocale(l.id)"
                    >
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: language !== l.id }"></i>
                        </td>
                        <td class="modal-menu-item-title">{{ l.name }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </DropdownContainer>
</template>

<script setup lang="ts">
import { clickOnEnter } from "@/utils/events";
import DropdownContainer from "./common/DropdownContainer.vue";
import { getLanguage, setLanguage } from "@/i18n";
import { ref } from "vue";
import { AVAILABLE_LANGUAGES } from "@/i18n";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Display
const display = defineModel<boolean>("display");

/// List of available languages
const languages = AVAILABLE_LANGUAGES.map((l) => {
    return {
        id: l.id,
        name: l.name,
    };
}).sort((a, b) => {
    if (a.name < b.name) {
        return -1;
    } else {
        return 1;
    }
});

// Current language
const language = ref(getLanguage());

/**
 * Changes the language
 * @param l The language
 */
const changeLocale = (l: string) => {
    language.value = l;
    setLanguage(l);
};
</script>
