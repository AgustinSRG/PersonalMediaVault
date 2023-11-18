// API models
// Typescript definitions for the API

"use strict";

// Albums

/**
 * Album information for the albums page
 */
export interface AlbumListItem {
    /**
     * Album ID
     */
    id: number;

    /**
     * Album name
     */
    name: string;

    /**
     * Number of elements in the album
     */
    size: number;

    /**
     * Path to the album thumbnail.
     */
    thumbnail: string;

    /**
     * Last modified timestamp
     */
    lm: number;
}

/**
 * Album information for the basic list.
 * Only the ID and the name.
 */
export interface AlbumListItemMin {
    /**
     * Album ID
     */
    id: number;

    /**
     * Album name
     */
    name: string;
}

/**
 * Album information to display the media list.
 */
export interface Album {
    /**
     * Album ID
     */
    id: number;

    /**
     * Album name
     */
    name: string;

    /**
     * List of elements in the album
     */
    list: MediaListItem[];

    /**
     * Last modified timestamp
     */
    lm: number;
}

// Config

/**
 * Video resolution
 */
export interface VideoResolution {
    /**
     * Width
     */
    width: number;

    /**
     * Height
     */
    height: number;

    /**
     * Frames per second
     */
    fps: number;
}

/**
 * Image resolution
 */
export interface ImageResolution {
    /**
     * Width
     */
    width: number;

    /**
     * Height
     */
    height: number;
}

/**
 * Vault user configuration
 */
export interface VaultUserConfig {
    /**
     * Custom vault title
     */
    title: string;

    /**
     * Max number of task allowed in parallel
     */
    max_tasks: number;

    /**
     * Max number of encoding threads per task
     */
    encoding_threads: number;

    /**
     * Interval in seconds to take video previews
     */
    video_previews_interval: number;

    /**
     * List of video resolutions
     */
    resolutions: VideoResolution[];

    /**
     * List of image resolutions
     */
    image_resolutions: ImageResolution[];

    /**
     * Custom stylesheet
     */
    css: string;
}

// Media

/**
 * Type value for images
 */
export const MEDIA_TYPE_IMAGE = 1;

/**
 * Type value for videos
 */
export const MEDIA_TYPE_VIDEO = 2;

/**
 * Type value for audios
 */
export const MEDIA_TYPE_AUDIO = 3;

/**
 * Media type.
 * Image = 1
 * Video = 2
 * Audio = 3
 * Deleted = 0
 */
export type MediaType = 0 | 1 | 2 | 3;

/**
 * Media information to show in a list
 */
export interface MediaListItem {
    /**
     * Media ID
     */
    id: number;

    /**
     * Media type
     */
    type: MediaType;

    /**
     * Title
     */
    title: string;

    /**
     * Description
     */
    description: string;

    /**
     * List of tags
     */
    tags: number[];

    /**
     * Thumbnail path
     */
    thumbnail: string;

    /**
     * Duration in seconds
     */
    duration: number;
}

/**
 * Full media information
 */
export interface MediaData {
    /**
     * Media ID
     */
    id: number;

    /**
     * Media type
     */
    type: MediaType;

    /**
     * Title
     */
    title: string;

    /**
     * Description
     */
    description: string;

    /**
     * List of tags
     */
    tags: number[];

    /**
     * Upload timestamp
     */
    upload_time: number;

    /**
     * Thumbnail path
     */
    thumbnail: string;

    /**
     * Duration in seconds
     */
    duration: number;

    /**
     * Width
     */
    width: number;

    /**
     * Height
     */
    height: number;

    /**
     * Frames per second
     */
    fps: number;

    /**
     * True if ready (fully uploaded)
     */
    ready: boolean;

    /**
     * If not ready, progress (0-100)
     */
    ready_p: number;

    /**
     * True if encoded (can be played)
     */
    encoded: boolean;

    /**
     * If not encoded, ID of the encoding task
     */
    task: number;

    /**
     * Path to the original media file
     */
    url: string;

    /**
     * Path to the video previews
     */
    video_previews: string;

    /**
     * Interval in seconds for the video previews
     */
    video_previews_interval: number;

    /**
     * List of available resolutions
     */
    resolutions: MediaResolution[];

    /**
     * List of available subtitles
     */
    subtitles: MediaSubtitle[];

    /**
     * List of available audio tracks
     */
    audios: MediaAudioTrack[];

