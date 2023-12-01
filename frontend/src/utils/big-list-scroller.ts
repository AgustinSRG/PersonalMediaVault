// Big list scroller

"use strict";

const SCROLL_BUFFER_SIZE_MULTIPLIER = 5;

/**
 * Utility to scroll big lists
 */
export class BigListScroller<T = any> {
    /**
     * Computes the window size from the number
     * of item fitting in the container without scrolling
     * @param itemsFit Number of items fitting in the container
     */
    public static GetWindowSize(itemsFit: number): number {
        return itemsFit + 2 * Math.floor(itemsFit * SCROLL_BUFFER_SIZE_MULTIPLIER);
    }

    /**
     * Computes the window step
     * @param windowSize The window size
     * @returns The window step
     */
    public static GetWindowStep(windowSize: number): number {
        return Math.floor(windowSize / (SCROLL_BUFFER_SIZE_MULTIPLIER * 2 + 1)) || 1;
    }

    /**
     * Obtains the list window
     */
    public getListWindow: () => T[];

    /**
     * Sets the list window
     */
    public setListWindow: (list: T[]) => void;

    /**
     * Minimal window size
     */
    public minWindowSize: number;

    /**
     * Max number of elements in the list window
     */
    public windowSize: number;

    /**
     * Full list of elements
     */
    public list: T[];

    /**
     * Position of the window
     */
    public windowPosition: number;

    /**
     * Constructor
     * @param windowSize Max number of elements in the list window
     * @param maxPages Max number of pages to keep
     * @param callbacks Callbacks to get and set the list window
     */
    constructor(windowSize: number, callbacks: { get: () => T[]; set: (list: T[]) => void }) {
        this.windowSize = windowSize;
        this.minWindowSize = windowSize;
        this.getListWindow = callbacks.get;
        this.setListWindow = callbacks.set;
        this.list = [];
        this.windowPosition = 0;
    }

    /**
     * Gets the current center position of the window
     * @returns The window center position
     */
    public getCenterPosition(): number {
        return this.windowPosition + Math.floor(this.windowSize / 2);
    }

    /**
     * Checks if the window is at the end
     * @returns True if the window position is at the end
     */
    public isAtTheEnd(): boolean {
        return this.windowPosition + this.windowSize >= this.list.length;
    }

    /**
     * Resets the list
     * After you do this, make sure to get the new listWindow
     */
    public reset() {
        this.setListWindow([]);
        this.list = [];
        this.windowPosition = 0;
    }

    /**
     * Adds a page to the edge of the list
     * @param page Page number
     * @param elements Page elements
     */
    public addElements(elements: T[]) {
        const listWindow = this.getListWindow();

        for (let i = 0; i < elements.length; i++) {
            this.list.push(elements[i]);

            if (listWindow.length < this.windowSize) {
                listWindow.push(elements[i]);
            }
        }
    }

    /**
     * Moves list window down
     */
    private moveWindowDown() {
        const listWindow = this.getListWindow();
        const step = BigListScroller.GetWindowStep(this.windowSize);

        let windowNext = this.windowPosition + listWindow.length;
        let moveCount = 0;

        while (moveCount < step && windowNext < this.list.length) {
            listWindow.shift();
            listWindow.push(this.list[windowNext]);
            this.windowPosition++;
            windowNext++;
            moveCount++;
        }
    }

    /**
     * Moves list window up
     */
    private moveWindowUp() {
        const listWindow = this.getListWindow();
        const step = BigListScroller.GetWindowStep(this.windowSize);
        let moveCount = 0;

        while (moveCount < step && this.windowPosition > 0) {
            listWindow.pop();
            listWindow.unshift(this.list[this.windowPosition - 1]);
            this.windowPosition--;
            moveCount++;
        }
    }

    /**
     * Checks the scroll of the container
     * and moves the window accordingly
     * @param elem The HTML element containing the elements
     */
    public checkElementScroll(elem: HTMLElement) {
        const elementBounds = elem.getBoundingClientRect();

        const overflowLength = elem.scrollHeight - elementBounds.height;

        if (overflowLength < 1) {
            return;
        }

        const relScroll = elem.scrollTop / overflowLength;

        if (relScroll == 0) {
            if (this.windowPosition > 0) {
                this.moveWindowToElement(0);
            }
        } else if (relScroll <= 0.1) {
            this.moveWindowUp();
        } else if (relScroll >= 0.9) {
            this.moveWindowDown();
        }
    }

    /**
     * Moves window to specified index
     * @param index The index
     */
    public moveWindowToElement(index: number) {
        this.windowPosition = Math.max(0, index - Math.floor(this.windowSize / 2));
        this.setListWindow(this.list.slice(this.windowPosition, this.windowPosition + this.windowSize));
    }

    /**
     * Increases window size
     * @param newSize The new size
     */
    public changeWindowSize(newSize: number): boolean {
        if (newSize <= 0 || newSize === this.windowSize) {
            return false;
        }

        this.windowSize = newSize;

        return true;
    }

    /**
     * Checks container height and increases window size if necessary
     * @param container The container
     * @param anyItem Any of the items
     */
    public checkScrollContainerHeight(container: HTMLElement, anyItem: HTMLElement): boolean {
        const containerWidth = container.getBoundingClientRect().width;
        const containerHeight = container.getBoundingClientRect().height;

        const itemHeight = anyItem.getBoundingClientRect().height || 1;
        const itemWidth = anyItem.getBoundingClientRect().width || 1;

        const itemsFitWidth = Math.round(containerWidth / itemWidth) || 1;
        const itemsFitHeight = Math.round(containerHeight / itemHeight) || 1;

        const minSize = BigListScroller.GetWindowSize(itemsFitWidth * itemsFitHeight);

        return this.changeWindowSize(Math.max(this.minWindowSize, minSize));
    }
}
