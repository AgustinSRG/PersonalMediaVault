<template>
    <DropdownContainer
        v-model:display="display"
        :position-class="'modal-container-help'"
        :focus-trap-exception-class="'top-bar-button-dropdown'"
    >
        <div class="modal-header-corner">
            <div class="modal-header-corner-title">{{ $t("Help") }}</div>
        </div>
        <div class="modal-body with-menu">
            <table class="modal-menu">
                <tbody>
                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('about')">
                        <td class="modal-menu-item-icon"><i class="fas fa-question"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("About PMV") }}
                        </td>
                    </tr>

                    <tr class="tr-link modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="openDocs">
                        <td class="modal-menu-item-icon"><i class="fas fa-book"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Documentation") }}
                        </td>
                    </tr>

                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('keyboard')">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-keyboard"></i>
                        </td>
                        <td class="modal-menu-item-title">
                            {{ $t("Keyboard shortcuts") }}
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </DropdownContainer>
</template>

<script setup lang="ts">
import { clickOnEnter } from "@/utils/events";
import DropdownContainer from "./common/DropdownContainer.vue";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Display
const display = defineModel<boolean>("display");

const emit = defineEmits<{
    /**
     * Event to go to an option of the dropdown menu
     */
    (e: "goto", option: string): void;
}>();

/**
 * Call when the user click on an option
 * @param o The option
 */
const clickOnOption = (o: string) => {
    emit("goto", o);
    display.value = false;
};

/**
 * Documentation URL
 */
const DOCS_URL = import.meta.env.VITE__DOCS_URL || "#";

/**
 * Opens the documentation in a new browser window
 */
const openDocs = () => {
    window.open(DOCS_URL);
    display.value = false;
};
</script>
