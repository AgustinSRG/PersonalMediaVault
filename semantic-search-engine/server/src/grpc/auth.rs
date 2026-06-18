// Auth for GRPC server

use crate::string_compare_time_safe;

pub struct GrpcServerAuth {
    api_key: String,
}

impl GrpcServerAuth {
    pub fn new(api_key: String) -> GrpcServerAuth {
        GrpcServerAuth { api_key }
    }

    pub fn check_key(&self, key: &str) -> bool {
        string_compare_time_safe(&self.api_key, key)
    }
}
