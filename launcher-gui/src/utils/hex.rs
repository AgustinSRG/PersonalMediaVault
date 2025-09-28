// Hexadecimal notation utils

/// Turns a byte array into a hexadecimal string
pub fn to_hex_string(bytes: &[u8]) -> String {
  let strs: Vec<String> = bytes.iter()
                               .map(|b| format!("{:02x}", b))
                               .collect();
  strs.join("")
}

#[cfg(test)]
mod test {
  use super::*;

  #[test]
  fn test_to_hex_string() {
    let bytes: Vec<u8> = vec![0xFF, 0, 0xAA];
    let actual = to_hex_string(&bytes);
    assert_eq!("ff00aa", actual);
  }
}