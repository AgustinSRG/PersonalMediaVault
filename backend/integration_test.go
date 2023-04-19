// Integration test (API calls)

package main

import "testing"

func TestAPIIntegration(t *testing.T) {
	server, err := GetTestServer()

	if err != nil {
		t.Error(err)
		panic(err)
	}

	defer server.Close()

	/* Login */

	session, _, err := LoginTest(server)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	/* Logout */

	err = LogoutTest(server, session)

	if err != nil {
		t.Error(err)
		panic(err)
	}
}
