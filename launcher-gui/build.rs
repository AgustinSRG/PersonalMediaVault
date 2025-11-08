// Build script

extern crate winres;

fn main() {
    let config =
        slint_build::CompilerConfiguration::new().with_bundled_translations("./translations");

    slint_build::compile_with_config("ui/main.slint", config).expect("Slint build failed");

    if cfg!(target_os = "windows") {
        let mut res = winres::WindowsResource::new();
        res.set_icon("./assets/favicon.ico");
        res.compile().unwrap();
    }
}
