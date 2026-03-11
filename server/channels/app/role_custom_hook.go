// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

/*
 * api/v4/roles/names エンドポイントへの Hook処理を行う関数
 *
 * ・TUNAG 側から、必要なカスタムロールが rolename に入った状態で APIが叩かれる想定
 * ・必要なカスタムロールが既に作成されているかを確認し、ない場合は新規作成する
 *
 * rolename : TUNAG からAPIが叩かれた際には、設定に必要な カスタムロール が全て含まれている配列
 */
func (a *App) CheckCustomRolesHook(c request.CTX, rolenames []string) *model.AppError {
	if len(rolenames) == 0 {
		return nil
	}

	hasTunagRole := false
	for _, rolename := range rolenames {
		if strings.Contains(rolename, "tunag") {
			hasTunagRole = true
			break
		}
	}

	if !hasTunagRole {
		return nil
	}

	roles, nErr := a.Srv().Store().Role().GetByNames(rolenames)
	if nErr != nil {
		return model.NewAppError("CheckCustomRoles", "app.role.get_by_names.app_error", nil, "failed to get roles by names from database", http.StatusInternalServerError).Wrap(nErr)
	}

	exist := make(map[string]bool)
	for _, role := range roles {
		exist[role.Name] = true
	}

	// check if each rolename already exist, if not, create new custom role
	for _, rolename := range rolenames {
		if exist[rolename] {
			continue
		}

		newRole, nErr := a.CreateRole(prepCustomRole(rolename))
		if nErr != nil {
			return model.NewAppError("CheckCustomRoles", "app.role.save.insert.app_error", nil, "failed to create custom role", http.StatusInternalServerError).Wrap(nErr)
		}
		c.Logger().Info("Created custom role for tunag",
			mlog.String("role_id", newRole.Id),
			mlog.String("rolename", newRole.Name),
			mlog.String("display_name", newRole.DisplayName),
			mlog.String("description", newRole.Description),
			mlog.Array("permission", newRole.Permissions),
			mlog.Bool("scheme_managed", newRole.SchemeManaged),
			mlog.Bool("built_in", newRole.BuiltIn),
		)
		exist[rolename] = true
	}

	return nil
}

func prepCustomRole(rolename string) *model.Role {
	role := model.Role{
		Name:          rolename,
		DisplayName:   rolename,
		Description:   "",
		SchemeManaged: false,
		BuiltIn:       false,
	}

	// grant appropriate permissions and description
	switch {
	case strings.HasPrefix(rolename, "system"):
		// system_tunag_admin
		role.Permissions = append(role.Permissions, model.PermissionCreateDirectChannel.Id)
		role.Permissions = append(role.Permissions, model.PermissionCreateGroupChannel.Id)
		role.Description = "TUNAGシステム管理者ロール。ダイレクトチャンネルとグループチャンネルの作成権限を持ちます。"

	case strings.HasPrefix(rolename, "team"):
		// team_tunag_admin
		role.Permissions = append(role.Permissions, model.PermissionCreatePrivateChannel.Id)
		role.Description = "TUNAGチーム管理者ロール。プライベートチャンネルの作成権限を持ちます。"
	}

	return &role
}
