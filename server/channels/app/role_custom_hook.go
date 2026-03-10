package app

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

/*
 * api/v4/roles/names エンドポイントへの Hook処理を行う関数
 *
 * ・TUNAG 側から、必要なカスタムロールが rolename に入った状態で APIが叩かれる
 * ・必要なカスタムロールが既に作成されているかを確認し、ない場合は新規作成する
 *
 * rolename : TUNAG の設定に必要な カスタムロール が全て含まれている配列
 */
func (a *App) CheckCustomRolesHook(c request.CTX, rolenames []string) *model.AppError {
	if rolenames == nil {
		return nil
	}

	contained := false
	for _, rolename := range rolenames {
		if strings.Contains(rolename, "tunag") {
			contained = true
		}
	}
	// rolenames does not include the custom role
	if !contained {
		return nil
	}

	roles, nErr := a.Srv().Store().Role().GetByNames(rolenames)
	if nErr != nil {
		return model.NewAppError("CheckCustomRoles", "", nil, "failed to get roles by names from database", http.StatusInternalServerError).Wrap(nErr)
	}

	// check if the rolename already exist, if not, create new custom role
	for _, rolename := range rolenames {
		for _, role := range roles {
			if rolename == role.Name {
				continue
			}
		}

		_, nErr := a.CreateRole(prepCustomRole(rolename))
		if nErr != nil {
			return model.NewAppError("CheckCustomRoles", "", nil, "failed to create custom role", http.StatusInternalServerError).Wrap(nErr)
		}
	}

	return nil
}

func prepCustomRole(rolename string) *model.Role {
	role := model.Role{
		Name:          rolename,
		DisplayName:   "",
		Description:   "",
		SchemeManaged: false,
		BuiltIn:       false,
	}

	// grant certain permissions
	switch {
	case strings.HasPrefix(rolename, "system"):
		// system_tunag_admin
		role.Permissions = append(role.Permissions, model.PermissionCreateDirectChannel.Id)
		role.Permissions = append(role.Permissions, model.PermissionCreateDirectChannel.Id)

	case strings.HasPrefix(rolename, "team"):
		// team_tunag_admin
		role.Permissions = append(role.Permissions, model.PermissionCreatePrivateChannel.Id)
	}

	return &role
}
