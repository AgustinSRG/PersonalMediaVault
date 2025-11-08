use std::{
    fmt::Display,
    fs::{self, read_to_string},
    path::Path,
};

use base64::{prelude::BASE64_STANDARD, Engine};
use pmv_encryption_rs::{decrypt, encrypt, DecryptionError, EncryptionMethod};
use rand::{rngs::StdRng, RngCore, SeedableRng};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

/// Current method for hashing credentials
const VAULT_CRED_METHOD_AES_SHA256: &str = "aes256/sha256/salt16";

/// Size for the salt
const SALT_SIZE: usize = 16;

/// Size of an AES key
pub const AES_KEY_SIZE: usize = 32;

/// An extra account of the vault
#[derive(Debug, Serialize, Deserialize, Default, Clone)]
pub struct VaultCredentialsExtraAccount {
    #[serde(default)]
    pub user: String,

    #[serde(default)]
    pub method: String,

    #[serde(default, rename = "pwhash")]
    pub password_hash: String,

    #[serde(default)]
    pub salt: String,

    #[serde(default, rename = "enckey")]
    pub encrypted_key: String,

    #[serde(default)]
    pub write: bool,
}

/// The credentials of the vault
#[derive(Debug, Serialize, Deserialize, Default, Clone)]
pub struct VaultCredentials {
    #[serde(default)]
    pub user: String,

    #[serde(default)]
    pub method: String,

    #[serde(default, rename = "pwhash")]
    pub password_hash: String,

    #[serde(default)]
    pub salt: String,

    #[serde(default, rename = "enckey")]
    pub encrypted_key: String,

    #[serde(default)]
    pub fingerprint: String,

    #[serde(default)]
    pub accounts: Vec<VaultCredentialsExtraAccount>,
}

/// Key export error
#[derive(Debug, Clone)]
pub enum KeyExportError {
    /// Invalid password
    InvalidPassword,
    /// Credentials file error
    CredentialsError(String),
}

impl Display for KeyExportError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            KeyExportError::InvalidPassword => write!(f, "Invalid password"),
            KeyExportError::CredentialsError(error) => write!(f, "{}", error),
        }
    }
}

impl From<String> for KeyExportError {
    fn from(value: String) -> Self {
        KeyExportError::CredentialsError(value)
    }
}

impl From<DecryptionError> for KeyExportError {
    fn from(value: DecryptionError) -> Self {
        KeyExportError::CredentialsError(value.to_string())
    }
}

impl VaultCredentials {
    pub fn load_from_file<P>(path: P) -> Result<VaultCredentials, String>
    where
        P: AsRef<Path>,
    {
        let file_str = match read_to_string(path) {
            Ok(s) => s,
            Err(e) => {
                return Err(e.to_string());
            }
        };

        let config: VaultCredentials = match serde_json::from_str(&file_str) {
            Ok(c) => c,
            Err(e) => {
                return Err(e.to_string());
            }
        };

        Ok(config)
    }

    pub fn write_to_file<P>(&self, path: P) -> Result<(), String>
    where
        P: AsRef<Path>,
    {
        let file_str = match serde_json::to_string(self) {
            Ok(s) => s,
            Err(e) => {
                return Err(e.to_string());
            }
        };

        if let Err(e) = fs::write(path, file_str) {
            return Err(e.to_string());
        }

        Ok(())
    }

    pub fn extract_encryption_key(&self, password: &str) -> Result<Vec<u8>, KeyExportError> {
        match self.method.as_str() {
            VAULT_CRED_METHOD_AES_SHA256 => self.extract_encryption_key_aes_sha256(password),
            _ => Err(KeyExportError::CredentialsError(
                "Unrecognized credentials method".to_string(),
            )),
        }
    }

    fn get_salt(&self) -> Result<Vec<u8>, String> {
        match BASE64_STANDARD.decode(&self.salt) {
            Ok(b) => Ok(b),
            Err(_) => Err("Invalid salt value in credentials file".to_string()),
        }
    }

    fn get_password_hash(&self) -> Result<Vec<u8>, String> {
        match BASE64_STANDARD.decode(&self.password_hash) {
            Ok(b) => Ok(b),
            Err(_) => Err("Invalid password hash value in credentials file".to_string()),
        }
    }

    fn get_encrypted_key(&self) -> Result<Vec<u8>, String> {
        match BASE64_STANDARD.decode(&self.encrypted_key) {
            Ok(b) => Ok(b),
            Err(_) => Err("Invalid encrypted key value in credentials file".to_string()),
        }
    }

    fn extract_encryption_key_aes_sha256(&self, password: &str) -> Result<Vec<u8>, KeyExportError> {
        // Compute password hash

        let password_bytes: Vec<u8> = password.bytes().collect();
        let salt = self.get_salt()?;

        let password_hash = Sha256::digest([password_bytes, salt].concat());

        let password_double_hash = Sha256::digest(password_hash);

        // Compare with expected password hash

        let expected_password_hash = self.get_password_hash()?;

        if password_double_hash.as_slice() != expected_password_hash.as_slice() {
            return Err(KeyExportError::InvalidPassword);
        }

        // Decrypt

        let encrypted_key = self.get_encrypted_key()?;

        let decrypted_key = decrypt(&encrypted_key, &password_hash)?;

        Ok(decrypted_key)
    }

    fn generate_random_salt() -> Vec<u8> {
        let mut salt: Vec<u8> = vec![0; SALT_SIZE];
        let mut rng = StdRng::from_os_rng();
        rng.fill_bytes(&mut salt);

        salt
    }

    pub fn recover_key(&mut self, key: &[u8], password: &str) -> Result<(), String> {
        // Compute password hash

        let password_bytes: Vec<u8> = password.bytes().collect();
        let salt = Self::generate_random_salt();

        let password_hash = Sha256::digest([password_bytes, salt.clone()].concat());

        let password_double_hash = Sha256::digest(password_hash);

        // Encrypt key

        let encrypted_key = match encrypt(key, EncryptionMethod::Aes256Flat, &password_hash) {
            Ok(k) => k,
            Err(e) => {
                return Err(e.to_string());
            }
        };

        // Update fields

        self.method = VAULT_CRED_METHOD_AES_SHA256.to_string();
        self.password_hash = BASE64_STANDARD.encode(password_double_hash);
        self.salt = BASE64_STANDARD.encode(&salt);
        self.encrypted_key = BASE64_STANDARD.encode(&encrypted_key);
        self.accounts = Vec::new();

        Ok(())
    }
}
