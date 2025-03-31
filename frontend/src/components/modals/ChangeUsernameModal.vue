<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Change username") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Current username") }}:</label>
                    <input
                        v-model="currentUsername"
                        type="text"
                        name="current-username"
                        :disabled="busy"
                        maxlength="255"
                        readonly
                        class="form-control form-control-full-width"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("New username") }}:</label>
                    <input
                        v-model="username"
                        type="text"
                        name="username"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("Password") }}:</label>
                    <input
                        v-model="password"
                        type="password"
                        name="password"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-check" :loading="busy"></LoadingIcon> {{ $t("Change username") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { apiAccountChangeUsername } from "@/api/api-account";
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";

export default defineComponent({
    name: "ChangeUsernameModal",
    components: {
        LoadingIcon,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            currentUsername: "",
            username: "",
            password: "",
            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.currentUsername = AuthController.Username;

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.usernameUpdated.bind(this));

        if (this.display) {
            this.error = "";
            this.autoFocus();
        }
    },
    methods: {
        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                if (elem) {
                    elem.focus();
                }
            });
        },

        submit: function (e) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            this.busy = true;
            this.error = "";

            makeApiRequest(apiAccountChangeUsername(this.username, this.password))
                .onSuccess(() => {
                    this.busy = false;
                    AuthController.UpdateUsername(this.username);
                    this.username = "";
                    this.password = "";
                    PagesController.ShowSnackBar(this.$t("Vault username changed!"));
                    this.forceCloseSignal++;
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
                            this.error = this.$t("Invalid password");
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

        usernameUpdated: function () {
            this.currentUsername = AuthController.Username;
        },
    },
});
</script>
