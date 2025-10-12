// Build script

extern crate winres;

fn main() {
    slint_build::compile("ui/main.slint").expect("Slint build failed");

    if cfg!(target_os = "windows") {
        let mut res = winres::WindowsResource::new();
        res.set_icon("./assets/favicon.ico");
        res.compile().unwrap();
    }
}
