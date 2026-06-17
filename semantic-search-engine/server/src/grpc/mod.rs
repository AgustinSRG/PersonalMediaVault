pub mod api {
    tonic::include_proto!("pmv.sse");
}

mod server;
pub use server::*;

mod methods;
