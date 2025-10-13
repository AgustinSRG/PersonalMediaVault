use std::{
    sync::{
        mpsc::{channel, Sender},
        Arc,
    },
    thread::{self, sleep},
    time::Duration,
};

use duct::{cmd, Handle};
use slint::Weak;

use crate::{
    worker::{run_vault, LauncherWorkerMessage, WorkerThreadStatus},
    MainWindow, VaultSelectedTool, VaultToolStatus,
};

pub fn cancel_vault_tool(status: &mut WorkerThreadStatus) {
    status.tool_id = status.tool_id.wrapping_add(1);

    if let Some(p) = &status.tool_process {
        let _ = p.kill();
    }

    if let Some(r) = &status.tool_process_wait_receiver {
        let _ = r.recv();
        status.tool_process_wait_receiver = None;
    }
}

pub fn run_vault_tool(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    tool: VaultSelectedTool,
) {
    cancel_vault_tool(status);

    let args: Vec<String> = match tool {
        VaultSelectedTool::None => {
            return;
        }
        VaultSelectedTool::Clean => vec![
            "--clean".to_string(),
            "--fix-consistency".to_string(),
            "--skip-lock".to_string(),
            "--vault-path".to_string(),
            status.vault_path.clone(),
        ],
        VaultSelectedTool::Recover => vec![
            "--recover".to_string(),
            "--skip-lock".to_string(),
            "--vault-path".to_string(),
            status.vault_path.clone(),
        ],
    };

    let cmd = cmd(status.daemon_binary.clone(), args)
        .stderr_null()
        .stdout_null()
        .stdin_null();

    let handle = match cmd.start() {
        Ok(p) => p,
        Err(e) => {
            let err_str = e.to_string();
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_tool_status(VaultToolStatus::Error);
                win.set_tool_error(err_str.into());
            });
            return;
        }
    };

    let tool_process = Arc::new(handle);

    let (tool_process_wait_sender, tool_process_wait_receiver) = channel::<bool>();

    status.tool_id = status.tool_id.wrapping_add(1);
    let tool_id = status.tool_id;

    wait_for_tool_process(
        sender.clone(),
        tool_process.clone(),
        tool_process_wait_sender,
        tool_id,
    );

    status.tool_process = Some(tool_process);
    status.tool_process_wait_receiver = Some(tool_process_wait_receiver);

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_tool_status(VaultToolStatus::Running);
    });
}

pub fn wait_for_tool_process(
    sender: Sender<LauncherWorkerMessage>,
    process: Arc<Handle>,
    daemon_process_wait_sender: Sender<bool>,
    tool_id: u64,
) {
    thread::spawn(move || {
        sleep(Duration::from_millis(1000));

        let r = process.wait();

        let _ = daemon_process_wait_sender.send(true);

        if let Err(e) = r {
            let _ = sender.send(LauncherWorkerMessage::ToolError {
                tool_id,
                error_details: e.to_string(),
            });
        } else {
            let _ = sender.send(LauncherWorkerMessage::ToolSuccess { tool_id });
        }
    });
}

pub fn on_tool_success(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    tool_id: u64,
) {
    if status.tool_id != tool_id {
        return;
    }

    status.tool_process = None;

    if let Some(r) = &status.tool_process_wait_receiver {
        let _ = r.recv();
        status.tool_process_wait_receiver = None;
    }

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_tool_status(VaultToolStatus::Success);
    });

    run_vault(status, sender, window_handle);
}

pub fn on_tool_error(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    tool_id: u64,
    error_details: String,
) {
    if status.tool_id != tool_id {
        return;
    }

    status.tool_process = None;

    if let Some(r) = &status.tool_process_wait_receiver {
        let _ = r.recv();
        status.tool_process_wait_receiver = None;
    }

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_tool_status(VaultToolStatus::Error);
        win.set_tool_error(error_details.into());
    });

    run_vault(status, sender, window_handle);
}
