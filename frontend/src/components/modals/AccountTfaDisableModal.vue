<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Disable two factor authentication") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body">
                <div class="tfa-label">{{ $t("Input your current one-time code for two factor authentication") }}</div>

                <SixDigitCodeInput v-model:val="code" :disabled="busy"></SixDigitCodeInput>

                <div class="form-error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button v-if="!busy && mustWait === 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait 1 second to try again") }}
                </button>
                <button v-else-if="!busy && mustWait > 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait $TIME seconds to try again").replace("$TIME", mustWait + "") }}
                </button>
                <button v-else type="submit" class="modal-footer-btn" :disabled="busy">
                    <LoadingIcon icon="fas fa-trash-alt" :loading="busy"></LoadingIcon> {{ $t("Disable two factor authentication") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import { apiAccountTfaDisable } from "@/api/api-account";

export default defineComponent({
    name: "AccountTfaDisableModal",
    components: {
        LoadingIcon,
        SixDigitCodeInput,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display", "done"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            timer: null as ReturnType<typeof setInterval> | null,
            saveRequestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            code: "",

            busy: false,

            error: "",

            cooldown: 0,
            mustWait: 0,
            now: Date.now(),

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        this.timer = setInterval(this.updateNow.bind(this), 200);

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        abortNamedApiRequest(this.saveRequestId);
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

        close: function () {
            this.closeSignal++;
        },

        updateNow: function () {
            this.now = Date.now();
            if (this.now < this.cooldown) {
                this.mustWait = Math.max(1, Math.round((this.cooldown - this.now) / 1000));
            } else {
                this.mustWait = 0;
            }
        },

        submit: function (e?: Event) {
            if (e) {
                e.preventDefault();
            }

            if (this.busy) {
                return;
            }

            this.busy = true;

            makeApiRequest(apiAccountTfaDisable(this.code))
                .onSuccess(() => {
                    this.busy = false;
                    this.error = "";
                    PagesController.ShowSnackBar(this.$t("Two factor authentication disabled"));
                    this.$emit("done");
                    this.forceCloseSignal++;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    this.error = "";
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        tfaNotEnabled: () => {
                            PagesController.ShowSnackBar(this.$t("Two factor authentication is not enabled"));
                            this.$emit("done");
                            this.forceCloseSignal++;
                        },
                        invalidCode: () => {
                            this.error = this.$t("Invalid one-time code");
                            this.code = "";
                            this.cooldown = Date.now() + 5000;
                            this.autoFocus();
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
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
                    this.busy = false;
                    console.error(err);
                    this.error = err.message;
                });
        },
    },
});
</script>
