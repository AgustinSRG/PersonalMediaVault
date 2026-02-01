<template>
    <DropdownContainer v-model:display="display" :focus-trap-exception-class="'top-bar-button-dropdown'">
        <div class="modal-header-corner">
            <div class="modal-header-corner-title">{{ $t("Select a theme for the app") }}</div>
        </div>
        <div class="modal-body with-menu limited-height">
            <table class="modal-menu">
                <tbody>
                    <tr class="modal-menu-item" tabindex="0" @click="setTheme('dark')" @keydown="clickOnEnter">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: theme !== 'dark' }"></i>
                        </td>
                        <td class="modal-menu-item-title">{{ $t("Dark Theme") }}</td>
                    </tr>
                    <tr class="modal-menu-item" tabindex="0" @click="setTheme('light')" @keydown="clickOnEnter">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-check" :class="{ unchecked: theme !== 'light' }"></i>
                        </td>
                        <td class="modal-menu-item-title">{{ $t("Light Theme") }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </DropdownContainer>
</template>

<script setup lang="ts">
import { clickOnEnter } from "@/utils/events";
import DropdownContainer from "./common/DropdownContainer.vue";
import { getTheme, setTheme } from "@/control/app-preferences";
import { ref } from "vue";
import { EVENT_NAME_THEME_CHANGED } from "@/control/app-events";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Display
const display = defineModel<boolean>("display");

// Current theme
const theme = ref(getTheme());

onApplicationEvent(EVENT_NAME_THEME_CHANGED, () => {
    theme.value = getTheme();
});
</script>
