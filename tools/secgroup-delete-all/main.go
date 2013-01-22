package main

import (
	"fmt"
	"io"
	"launchpad.net/gnuflag"
	"launchpad.net/goose/client"
	"launchpad.net/goose/identity"
	"launchpad.net/goose/nova"
	"os"
)

// DeleteAll destroys all security groups except the default
func DeleteAll(w io.Writer, osn *nova.Client) (err error) {
	groups, err := osn.ListSecurityGroups()
	if err != nil {
		return err
	}
	deleted := 0
	failed := 0
	for _, group := range groups {
		if group.Name != "default" {
			err := osn.DeleteSecurityGroup(group.Id)
			if err != nil {
				failed += 1
			} else {
				deleted += 1
			}
		}
	}
	if deleted != 0 {
		fmt.Fprintf(w, "%d security groups deleted.\n", deleted)
	} else if failed == 0 {
		fmt.Fprint(w, "No security groups to delete.\n")
	}
	if failed != 0 {
		fmt.Fprintf(w, "%d security groups could not be deleted.\n", failed)
	}
	return nil
}

func createNovaClient(authMode identity.AuthMethod) (osn *nova.Client, err error) {
	creds, err := identity.CompleteCredentialsFromEnv()
	if err != nil {
		return nil, err
	}
	osc := client.NewClient(creds, authMode, nil)
	return nova.New(osc), nil
}

var authModeFlag = gnuflag.String("auth-mode", "userpass", "type of authentication to use")

var authModes = map[string]identity.AuthMethod{
	"userpass": identity.AuthUserPass,
	"legacy":   identity.AuthLegacy,
}

func main() {
	gnuflag.Parse(true)
	mode, ok := authModes[*authModeFlag]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: no such auth-mode %q\n", *authModeFlag)
		os.Exit(1)
	}
	novaclient, err := createNovaClient(mode)
	if err == nil {
		err = DeleteAll(os.Stdout, novaclient)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
