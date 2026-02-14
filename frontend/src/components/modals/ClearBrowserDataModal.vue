<template>
    <ModalDialogContainer ref="container" v-model:display="display">
        <div class="modal-dialog modal-md" role="document">
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
                                <ToggleSwitch v-model:val="clearPlayerPreferencesOpt"></ToggleSwitch>
                            </td>

                            <td>
                                {{ $t("Clear player preferences (volume, resolution, subtitles, etc.)") }}
                            </td>
                        </tr>

                        <tr>
                            <td class="td-shrink">
                                <ToggleSwitch v-model:val="clearPagePreferencesOpt"></ToggleSwitch>
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

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { ref, useTemplateRef } from "vue";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { clearCachedAlbumPositions, clearCachedTimes, clearPlayerPreferences } from "@/control/player-preferences";
import {
    clearAlbumsOrderMap,
    clearFavAlbums,
    clearLastUsedTags,
    clearPagePreferences,
    clearSearchPreferences,
    clearUploadPreferences,
} from "@/control/app-preferences";
import { PagesController } from "@/control/pages";
import { clearLanguageSetting } from "@/i18n";
import { clearLocalStorage } from "@/utils/local-storage";
import { emitAppEvent, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close } = useModal(display, container);

// Options to clear data
const clearSavedTimestamps = ref(false);
const clearSavedAlbumPositions = ref(false);
const clearFavorites = ref(false);
const clearRecentlyAccessedAlbums = ref(false);
const clearRecentlyUsedTags = ref(false);
const clearPlayerPreferencesOpt = ref(false);
const clearPagePreferencesOpt = ref(false);
const clearResizableWidgets = ref(false);

/**
 * Submits the form
 */
const submit = () => {
    if (clearSavedTimestamps.value) {
        clearCachedTimes();
    }

    if (clearSavedAlbumPositions.value) {
        clearCachedAlbumPositions();
    }

    if (clearFavorites.value) {
        clearFavAlbums();
    }

    if (clearRecentlyAccessedAlbums.value) {
        clearAlbumsOrderMap();
        emitAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, AlbumsController.AlbumsMap);
    }

    if (clearRecentlyUsedTags.value) {
        clearLastUsedTags();
    }

    if (clearPlayerPreferencesOpt.value) {
        clearPlayerPreferences();
    }

    if (clearPagePreferencesOpt.value) {
        clearPagePreferences();
        clearLanguageSetting();
        clearUploadPreferences();
        clearSearchPreferences();
    }

    if (clearResizableWidgets.value) {
        ["desc-widget-pos", "tags-edit-helper-pos", "time-slices-helper-pos"].forEach(clearLocalStorage);
    }

    close();

    PagesController.ShowSnackBar($t("Successfully cleared browser data!"));
};
</script>
