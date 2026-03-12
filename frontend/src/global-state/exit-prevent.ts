// Exit preventer

"use strict";

const ExitPreventState = {
    /**
     * Current function to call to check if if the exit should be prevented
     */
    currentChecker: null as (() => boolean) | null,

    /**
     * Current function to call before running any exit action
     */
    currentExitFunc: null as ((callback: () => void) => void) | null,
};

window.addEventListener("beforeunload", function (e: BeforeUnloadEvent) {
    if (ExitPreventState.currentChecker && ExitPreventState.currentChecker()) {
        // Cancel the event
        e.preventDefault(); // If you prevent default behavior in Mozilla Firefox prompt will always be shown
        // Chrome requires returnValue to be set
        e.returnValue = "";
    }
});

/**
 * Set up the functions to manage the exit prevention
 * @param checker Checker function
 * @param exitFunc Exit handler function
 */
export function setupExitPrevent(checker: () => boolean, exitFunc: (callback: () => void) => void) {
    ExitPreventState.currentChecker = checker;
    ExitPreventState.currentExitFunc = exitFunc;
}

/**
 * Removes the exit prevention logic
 * Call on component beforeUnmount
 */
export function removeExitPrevent() {
    ExitPreventState.currentChecker = null;
    ExitPreventState.currentExitFunc = null;
}

/**
 * Call for every exit action
 * @param callback The callback function to call in order to exit
 */
export function tryPreventableExit(callback: () => void) {
    if (!ExitPreventState.currentChecker || !ExitPreventState.currentExitFunc || !ExitPreventState.currentChecker()) {
        return callback();
    }

    return ExitPreventState.currentExitFunc(callback);
}
