// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import "net/http"

func (api *API) InitEngageChatLocal() {
	api.BaseRoutes.EngageChat.Handle("/roles", api.APILocal(enableCustomRoles)).Methods(http.MethodPost)
}
