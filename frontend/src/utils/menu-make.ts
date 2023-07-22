// Menu creation for pagination result

"use strict";

const MAX_MENU_ITEMS = 9;

interface PageNode {
    type: "page" | "skip";
    current: boolean;
    page: number;
}

export function generateMenuForPages(page: number, totalPages: number): PageNode[] {
    if (totalPages <= 0 || page < 0 || page >= totalPages) {
        return [];
    }

    const currentPageNode: PageNode[] = [
        {
            type: "page",
            current: true,
            page: page,
        },
    ];

    const leftNodes: PageNode[] = [];
    const rightNodes: PageNode[] = [];

    let remainingNodes = MAX_MENU_ITEMS - 1;

    let amountNodesLeft = Math.max(0, page);
    let amountNodesRight = Math.max(0, totalPages - page - 1);

    if (amountNodesLeft > 0 && amountNodesRight > 0) {
        if (amountNodesLeft + amountNodesRight <= remainingNodes) {
            for (let i = 0; i < amountNodesLeft; i++) {
                leftNodes.push({
                    type: "page",
                    current: false,
                    page: i,
                });
            }
            for (let i = 0; i < amountNodesRight; i++) {
                rightNodes.unshift({
                    type: "page",
                    current: false,
                    page: totalPages - 1 - i,
                });
            }
        } else {
            amountNodesLeft--;
            amountNodesRight--;

            remainingNodes -= 2;

            let swap = false;
            let currentRight = page + 1;
            let currentLeft = page - 1;

            while (remainingNodes > 2) {
                swap = !swap;

                if (swap) {
                    if (amountNodesRight > 0) {
                        rightNodes.push({
                            type: "page",
                            current: false,
                            page: currentRight,
                        });
                        currentRight++;
                        amountNodesRight--;
                        remainingNodes--;
                    }
                } else {
                    if (amountNodesLeft > 0) {
                        leftNodes.unshift({
                            type: "page",
                            current: false,
                            page: currentLeft,
                        });
                        currentLeft--;
                        amountNodesLeft--;
                        remainingNodes--;
                    }
                }
            }

            if (amountNodesLeft > 1 && amountNodesRight > 1) {
                leftNodes.unshift({
                    type: "skip",
                    current: false,
                    page: 0,
                });
                rightNodes.push({
                    type: "skip",
                    current: false,
                    page: totalPages - 1,
                });
            } else if (amountNodesLeft > 1) {
                if (amountNodesRight === 1) {
                    leftNodes.unshift({
                        type: "skip",
                        current: false,
                        page: 0,
                    });
                    rightNodes.push({
                        type: "page",
                        current: false,
                        page: currentRight,
                    });
                } else if (amountNodesLeft > 2) {
                    leftNodes.unshift({
                        type: "page",
                        current: false,
                        page: currentLeft,
                    });
                    leftNodes.unshift({
                        type: "skip",
                        current: false,
                        page: 0,
                    });
                } else {
                    leftNodes.unshift({
                        type: "page",
                        current: false,
                        page: currentLeft,
                    });
                    leftNodes.unshift({
                        type: "page",
                        current: false,
                        page: currentLeft - 1,
                    });
                }
            } else if (amountNodesRight > 1) {
                if (amountNodesLeft === 1) {
                    leftNodes.unshift({
                        type: "page",
                        current: false,
                        page: currentLeft,
                    });
                    rightNodes.push({
                        type: "skip",
                        current: false,
                        page: totalPages - 1,
                    });
                } else if (amountNodesRight > 2) {
                    rightNodes.push({
                        type: "page",
                        current: false,
                        page: currentRight,
                    });
                    rightNodes.push({
                        type: "skip",
                        current: false,
                        page: totalPages - 1,
                    });
                } else {
                    rightNodes.push({
                        type: "page",
                        current: false,
                        page: currentRight,
                    });
                    rightNodes.push({
                        type: "page",
                        current: false,
                        page: currentRight + 1,
                    });
                }
            }

            rightNodes.push({
                type: "page",
                current: false,
                page: totalPages - 1,
            });

            leftNodes.unshift({
                type: "page",
                current: false,
                page: 0,
            });
        }
    } else if (amountNodesLeft > 0) {
        if (amountNodesLeft <= remainingNodes) {
            for (let i = 0; i < amountNodesLeft; i++) {
                leftNodes.push({
                    type: "page",
                    current: false,
                    page: i,
                });
            }
        } else {
            leftNodes.push({
                type: "page",
                current: false,
                page: 0,
            });
            leftNodes.push({
                type: "skip",
                current: false,
                page: 0,
            });

            remainingNodes -= 2;

            for (let i = 0; i < remainingNodes; i++) {
                leftNodes.push({
                    type: "page",
                    current: false,
                    page: page - remainingNodes + i,
                });
            }
        }
    } else if (amountNodesRight > 0) {
        if (amountNodesRight <= remainingNodes) {
            for (let i = 0; i < amountNodesRight; i++) {
                rightNodes.unshift({
                    type: "page",
                    current: false,
                    page: totalPages - 1 - i,
                });
            }
        } else {
            remainingNodes -= 2;

            for (let i = 0; i < remainingNodes; i++) {
                rightNodes.push({
                    type: "page",
                    current: false,
                    page: page + i + 1,
                });
            }

            rightNodes.push({
                type: "skip",
                current: false,
                page: totalPages - 1,
            });

            rightNodes.push({
                type: "page",
                current: false,
                page: totalPages - 1,
            });
        }
    }

    return leftNodes.concat(currentPageNode).concat(rightNodes);
}
