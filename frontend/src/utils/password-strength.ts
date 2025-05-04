// Password strength calculator

"use strict";

/**
 * Password strength tag
 */
export type PasswordStrengthTag = "very-weak" | "weak" | "medium" | "strong" | "very-strong" | "crypto-secure";

/**
 * Gets password strength tag
 * @param strength The password strength
 * @returns The tag
 */
export function getPasswordStrengthTag(strength: number): PasswordStrengthTag {
    if (strength < 16) {
        return "very-weak";
    } else if (strength < 32) {
        return "weak";
    } else if (strength < 48) {
        return "medium";
    } else if (strength < 64) {
        return "strong";
    } else if (strength < 128) {
        return "very-strong";
    } else {
        return "crypto-secure";
    }
}

/**
 * Computes the password strength in bits
 * @param password The password
 * @returns The computed bits
 */
export function computePasswordStrength(password: string): number {
    // Transform the password to remove repeated sequences

    let tempPassword = password;

    let halfPasswordLength = Math.floor(password.length / 2);

    for (let i = halfPasswordLength; i > 0; i--) {
        tempPassword = removeRepeatedParts(tempPassword, i);
    }

    // Now, remove incremental sequences (123456789 for example)

    halfPasswordLength = Math.floor(tempPassword.length / 2);

    for (let i = halfPasswordLength; i > 0; i--) {
        tempPassword = removeIncrementalSequences(tempPassword, i);
    }

    // After the password is transformed, we calculate its entropy

    if (tempPassword.length === 0) {
        return 0;
    }

    // First, calculate the number of unique characters

    const charSet = new Set<string>();

    for (let i = 0; i < tempPassword.length; i++) {
        charSet.add(tempPassword.charAt(i));
    }

    const charSetSize = charSet.size;

    // Now, we calculate the number of bits necessary to represent one character

    const charBits = Math.log2(charSetSize) || 0;

    // Finally, we multiple by the number of characters

    return charBits * tempPassword.length;
}

/**
 * Removes repeated parts from a password
 * @param password The password
 * @param n The size of the parts
 * @returns The password with no repeated parts
 */
function removeRepeatedParts(password: string, n: number): string {
    const cycles = Math.ceil(password.length / n);

    let result = "";
    let previousPart = "";

    for (let i = 0; i < cycles; i++) {
        const part = password.substring(n * i, n * (i + 1));

        if (previousPart !== part) {
            previousPart = part;
            result += part;
        }
    }

    return result;
}

const INCREMENTAL_NUMBERS = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9"];
const INCREMENTAL_LETTERS = [
    "a",
    "b",
    "c",
    "d",
    "e",
    "f",
    "g",
    "h",
    "i",
    "j",
    "k",
    "l",
    "m",
    "n",
    "o",
    "p",
    "q",
    "r",
    "s",
    "t",
    "u",
    "v",
    "w",
    "x",
    "y",
    "z",
];
const INCREMENTAL_LETTERS_UPPER = INCREMENTAL_LETTERS.map((l) => l.toUpperCase());

/**
 * Computes the character distance, to distinguish incremental sequences
 * @param c1 Character 1
 * @param c2 Character 2
 * @returns The distance (-1 if not related)
 */
function charDistance(c1: string, c2: string): number {
    let i1: number;
    let i2: number;

    i1 = INCREMENTAL_NUMBERS.indexOf(c1);
    i2 = INCREMENTAL_NUMBERS.indexOf(c2);

    if (i1 !== -1 && i2 !== -1) {
        return Math.abs(i2 - i1);
    }

    i1 = INCREMENTAL_LETTERS.indexOf(c1);
    i2 = INCREMENTAL_LETTERS.indexOf(c2);

    if (i1 !== -1 && i2 !== -1) {
        return Math.abs(i2 - i1);
    }

    i1 = INCREMENTAL_LETTERS_UPPER.indexOf(c1);
    i2 = INCREMENTAL_LETTERS_UPPER.indexOf(c2);

    if (i1 !== -1 && i2 !== -1) {
        return Math.abs(i2 - i1);
    }

    return -1;
}

/**
 * Checks if 2 parts with the same size are in sequence
 * @param part1 The part 1
 * @param part2 The part 2
 * @returns True if the parts are in sequence
 */
function partsAreSequence(part1: string, part2: string): boolean {
    if (!part1 || !part2 || part1.length !== part2.length) {
        return false;
    }

    if (part1.length > 1) {
        const prefix1 = part1.substring(0, part1.length - 1);
        const prefix2 = part2.substring(0, part2.length - 1);

        if (prefix1 !== prefix2) {
            return false;
        }

        const lastCharPart1 = part1.charAt(part1.length - 1);
        const lastCharPart2 = part2.charAt(part2.length - 1);

        return charDistance(lastCharPart1, lastCharPart2) === 1;
    } else {
        return charDistance(part1, part2) === 1;
    }
}

/**
 * Removes incremental sequences from password
 * @param password The password
 * @param n The length of the parts
 * @returns The password with no sequences
 */
function removeIncrementalSequences(password: string, n: number): string {
    const cycles = Math.ceil(password.length / n);

    let result = "";
    let previousPart = "";

    for (let i = 0; i < cycles; i++) {
        const part = password.substring(n * i, n * (i + 1));

        if (!partsAreSequence(previousPart, part)) {
            result += part;
        }

        previousPart = part;
    }

    return result;
}
