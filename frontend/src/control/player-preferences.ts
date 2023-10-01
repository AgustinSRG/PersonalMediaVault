// Player preferences

import { LocalStorage } from "./local-storage";

const MAX_CACHE_PLAY_TIME_SIZE = 100;
const MAX_CACHE_ALBUM_POS_SIZE = 100;

export class PlayerPreferences {
    public static UserSelectedResolution = {
        original: true,
        width: 0,
        height: 0,
        fps: 0,
    };

    public static UserSelectedResolutionImage = {
        original: true,
        width: 0,
        height: 0,
    };

    public static PlayTimeCache: { mid: number; time: number }[] = [];

    public static AlbumCurrentCache: { id: number; pos: number }[] = [];

    public static PlayerVolume = 1;
    public static PlayerMuted = false;

    public static PlayerScale = 0;
    public static PlayerFit = true;

    public static AudioAnimationStyle = "gradient";

    public static ImagePlayerBackground = "default";

    public static NextOnEnd = true;
    public static ImageAutoNext = 0;

    public static ImageNotesVisible = true;

    public static SelectedSubtitles = "";
    public static SubtitlesSize = "l";
    public static SubtitlesBackground = "75";
    public static SubtitlesHTML = false;

    public static SelectedAudioTrack = "";

    public static PlayerTogglePlayDelay = 250;

    public static ExtendedDescriptionSize = "xl";

    public static LoadPreferences() {
        const userRes = LocalStorage.Get("player-pref-resolution", PlayerPreferences.UserSelectedResolution);
        if (userRes) {
            PlayerPreferences.UserSelectedResolution = userRes;
        }

        const userResImage = LocalStorage.Get("player-pref-resolution-img", PlayerPreferences.UserSelectedResolutionImage);
        if (userResImage) {
            PlayerPreferences.UserSelectedResolutionImage = userResImage;
        }

        const playTimeCache = LocalStorage.Get("player-play-time-cache", []);

        if (playTimeCache) {
            PlayerPreferences.PlayTimeCache = playTimeCache;
        }

        const albumPosCache = LocalStorage.Get("player-album-pos-cache", []);

        if (albumPosCache) {
            PlayerPreferences.AlbumCurrentCache = albumPosCache;
        }

        PlayerPreferences.PlayerVolume = LocalStorage.Get("player-pref-volume", 1);
        PlayerPreferences.PlayerMuted = LocalStorage.Get("player-pref-muted", false);

        PlayerPreferences.PlayerScale = LocalStorage.Get("player-pref-scale", 0);
        PlayerPreferences.PlayerFit = LocalStorage.Get("player-pref-fit", true);

        PlayerPreferences.AudioAnimationStyle = LocalStorage.Get("player-pref-audio-anim", "gradient");

        PlayerPreferences.ImagePlayerBackground = LocalStorage.Get("player-pref-img-bg", "default");

        PlayerPreferences.ImageAutoNext = LocalStorage.Get("player-pref-img-auto-next", 0);
        PlayerPreferences.NextOnEnd = LocalStorage.Get("player-pref-next-end", true);

        PlayerPreferences.ImageNotesVisible = LocalStorage.Get("player-pref-img-notes-v", true);

        PlayerPreferences.SelectedSubtitles = LocalStorage.Get("player-pref-subtitles", "");
        PlayerPreferences.SubtitlesSize = LocalStorage.Get("player-pref-subtitles-size", "l");
        PlayerPreferences.SubtitlesBackground = LocalStorage.Get("player-pref-subtitles-bg", "75");
        PlayerPreferences.SubtitlesHTML = LocalStorage.Get("player-pref-subtitles-html", false);

        PlayerPreferences.PlayerTogglePlayDelay = LocalStorage.Get("player-pref-toggle-delay", 250);

        PlayerPreferences.SelectedAudioTrack = LocalStorage.Get("player-pref-audio-track", "");

        PlayerPreferences.ExtendedDescriptionSize = LocalStorage.Get("player-pref-ext-desc-size", "xl");
    }

