# Personal media vault (GUI Launcher)

This laucnher is meant for local versions of personal media vault. It allows the user to open vaults with a GUI application.

# Building from source code

In order to compile the project, you'll need the following:

 - A C++ compiler
 - [CMAKE](https://cmake.org/install/) installed.
 - [wxWidgets](https://www.wxwidgets.org/) installed and available for your C++ compiler.

In order to build the release version with CMAKE use:

```sh
cmake -DCMAKE_BUILD_TYPE=Release -Ssrc -Brelease

cmake --build release --config Release
```

After it is successfully built, your binaries will be available in the `release` folder.
