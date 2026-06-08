use std::time::Duration;

use tokio::sync::{mpsc, oneshot};

type RequestPayload<S, R> = (S, oneshot::Sender<R>);

/// Represents the sending side of a roundtrip channel.
pub(crate) struct RoundtripSender<S, R> {
    inner: mpsc::Sender<RequestPayload<S, R>>,
}

impl<S, R> RoundtripSender<S, R> {
    /// Execute the roundtrip, i.e. send the request and get the response.
    pub(crate) async fn roundtrip(
        &self,
        request: S,
        send_timeout: Duration,
        receive_timeout: Duration,
    ) -> Result<R, String> {
        let (reply_tx, reply_rx) = oneshot::channel();
        self.inner
            .send_timeout((request, reply_tx), send_timeout)
            .await
            .map_err(|err| format!("error sending request: {err}"))?;
        let reply = tokio::time::timeout(receive_timeout, reply_rx)
            .await
            .map_err(|_| "reply receiving timeout".to_string())?
            .map_err(|err| format!("error receiving reply: {err}"))?;
        Ok(reply)
    }
}

/// Create and return a roundtrip channel, i.e. a tuple of [`RoundtripSender`] and [`mpsc::Receiver`].
pub(crate) fn roundtrip_channel<S, R>(
    buffer: usize,
) -> (RoundtripSender<S, R>, mpsc::Receiver<RequestPayload<S, R>>) {
    let (inner, rx) = mpsc::channel(buffer);

    let sender = RoundtripSender { inner };

    (sender, rx)
}

#[cfg(test)]
mod tests {
    use super::*;

    mod roundtrip {
        use super::*;

        #[tokio::test]
        async fn request_receiver_dropped() {
            let (tx, _) = roundtrip_channel::<(), ()>(1);
            let result = tx
                .roundtrip((), Duration::from_millis(100), Duration::from_millis(200))
                .await;
            assert!(result.is_err());
        }

        #[tokio::test]
        async fn request_send_timeout() {
            let (tx, _rx) = roundtrip_channel::<(), ()>(1);
            let (reply_tx, _) = oneshot::channel::<()>();
            tx.inner.send(((), reply_tx)).await.unwrap();
            let result = tx
                .roundtrip((), Duration::from_millis(100), Duration::from_millis(200))
                .await;
            assert!(result.is_err());
        }

        #[tokio::test]
        async fn reply_sender_dropped() {
            let (tx, mut rx) = roundtrip_channel::<(), ()>(1);
            tokio::spawn(async move {
                let (_, _) = rx.recv().await.unwrap();
            });
            let result = tx
                .roundtrip((), Duration::from_millis(100), Duration::from_millis(200))
                .await;
            assert!(result.is_err());
        }

        #[tokio::test]
        async fn reply_timeout() {
            let (tx, mut rx) = roundtrip_channel::<(), ()>(1);
            tokio::spawn(async move {
                let (_, _reply_tx) = rx.recv().await.unwrap();
                tokio::time::sleep(Duration::from_millis(120)).await;
            });
            let result = tx
                .roundtrip((), Duration::from_millis(200), Duration::from_millis(100))
                .await;
            assert!(result.is_err());
        }

        #[tokio::test]
        async fn success() {
            let (tx, mut rx) = roundtrip_channel::<u8, u8>(1);
            tokio::spawn(async move {
                let (request, reply_tx) = rx.recv().await.unwrap();
                assert_eq!(request, 54);
                reply_tx.send(63).unwrap();
            });
            let response = tx
                .roundtrip(54, Duration::from_millis(100), Duration::from_millis(200))
                .await
                .unwrap();
            assert_eq!(response, 63);
        }
    }
}