    public static GetResolutionIndex(metadata: any): number {
        if (PlayerPreferences.UserSelectedResolution.original || !metadata.resolutions || metadata.resolutions.length === 0) {
            return -1;
        }
        let currentVal = metadata.width * metadata.height * metadata.fps;
        const prefVal =
            PlayerPreferences.UserSelectedResolution.width *
            PlayerPreferences.UserSelectedResolution.height *
            PlayerPreferences.UserSelectedResolution.fps;
        let currenRes = -1;
        for (let i = 0; i < metadata.resolutions.length; i++) {
            const res = metadata.resolutions[i];
            if (!res.ready) {
                continue;
            }
            const resVal = res.width * res.height * res.fps;
            if (Math.abs(resVal - prefVal) < Math.abs(currentVal - prefVal)) {
                currentVal = resVal;
                currenRes = i;
            }
        }

        return currenRes;
    }

    public static GetResolutionIndexImage(metadata: any): number {
        if (PlayerPreferences.UserSelectedResolutionImage.original || !metadata.resolutions || metadata.resolutions.length === 0) {
            return -1;
        }
        let currentVal = metadata.width * metadata.height;
        const prefVal = PlayerPreferences.UserSelectedResolutionImage.width * PlayerPreferences.UserSelectedResolutionImage.height;
        let currenRes = -1;
        for (let i = 0; i < metadata.resolutions.length; i++) {
            const res = metadata.resolutions[i];
            if (!res.ready) {
                continue;
            }
            const resVal = res.width * res.height;
            if (Math.abs(resVal - prefVal) < Math.abs(currentVal - prefVal)) {
                currentVal = resVal;
                currenRes = i;
            }
        }

        return currenRes;
    }

    public static SetResolutionIndex(metadata: any, index: number) {
        if (index < 0) {
            PlayerPreferences.UserSelectedResolution = {
                original: true,
                width: 0,
                height: 0,
                fps: 0,
            };
        } else if (metadata && metadata.resolutions && metadata.resolutions[index] && metadata.resolutions[index].ready) {
            PlayerPreferences.UserSelectedResolution = {
                original: false,
                width: metadata.resolutions[index].width,
                height: metadata.resolutions[index].height,
                fps: metadata.resolutions[index].fps,
            };
        }

        LocalStorage.Set("player-pref-resolution", PlayerPreferences.UserSelectedResolution);
    }

    public static SetResolutionIndexImage(metadata: any, index: number) {
        if (index < 0) {
            PlayerPreferences.UserSelectedResolutionImage = {
                original: true,
                width: 0,
                height: 0,
            };
        } else if (metadata && metadata.resolutions && metadata.resolutions[index] && metadata.resolutions[index].ready) {
            PlayerPreferences.UserSelectedResolutionImage = {
                original: false,
                width: metadata.resolutions[index].width,
                height: metadata.resolutions[index].height,
            };
        }

        LocalStorage.Set("player-pref-resolution-img", PlayerPreferences.UserSelectedResolutionImage);
    }

    public static GetInitialTime(mid: number) {
        PlayerPreferences.PlayTimeCache = LocalStorage.Get("player-play-time-cache", []); // Update
        for (const entry of PlayerPreferences.PlayTimeCache) {
            if (entry.mid === mid) {
                const time = entry.time;
                if (typeof time === "number" && !isNaN(time) && isFinite(time) && time >= 0) {
                    return time;
                } else {
                    return 0;
                }
            }
        }

        return 0;
    }

    public static SetInitialTime(mid: number, time: number) {
        // Remove if found
        PlayerPreferences.PlayTimeCache = LocalStorage.Get("player-play-time-cache", []).filter((e) => {
            return e.mid !== mid;
        });

        while (PlayerPreferences.PlayTimeCache.length >= MAX_CACHE_PLAY_TIME_SIZE) {
            PlayerPreferences.PlayTimeCache.shift();
        }

        PlayerPreferences.PlayTimeCache.push({
            mid: mid,
            time: time,
        });

        LocalStorage.Set("player-play-time-cache", PlayerPreferences.PlayTimeCache);
    }

