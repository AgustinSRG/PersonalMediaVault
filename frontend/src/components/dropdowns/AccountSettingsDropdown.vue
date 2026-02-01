<template>
    <DropdownContainer
        v-model:display="display"
        :position-class="'modal-container-account'"
        :focus-trap-exception-class="'top-bar-button-dropdown'"
    >
        <div class="modal-header-corner">
            <div class="modal-header-corner-title">{{ $t("Account settings") }}</div>
        </div>
        <div class="modal-body with-menu">
            <table class="modal-menu">
                <tbody>
                    <tr v-if="isRoot" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('username')">
                        <td class="modal-menu-item-icon"><i class="fas fa-user-tag"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Change username") }}
                        </td>
                    </tr>

                    <tr v-if="!isGuest" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('password')">
                        <td class="modal-menu-item-icon"><i class="fas fa-key"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Change password") }}
                        </td>
                    </tr>

                    <tr v-if="!isGuest" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('security')">
                        <td class="modal-menu-item-icon"><i class="fas fa-lock"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Account security") }}
                        </td>
                    </tr>

                    <tr v-if="!isGuest" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('invite')">
                        <td class="modal-menu-item-icon"><i class="fas fa-user-check"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Invite") }}
                        </td>
                    </tr>

                    <tr v-if="isRoot" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('admin')">
                        <td class="modal-menu-item-icon"><i class="fas fa-users"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Administrate accounts") }}
                        </td>
                    </tr>

                    <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('logout')">
                        <td class="modal-menu-item-icon"><i class="fas fa-sign-out-alt"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Close vault") }}
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
import { useUserPermissions } from "@/composables/use-user-permissions";
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

// User permissions
const { isGuest, isRoot } = useUserPermissions();
</script>
