/*
 * Main file
 */

#include "main_win.h"
#include <wx/msgdlg.h>

using namespace std;

int mainEntryPoint(App &app, int argc, char *argv[]);

wxIMPLEMENT_APP(App);
// clang-format on

bool App::OnInit()
{
    if (mainEntryPoint(*this, this->argc, this->argv) != 0) {
        this->Exit();
    }

    return true;
}

int mainEntryPoint(App &app, int argc, char *argv[])
{
     wxMessageBox(wxString("Hello world"), wxT("Test"), wxICON_ERROR);
    return 0;
}