    public static ClearInitialTime(mid: number) {
        // Remove if found
        PlayerPreferences.PlayTimeCache = LocalStorage.Get("player-play-time-cache", []).filter((e) => {
            return e.mid !== mid;
        });

        LocalStorage.Set("player-play-time-cache", PlayerPreferences.PlayTimeCache);
    }

    public static GetAlbumPos(id: number): number {
        PlayerPreferences.AlbumCurrentCache = LocalStorage.Get("player-album-pos-cache", []);
        for (const entry of PlayerPreferences.AlbumCurrentCache) {
            if (entry.id === id) {
                const pos = entry.pos;
                if (typeof pos === "number" && !isNaN(pos) && isFinite(pos) && pos >= 0) {
                    return pos;
                } else {
                    return 0;
                }
            }
        }

        return 0;
    }

    public static SetAlbumPos(id: number, pos: number) {
        PlayerPreferences.AlbumCurrentCache = LocalStorage.Get("player-album-pos-cache", []).filter((e) => {
            return e.id !== id;
        });

        while (PlayerPreferences.AlbumCurrentCache.length >= MAX_CACHE_ALBUM_POS_SIZE) {
            PlayerPreferences.AlbumCurrentCache.shift();
        }

        PlayerPreferences.AlbumCurrentCache.push({
            id: id,
            pos: pos,
        });

        LocalStorage.Set("player-album-pos-cache", PlayerPreferences.AlbumCurrentCache);
    }

    public static SetVolume(vol: number) {
        PlayerPreferences.PlayerVolume = vol;
        LocalStorage.Set("player-pref-volume", vol);
    }

    public static SetMuted(m: boolean) {
        PlayerPreferences.PlayerMuted = m;
        LocalStorage.Set("player-pref-muted", m);
    }

    public static SetScale(s: number) {
        PlayerPreferences.PlayerScale = s;
        LocalStorage.Set("player-pref-scale", s);
    }

    public static SetFit(f: boolean) {
        PlayerPreferences.PlayerFit = f;
        LocalStorage.Set("player-pref-fit", f);
    }

    public static SetAudioAnimationStyle(s: string) {
        PlayerPreferences.AudioAnimationStyle = s;
        LocalStorage.Set("player-pref-audio-anim", s);
    }

    public static SetImagePlayerBackground(s: string) {
        PlayerPreferences.ImagePlayerBackground = s;
        LocalStorage.Set("player-pref-img-bg", s);
    }

    public static SetImageAutoNext(s: number) {
        PlayerPreferences.ImageAutoNext = s;
        LocalStorage.Set("player-pref-img-auto-next", s);
    }

    public static SetNextOnEnd(s: boolean) {
        PlayerPreferences.NextOnEnd = s;
        LocalStorage.Set("player-pref-next-end", s);
    }

    public static SetImageNotesVisible(v: boolean) {
        PlayerPreferences.ImageNotesVisible = v;
        LocalStorage.Set("player-pref-img-notes-v", v);
    }

    public static SetSubtitles(s: string) {
        PlayerPreferences.SelectedSubtitles = s;
        LocalStorage.Set("player-pref-subtitles", s);
    }

    public static SetAudioTrack(s: string) {
        PlayerPreferences.SelectedAudioTrack = s;
        LocalStorage.Set("player-pref-audio-track", s);
    }

    public static SetSubtitlesSize(s: string) {
        PlayerPreferences.SubtitlesSize = s;
        LocalStorage.Set("player-pref-subtitles-size", s);
    }

    public static SetSubtitlesBackground(s: string) {
        PlayerPreferences.SubtitlesBackground = s;
        LocalStorage.Set("player-pref-subtitles-bg", s);
    }

    public static SetSubtitlesHTML(s: boolean) {
        PlayerPreferences.SubtitlesHTML = s;
        LocalStorage.Set("player-pref-subtitles-html", s);
    }

    public static SetPlayerToggleDelay(d: number) {
        PlayerPreferences.PlayerTogglePlayDelay = d;
        LocalStorage.Set("player-pref-toggle-delay", d);
    }

    public static SetExtendedDescriptionSize(s: string) {
        PlayerPreferences.ExtendedDescriptionSize = s;
        LocalStorage.Set("player-pref-ext-desc-size", s);
    }
}
