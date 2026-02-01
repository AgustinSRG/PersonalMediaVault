<template>
    <DropdownContainer v-model:display="display" :focus-trap-exception-class="'top-bar-button-dropdown'">
        <div class="modal-header-corner">
            <div class="modal-header-corner-title">{{ $t("Vault settings") }}</div>
        </div>
        <div class="modal-body with-menu">
            <table class="modal-menu">
                <tbody>
                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('theme')">
                        <td class="modal-menu-item-icon"><i class="fas fa-moon"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Change theme (Dark / Light)") }}
                        </td>
                    </tr>

                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('lang')">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-language"></i>
                        </td>
                        <td class="modal-menu-item-title">
                            {{ $t("Change language") }}
                        </td>
                    </tr>

                    <tr v-if="isRoot" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('tasks')">
                        <td class="modal-menu-item-icon">
                            <i class="fas fa-bars-progress"></i>
                        </td>
                        <td class="modal-menu-item-title">
                            {{ $t("Tasks") }}
                        </td>
                    </tr>

                    <tr v-if="isRoot" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('advanced')">
                        <td class="modal-menu-item-icon"><i class="fas fa-cog"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Advanced settings") }}
                        </td>
                    </tr>

                    <tr v-if="isRoot" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('batch')">
                        <td class="modal-menu-item-icon"><i class="fas fa-list"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Batch operation") }}
                        </td>
                    </tr>

                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('clear-browser-data')">
                        <td class="modal-menu-item-icon"><i class="fas fa-broom"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Clear browser data") }}
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
import { useUserPermissions } from "@/composables/use-user-permissions";

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

// User permissions
const { isRoot } = useUserPermissions();
</script>
