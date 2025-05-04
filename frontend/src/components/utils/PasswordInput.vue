<template>
    <div class="password-input-container">
        <div class="password-input">
            <input
                v-model="valState"
                :type="hidden ? 'password' : 'text'"
                class="form-control form-control-full-width"
                :disabled="disabled"
                maxlength="255"
                :name="name"
                :autocomplete="!hidden ? 'off' : isNewPassword ? 'new-password' : undefined"
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

<script lang="ts">
import { defineComponent } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "@/utils/focus-trap";

export default defineComponent({
    name: "PasswordInput",
    props: {
        val: String,
        disabled: Boolean,
        autoFocus: Boolean,
        name: String,
        isNewPassword: Boolean,
    },
    emits: ["update:val", "tab-skip"],
    setup(props) {
        return {
            valState: useVModel(props, "val"),
            focusTrap: null as FocusTrap,
        };
    },
    data: function () {
        return {
            hidden: true,
        };
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.onBlur.bind(this));
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        toggleHide: function () {
            this.hidden = !this.hidden;

            const inputElement = this.$el.querySelector(".form-control");

            if (inputElement) {
                inputElement.focus();
            }
        },

        onFocus: function () {
            this.focusTrap.activate();
        },

        onBlur: function () {
            this.hidden = true;
            this.focusTrap.deactivate();
        },

        onKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && !event.shiftKey) {
                this.$emit("tab-skip", event);
            }
        },
    },
});
</script>
