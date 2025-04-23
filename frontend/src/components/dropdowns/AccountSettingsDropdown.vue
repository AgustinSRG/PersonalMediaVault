<template>
    <div
        class="modal-container modal-container-corner modal-container-account no-transition"
        :class="{ hidden: !display }"
        tabindex="-1"
        role="dialog"
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

                        <tr class="modal-menu-item" tabindex="0" @keydown="clickOnEnter" @click="clickOnOption('logout')">
                            <td class="modal-menu-item-icon"><i class="fas fa-sign-out-alt"></i></td>
                            <td class="modal-menu-item-title">
                                {{ $t("Close vault") }}
                            </td>
                        </tr>
                    </tbody>
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
    name: "AccountSettingsDropdown",
    props: {
        display: Boolean,
    },
    emits: ["update:display", "goto"],
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
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        clickOnOption: function (o: string) {
            this.$emit("goto", o);
            this.close();
        },

        updateAuthInfo: function () {
            this.isRoot = AuthController.IsRoot;
            this.canWrite = AuthController.CanWrite;
            this.username = AuthController.Username;
        },

        keyDownHandle: function (e: KeyboardEvent) {
            e.stopPropagation();
            if (e.key === "Escape") {
                this.close();
            }
        },
    },
});
</script>
