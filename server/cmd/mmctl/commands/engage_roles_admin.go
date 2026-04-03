// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost/server/v8/cmd/mmctl/client"
	"github.com/mattermost/mattermost/server/v8/cmd/mmctl/printer"
)

var RolesEngageAdminCmd = &cobra.Command{
	Use:     "engage-admin",
	Aliases: []string{"engage_admin"},
	Short:   "Enable the engage admin role",
	Long:    "Create or restore the system_engage_admin custom role.",
	Example: `  $ mmctl roles engage-admin`,
	RunE:    withClient(rolesEngageAdminCmdF),
}

func init() {
	RolesCmd.AddCommand(RolesEngageAdminCmd)
}

func rolesEngageAdminCmdF(c client.Client, _ *cobra.Command, _ []string) error {
	if _, _, err := c.EnableCustomRoles(context.TODO(), []string{model.SystemEngageAdmin}); err != nil {
		return fmt.Errorf("unable to enable %q role: %w", model.SystemEngageAdmin, err)
	}

	printer.Print(fmt.Sprintf("Role %q enabled successfully.", model.SystemEngageAdmin))
	return nil
}
