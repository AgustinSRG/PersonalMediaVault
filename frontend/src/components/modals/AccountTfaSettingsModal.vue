<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal" @close="onClose">
        <form v-if="display" class="modal-dialog modal-lg" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Account security settings") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Issuer") }}:</label>
                    <input
                        v-model="issuerStatus"
                        type="text"
                        maxlength="100"
                        class="form-control form-control-full-width auto-focus"
                        @input="markDirty"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Account name") }}:</label>
                    <input
                        v-model="accountStatus"
                        type="text"
                        maxlength="100"
                        class="form-control form-control-full-width"
                        @input="markDirty"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Hashing algorithm") }}:</label>
                    <select v-model="algorithmStatus" class="form-control form-control-full-width form-select" @change="markDirty">
                        <option :value="'sha1'">SHA-1</option>
                        <option :value="'sha256'">SHA-256</option>
                        <option :value="'sha512'">SHA-512</option>
                    </select>
                </div>

                <div class="form-group">
                    <label>{{ $t("One-time password period") }}:</label>
                    <select v-model="periodStatus" class="form-control form-control-full-width form-select" @change="markDirty">
                        <option :value="'30'">{{ $t("30 seconds") }}</option>
                        <option :value="'60'">{{ $t("60 seconds") }}</option>
                        <option :value="'120'">{{ $t("120 seconds") }}</option>
                    </select>
                </div>

                <div class="form-group">
                    <table class="table no-margin no-border">
                        <tbody>
                            <tr>
                                <td class="text-right td-shrink no-padding">
                                    <ToggleSwitch v-model:val="skewStatus" @update:val="markDirty"></ToggleSwitch>
                                </td>
                                <td>
                                    {{ $t("Allow clock skew of one period") }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <div class="modal-footer no-padding">
                <button type="submit" :disabled="!dirty" class="modal-footer-btn"><i class="fas fa-check"></i> {{ $t("Done") }}</button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick, PropType } from "vue";
import { useVModel } from "../../utils/v-model";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { TimeOtpAlgorithm, TimeOtpPeriod } from "@/api/api-account";

export default defineComponent({
    name: "AccountTfaSettingsModal",
    components: {
        ToggleSwitch,
    },
    props: {
        display: Boolean,

        issuer: String,
        account: String,
        algorithm: String as PropType<TimeOtpAlgorithm>,
        period: String as PropType<TimeOtpPeriod>,
        skew: Boolean,
    },
    emits: ["update:display", "update:issuer", "update:account", "update:algorithm", "update:period", "update:skew", "done"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),

            issuerStatus: useVModel(props, "issuer"),
            accountStatus: useVModel(props, "account"),
            algorithmStatus: useVModel(props, "algorithm"),
            periodStatus: useVModel(props, "period"),
            skewStatus: useVModel(props, "skew"),
        };
    },
    data: function () {
        return {
            dirty: false,
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

        onClose: function () {
            if (this.dirty) {
                this.dirty = false;
                this.$emit("done");
            }
        },

        markDirty: function () {
            this.dirty = true;
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e?: Event) {
            if (e) {
                e.preventDefault();
            }

            this.dirty = false;

            this.$emit("done");

            this.close();
        },
    },
});
</script>
