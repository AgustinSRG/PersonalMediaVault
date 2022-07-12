# CMakeList.txt: Project
cmake_minimum_required (VERSION 3.8)

project ("PersonalMediaVault")

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# hack for buggy CMake's FindwxWidgets script
if (DEFINED ENV_WX_CONFIG)
    set (ENV{WX_CONFIG} ${ENV_WX_CONFIG})
endif()

set(wxWidgets_USE_STATIC 1)
SET(wxWidgets_USE_LIBS)

find_package(wxWidgets REQUIRED)
include(${wxWidgets_USE_FILE})

SET(SOURCES_RC 
    "resources/resources.rc"
)

# Win app
add_executable (PersonalMediaVault WIN32 
    "main_win.cpp" "main_win.h"
    ${SOURCES_RC}
)

target_link_libraries(PersonalMediaVault ${wxWidgets_LIBRARIES})
