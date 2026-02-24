// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package platform

import (
// 	"strings"
// 	"sync"
// 	"time"
//
// 	"github.com/mattermost/mattermost/server/public/model"
// 	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (h *callTimeoutEventHook) rebuildMap() {
	h.mu.Lock()
	defer h.mu.Unlock()

	newMap := make(map[string]*CallRoom, len(h.rooms))
	for k, v := range h.rooms {
		newMap[k] = v
	}
	h.rooms = newMap
}
