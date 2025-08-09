<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal">
        <div v-if="display" class="modal-dialog modal-md" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Clear browser data") }}</div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body">
                <p>{{ $t("You can use this option to remove data stored in your browser while using the application") }}.</p>

                <table class="table no-margin table-invisible">
                    <tbody>
                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearSavedTimestamps"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear saved media current times (used to resume playing media content)") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearSavedAlbumPositions"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear saved album current positions (used to resume navigating an album)") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearFavorites"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear list of favorite albums") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearRecentlyAccessedAlbums"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear list recently accessed albums") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearRecentlyUsedTags"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear list of recently used tags") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearPlayerPreferences"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear player preferences (volume, resolution, subtitles, etc.)") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearPagePreferences"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear page preferences (theme, language and page settings)") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearResizableWidgets"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear position and size of player resizable widgets") }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div class="modal-footer no-padding">
                <button
                    type="button"
                    class="modal-footer-btn auto-focus"
                    :disabled="
                        !clearSavedTimestamps &&
                        !clearSavedAlbumPositions &&
                        !clearFavorites &&
                        !clearPlayerPreferences &&
                        !clearPagePreferences
                    "
                    @click="submit"
                >
                    <i class="fas fa-broom"></i> {{ $t("Clear browser data") }}
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { clearCachedAlbumPositions, clearCachedTimes, clearPlayerPreferences } from "@/control/player-preferences";
import { clearAlbumsOrderMap, clearFavAlbums, clearLastUsedTags, clearPagePreferences } from "@/control/app-preferences";
import { PagesController } from "@/control/pages";
import { clearLanguageSetting } from "@/i18n";
import { clearLocalStorage } from "@/utils/local-storage";
import { AppEvents } from "@/control/app-events";
import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";

export default defineComponent({
    name: "ClearBrowserDataModal",
    components: { ToggleSwitch },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            closeSignal: 0,

            clearSavedTimestamps: false,
            clearSavedAlbumPositions: false,
            clearFavorites: false,
            clearRecentlyAccessedAlbums: false,
            clearRecentlyUsedTags: false,
            clearPlayerPreferences: false,
            clearPagePreferences: false,
            clearResizableWidgets: false,
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
        close: function () {
            this.closeSignal++;
        },

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

        submit: function () {
            if (this.clearSavedTimestamps) {
                clearCachedTimes();
            }

            if (this.clearSavedAlbumPositions) {
                clearCachedAlbumPositions();
            }

            if (this.clearFavorites) {
                clearFavAlbums();
            }

            if (this.clearRecentlyAccessedAlbums) {
                clearAlbumsOrderMap();
                AppEvents.Emit(EVENT_NAME_ALBUMS_LIST_UPDATE, AlbumsController.AlbumsMap);
            }

            if (this.clearRecentlyUsedTags) {
                clearLastUsedTags();
            }

            if (this.clearPlayerPreferences) {
                clearPlayerPreferences();
            }

            if (this.clearPagePreferences) {
                clearPagePreferences();
                clearLanguageSetting();
            }

            if (this.clearResizableWidgets) {
                ["desc-widget-pos", "tags-edit-helper-pos", "time-slices-helper-pos"].forEach(clearLocalStorage);
            }

            this.close();

            PagesController.ShowSnackBar(this.$t("Successfully cleared browser data!"));
        },
    },
});
</script>
