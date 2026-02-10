<template>
    <ModalDialogContainer ref="container" v-model:display="display" @close="onClose">
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
                        v-model="issuer"
                        type="text"
                        maxlength="100"
                        class="form-control form-control-full-width auto-focus"
                        @input="markDirty"
                    />
                </div>

                <div class="form-group">
                    <label>{{ $t("Account name") }}:</label>
                    <input v-model="account" type="text" maxlength="100" class="form-control form-control-full-width" @input="markDirty" />
                </div>

                <div class="form-group">
                    <label>{{ $t("Hashing algorithm") }}:</label>
                    <select v-model="algorithm" class="form-control form-control-full-width form-select" @change="markDirty">
                        <option :value="'sha1'">SHA-1</option>
                        <option :value="'sha256'">SHA-256</option>
                        <option :value="'sha512'">SHA-512</option>
                    </select>
                </div>

                <div class="form-group">
                    <label>{{ $t("One-time password period") }}:</label>
                    <select v-model="period" class="form-control form-control-full-width form-select" @change="markDirty">
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
                                    <ToggleSwitch v-model:val="skew" @update:val="markDirty"></ToggleSwitch>
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

<script setup lang="ts">
import { ref, useTemplateRef } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import type { TimeOtpAlgorithm, TimeOtpPeriod } from "@/api/api-account";
import { useModal } from "@/composables/use-modal";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Events
const emit = defineEmits<{
    /**
     * Emitted when the request is done successfully.
     */
    (e: "done"): void;
}>();

// TFA issuer name
const issuer = defineModel<string>("issuer");

// TFA account name
const account = defineModel<string>("account");

// TFA algorithm
const algorithm = defineModel<TimeOtpAlgorithm>("algorithm");

// TOTP period
const period = defineModel<TimeOtpPeriod>("period");

// Allow clock skew?
const skew = defineModel<boolean>("skew");

// Changed made?
const dirty = ref(false);

/**
 * Indicates changes were made to the settings
 */
const markDirty = () => {
    dirty.value = true;
};

/**
 * Handles for the modal 'close' event
 */
const onClose = () => {
    if (!dirty.value) {
        return;
    }

    dirty.value = false;

    emit("done");
};

/**
 * Handler for the form 'submit' event
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    dirty.value = false;

    emit("done");

    close();
};
</script>
