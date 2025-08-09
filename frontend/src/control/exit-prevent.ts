// Exit preventer

"use strict";

/**
 * Exit preventer
 */
export class ExitPreventer {
    /**
     * Current function to call to check if if the exit should be prevented
     */
    public static CurrentChecker: (() => boolean) | null = null;

    /**
     * Current function to call before running any exit action
     */
    public static CurrentExitFunc: ((callback: () => void) => void) | null = null;

    /**
     * Initialization logic
     */
    public static Initialize() {
        window.addEventListener("beforeunload", function (e: BeforeUnloadEvent) {
            if (ExitPreventer.CurrentChecker && ExitPreventer.CurrentChecker()) {
                // Cancel the event
                e.preventDefault(); // If you prevent default behavior in Mozilla Firefox prompt will always be shown
                // Chrome requires returnValue to be set
                e.returnValue = "";
            }
        });
    }

    /**
     * Set up the functions to manage the exit prevention
     * @param checker Checker function
     * @param exitFunc Exit handler function
     */
    public static SetupExitPrevent(checker: () => boolean, exitFunc: (callback: () => void) => void) {
        ExitPreventer.CurrentChecker = checker;
        ExitPreventer.CurrentExitFunc = exitFunc;
    }

    /**
     * Removes the exit prevention logic
     * Call on component beforeUnmount
     */
    public static RemoveExitPrevent() {
        ExitPreventer.CurrentChecker = null;
        ExitPreventer.CurrentExitFunc = null;
    }

    /**
     * Call for every exit action
     * @param callback The callback function to call in order to exit
     */
    public static TryExit(callback: () => void) {
        if (!ExitPreventer.CurrentChecker || !ExitPreventer.CurrentExitFunc || !ExitPreventer.CurrentChecker()) {
            return callback();
        }

        return ExitPreventer.CurrentExitFunc(callback);
    }
}

ExitPreventer.Initialize();
