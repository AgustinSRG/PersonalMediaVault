<template>
  <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" :static="true">
    <div v-if="display" class="modal-dialog modal-xl modal-height-100" role="document">
      <div class="modal-header">
        <div class="modal-title" v-if="!isUpload">
          {{ $t("Search media to add to the album") }}
        </div>
        <div class="modal-title" v-if="isUpload">
          {{ $t("Upload media to add to the album") }}
        </div>
        <button class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body no-padding">
        <div class="modal-top-menu items-2">
          <a href="javascript:;" @click="changeToSearch" class="modal-top-menu-item" :class="{ selected: !isUpload }"><i class="fas fa-search"></i> {{ $t("Search") }}</a>
          <a href="javascript:;" @click="changeToUpload" class="modal-top-menu-item" :class="{ selected: isUpload }"><i class="fas fa-upload"></i> {{ $t("Upload") }}</a>
        </div>
        <PageAdvancedSearch v-if="!isUpload" :display="true" :inModal="true" :noAlbum="aid" @select-media="selectMedia"></PageAdvancedSearch>
        <PageUpload v-if="isUpload" :display="true" :inModal="true" :fixedAlbum="aid" @media-go="close"></PageUpload>
      </div>
    </div>
  </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import PageAdvancedSearch from "@/components/pages/PageAdvancedSearch.vue";
import PageUpload from "@/components/pages/PageUpload.vue";
import { Request } from "@/utils/request";
import { AlbumsAPI } from "@/api/api-albums";
import { AppEvents } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";

export default defineComponent({
    components: {
        PageAdvancedSearch,
        PageUpload,
    },
    name: "AlbumAddMediaModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
        aid: Number,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            busy: false,

            isUpload: false,
        };
    },
    methods: {
        close: function () {
            this.$refs.modalContainer.close();
        },

        changeToUpload: function () {
            this.isUpload = true;
        },

        changeToSearch: function () {
            this.isUpload = false;
        },

        selectMedia: function (mid, callback) {
            if (this.busy) {
                return;
            }
            const albumId = this.aid;
            this.busy = true;
            // Add
            Request.Do(AlbumsAPI.AddMediaToAlbum(albumId, mid))
                .onSuccess(() => {
                    this.busy = false;
                    AppEvents.Emit("snack", this.$t("Successfully added to album"));
                    AlbumsController.OnChangedAlbum(albumId, true);
                    callback();
                })
                .onRequestError((err) => {
                    this.busy = false;
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized");
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    this.busy = false;
                    console.error(err);
                });
        },
    },
    mounted: function () {
        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
        }
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
});
</script>
