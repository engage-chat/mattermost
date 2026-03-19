// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// EnableCustomRoles ensures that the given custom roles are active.
// It creates them if they don't exist, or restores them if they were soft-deleted.
func (a *App) EnableCustomRoles(c request.CTX, roleNames []string) ([]*model.Role, *model.AppError) {
	if len(roleNames) == 0 {
		return []*model.Role{}, nil
	}
	// Initialize custom role templates
	customRoleTemplates := model.GetAllCustomRoleTemplates()

	// Get existing roles from the DB in a single batch.
	existingRoles, err := a.GetRolesByNames(roleNames)
	if err != nil {
		return nil, err
	}

	roleMap := make(map[string]*model.Role)
	for _, role := range existingRoles {
		roleMap[role.Name] = role
	}

	enabledRoles := make([]*model.Role, 0, len(roleNames))

	for _, rolename := range roleNames {
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

// DisableCustomRoles soft-deletes all custom roles specified by name.
func (a *App) DisableCustomRoles(c request.CTX, roleNames []string) *model.AppError {
	if len(roleNames) == 0 {
		return nil
	}
	c.Logger().Debug("Attempting to disable custom roles", mlog.Strings("roles", roleNames))

	customRoles, err := a.GetRolesByNames(roleNames)
	if err != nil {
		c.Logger().Error("Failed to get roles by names", mlog.Err(err))
		return err
	}
	c.Logger().Debug("Found roles to delete", mlog.Int("count", len(customRoles)))

	var deleted []string
	for _, role := range customRoles {
		c.Logger().Debug("Attempting to delete role", mlog.String("role_id", role.Id), mlog.String("role_name", role.Name))
		_, err = a.DeleteRole(role.Id)
		if err != nil {
			c.Logger().Error("Failed to delete role", mlog.String("role_id", role.Id), mlog.Err(err))
			return err
		}
		c.Logger().Debug("Successfully deleted role (in app layer)", mlog.String("role_id", role.Id))
		deleted = append(deleted, role.Name)
	}

	c.Logger().Info("Deleted custom roles for tunag",
		mlog.Array("roles", deleted),
	)
	return nil
}
