use std::sync::{
    mpsc::{channel, Receiver, Sender},
    Arc, Mutex,
};

pub struct CancellableTaskStatus {
    /// Task cancelled
    pub cancelled: bool,

    /// Task ended
    pub ended: bool,
}

impl CancellableTaskStatus {
    pub fn new() -> Self {
        Self {
            cancelled: false,
            ended: false,
        }
    }
}

#[derive(Clone)]
pub struct CancellableTaskController {
    /// Status
    status: Arc<Mutex<CancellableTaskStatus>>,

    /// Sender for end message
    end_sender: Sender<()>,

    /// Receiver for end message
    end_receiver: Arc<Mutex<Receiver<()>>>,
}

impl CancellableTaskController {
    /// Creates a new instance of CancellableTaskController
    pub fn new() -> Self {
        let (end_sender, end_receiver) = channel();
        Self {
            status: Arc::new(Mutex::new(CancellableTaskStatus::new())),
            end_sender,
            end_receiver: Arc::new(Mutex::new(end_receiver)),
        }
    }

    /// Checks if the task is cancelled
    pub fn is_cancelled(&self) -> bool {
        let status = self.status.lock().unwrap();
        status.cancelled
    }

    /// Cancels the task
    pub fn cancel(&self) {
        let mut status = self.status.lock().unwrap();

        if status.ended || status.cancelled {
            return;
        }

        status.cancelled = true;

        drop(status);

        let end_receiver = self.end_receiver.lock().unwrap();

        let _ = end_receiver.recv();
    }

    /// Indicates the ending of the task
    pub fn end(&self) {
        let mut status = self.status.lock().unwrap();

        if status.ended {
            return;
        }

        status.ended = true;

        let was_cancelled = status.cancelled;

        drop(status);

        if was_cancelled {
            let _ = self.end_sender.send(());
        }
    }
}
