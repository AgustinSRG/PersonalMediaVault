<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy">
        <div class="modal-dialog modal-sm" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Albums") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div v-if="loading" class="modal-body"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</div>

            <div v-if="!loading" class="modal-body with-menu">
                <div v-if="!canWrite" class="albums-modal-filter">
                    <input
                        v-model="filter"
                        type="text"
                        autocomplete="off"
                        :disabled="busy"
                        class="form-control form-control-full-width auto-focus"
                        :placeholder="$t('Filter by name') + '...'"
                        @keydown="onFilterKeyDown"
                        @input="updateAlbums"
                    />
                </div>
                <div v-if="canWrite" class="albums-modal-filter with-edit-mode">
                    <input
                        v-model="filter"
                        type="text"
                        autocomplete="off"
                        :disabled="busy"
                        class="form-control auto-focus"
                        :placeholder="$t('Filter by name') + '...'"
                        @keydown="onFilterKeyDown"
                        @input="updateAlbums"
                    />
                    <button
                        v-if="!editMode"
                        type="button"
                        :disabled="busy"
                        class="album-edit-mode-btn"
                        :title="$t('Change to edit mode')"
                        @click="changeEditMode"
                    >
                        <i class="fas fa-pencil-alt"></i>
                    </button>
                    <button
                        v-if="editMode"
                        type="button"
                        :disabled="busy"
                        class="album-edit-mode-btn"
                        :title="$t('Change to view mode')"
                        @click="changeEditMode"
                    >
                        <i class="fas fa-eye"></i>
                    </button>
                </div>
                <div class="albums-list-table-container">
                    <div v-if="editMode" class="albums-modal-menu" @scroll="onScroll">
                        <tr v-if="albums.length === 0">
                            <td colspan="2" class="albums-menu-empty">
                                {{ $t("No albums found") }}
                            </td>
                        </tr>
                        <a
                            v-for="a in albums"
                            :key="a.id"
                            :href="getAlbumURL(a.id, mid)"
                            :title="a.name"
                            class="albums-modal-menu-item"
                            :class="{ disabled: busy }"
                            @click="clickOnAlbum(a, false, $event)"
                            @keydown="clickOnEnter"
                        >
                            <div class="albums-modal-menu-item-icon">
                                <LoadingIcon
                                    :icon="a.added ? 'far fa-square-check' : 'far fa-square'"
                                    :loading="busy && busyTarget === a.id"
                                ></LoadingIcon>
                            </div>
                            <div class="albums-modal-menu-item-title">
                                {{ a.name }}
                            </div>
                        </a>
                    </div>
                    <div v-if="!editMode" class="albums-modal-menu" @scroll="onScroll">
                        <div v-if="albums.length === 0">
                            <div class="albums-menu-empty">
                                {{ $t("No albums found") }}
                            </div>
                        </div>
                        <a
                            v-for="a in albums"
                            :key="a.id"
                            :href="getAlbumURL(a.id, mid)"
                            :title="a.name"
                            class="albums-modal-menu-item"
                            @click="goToAlbum(a, $event)"
                            @keydown="clickOnEnter"
                        >
                            <div class="albums-modal-menu-item-icon">
                                <i class="fas fa-list-ol"></i>
                            </div>
                            <div class="albums-modal-menu-item-title">
                                {{ a.name }}
                            </div>
                        </a>
                    </div>
                </div>
            </div>

            <div v-if="!loading && editMode" class="modal-footer no-padding">
                <button type="button" :disabled="busy" class="modal-footer-btn" @click="createAlbum">
                    <i class="fas fa-plus"></i> {{ $t("Create album") }}
                </button>
            </div>
        </div>
        <AlbumCreateModal
            v-model:display="displayAlbumCreate"
            @new-album="onNewAlbum"
            @update:display="afterModalCreateClosed"
        ></AlbumCreateModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import { AlbumsController } from "@/control/albums";
import { emitAppEvent, EVENT_NAME_ALBUMS_LIST_UPDATE, EVENT_NAME_APP_STATUS_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineAsyncComponent, onMounted, ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import { apiAlbumsAddMediaToAlbum, apiAlbumsRemoveMediaFromAlbum } from "@/api/api-albums";
import { apiMediaGetMediaAlbums } from "@/api/api-media";
import { getFrontendUrl } from "@/utils/api";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { BigListScroller } from "@/utils/big-list-scroller";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useRequestId } from "@/composables/use-request-id";

const AlbumCreateModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumCreateModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

/**
 * Album metadata to use for the list
 */
type AlbumModalListItem = {
    id: number;
    name: string;
    nameLowerCase: string;
    added: boolean;
    starts: boolean;
    contains: boolean;
};

// User permissions
const { canWrite } = useUserPermissions();

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose, focus } = useModal(display, container);

// Edit mode
const editMode = ref(canWrite.value);

// True if the edit mode was changed by the user
const editModeChanged = ref(false);

/**
 * Changes edit mode
 */
const changeEditMode = () => {
    editMode.value = !editMode.value;
    editModeChanged.value = true;
    updateAlbums();
    focus();
};

// Display modal to create a new album?
const displayAlbumCreate = ref(false);

const createAlbum = () => {
    displayAlbumCreate.value = true;
};

/**
 * Handler for display change events
 * on the album creation modal.
 * After the modal closes, the parent modal should be refocused.
 * @param modalDisplay The new display status
 */
const afterModalCreateClosed = (modalDisplay: boolean) => {
    if (!modalDisplay && display.value) {
        focus();
    }
};

// List of albums to display in the list
const albums = ref<AlbumModalListItem[]>([]);

// Albums filter
const filter = ref("");

// Current media ID
const mid = ref(AppStatus.CurrentMedia);

// List of albums the media is in
const mediaAlbums = ref<number[]>([]);

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    const changed = mid.value !== AppStatus.CurrentMedia;
    mid.value = AppStatus.CurrentMedia;
    if (changed) {
        updateAlbums();
    }
});

/**
 * Called when a new album is created by the user
 * @param _albumId The album ID
 * @param albumName The album name
 */
const onNewAlbum = (_albumId: number, albumName: string) => {
    filter.value = albumName;
    updateAlbums();
};

// Max number of items that will fit in the visible section of the scroller
const LIST_SCROLLER_ITEMS_FIT = 6;

// Big list scroller
const bigListScroller = new BigListScroller(BigListScroller.GetWindowSize(LIST_SCROLLER_ITEMS_FIT), {
    get: () => {
        return albums.value;
    },
    set: (list) => {
        albums.value = list;
    },
});

/**
 * Updates the albums list
 */
const updateAlbums = () => {
    const mid = AppStatus.CurrentMedia;

    const albumFilter = normalizeString(filter.value).trim().toLowerCase();
    const albumFilterWords = filterToWords(albumFilter);

    const albums = AlbumsController.GetAlbumsListMin()
        .map((a) => {
            const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
            return {
                ...a,
                added: mid >= 0 && mediaAlbums.value.indexOf(a.id) >= 0,
                starts: i === 0,
                contains: i >= 0,
            };
        })
        .filter((a) => {
            return (a.starts || a.contains) && (editMode.value || a.added);
        })
        .sort((a, b) => {
            if (a.starts && !b.starts) {
                return -1;
            } else if (b.starts && !a.starts) {
                return 1;
            } else if (a.nameLowerCase < b.nameLowerCase) {
                return -1;
            } else {
                return 1;
            }
        });

    bigListScroller.reset();
    bigListScroller.addElements(albums);
};

onApplicationEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, updateAlbums);

// Load request ID
const loadRequestId = useRequestId();

// Loading status
const loading = ref(true);

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the list of media albums
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    loading.value = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiMediaGetMediaAlbums(mid.value))
        .onSuccess((result) => {
            mediaAlbums.value = result;

            loading.value = false;

            if (!canWrite.value) {
                editMode.value = false;
            } else if (!editModeChanged.value) {
                editMode.value = result.length === 0;
            }

            updateAlbums();
            focus();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    forceClose();
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
        AlbumsController.Refresh();

        updateAlbums();
        load();
    }
});

watch(display, () => {
    displayAlbumCreate.value = false;

    if (display.value) {
        AlbumsController.Refresh();

        updateAlbums();
        load();
    }
});

