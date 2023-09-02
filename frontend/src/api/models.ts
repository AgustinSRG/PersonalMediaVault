// API models
// Typescript definitions for the API

// Albums

export interface AlbumListItem {
    id: number;
    name: string;
    size: number;
    thumbnail: string;
    lm: number;
}

export interface AlbumListItemMin {
    id: number;
    name: string;
}

export interface Album {
    id: number;
    name: string;
    list: MediaListItem[];
    lm: number;
}

// Config

export interface VaultUserConfig {
    title: string;
    max_tasks: number;
    encoding_threads: number;
    video_previews_interval: number;
    resolutions: {
        width: number;
        height: number;
        fps: number;
    }[];
    image_resolutions: {
        width: number;
        height: number;
    }[];
    css: string;
}

// Media

export interface MediaListItem {
    id: number;
    type: 0 | 1 | 2 | 3;
    title: string;
    description: string;
    tags: number[];
    thumbnail: string;
    duration: number;
}

export interface MediaData {
    id: number;
    type: 0 | 1 | 2 | 3;
    title: string;
    description: string;
    tags: number[];
    upload_time: number;
    thumbnail: string;
    duration: number;
    width: number;
    height: number;
    fps: number;
    ready: boolean;
    ready_p: number;
    encoded: boolean;
    task: number;
    url: string;
    video_previews: string;
    video_previews_interval: number;
    resolutions: MediaResolution[];
    subtitles: MediaSubtitle[];
    audios: MediaAudioTrack[];
    attachments: MediaAttachment[];
    force_start_beginning: boolean;
    img_notes: boolean;
    img_notes_url: string;
    ext_desc_url: string;
    time_slices: {
        time: number;
        name: string;
    }[];
}

export interface MediaResolution {
    width: number;
    height: number;
    fps: number;
    ready: boolean;
    task: number;
    url: string;
}

export interface MediaSubtitle {
    id: string;
    name: string;
    url: string;
}

export interface MediaAudioTrack {
    id: string;
    name: string;
    url: string;
}

export interface MediaAttachment {
    id: number;
    name: string;
    size: number;
    url: string;
}

export interface MediaSizeStats {
    meta_size: number;
    assets: {
        id: number;
        type: "s" | "m";
        name: string;
        size: number;
    }[];
}

// Search

export interface SearchResults {
    total_count: number;
    page_index: number;
    page_count: number;
    page_size: number;
    page_items: MediaListItem[];
}

export interface RandomResults {
    seed: number;
    page_size: number;
    page_items: MediaListItem[];
}

// Tags

export interface MediaTag {
    id: number;
    name: string;
}

// Tasks

export interface TaskStatus {
    id: number;
    running: boolean;
    media_id: number;
    type: number;
    resolution: {
        width: number;
        height: number;
        fps: number;
    };
    stage: number;
    stage_start: number;
    time_now: number;
    stage_progress: number;
}
