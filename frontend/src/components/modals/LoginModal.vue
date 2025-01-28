<template>
    <div class="modal-container modal-container-login" :class="{ hidden: !display }" tabindex="-1" role="dialog">
        <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("The media vault is locked") }}
                </div>
            </div>
            <div class="modal-body">
                <div class="horizontal-filter-menu two-child no-border">
                    <a href="javascript:;" @click="changeToCredentials" class="horizontal-filter-menu-item" :class="{ selected: !isCode }"
                        ><i class="fas fa-key"></i> {{ $t("Credentials") }}</a
                    >
                    <a href="javascript:;" @click="changeToCode" class="horizontal-filter-menu-item" :class="{ selected: isCode }"
                        ><i class="fas fa-user-check"></i> {{ $t("Invite code") }}</a
                    >
                </div>
                <div v-if="!isCode" class="div-pt">
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
                    <div class="form-group">
                        <label>{{ $t("Session duration") }}:</label>
                        <select v-model="duration" class="form-control form-control-full-width form-select">
                            <option :value="'day'">1 {{ $t("day") }}</option>
                            <option :value="'week'">1 {{ $t("week") }} | 7 {{ $t("days") }}</option>
                            <option :value="'month'">1 {{ $t("month") }} | 30 {{ $t("days") }}</option>
                            <option :value="'year'">1 {{ $t("year") }} | 365 {{ $t("days") }}</option>
                        </select>
                    </div>
                </div>
                <div v-if="isCode" class="div-pt">
                    <div class="invite-code-label">{{ $t("Input your invite code below") }}</div>
                    <div class="invite-code-multi-input">
                        <div v-for="(c, i) in code" :key="i" class="invite-code-char-input">
                            <input
                                type="text"
                                :disabled="busy"
                                :class="'form-control auto-focus code-char-' + i"
                                @input="goNextChar(c, i)"
                                @paste="onPaste($event, i)"
                                v-model="c.c"
                                maxlength="1"
                            />
                        </div>
                    </div>
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
import { SessionDuration, apiAuthLogin } from "@/api/api-auth";
import { apiInvitesLogin } from "@/api/api-invites";
import { AuthController } from "@/control/auth";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";

export default defineComponent({
    name: "LoginModal",
    props: {
        display: Boolean,
    },
    setup() {
        return {
            timer: null,
        };
    },
    data: function () {
        return {
            username: "",
            password: "",
            duration: "day" as SessionDuration,

            code: [{ c: "" }, { c: "" }, { c: "" }, { c: "" }, { c: "" }, { c: "" }] as { c: string }[],

            isCode: false,

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
                    if (this.isCode) {
                        elem.select && elem.select();
                    }
                }
            });
        },

        submit: function (e) {
            e.preventDefault();

            if (this.isCode) {
                this.loginWithCode();
            } else {
                this.loginWithCredentials();
            }
        },

        loginWithCredentials: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;
            this.error = "";

            makeApiRequest(apiAuthLogin(this.username, this.password, this.duration))
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
                            this.autoFocus();
                        },
                        wrongCredentials: () => {
                            this.error = this.$t("Invalid username or password");
                            this.cooldown = Date.now() + 5000;
                            this.autoFocus();
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

        loginWithCode: function () {
            if (this.busy) {
                return;
            }

            this.busy = true;
            this.error = "";

            makeApiRequest(apiInvitesLogin(this.code.map((c) => c.c).join("")))
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
                        invalidCode: () => {
                            this.error = this.$t("Invalid invite code");
                            this.autoFocus();
                        },
                        wrongCode: () => {
                            this.error = this.$t("Invalid invite code");
                            this.cooldown = Date.now() + 5000;
                            this.autoFocus();
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

        changeToCode: function () {
            this.isCode = true;
            this.autoFocus();
        },

        changeToCredentials: function () {
            this.isCode = false;
            this.autoFocus();
        },

        goNextChar: function (c: { c: string }, i: number) {
            c.c = c.c.charAt(0).toUpperCase();

            if (!c.c) {
                // Go back
                if (i > 0) {
                    const nextInput = this.$el.querySelector(".code-char-" + (i - 1));

                    if (nextInput) {
                        nextInput.focus();
                        nextInput.select && nextInput.select();
                    }
                }

                return;
            }

            if (i < this.code.length - 1) {
                const nextInput = this.$el.querySelector(".code-char-" + (i + 1));

                if (nextInput) {
                    nextInput.focus();
                    nextInput.select && nextInput.select();
                }
            }
        },

        onPaste: function (ev: ClipboardEvent, i: number) {
            ev.preventDefault();

            const text = ev.clipboardData.getData("text/plain").replace(/[^a-z0-9]+/gi, "");

            let k = 0;
            for (let j = i; j < this.code.length; j++) {
                if (k >= text.length) {
                    break;
                }

                const c = text.charAt(k).toUpperCase();
                k++;

                this.code[j].c = c;
                this.goNextChar(this.code[j], j);
            }
        },
    },
    mounted: function () {
        this.autoFocus();

        this.timer = setInterval(this.updateNow.bind(this), 200);
    },
    watch: {
        display: function () {
            this.error = "";
            this.autoFocus();
        },
    },
    beforeUnmount: function () {
        clearInterval(this.timer);
    },
});
</script>
