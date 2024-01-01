<template>
    <div
        class="modal-container modal-container-corner no-transition"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
        :aria-hidden="!display"
        @mousedown="close"
        @touchstart="close"
        @keydown="keyDownHandle"
    >
        <div
            v-if="display"
            class="modal-dialog modal-sm"
            role="document"
            @click="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
        >
            <div class="modal-header-corner">
                <div class="modal-header-corner-title">{{ $t("Vault settings") }}</div>
            </div>
            <div class="modal-body with-menu">
                <table class="modal-menu">
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

                    <tr v-if="isRoot" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('username')">
                        <td class="modal-menu-item-icon"><i class="fas fa-user"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Change username") }}
                        </td>
                    </tr>

                    <tr v-if="username" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('password')">
                        <td class="modal-menu-item-icon"><i class="fas fa-key"></i></td>
                        <td class="modal-menu-item-title">
                            {{ $t("Change password") }}
                        </td>
                    </tr>

                    <tr v-if="username" class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('invite')">
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
                </table>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { AuthController, EVENT_NAME_AUTH_CHANGED } from "@/control/auth";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
    name: "SettingsDropdown",
    emits: ["update:display", "goto"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            isRoot: AuthController.IsRoot,
            canWrite: AuthController.CanWrite,
            username: AuthController.Username,
        };
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        clickOnOption: function (o: string) {
            this.$emit("goto", o);
            this.close();
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },

        updateAuthInfo: function () {
            this.isRoot = AuthController.IsRoot;
            this.canWrite = AuthController.CanWrite;
            this.username = AuthController.Username;
        },

        keyDownHandle: function (e) {
            e.stopPropagation();
            if (e.key === "Escape") {
                this.close();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "top-bar-button-dropdown");

        if (this.display) {
            this.focusTrap.activate();
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    watch: {
        display: function () {
            if (this.display) {
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },
    },
});
</script>
