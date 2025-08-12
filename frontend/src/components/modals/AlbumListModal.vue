<template>
    <ModalDialogContainer
        v-model:display="displayStatus"
        :close-signal="closeSignal"
        :force-close-signal="forceCloseSignal"
        :lock-close="busy"
    >
        <div v-if="display" class="modal-dialog modal-sm" role="document">
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

<script lang="ts">
import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import AlbumCreateModal from "../modals/AlbumCreateModal.vue";
import { getUniqueStringId } from "@/utils/unique-id";
import { PagesController } from "@/control/pages";
import { apiAlbumsAddMediaToAlbum, apiAlbumsRemoveMediaFromAlbum } from "@/api/api-albums";
import { apiMediaGetMediaAlbums } from "@/api/api-media";
import { getFrontendUrl } from "@/utils/api";
import LoadingIcon from "@/components/utils/LoadingIcon.vue";
import { BigListScroller } from "@/utils/big-list-scroller";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";

interface AlbumModalListItem {
    id: number;
    name: string;
    nameLowerCase: string;
    added: boolean;
    starts: boolean;
    contains: boolean;
}

export default defineComponent({
    name: "AlbumListModal",
    components: {
        AlbumCreateModal,
        LoadingIcon,
    },
    props: {
        display: Boolean,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            displayStatus: useVModel(props, "display"),

            bigListScroller: null as BigListScroller<AlbumModalListItem>,
        };
    },
    data: function () {
        return {
            albums: [] as AlbumModalListItem[],
            filter: "",

            mid: AppStatus.CurrentMedia,
            mediaAlbums: [] as number[],

            loading: true,
            busy: false,
            busyTarget: -1,

            displayAlbumCreate: false,

            editMode: AuthController.CanWrite,
            canWrite: AuthController.CanWrite,
            editModeChanged: false,

            closeSignal: 0,
            forceCloseSignal: 0,
        };
    },
    watch: {
        display: function () {
            this.displayAlbumCreate = false;
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
                AlbumsController.Load();
                this.load();
            } else {
                clearNamedTimeout(this.loadRequestId);
                abortNamedApiRequest(this.loadRequestId);
            }
        },
    },
    mounted: function () {
        this.bigListScroller = new BigListScroller(BigListScroller.GetWindowSize(6), {
            get: () => {
                return this.albums;
            },
            set: (list) => {
                this.albums = list;
            },
        });

        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, this.updateAlbums.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onUpdateStatus.bind(this));

        this.updateAlbums();
        this.load();

        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
            AlbumsController.Load();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    methods: {
        autoFocus: function () {
            nextTick(() => {
                const el = this.$el.querySelector(".auto-focus");

                if (el) {
                    el.focus();
                }
            });
        },

        afterModalCreateClosed: function (display: boolean) {
            if (!display && this.display) {
                this.autoFocus();
            }
        },

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

            makeNamedApiRequest(this.loadRequestId, apiMediaGetMediaAlbums(this.mid))
                .onSuccess((result) => {
                    this.mediaAlbums = result;
                    this.loading = false;
                    this.canWrite = AuthController.CanWrite;
                    if (!this.canWrite) {
                        this.editMode = false;
                    } else if (!this.editModeChanged) {
                        this.editMode = result.length === 0;
                    }
                    this.updateAlbums();
                    this.autoFocus();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        notFound: () => {
                            this.forceCloseSignal++;
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

        close: function () {
            this.closeSignal++;
        },

        createAlbum: function () {
            this.displayAlbumCreate = true;
        },

        onNewAlbum: function (albumId: number, albumName: string) {
            this.filter = albumName;
            this.updateAlbums();
        },

        goToAlbum: function (album: AlbumModalListItem, event?: Event) {
            if (event) {
                event.preventDefault();
            }

            this.forceCloseSignal++;
            AppStatus.ClickOnAlbumByMedia(album.id, this.mid);
        },

        changeEditMode: function () {
            this.editMode = !this.editMode;
            this.editModeChanged = true;
            this.updateAlbums();
            this.autoFocus();
        },

        clickOnAlbum: function (album: AlbumModalListItem, backToText?: boolean, event?: Event) {
            if (event) {
                event.preventDefault();
            }

            if (this.busy) {
                return;
            }

            this.busy = true;
            this.busyTarget = album.id;

            if (album.added) {
                // Remove
                makeApiRequest(apiAlbumsRemoveMediaFromAlbum(album.id, this.mid))
                    .onSuccess(() => {
                        this.busy = false;
                        album.added = false;
                        PagesController.ShowSnackBar(this.$t("Successfully removed from album"));
                        if (this.mediaAlbums.indexOf(album.id) >= 0) {
                            this.mediaAlbums.splice(this.mediaAlbums.indexOf(album.id), 1);
                        }
                        this.updateAlbums();
                        AlbumsController.OnChangedAlbum(album.id, true);
                        if (backToText && this.editMode) {
                            this.autoFocus();
                        }
                    })
                    .onRequestError((err, handleErr) => {
                        this.busy = false;
                        handleErr(err, {
                            unauthorized: () => {
                                AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            },
                            accessDenied: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                                AuthController.CheckAuthStatusSilent();
                            },
                            notFound: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                                AlbumsController.Load();
                            },
                            serverError: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                            },
                            networkError: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                            },
                        });
                    })
                    .onUnexpectedError((err) => {
                        this.busy = false;
                        console.error(err);
                    });
            } else {
                // Add
                makeApiRequest(apiAlbumsAddMediaToAlbum(album.id, this.mid))
                    .onSuccess(() => {
                        this.busy = false;
                        album.added = true;
                        PagesController.ShowSnackBar(this.$t("Successfully added to album"));
                        if (this.mediaAlbums.indexOf(album.id) === -1) {
                            this.mediaAlbums.push(album.id);
                        }
                        this.updateAlbums();
                        AlbumsController.OnChangedAlbum(album.id, true);
                        if (backToText && this.editMode) {
                            this.changeEditMode();
                        }
                    })
                    .onRequestError((err, handleErr) => {
                        this.busy = false;
                        handleErr(err, {
                            unauthorized: () => {
                                AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            },
                            maxSizeReached: () => {
                                PagesController.ShowSnackBar(
                                    this.$t("Error") +
                                        ": " +
                                        this.$t("The album reached the limit of 1024 elements. Please, consider creating another album."),
                                );
                            },
                            badRequest: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                            },
                            accessDenied: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                                AuthController.CheckAuthStatusSilent();
                            },
                            notFound: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                                AlbumsController.Load();
                            },
                            serverError: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                            },
                            networkError: () => {
                                PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                            },
                        });
                    })
                    .onUnexpectedError((err) => {
                        this.busy = false;
                        console.error(err);
                    });
            }
        },

        onUpdateStatus: function () {
            const changed = this.mid !== AppStatus.CurrentMedia;
            this.mid = AppStatus.CurrentMedia;
            if (changed) {
                this.updateAlbums();
            }
        },

        updateAlbums: function () {
            const mid = AppStatus.CurrentMedia;

            const albumFilter = normalizeString(this.filter).trim().toLowerCase();
            const albumFilterWords = filterToWords(albumFilter);

            const albums = AlbumsController.GetAlbumsListMin()
                .map((a) => {
                    const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
                    return {
                        ...a,
                        added: mid >= 0 && this.mediaAlbums.indexOf(a.id) >= 0,
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    return (a.starts || a.contains) && (this.editMode || a.added);
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

            this.bigListScroller.reset();
            this.bigListScroller.addElements(albums);
        },

        onFilterKeyDown: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();

                if (!this.filter) {
                    return;
                }

                if (this.albums.length === 0) {
                    return;
                }

                if (this.editMode) {
                    this.clickOnAlbum(this.albums[0], true);
                } else {
                    this.goToAlbum(this.albums[0]);
                }
            } else if (e.key === "Tab") {
                if (this.albums.length === 0) {
                    if (this.filter) {
                        e.preventDefault();
                    }
                    return;
                }

                if (this.filter === this.albums[0].name || !this.filter) {
                    return;
                }

                e.preventDefault();

                this.filter = this.albums[0].name;
            }
        },

        getAlbumURL: function (albumId: number, mid: number): string {
            return getFrontendUrl({
                media: mid,
                album: albumId,
            });
        },

        onScroll: function (e: Event) {
            this.bigListScroller.checkElementScroll(e.target as HTMLElement);
        },
    },
});
</script>
