<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div v-if="display" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Administrate accounts") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body no-padding">
                <div class="table-responsive">
                    <table class="table no-margin">
                        <thead>
                            <tr>
                                <th class="text-left">{{ $t("Username") }}</th>
                                <th class="text-left">{{ $t("Account type") }}</th>
                                <th class="text-right td-shrink"></th>
                                <th class="text-right td-shrink"></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="a in accounts" :key="a.username" class="tr-align-middle">
                                <td class="bold">{{ a.username }}</td>
                                <td v-if="!a.write">{{ $t("Read only") }}</td>
                                <td v-if="a.write">{{ $t("Read / Write") }}</td>
                                <td class="text-right">
                                    <button type="button" class="btn btn-primary btn-xs" @click="modifyAccount(a)">
                                        <i class="fas fa-pencil-alt"></i> {{ $t("Edit") }}
                                    </button>
                                </td>
                                <td class="text-right">
                                    <button type="button" class="btn btn-danger btn-xs" @click="deleteAccount(a)">
                                        <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                                    </button>
                                </td>
                            </tr>
                            <tr class="tr-align-middle">
                                <td colspan="4" class="text-right">
                                    <button type="button" class="btn btn-primary btn-xs" @click="createAccount">
                                        <i class="fas fa-plus"></i> {{ $t("Create account") }}
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <AccountModifyModal
            v-if="displayAccountModify"
            v-model:display="displayAccountModify"
            :username="selectedAccount?.username"
            :write="selectedAccount?.write"
            @update:display="afterSubModalClosed"
            @done="load"
        ></AccountModifyModal>

        <AccountDeleteModal
            v-if="displayAccountDelete"
            v-model:display="displayAccountDelete"
            :username="selectedAccount?.username"
            @update:display="afterSubModalClosed"
            @done="load"
        ></AccountDeleteModal>

        <AccountCreateModal
            v-if="displayAccountCreate"
            v-model:display="displayAccountCreate"
            @account-created="load"
            @update:display="afterSubModalClosed"
        ></AccountCreateModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineAsyncComponent, onMounted, ref, useTemplateRef, watch } from "vue";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import type { VaultAccount } from "@/api/api-admin";
import { apiAdminListAccounts } from "@/api/api-admin";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useRequestId } from "@/composables/use-request-id";

const AccountCreateModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountCreateModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountModifyModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountModifyModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AccountDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AccountDeleteModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, focus } = useModal(display, container);

// List of accounts
const accounts = ref<VaultAccount[]>([]);

// Loading status
const loading = ref(true);

// Request ID
const loadRequestId = useRequestId();

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads account list
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    loading.value = true;

    makeNamedApiRequest(loadRequestId, apiAdminListAccounts())
        .onSuccess((res) => {
            accounts.value = res;

            loading.value = false;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    close();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
        });
};

onMounted(() => {
    if (display.value) {
        load();
    }
});

// Selected account
const selectedAccount = ref<VaultAccount | null>(null);

// Create account modal
const displayAccountCreate = ref(false);

// Modify account modal
const displayAccountModify = ref(false);

// Delete account modal
const displayAccountDelete = ref(false);

watch(display, () => {
    if (display.value) {
        displayAccountCreate.value = false;
        displayAccountModify.value = false;
        displayAccountDelete.value = false;

        load();
    }
});

/**
 * Opens the modal to create an account
 */
const createAccount = () => {
    displayAccountCreate.value = true;
};

/**
 * Opens the modal to modify an account
 * @param a The account
 */
const modifyAccount = (a: VaultAccount) => {
    selectedAccount.value = a;
    displayAccountModify.value = true;
};

/**
 * Opens the modal to delete an account
 * @param a The account
 */
const deleteAccount = (a: VaultAccount) => {
    selectedAccount.value = a;
    displayAccountDelete.value = true;
};

/**
 * Called when a sub-modal display status changes
 * @param subModalDisplay Sub-modal display status
 */
const afterSubModalClosed = function (subModalDisplay: boolean) {
    if (!subModalDisplay && display.value) {
        focus();
    }
};
</script>
