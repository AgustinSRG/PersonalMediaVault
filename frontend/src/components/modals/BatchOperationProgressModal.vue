<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :static="true"
        :lock-close="status === 'search' || status === 'action'"
    >
        <div v-if="display" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div v-if="status === 'search'" class="modal-title">{{ $t("Searching") }}...</div>
                <div v-if="status === 'action'" class="modal-title">{{ $t("Applying batch action") }}...</div>
                <div v-if="status === 'confirmation' || status === 'confirmation-delete'" class="modal-title">
                    {{ $t("Confirmation") }}
                </div>
                <div v-if="status === 'error'" class="modal-title">
                    {{ $t("Error") }}
                </div>
                <div v-if="status === 'success'" class="modal-title">
                    {{ $t("Success") }}
                </div>
                <button
                    v-if="status === 'search' || status === 'action'"
                    type="button"
                    class="modal-close-btn"
                    :title="$t('Close')"
                    @click="cancel"
                >
                    <i class="fas fa-times"></i>
                </button>
                <button
                    v-if="status === 'confirmation' || status === 'confirmation-delete' || status === 'error' || status === 'success'"
                    type="button"
                    class="modal-close-btn"
                    :title="$t('Close')"
                    @click="close"
                >
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="status === 'search' || status === 'action'" class="modal-body">
                <div class="batch-progress-bar">
                    <div class="batch-progress-bar-current" :style="{ width: cssProgress(progress) }"></div>
                    <div class="batch-progress-bar-text">{{ renderStatus(status, progress) }}</div>
                </div>
            </div>
            <div v-if="status === 'confirmation'" class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to update $N elements?").replace("$N", "" + actionCount) }}</label>
                </div>
            </div>
            <div v-if="status === 'confirmation-delete'" class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Do you want to delete $N elements?").replace("$N", "" + actionCount) }}</label>
                </div>

                <table class="table no-border">
                    <tbody>
                        <tr>
                            <td class="text-right td-shrink no-padding">
                                <ToggleSwitch v-model:val="confirmationDelete"></ToggleSwitch>
                            </td>
                            <td>
                                {{ $t("Remember. If you delete the media by accident you would have to re-upload it.") }}
                                <br />
                                {{ $t("Make sure you actually want to delete it.") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div v-if="status === 'error'" class="modal-body">
                <div class="form-group">
                    <label>{{ error }}</label>
                </div>
            </div>
            <div v-if="status === 'success'" class="modal-body">
                <div class="form-group">
                    <label>{{ $t("The batch operation was completed successfully.") }}</label>
                </div>
            </div>
            <div v-if="status === 'confirmation'" class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="confirm">
                    <i class="fas fa-check"></i> {{ $t("Continue") }}
                </button>
            </div>
            <div v-if="status === 'confirmation-delete'" class="modal-footer no-padding">
                <button type="button" :disabled="!confirmationDelete" class="modal-footer-btn auto-focus" @click="confirm">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                </button>
            </div>
            <div v-if="status === 'error'" class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="close">
                    <i class="fas fa-times"></i> {{ $t("Close") }}
                </button>
            </div>
            <div v-if="status === 'success'" class="modal-footer no-padding">
                <button type="button" class="modal-footer-btn auto-focus" @click="close">
                    <i class="fas fa-check"></i> {{ $t("Close") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import ToggleSwitch from "../utils/ToggleSwitch.vue";

export default defineComponent({
    name: "BatchOperationProgressModal",
    components: {
        ToggleSwitch,
    },
    props: {
        display: Boolean,

        status: String,
        progress: Number,

        actionCount: Number,

        error: String,
    },
    emits: ["update:display", "confirm", "cancel"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            confirmationDelete: false,

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        cancel: function () {
            this.$emit("cancel");
            this.forceCloseSignal++;
        },

        confirm: function () {
            this.$emit("confirm");
        },

        cssProgress: function (p: number) {
            return Math.round(p) + "%";
        },

        renderStatus(status: string, p: number) {
            const renderP = Math.round(p * 100) / 100;
            switch (status) {
                case "search":
                    if (p > 0) {
                        return this.$t("Searching") + "... (" + renderP + "%)";
                    } else {
                        return this.$t("Searching") + "...";
                    }
                case "action":
                    if (p > 0) {
                        return this.$t("Applying") + "... (" + renderP + "%)";
                    } else {
                        return this.$t("Applying") + "...";
                    }
                default:
                    return "-";
            }
        },
    },
});
</script>
