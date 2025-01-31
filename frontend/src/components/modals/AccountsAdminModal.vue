<template>
    <ModalDialogContainer :closeSignal="closeSignal" v-model:display="displayStatus">
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
                    <table class="table">
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
                                    <button type="button" class="btn btn-primary btn-xs" @click="openEditAccount(a)">
                                        <i class="fas fa-pencil-alt"></i> {{ $t("Edit") }}
                                    </button>
                                </td>
                                <td class="text-right">
                                    <button type="button" class="btn btn-danger btn-xs" @click="askDeleteAccount(a.username)">
                                        <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                                    </button>
                                </td>
                            </tr>
                            <tr class="tr-align-middle">
                                <td colspan="4" class="text-right">
                                    <button type="button" @click="createAccount" :disabled="busy" class="btn btn-primary btn-xs">
                                        <i class="fas fa-plus"></i> {{ $t("Create account") }}
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <div class="modal-footer no-padding">
                <button type="button" @click="close" :disabled="busy" class="modal-footer-btn">
                    <i class="fas fa-check"></i> {{ $t("Done") }}
                </button>
            </div>
        </div>

        <AccountModifyModal
            v-if="displayAccountModify"
            v-model:display="displayAccountModify"
            @update:display="afterSubModalClosed"
            :username="accountModifyUsername"
            :write="accountModifyWrite"
            @done="load"
        ></AccountModifyModal>

        <AccountDeleteModal
            v-if="displayAccountDelete"
            v-model:display="displayAccountDelete"
            @update:display="afterSubModalClosed"
            :username="accountToDelete"
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

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import AccountModifyModal from "../modals/AccountModifyModal.vue";
import AccountDeleteModal from "../modals/AccountDeleteModal.vue";
import AccountCreateModal from "../modals/AccountCreateModal.vue";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiAdminListAccounts, VaultAccount } from "@/api/api-admin";

export default defineComponent({
    components: {
        AccountDeleteModal,
        AccountModifyModal,
        AccountCreateModal,
    },
    name: "AccountsAdminModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            accounts: [] as VaultAccount[],

            accountUsername: "",
            accountPassword: "",
            accountPassword2: "",
            accountWrite: false,

            displayAccountModify: false,
            accountModifyUsername: "",
            accountModifyWrite: false,

            displayAccountDelete: false,
            accountToDelete: "",

            displayAccountCreate: false,

            loading: true,
            busy: false,
            error: "",

            closeSignal: 0,
        };
    },
    methods: {
        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            this.loading = true;

            makeNamedApiRequest(this.loadRequestId, apiAdminListAccounts())
                .onSuccess((accounts) => {
                    this.accounts = accounts;
                    this.loading = false;
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            this.displayStatus = false;
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        openEditAccount: function (a: VaultAccount) {
            this.accountModifyUsername = a.username;
            this.accountModifyWrite = a.write;
            this.displayAccountModify = true;
        },

        askDeleteAccount: function (username: string) {
            this.accountToDelete = username;
            this.displayAccountDelete = true;
        },

        createAccount: function () {
            this.displayAccountCreate = true;
        },

        close: function () {
            this.displayAccountDelete = false;
            this.displayAccountCreate = false;
            this.closeSignal++;
        },

        afterSubModalClosed: function (display: boolean) {
            if (!display && this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
    },
    mounted: function () {
        this.load();

        if (this.display) {
            this.error = "";
            nextTick(() => {
                this.$el.focus();
            });
            this.displayAccountDelete = false;
            this.displayAccountCreate = false;
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                nextTick(() => {
                    this.$el.focus();
                });
                this.displayAccountDelete = false;
                this.displayAccountCreate = false;
                this.load();
            }
        },
    },
});
</script>
