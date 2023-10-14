// Big list scroller

"use strict";

/**
 * Utility to scroll big lists
 */
export class BigListScroller<T = any> {
    /**
     * Obtains the list window
     */
    public getListWindow: () => T[];

    /**
     * Sets the list window
     */
    public setListWindow: (list: T[]) => void;

    /**
     * Max number of elements in the list window
     */
    public windowSize: number;

    /**
     * Full list of elements
     */
    private list: T[];

    /**
     * Position of the window
     */
    private windowPosition: number;

    /**
     * Constructor
     * @param windowSize Max number of elements in the list window
     * @param maxPages Max number of pages to keep
     * @param callbacks Callbacks to get and set the list window
     */
    constructor(windowSize: number, callbacks: {get: () => T[], set:  (list: T[]) => void}) {
        this.windowSize = windowSize;
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
        const step = Math.floor(this.windowSize / 4) || 1;

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
        const step = Math.floor(this.windowSize / 4) || 1;
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


        if (relScroll <= 0.1) {
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

        const itemsFitWidth = Math.floor(containerWidth / itemWidth) || 1;
        const itemsFitHeight = Math.floor(containerHeight / itemHeight) || 1;
        
        const minSize = itemsFitWidth * itemsFitHeight * 11;

        return this.changeWindowSize(minSize);
    }
}
