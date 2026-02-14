<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <form class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Add new row") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>{{ $t("Row name") }}:</label>
                    <input
                        v-model="name"
                        type="text"
                        name="row-name"
                        autocomplete="off"
                        :disabled="busy"
                        :placeholder="getDefaultGroupName(rowType, $t)"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />

                    <div v-if="nameError" class="form-error form-error-pt">{{ nameError }}</div>
                </div>

                <div class="form-group">
                    <label>{{ $t("Row type") }}:</label>
                    <select v-model="rowType" class="form-control form-control-full-width form-select" :disabled="busy">
                        <option v-for="t in ROW_TYPES" :key="t" :value="t">{{ getDefaultGroupName(t, $t) }}</option>
                    </select>

                    <div v-if="rowTypeError" class="form-error form-error-pt">{{ rowTypeError }}</div>
                </div>

                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-plus" :loading="busy"></LoadingIcon> {{ $t("Add new row") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { makeApiRequest } from "@asanrom/request-browser";
import { ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import type { HomePageGroup } from "@/api/api-home";
import { apiHomeAddGroup } from "@/api/api-home";
import { getDefaultGroupName, HomePageGroupTypes } from "@/utils/home";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * True to prepend the row, false to append it.
     */
    prepend: Boolean,
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when the new row is added
     */
    (e: "new-row", group: HomePageGroup, prepend: boolean): void;
}>();

// Home page tow types
const ROW_TYPES = [HomePageGroupTypes.Custom, HomePageGroupTypes.RecentMedia, HomePageGroupTypes.RecentAlbums];

// Row name
const name = ref("");

// Row type
const rowType = ref(HomePageGroupTypes.Custom);

// Busy (request in progress)
const busy = ref(false);

// Request error
const { error, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

// Other errors
const nameError = ref("");
const rowTypeError = ref("");

// Resets the error messages
const resetErrors = () => {
    error.value = "";

    nameError.value = "";
    rowTypeError.value = "";
};

// Reset error when modal opens
watch(display, () => {
    if (display.value) {
        resetErrors();
    }
});

/**
 * Event handler for 'submit'
 * @param e The event
 */
const submit = (e: Event) => {
    e.preventDefault();

    if (busy.value) {
        return;
    }

    resetErrors();

    busy.value = true;

    const prepend = props.prepend;

    makeApiRequest(
        apiHomeAddGroup({
            name: name.value,
            type: rowType.value,
            prepend,
        }),
    )
        .onSuccess((response) => {
            PagesController.ShowSnackBar($t("Row added") + ": " + (response.name || getDefaultGroupName(response.type, $t)));

            busy.value = false;
            name.value = "";

            forceClose();

            emit("new-row", response, prepend);
        })
        .onCancel(() => {
            busy.value = false;
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                invalidName: () => {
                    nameError.value = $t("Invalid row name provided");
                },
                invalidGroupType: () => {
                    rowTypeError.value = $t("Invalid row type provided");
                },
                tooManyGroups: () => {
                    error.value = $t("There are already too many rows in the home page");
                },
                accessDenied,
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;

            error.value = err.message;

            console.error(err);
        });
};
</script>
