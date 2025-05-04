<template>
    <div class="password-input-container">
        <div class="password-input">
            <input
                v-model="valState"
                class="form-control form-control-full-width"
                :type="hidden ? 'password' : 'text'"
                :disabled="disabled"
                maxlength="255"
                :name="name"
                :autocomplete="!hidden ? 'off' : undefined"
                :class="{ 'auto-focus': !!autoFocus }"
                @keydown="onKeyDown"
                @blur="onBlur"
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

export default defineComponent({
    name: "PasswordInput",
    props: {
        val: String,
        disabled: Boolean,
        autoFocus: Boolean,
        name: String,
    },
    emits: ["update:val", "tab-skip"],
    setup(props) {
        return {
            valState: useVModel(props, "val"),
        };
    },
    data: function () {
        return {
            hidden: true,
        };
    },
    mounted: function () {},
    beforeUnmount: function () {},
    methods: {
        toggleHide: function () {
            this.hidden = !this.hidden;

            const inputElement = this.$el.querySelector(".form-control");

            if (inputElement) {
                inputElement.focus();
            }
        },

        onBlur: function () {
            this.hidden = true;
        },

        onKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && !event.shiftKey) {
                this.$emit("tab-skip", event);
            }
        },
    },
});
</script>
