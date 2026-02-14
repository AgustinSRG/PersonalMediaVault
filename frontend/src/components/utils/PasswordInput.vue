<template>
    <div ref="container" class="password-input-container">
        <div class="password-input">
            <input
                v-model="val"
                :type="hidden ? 'password' : 'text'"
                class="form-control form-control-full-width"
                :disabled="disabled"
                maxlength="255"
                :name="name"
                :autocomplete="!hidden ? 'off' : isNewPassword ? 'new-password' : 'current-password'"
                :class="{ 'auto-focus': !!autoFocus }"
                @keydown="onKeyDown"
                @focus="onFocus"
            />
            <button
                type="button"
                :disabled="disabled"
                class="password-input-hide-btn"
                :title="hidden ? $t('Show password') : $t('Hide password')"
                @click="toggleHide"
            >
                <i v-if="hidden" class="fas fa-eye"></i>
                <i v-else class="fas fa-eye-slash"></i>
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useFocusTrap } from "@/composables/use-focus-trap";
import { useI18n } from "@/composables/use-i18n";
import { ref, useTemplateRef } from "vue";

// Translation function
const { $t } = useI18n();

const emit = defineEmits<{
    /**
     * Event emitted when the user tabs out of the element
     */
    (e: "tab-skip", keyboardEvent: KeyboardEvent): void;
}>();

defineProps({
    /**
     * True if the input is disabled
     */
    disabled: Boolean,

    /**
     * True if the input has the class 'auto-focus'
     */
    autoFocus: Boolean,

    /**
     * The 'name' attribute for the input
     */
    name: String,

    /**
     * True if the input is a new password,
     * in order to indicate to the browser autocomplete
     */
    isNewPassword: Boolean,
});

// Value model
const val = defineModel<string>("val");

/// True if the password input is hidden, false if visible
const hidden = ref(true);

// Ref to the container element
const container = useTemplateRef("container");

/**
 * Toggles the password visibility
 */
const toggleHide = () => {
    hidden.value = !hidden.value;

    const inputElement = container.value.querySelector(".form-control") as HTMLInputElement;

    if (inputElement) {
        inputElement.focus();
    }
};

// Focused state
const focused = ref(false);

// Called on focus
const onFocus = () => {
    focused.value = true;
};

// Called on blur
const onBlur = () => {
    hidden.value = true;
    focused.value = false;
};

// Focus trap
useFocusTrap(container, focused, onBlur);

/**
 * Event listener for 'keydown' on the input
 * @param event The event
 */
const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Tab" && !event.shiftKey) {
        emit("tab-skip", event);
    }
};
</script>
