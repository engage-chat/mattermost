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

var EngageChatCmd = &cobra.Command{
	Use:   "engage-chat",
	Short: "Management of engage-chat features",
}

var EngageChatSystemEngageAdminCmd = &cobra.Command{
	Use:     "system-engage-admin",
	Aliases: []string{"system_engage_admin"},
	Short:   "Enable the system_engage_admin role",
	Long:    "Create or restore the system_engage_admin custom role.",
	Example: `  $ mmctl engage-chat system-engage-admin`,
	RunE:    withClient(engageChatEnableRolesCmdF),
}

func init() {
	EngageChatCmd.AddCommand(EngageChatSystemEngageAdminCmd)
	RootCmd.AddCommand(EngageChatCmd)
}

func engageChatEnableRolesCmdF(c client.Client, _ *cobra.Command, _ []string) error {
	if _, _, err := c.EnableCustomRoles(context.TODO(), []string{model.SystemEngageAdmin}); err != nil {
		return fmt.Errorf("unable to enable engage-chat roles: %w", err)
	}

	printer.Print(fmt.Sprintf("Role %q enabled successfully.", model.SystemEngageAdmin))
	return nil
}
