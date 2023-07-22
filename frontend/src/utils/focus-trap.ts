// Focus trap

"use strict";

/**
 * Traps focus
 * Calls a handler when focus is lost
 */
export class FocusTrap {
    public element: Node;

    private focusHandler: (event: FocusEvent) => void;

    private exitHandler: () => void;

    private active: boolean;

    private exceptClass: string;

    constructor(element: Node, onExitFocus: () => void, exceptClass?: string) {
        this.element = element;
        this.focusHandler = this.handleFocus.bind(this);
        this.exitHandler = onExitFocus;
        this.exceptClass = exceptClass || "";
    }

    private handleFocus(event: FocusEvent) {
        if (!event.target) {
            return;
        }
        if (this.exceptClass && (<Element>event.target).classList.contains(this.exceptClass)) {
            return;
        }
        if (event.target !== this.element && !this.element.contains(<Node>event.target)) {
            this.exitHandler();
        }
    }

    public activate() {
        if (this.active) {
            return;
        }
        this.active = true;
        document.addEventListener("focus", this.focusHandler, true);
    }

    public deactivate() {
        if (!this.active) {
            return;
        }
        this.active = false;
        document.removeEventListener("focus", this.focusHandler, true);
    }

    public destroy() {
        this.deactivate();
    }
}
