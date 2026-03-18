// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// GetCustomRolesForGroup retrieves custom roles from the database.
// If customRoleGroup is empty, all custom roles are returned.
// Otherwise, only roles from the specified group are returned.
// This will return roles even if they are soft-deleted.
func (a *App) GetCustomRolesForGroup(c request.CTX, customRoleGroup string) ([]*model.Role, *model.AppError) {
	var targetRoleNames []string

	if customRoleGroup == "" {
		for _, group := range model.AllCustomRoleGroups() {
			targetRoleNames = append(targetRoleNames, model.CustomRoleNamesForGroup(group)...)
		}
	} else {
		targetRoleNames = model.CustomRoleNamesForGroup(customRoleGroup)
	}

	return a.GetRolesByNames(targetRoleNames)
}

// EnableCustomRoles ensures that the custom roles for the given group are active.
// It creates them if they don't exist, or restores them if they were soft-deleted.
func (a *App) EnableCustomRoles(c request.CTX, customRoleGroup string) ([]*model.Role, *model.AppError) {
	customRoleNames := model.CustomRoleNamesForGroup(customRoleGroup)
	if len(customRoleNames) == 0 {
		return []*model.Role{}, nil
	}
	// Initialize custom role templates
	customRoleTemplates := model.MakeTunagCustomRoles(customRoleGroup)

	// Get existing roles from the DB in a single batch.
	existingRoles, err := a.GetRolesByNames(customRoleNames)
	if err != nil {
		return nil, err
	}

	roleMap := make(map[string]*model.Role)
	for _, role := range existingRoles {
		roleMap[role.Name] = role
	}

	enabledRoles := make([]*model.Role, 0, len(customRoleNames))

	for _, rolename := range customRoleNames {
		role, exists := roleMap[rolename]

		if exists && role.DeleteAt == 0 {
			enabledRoles = append(enabledRoles, role)
			continue
		} else if exists && role.DeleteAt > 0 {
			restoredRole, err := a.restoreCustomRole(c, role)
			if err != nil {
				return nil, err
			}
			enabledRoles = append(enabledRoles, restoredRole)
		} else if !exists {
			customRole, ok := customRoleTemplates[rolename]
			if !ok {
				c.Logger().Warn("Custom role definition not found, skipping creation.", mlog.String("rolename", rolename))
				continue
			}
			createdRole, err := a.createCustomRole(c, &customRole)
			if err != nil {
				return nil, err
			}
			enabledRoles = append(enabledRoles, createdRole)
		}
	}

	return enabledRoles, nil
}

func (a *App) createCustomRole(c request.CTX, role *model.Role) (*model.Role, *model.AppError) {
	role, err := a.CreateRole(role)
	if err != nil {
		return nil, err
	}

	c.Logger().Info("Created custom role for tunag",
		mlog.String("role_id", role.Id),
		mlog.String("rolename", role.Name),
		mlog.String("display_name", role.DisplayName),
		mlog.String("description", role.Description),
		mlog.Array("permission", role.Permissions),
		mlog.Bool("scheme_managed", role.SchemeManaged),
		mlog.Bool("built_in", role.BuiltIn),
	)
	return role, nil
}

func (a *App) restoreCustomRole(c request.CTX, role *model.Role) (*model.Role, *model.AppError) {
	role.DeleteAt = 0
	role, err := a.UpdateRole(role)
	if err != nil {
		return nil, err
	}

	c.Logger().Info("Restored custom role for tunag",
		mlog.String("role_id", role.Id),
		mlog.String("rolename", role.Name),
		mlog.String("display_name", role.DisplayName),
		mlog.String("description", role.Description),
		mlog.Array("permission", role.Permissions),
		mlog.Bool("scheme_managed", role.SchemeManaged),
		mlog.Bool("built_in", role.BuiltIn),
	)
	return role, nil
}

// DisableCustomRoles soft-deletes all custom roles belonging to the given group.
func (a *App) DisableCustomRoles(c request.CTX, customRoleGroup string) *model.AppError {
	customRoles, err := a.GetRolesByNames(model.CustomRoleNamesForGroup(customRoleGroup))
	if err != nil {
		return err
	}

	var deleted []string
	for _, role := range customRoles {
		_, err = a.DeleteRole(role.Id)
		if err != nil {
			return err
		}
		deleted = append(deleted, role.Name)
	}

	c.Logger().Info("Deleted custom roles for tunag",
		mlog.String("custom_role_group", customRoleGroup),
		mlog.Array("roles", deleted),
	)
	return nil
}
