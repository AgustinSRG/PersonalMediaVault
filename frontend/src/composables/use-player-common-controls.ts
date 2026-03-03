// Player common controls composable

"use strict";

import type { Ref, ShallowRef } from "vue";
import { onMounted, ref, watch } from "vue";
import type { MediaData } from "@/api/models";
import { useUserPermissions } from "./use-user-permissions";

/**
 * Required properties for common controls
 */
export type PlayerCommonControlsProps = {
    /**
     * Media metadata
     */
    metadata: MediaData | null;

    /**
     * Reload tick
     */
    rTick: number;
};

/**
 * Required emits for common controls
 */
export type PlayerCommonControlsEmits = {
    /**
     * The user wants to open the albums list modal
     */
    (e: "albums-open"): void;

    /**
     * The user wants to upen the size stats modal
     */
    (e: "stats-open"): void;
};

/**
 * Player common controls composable
 */
export type PlayerCommonControlsComposable = {
    /**
     * Has description?
     */
    hasDescription: Ref<boolean>;

    /**
     * Has attachments?
     */
    hasAttachments: Ref<boolean>;

    /**
     * Display attachments
     */
    displayAttachments: Ref<boolean>;

    /**
     * Has related media?
     */
    hasRelatedMedia: Ref<boolean>;

    /**
     * Display related media?
     */
    displayRelatedMedia: Ref<boolean>;

    /**
     * Display player config
     */
    displayConfig: Ref<boolean>;

    /**
     * Display the context menu?
     */
    contextMenuShown: Ref<boolean>;

    /**
     * X coordinate of context menu
     */
    contextMenuX: Ref<number>;

    /**
     * Y coordinate of context menu
     */
    contextMenuY: Ref<number>;

    /**
     * Event handler for context menu
     * @param e The mouse event
     */
    onContextMenu: (e: MouseEvent) => void;

    /**
     * Opens the albums modal
     */
    manageAlbums: () => void;

    /**
     * Opens the stats modal
     */
    openStats: () => void;

    /**
     * Opens the tags widget
     */
    openTags: () => void;

    /**
     * Opens the description widget
     */
    openDescription: () => void;

    /**
     * Refreshes the description from the metadata
     */
    refreshDescription: () => void;

    /**
     * Opens the attachments list
     */
    showAttachments: () => void;

    /**
     * Opens the related media list
     */
    showRelatedMedia: () => void;

    /**
     * Displays the player config
     */
    showConfig: () => void;

    /**
     * Called when the user clicks the controls,
     * in order to hide any menus
     * @param e The click event (optional)
     */
    clickControls: (e?: Event) => void;
};

/**
 * Gets a composable for common controls for all 3 players (image, audio, video)
 * @param props The props
 * @param emit The emit function
 * @param refs Required references to the component models and template
 */
export function usePlayerCommonControls(
    props: PlayerCommonControlsProps,
    emit: PlayerCommonControlsEmits,
    refs: {
        displayTagList: Ref<boolean>;
        displayDescription: Ref<boolean>;
        contextMenu: Readonly<ShallowRef<{ show: () => void; hide: () => void }>>;
    },
): PlayerCommonControlsComposable {
    // User permissions
    const { canWrite } = useUserPermissions();

    // Does the media has a description?
    const hasDescription = ref(false);

    // Does the media have attachments
    const hasAttachments = ref(false);

    // Display the attachments list
    const displayAttachments = ref(false);

    // Does the media has related media?
    const hasRelatedMedia = ref(false);

    // Display related media list?
    const displayRelatedMedia = ref(false);

    // Display player config?
    const displayConfig = ref(false);

    /**
     * Called when metadata is updated
     */
    const updateMetadata = () => {
        if (!props.metadata) {
            hasDescription.value = false;
            hasAttachments.value = false;
            hasRelatedMedia.value = false;
            return;
        }

        hasDescription.value = !!props.metadata.description_url;
        hasAttachments.value = props.metadata.attachments && props.metadata.attachments.length > 0;
        hasRelatedMedia.value = props.metadata.related && props.metadata.related.length > 0;
    };

    onMounted(updateMetadata);
    watch(() => props.rTick, updateMetadata);

    // Display context menu?
    const contextMenuShown = ref(false);

    // Context menu coordinates
    const contextMenuX = ref(0);
    const contextMenuY = ref(0);

    /**
     * Context menu event handler
     * @param e The event
     */
    const onContextMenu = (e: MouseEvent) => {
        contextMenuX.value = e.pageX;
        contextMenuY.value = e.pageY;
        contextMenuShown.value = true;
        refs.contextMenu.value?.show();
        e.preventDefault();
    };

    /**
     * Opens the albums modal
     */
    const manageAlbums = () => {
        emit("albums-open");
    };

    /**
     * Opens the size stats modal
     */
    const openStats = () => {
        emit("stats-open");
    };

    /**
     * Opens the tags widget
     */
    const openTags = () => {
        refs.displayTagList.value = true;
    };

    /**
     * Opens the description widget
     */
    const openDescription = () => {
        if (!hasDescription.value && !canWrite.value) {
            return;
        }
        refs.displayDescription.value = true;
    };

    /**
     * Refreshes the description from the media metadata
     */
    const refreshDescription = () => {
        hasDescription.value = !!props.metadata?.description_url;
    };

    /**
     * Opens the attachments list
     */
    const showAttachments = () => {
        displayAttachments.value = !displayAttachments.value;
        displayRelatedMedia.value = false;
        displayConfig.value = false;
    };

    /**
     * Opens the related media list
     */
    const showRelatedMedia = () => {
        displayRelatedMedia.value = !displayRelatedMedia.value;
        displayAttachments.value = false;
        displayConfig.value = false;
    };

    /**
     * Opens the player configuration
     */
    const showConfig = () => {
        displayConfig.value = !displayConfig.value;
        displayAttachments.value = false;
        displayRelatedMedia.value = false;
    };

    /**
     * Called when the user clicked the controls
     * @param e The click event
     */
    const clickControls = (e?: Event) => {
        displayConfig.value = false;
        refs.contextMenu.value?.hide();
        displayAttachments.value = false;
        displayRelatedMedia.value = false;
        if (e) {
            e.stopPropagation();
        }
    };

    return {
        hasAttachments,
        displayAttachments,
        hasDescription,
        hasRelatedMedia,
        displayRelatedMedia,
        displayConfig,
        contextMenuShown,
        contextMenuX,
        contextMenuY,
        onContextMenu,
        openDescription,
        openStats,
        openTags,
        showAttachments,
        showRelatedMedia,
        showConfig,
        manageAlbums,
        refreshDescription,
        clickControls,
    };
}
