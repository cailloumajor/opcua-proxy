use tonic_prost_build::configure;

const API_PROTO_FILE: &str = "proto/centrifugo/api.proto";
const PROXY_PROTO_FILE: &str = "proto/centrifugo/proxy.proto";

fn main() {
    configure()
        .build_server(false)
        .compile_protos(&[API_PROTO_FILE], &["proto/centrifugo"])
        .expect("compiling protobuf should not fail");

    configure()
        .build_client(false)
        .generate_default_stubs(true)
        .compile_protos(&[PROXY_PROTO_FILE], &["proto/centrifugo"])
        .expect("compiling protobuf should not fail");

    println!("cargo:rerun-if-changed={API_PROTO_FILE}");
    println!("cargo:rerun-if-changed={PROXY_PROTO_FILE}");
}
