// Player preferences

import { LocalStorage } from "./local-storage";

const MAX_CACHE_PLAY_TIME_SIZE = 100;

export class PlayerPreferences {
    public static UserSelectedResolution = {
        original: true,
        width: 0,
        height: 0,
        fps: 0,
    };

    public static PlayTimeCache: { mid: number, time: number }[] = [];

    public static PlayerVolume = 1;

    public static PlayerMuted = false;

    public static LoadPreferences() {
        const userRes = LocalStorage.Get("player-pref-resolution", PlayerPreferences.UserSelectedResolution)
        if (userRes) {
            PlayerPreferences.UserSelectedResolution = userRes;
        }

        const playTimeCache = LocalStorage.Get("player-play-time-cache", []);

        if (playTimeCache) {
            PlayerPreferences.PlayTimeCache = playTimeCache;
        }

        PlayerPreferences.PlayerVolume = LocalStorage.Get("player-pref-volume", 1);
        PlayerPreferences.PlayerMuted = LocalStorage.Get("player-pref-muted", false);
    }

    public static GetResolutionIndex(metadata: any): number {
        if (PlayerPreferences.UserSelectedResolution.original || !metadata.resolutions || metadata.resolutions.length === 0) {
            return -1;
        }
        let currentVal = metadata.width * metadata.height * metadata.fps;
        const prefVal = PlayerPreferences.UserSelectedResolution.width * PlayerPreferences.UserSelectedResolution.height * PlayerPreferences.UserSelectedResolution.fps;
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

    public static SetResolutionIndex(metadata: any, index: number) {
        if (index < 0) {
            PlayerPreferences.UserSelectedResolution = {
                original: true,
                width: 0,
                height: 0,
                fps: 0,
            };
        } else if (metadata  && metadata.resolutions && metadata.resolutions[index] && metadata.resolutions[index].ready) {
            PlayerPreferences.UserSelectedResolution = {
                original: false,
                width: metadata.resolutions[index].width,
                height: metadata.resolutions[index].height,
                fps: metadata.resolutions[index].fps,
            };
        }

        LocalStorage.Set("player-pref-resolution", PlayerPreferences.UserSelectedResolution)
    }

    public static GetInitialTime(mid: number) {
        for (const entry of PlayerPreferences.PlayTimeCache) {
            if (entry.mid === mid) {
                return entry.time;
            }
        }
    }

    public static SetInitialTime(mid: number, time: number) {
        // Remove if found
        PlayerPreferences.PlayTimeCache = PlayerPreferences.PlayTimeCache.filter(e => {
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
        PlayerPreferences.PlayTimeCache = PlayerPreferences.PlayTimeCache.filter(e => {
            return e.mid !== mid;
        });

        LocalStorage.Set("player-play-time-cache", PlayerPreferences.PlayTimeCache);
    }

    public static SetVolume(vol: number) {
        PlayerPreferences.PlayerVolume = vol;
        LocalStorage.Set("player-pref-volume", vol);
    }

    public static SetMuted(m: boolean) {
        PlayerPreferences.PlayerMuted = m;
        LocalStorage.Set("player-pref-muted", m);
    }
}
