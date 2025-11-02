
const UNITS: &'static str = "kMGTPE";
const UNIT: u64 = 1024;

pub fn display_size(bytes: u64) -> String {
    if bytes < UNIT {
        return format!("{} B", bytes)
    }

    let mut div = UNIT;
    let mut exp: usize = 0;

    let mut n = bytes / UNIT;

    while n >= UNIT {
        div *= UNIT;
        exp += 1;
        n /= UNIT;
    }

    let result = (bytes as f64) / (div as f64);
    let unit = UNITS.chars().nth(exp).unwrap_or('?');

    format!("{:.1} {}B", result, unit)
}