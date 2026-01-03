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
    utils::command_no_window,
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
    username: String,
    password: String,
) {
    cancel_vault_tool(status);

    let args: Vec<String> = match tool {
        VaultSelectedTool::None => {
            return;
        }
        VaultSelectedTool::Clean => vec![
            "--clean".to_string(),
            "--remove-trash".to_string(),
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

    let mut cmd = cmd(status.daemon_binary.clone(), args)
        .stderr_to_stdout()
        .stdout_capture()
        .before_spawn(command_no_window)
        .stdin_null();

    match tool {
        VaultSelectedTool::Clean => {
            cmd = cmd
                .env("VAULT_USER", username)
                .env("VAULT_PASSWORD", password)
        }
        _ => {}
    }

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

fn find_tool_error_in_output(o: &str) -> Option<String> {
    let lines: Vec<&str> = o
        .split("\n")
        .map(|l| l.trim())
        .filter(|l| l.len() > 0)
        .collect();

    for line in lines {
        let line_parts: Vec<&str> = line.split("[ERROR]").collect();

        if line_parts.len() < 2 {
            continue;
        }

        let err_msg_joined = line_parts[1..].join("[ERROR]");

        let err_msg_trim = err_msg_joined.trim();

        return Some(err_msg_trim.to_string());
    }

    None
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

        match r {
            Ok(o) => {
                let out_utf8 = String::from_utf8_lossy(&o.stdout).to_string();

                let err_option = find_tool_error_in_output(&out_utf8);

                match err_option {
                    Some(err_msg) => {
                        let _ = sender.send(LauncherWorkerMessage::ToolError {
                            tool_id,
                            error_details: err_msg,
                        });
                    }
                    None => {
                        let _ = sender.send(LauncherWorkerMessage::ToolSuccess { tool_id });
                    }
                }
            }
            Err(e) => {
                let _ = sender.send(LauncherWorkerMessage::ToolError {
                    tool_id,
                    error_details: e.to_string(),
                });
            }
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

    run_vault(status, sender, window_handle, false);
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

    run_vault(status, sender, window_handle, false);
}
