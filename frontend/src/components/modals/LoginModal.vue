<template>
    <div class="modal-container modal-container-login" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("The media vault is locked") }}
                </div>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Username") }}:</label>
                    <input
                        type="text"
                        name="username"
                        v-model="username"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-group">
                    <label>{{ $t("Password") }}:</label>
                    <input
                        type="password"
                        name="password"
                        v-model="password"
                        :disabled="busy"
                        maxlength="255"
                        class="form-control form-control-full-width"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button v-if="!busy && mustWait <= 0" type="submit" class="modal-footer-btn">
                    <i class="fas fa-unlock"></i> {{ $t("Unlock vault") }}
                </button>
                <button v-if="!busy && mustWait === 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait 1 second to try again") }}
                </button>
                <button v-if="!busy && mustWait > 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait $TIME seconds to try again").replace("$TIME", mustWait + "") }}
                </button>
                <button v-if="busy" type="button" disabled class="modal-footer-btn">
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Unlocking vault") }}...
                </button>
            </div>
        </form>
    </div>
</template>

<script lang="ts">
import { apiAuthLogin } from "@/api/api-auth";
import { AuthController } from "@/control/auth";
import { makeApiRequest } from "@/api/request";
import { defineComponent, nextTick } from "vue";

export default defineComponent({
    name: "LoginModal",
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            username: "",
            password: "",
            cooldown: 0,
            mustWait: 0,
            now: Date.now(),
            busy: false,
            error: "",
        };
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

            makeApiRequest(apiAuthLogin(this.username, this.password))
                .onSuccess((response) => {
                    this.busy = false;
                    this.username = "";
                    this.password = "";
                    AuthController.SetSession(response.session_id, response.vault_fingerprint);
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        invalidCredentials: () => {
                            this.error = this.$t("Invalid username or password");
                        },
                        wrongCredentials: () => {
                            this.error = this.$t("Invalid username or password");
                            this.cooldown = Date.now() + 5000;
                        },
                        cooldown: () => {
                            this.error = this.$t("You must wait 5 seconds to try again");
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

        updateNow: function () {
            this.now = Date.now();
            if (this.now < this.cooldown) {
                this.mustWait = Math.max(1, Math.round((this.cooldown - this.now) / 1000));
            } else {
                this.mustWait = 0;
            }
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this.autoFocus();

        this._handles.timer = setInterval(this.updateNow.bind(this), 200);
    },
    watch: {
        display: function () {
            this.error = "";
            this.autoFocus();
        },
    },
    beforeUnmount: function () {
        clearInterval(this._handles.timer);
    },
});
</script>