/**
 * Handler for scroll events
 * @param e The event
 */
const onScroll = (e: Event) => {
    bigListScroller.checkElementScroll(e.target as HTMLElement);
};

/**
 * Navigates to an album the media is into.
 * @param album The album
 * @param event The click event
 */
const goToAlbum = (album: AlbumModalListItem, event?: Event) => {
    if (event) {
        event.preventDefault();
    }

    forceClose();
    AppStatus.ClickOnAlbumByMedia(album.id, mid.value);
};

/**
 * Resolves the URL of an album
 * @param albumId The album ID
 * @param mid The current media ID
 */
const getAlbumURL = (albumId: number, mid: number): string => {
    return getFrontendUrl({
        media: mid,
        album: albumId,
    });
};

// Busy status
const busy = ref(false);

// ID of the album being targeted
const busyTarget = ref(-1);

/**
 * Performs a request to
 * remove the media from the album
 */
const removeFromAlbum = (album: AlbumModalListItem, backToText?: boolean) => {
    if (busy.value) {
        return;
    }

    busy.value = true;
    busyTarget.value = album.id;

    makeApiRequest(apiAlbumsRemoveMediaFromAlbum(album.id, mid.value))
        .onSuccess(() => {
            busy.value = false;

            album.added = false;

            PagesController.ShowSnackBar($t("Successfully removed from album"));

            if (mediaAlbums.value.indexOf(album.id) >= 0) {
                mediaAlbums.value.splice(mediaAlbums.value.indexOf(album.id), 1);
            }

            updateAlbums();

            AlbumsController.OnChangedAlbum(album.id, true);

            if (backToText && editMode.value) {
                focus();
            }
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    AuthController.CheckAuthStatusSilent();
                },
                notFound: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Not found"));
                    AlbumsController.Load();
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            console.error(err);
        });
};

/**
 * Performs a request to
 * add the media into the album
 */
const addIntoAlbum = (album: AlbumModalListItem, backToText?: boolean) => {
    if (busy.value) {
        return;
    }

    busy.value = true;
    busyTarget.value = album.id;

    makeApiRequest(apiAlbumsAddMediaToAlbum(album.id, mid.value))
        .onSuccess(() => {
            busy.value = false;

            album.added = true;

            PagesController.ShowSnackBar($t("Successfully added to album"));

            if (mediaAlbums.value.indexOf(album.id) === -1) {
                mediaAlbums.value.push(album.id);
            }

            updateAlbums();

            AlbumsController.OnChangedAlbum(album.id, true);

            if (backToText && editMode.value) {
                changeEditMode();
            }
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                maxSizeReached: () => {
                    PagesController.ShowSnackBar(
                        $t("Error") + ": " + $t("The album reached the limit of 1024 elements. Please, consider creating another album."),
                    );
                },
                badRequest: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Bad request"));
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    AuthController.CheckAuthStatusSilent();
                },
                notFound: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Not found"));
                    AlbumsController.Load();
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            console.error(err);
        });
};

/**
 * Call when the user clicks in an album checkbox
 * @param album The album element
 * @param backToText True to go back to text input after adding
 * @param event The click event
 */
const clickOnAlbum = (album: AlbumModalListItem, backToText?: boolean, event?: Event) => {
    if (event) {
        event.preventDefault();
    }

    if (busy.value) {
        return;
    }

    if (album.added) {
        // Remove
        removeFromAlbum(album, backToText);
    } else {
        // Add
        addIntoAlbum(album, backToText);
    }
};

/**
 * Handler for 'keydown' in the filter input element
 * @param e The keyboard event
 */
const onFilterKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Enter") {
        e.preventDefault();

        if (!filter.value) {
            return;
        }

        if (albums.value.length === 0) {
            return;
        }

        if (editMode.value) {
            clickOnAlbum(albums.value[0], true);
        } else {
            goToAlbum(albums.value[0]);
        }
    } else if (e.key === "Tab") {
        if (albums.value.length === 0) {
            if (filter.value) {
                e.preventDefault();
            }
            return;
        }

        if (filter.value === albums.value[0].name || !filter.value) {
            return;
        }

        e.preventDefault();

        filter.value = albums.value[0].name;
    }
};
</script>
