// Focus trap

"use strict"

/**
 * Traps focus
 * Calls a handler when focus is lost
 */
export class FocusTrap {
    public element: Node;

    private focusHandler: (event: FocusEvent) => void;

    private exitHandler: () => void;

    private active: boolean;

    constructor(element: Node, onExitFocus: () => void) {
        this.element = element;
        this.focusHandler = this.handleFocus.bind(this);
        this.exitHandler = onExitFocus;
    }

    private handleFocus(event: FocusEvent) {
        if (!event.target) {
            return;
        }
        if (event.target !== this.element && !this.element.contains(<Node>event.target)) {
            console.log(event);
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

