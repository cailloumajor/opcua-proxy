pub mod api {
    tonic::include_proto!("centrifugal.centrifugo.api");
}

pub mod proxy {
    tonic::include_proto!("centrifugal.centrifugo.proxy");
}
