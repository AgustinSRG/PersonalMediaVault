<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Authentication confirmation") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body">
                <div v-if="tfa" class="tfa-label">{{ $t("Input your current one-time code for two factor authentication") }}</div>
                <SixDigitCodeInput v-if="tfa" v-model:val="code"></SixDigitCodeInput>

                <div v-if="!tfa" class="form-group">
                    <label>{{ $t("To confirm the operation, type your account password") }}:</label>
                    <PasswordInput v-model:val="password" :name="'password'" :auto-focus="true" @tab-skip="passwordTabSkip"></PasswordInput>
                </div>

                <div class="form-error" v-if="error">{{ error }}</div>
            </div>

            <div class="modal-footer no-padding">
                <button v-if="mustWait === 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait 1 second to try again") }}
                </button>
                <button v-else-if="mustWait > 1" type="button" disabled class="modal-footer-btn">
                    <i class="fas fa-hourglass"></i>
                    {{ $t("You must wait $TIME seconds to try again").replace("$TIME", mustWait + "") }}
                </button>
                <button v-else type="submit" class="modal-footer-btn" :disabled="tfa ? !code : !password">
                    <i class="fas fa-check"></i> {{ $t("Confirm") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import SixDigitCodeInput from "../utils/SixDigitCodeInput.vue";
import PasswordInput from "../utils/PasswordInput.vue";
import { ProvidedAuthConfirmation } from "@/api/api-auth";

export default defineComponent({
    name: "AuthConfirmationModal",
    components: {
        SixDigitCodeInput,
        PasswordInput,
    },
    props: {
        display: Boolean,
        error: String,
        tfa: Boolean,
        cooldown: Number,
    },
    emits: ["update:display", "confirm"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
            timer: null as ReturnType<typeof setInterval> | null,
        };
    },
    data: function () {
        return {
            code: "",
            password: "",

            mustWait: 0,
            now: Date.now(),

            closeSignal: 0,
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

        passwordTabSkip: function (e: KeyboardEvent) {
            const nextElement = this.$el.querySelector(".modal-footer-btn");

            if (nextElement) {
                e.preventDefault();
                nextElement.focus();
            }
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

            const providedAuthConfirmation: ProvidedAuthConfirmation = {
                password: this.password,
                tfaCode: this.code,
            };

            this.$emit("confirm", providedAuthConfirmation);
            this.displayStatus = false;
        },
    },
});
</script>
