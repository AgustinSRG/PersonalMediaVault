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
                </div>

                <div class="form-group">
                    <label>{{ $t("Row type") }}:</label>
                    <select v-model="rowType" class="form-control form-control-full-width form-select" :disabled="busy">
                        <option v-for="t in rowTypes" :key="t" :value="t">{{ getDefaultGroupName(t, $t) }}</option>
                    </select>
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

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { apiHomeAddGroup, HomePageGroupTypes } from "@/api/api-home";
import { getDefaultGroupName } from "@/utils/home";

export default defineComponent({
    name: "HomePageCreateRowModal",
    components: {
        LoadingIcon,
    },
    props: {
        display: Boolean,

        prepend: Boolean,
    },
    emits: ["update:display", "new-row"],
    setup(props) {
        return {
            rowTypes: [HomePageGroupTypes.Custom, HomePageGroupTypes.RecentMedia, HomePageGroupTypes.RecentAlbums],

            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            name: "",
            rowType: HomePageGroupTypes.Custom,

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
                this.autoFocus();
            }
        },
    },
    mounted: function () {
        if (this.display) {
            this.error = "";
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
                if (elem) {
                    elem.focus();
                }
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

            this.busy = true;
            this.error = "";

            const prepend = this.prepend;

            makeApiRequest(
                apiHomeAddGroup({
                    name: this.name,
                    type: this.rowType,
                    prepend,
                }),
            )
                .onSuccess((response) => {
                    PagesController.ShowSnackBar(
                        this.$t("Row added") + ": " + (response.name || this.getDefaultGroupName(response.type, this.$t)),
                    );
                    this.busy = false;
                    this.name = "";
                    this.forceCloseSignal++;
                    this.$emit("new-row", response, prepend);
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
                            this.error = this.$t("Invalid row name provided");
                        },
                        invalidGroupType: () => {
                            this.error = this.$t("Invalid row type provided");
                        },
                        tooManyGroups: () => {
                            this.error = this.$t("There are already too many rows in the home page");
                        },
                        accessDenied: () => {
                            this.error = this.$t("Access denied");
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
