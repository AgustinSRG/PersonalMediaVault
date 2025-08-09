<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <form v-if="display" class="modal-dialog modal-md" role="document" @submit="submit">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Rename row") }}
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
                        :placeholder="getDefaultGroupName(selectedRowType, $t)"
                        maxlength="255"
                        class="form-control form-control-full-width auto-focus"
                    />
                </div>
                <div class="form-error">{{ error }}</div>
            </div>
            <div class="modal-footer no-padding">
                <button :disabled="busy" type="submit" class="modal-footer-btn">
                    <LoadingIcon icon="fas fa-pencil-alt" :loading="busy"></LoadingIcon> {{ $t("Rename row") }}
                </button>
            </div>
        </form>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { getDefaultGroupName } from "@/utils/home";
import { apiHomeGroupRename } from "@/api/api-home";

export default defineComponent({
    name: "HomePageRenameRowModal",
    components: {
        LoadingIcon,
    },
    props: {
        display: Boolean,

        selectedRowType: Number,
        selectedRow: Number,
        selectedRowName: String,
    },
    emits: ["update:display", "renamed", "must-reload"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            name: "",

            busy: false,
            error: "",

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.error = "";
                this.name = this.selectedRowName;
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
            this.name = this.selectedRowName;
            this.autoFocus();
        }
    },
    methods: {
        getDefaultGroupName: getDefaultGroupName,

        autoFocus: function () {
            if (!this.display) {
                return;
            }
            nextTick(() => {
                const elem = this.$el.querySelector(".auto-focus");
                elem.focus();
                elem.select();
            });
        },

        close: function () {
            this.closeSignal++;
        },

        submit: function (e: Event) {
            e.preventDefault();

            if (this.busy) {
                return;
            }

            if (this.name === this.selectedRowName) {
                this.forceCloseSignal++;
                return;
            }

            this.busy = true;
            this.error = "";

            const rowId = this.selectedRow;
            const newName = this.name;

            makeApiRequest(apiHomeGroupRename(rowId, newName))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(
                        this.$t("Row renamed") + ": " + (newName || this.getDefaultGroupName(this.selectedRowType, this.$t)),
                    );
                    this.busy = false;
                    this.name = "";
                    this.forceCloseSignal++;
                    this.$emit("renamed", rowId, newName);
                })
                .onCancel(() => {
                    this.busy = false;
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            this.error = this.$t("Access denied");
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        invalidName: () => {
                            this.error = this.$t("Invalid album name provided");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
                        },
                        notFound: () => {
                            this.forceCloseSignal++;
                            this.$emit("must-reload");
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
    },
});
</script>
