<template>
    <ModalDialogContainer :closeSignal="closeSignal" v-model:display="displayStatus">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Create account") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Account name") }}:</label>
                    <input
                        type="text"
                        autocomplete="off"
                        v-model="accountUsername"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Account password") }}:</label>
                    <input
                        type="password"
                        autocomplete="new-password"
                        v-model="accountPassword"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Account password") }} ({{ $t("Again") }}):</label>
                    <input
                        type="password"
                        autocomplete="new-password"
                        v-model="accountPassword2"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Account type") }}:</label>
                    <select v-model="accountWrite" class="form-control form-select form-control-full-width">
                        <option :value="false">{{ $t("Read only") }}</option>
                        <option :value="true">{{ $t("Read / Write") }}</option>
                    </select>
                </div>

                <div class="form-group form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button type="submit" :disabled="busy" class="modal-footer-btn">
                    <i class="fas fa-plus"></i> {{ $t("Create account") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAdminCreateAccount } from "@/api/api-admin";

export default defineComponent({
    components: {},
    name: "AccountCreateModal",
    emits: ["update:display", "account-created"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            accountUsername: "",
            accountPassword: "",
            accountPassword2: "",
            accountWrite: false,

            busy: false,
            error: "",

            closeSignal: 0,
        };
    },
    methods: {
        autoFocus: function () {
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        },

        submit: function (e) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            if (this.accountPassword !== this.accountPassword2) {
                this.error = this.$t("The passwords do not match");
                return;
            }

            const username = this.accountUsername;
            const password = this.accountPassword;
            const write = this.accountWrite;

            this.busy = true;
            this.error = "";

            makeApiRequest(apiAdminCreateAccount(username, password, write))
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Account created") + ": " + username);
                    this.accountUsername = "";
                    this.accountPassword = "";
                    this.accountPassword2 = "";
                    this.accountWrite = false;
                    this.$emit("account-created");
                    this.close();
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidUsername: () => {
                            this.error = this.$t("Invalid username provided");
                        },
                        usernameInUse: () => {
                            this.error = this.$t("The username is already in use");
                        },
                        invalidPassword: () => {
                            this.error = this.$t("Invalid password provided");
                        },
                        badRequest: () => {
                            this.error = this.$t("Bad request");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        serverError: () => {
                            this.error = this.$t("Internal server error");
                        },
                        networkError: () => {
                            this.error = this.$t("Could not connect to the server");
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.error = err.message;
                    console.error(err);
                    this.busy = false;
                });
        },

        close: function () {
            this.closeSignal++;
        },
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
            this.autoFocus();
        }
    },
    beforeUnmount: function () {},
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.autoFocus();
            }
        },
    },
});
</script>
