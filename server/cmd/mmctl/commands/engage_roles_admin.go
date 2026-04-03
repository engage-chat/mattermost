// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost/server/v8/cmd/mmctl/client"
	"github.com/mattermost/mattermost/server/v8/cmd/mmctl/printer"
)

var RolesEngageAdminCmd = &cobra.Command{
	Use:     "engage-admin [users]",
	Aliases: []string{"engage_admin"},
	Short:   "Set a user as engage admin",
	Long:    "Assign the system_engage_admin role to one or more users. The role is created if it does not yet exist.",
	Example: `  # Assign engage admin role to a single user
  $ mmctl roles engage-admin john_doe

  # Or assign to multiple users at the same time
  $ mmctl roles engage-admin john_doe jane_doe`,
	RunE: withClient(rolesEngageAdminCmdF),
	Args: cobra.MinimumNArgs(1),
}

func init() {
	RolesCmd.AddCommand(RolesEngageAdminCmd)
}

func rolesEngageAdminCmdF(c client.Client, _ *cobra.Command, args []string) error {
	if _, _, err := c.EnableCustomRoles(context.TODO(), []string{model.SystemEngageAdmin}); err != nil {
		return fmt.Errorf("unable to enable %q role: %w", model.SystemEngageAdmin, err)
	}

	var errs *multierror.Error
	users := getUsersFromUserArgs(c, args)
	for i, user := range users {
		if user == nil {
			userErr := fmt.Errorf("unable to find user %q", args[i])
			errs = multierror.Append(errs, userErr)
			printer.PrintError(userErr.Error())
			continue
		}

		alreadyEngageAdmin := false
		roles := strings.Fields(user.Roles)
		for _, role := range roles {
			if role == model.SystemEngageAdmin {
				alreadyEngageAdmin = true
			}
		}

		if !alreadyEngageAdmin {
			roles = append(roles, model.SystemEngageAdmin)
			if _, err := c.UpdateUserRoles(context.TODO(), user.Id, strings.Join(roles, " ")); err != nil {
				updateErr := fmt.Errorf("can't update roles for user %q: %w", args[i], err)
				errs = multierror.Append(errs, updateErr)
				printer.PrintError(updateErr.Error())
				continue
			}

			printer.Print(fmt.Sprintf("Engage admin role assigned to user %q. Current roles are: %s", args[i], strings.Join(roles, ", ")))
		}
	}

	return errs.ErrorOrNil()
}