    /**
     * List of attachments
     */
    attachments: MediaAttachment[];

    /**
     * True to always start from the beginning
     */
    force_start_beginning: boolean;

    /**
     * True if image notes are available
     */
    img_notes: boolean;

    /**
     * Path to the image notes file
     */
    img_notes_url: string;

    /**
     * Path to the extended description file
     */
    ext_desc_url: string;

    /**
     * List of time slices
     */
    time_slices: MediaTimeSlice[];
}

/**
 * Media resolution
 */
export interface MediaResolution {
    /**
     * Width
     */
    width: number;

    /**
     * Height
     */
    height: number;

    /**
     * Frames per second
     */
    fps: number;

    /**
     * True if ready to the played
     */
    ready: boolean;

    /**
     * If not ready, ID of the encoding task
     */
    task: number;

    /**
     * Path to the media file
     */
    url: string;
}

/**
 * Media subtitles file
 */
export interface MediaSubtitle {
    /**
     * File ID
     */
    id: string;

    /**
     * File name
     */
    name: string;

    /**
     * Path to the subtitles file
     */
    url: string;
}

/**
 * Media audio track
 */
export interface MediaAudioTrack {
    /**
     * File ID
     */
    id: string;

    /**
     * File name
     */
    name: string;

    /**
     * Path to the audio file
     */
    url: string;
}

/**
 * Media attachment
 */
export interface MediaAttachment {
    /**
     * Attachment ID
     */
    id: number;

    /**
     * File name
     */
    name: string;

    /**
     * File size (bytes)
     */
    size: number;

    /**
     * Path to the file
     */
    url: string;
}

/**
 * Media time slice
 */
export interface MediaTimeSlice {
    /**
     * Timestamp in seconds of the start of the slice.
     */
    time: number;

    /**
     * Slice name
     */
    name: string;
}

/**
 * Media size stats
 */
export interface MediaSizeStats {
    /**
     * Size of the metadata (bytes)
     */
    meta_size: number;

    /**
     * Size of the assets
     */
    assets: MediaSizeStatsAsset[];
}

/**
 * Size starts for a media asset
 */
export interface MediaSizeStatsAsset {
    /**
     * Asset ID
     */
    id: number;

    /**
     * Asset type.
     * Single file = s
     * Multi file = m
     */
    type: "s" | "m";

    /**
     * Asset name
     */
    name: string;

    /**
     * Asset size (bytes)
     */
    size: number;
}

// Search

/**
 * Search results for a page
 */
export interface SearchResults {
    /**
     * Total number of available items
     */
    total_count: number;

    /**
     * Current page index
     */
    page_index: number;

    /**
     * Total number of pages
     */
    page_count: number;

    /**
     * Page size
     */
    page_size: number;

    /**
     * List of page items
     */
    page_items: MediaListItem[];
}

/**
 * Random results for a page
 */
export interface RandomResults {
    /**
     * Used seed
     */
    seed: number;

    /**
     * Page size
     */
    page_size: number;

    /**
     * List of page items
     */
    page_items: MediaListItem[];
}

// Tags

/**
 * Information of a tag
 */
export interface MediaTag {
    /**
     * Tag ID
     */
    id: number;

    /**
     * Tag name
     */
    name: string;
}

// Tasks

/**
 * Status of a task
 */
export interface TaskStatus {
    /**
     * Task ID
     */
    id: number;

    /**
     * True if the task is running
     */
    running: boolean;

    /**
     * ID of the media the task was created for
     */
    media_id: number;

    /**
     * Type of task.
     * Encode original = 0
     * Encode resolution = 1
     * Encode previews = 2
     */
    type: number;

    /**
     * Resolution data
     */
    resolution: VideoResolution;

    /**
     * Name of the current stage
     * PREPARE - Preparing task resources
     * COPY - Decrypting and copying media assets
     * PROBE - Extracting metadata from media assets
     * ENCODE - Encoding
     * ENCRYPT - Encrypting new assets
     * UPDATE - Updating vault
     * FINISH - Finishing (clearing temp files, etc)
     */
    stage: string;

    /**
     * Timestamp of the beginning of the stage
     */
    stage_start: number;

    /**
     * Current server timestamp
     */
    time_now: number;

    /**
     * Progress of the stage (0-100)
     */
    stage_progress: number;
}
