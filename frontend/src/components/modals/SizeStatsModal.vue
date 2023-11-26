<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
        <div v-if="display" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Size Statistics") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div v-if="loading" class="modal-body">
                <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
            </div>
            <div v-if="!loading" class="modal-body no-padding table-responsive">
                <table class="table table-text-overflow">
                    <thead>
                        <tr>
                            <th class="text-left">{{ $t("Asset") }}</th>
                            <th class="text-left">{{ $t("Size") }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>METADATA</td>
                            <td>{{ renderSize(metaSize) }}</td>
                        </tr>
                        <tr v-for="a in assets" :key="a.key">
                            <td>{{ a.name }}</td>
                            <td>{{ renderSize(a.size) }}</td>
                        </tr>
                        <tr>
                            <td class="bold">{{ $t("Total") }}</td>
                            <td class="bold">{{ renderSize(total) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { AppEvents } from "@/control/app-events";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiMediaGetMediaSizeStats } from "@/api/api-media";

export default defineComponent({
    name: "SizeStatsModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
        mid: Number,
    },
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            loading: false,
            metaSize: 0,
            assets: [],
            total: 0,
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

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiMediaGetMediaSizeStats(this.mid))
                .onSuccess((result) => {
                    this.loading = false;
                    this.metaSize = result.meta_size;
                    this.assets = result.assets;

                    let total = 0;

                    total += result.meta_size;

                    for (const asset of result.assets) {
                        total += asset.size;
                    }

                    this.total = total;
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        notFound: () => {
                            this.close();
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

        renderSize: function (bytes: number): string {
            if (bytes > 1024 * 1024 * 1024) {
                let gb = bytes / (1024 * 1024 * 1024);
                gb = Math.floor(gb * 100) / 100;
                return gb + " GB";
            } else if (bytes > 1024 * 1024) {
                let mb = bytes / (1024 * 1024);
                mb = Math.floor(mb * 100) / 100;
                return mb + " MB";
            } else if (bytes > 1024) {
                let kb = bytes / 1024;
                kb = Math.floor(kb * 100) / 100;
                return kb + " KB";
            } else {
                return bytes + " Bytes";
            }
        },

        close: function () {
            this.$refs.modalContainer.close();
        },
    },
    mounted: function () {
        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
            this.load();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
                this.load();
            }
        },

        mid: function () {
            this.load();
        },
    },
});
</script>
