// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package platform

import (
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

const callsPluginId = "com.mattermost.calls"

const (
	interceptorCallTimeout = "call_timeout"
)

type WebSocketEventHook interface {
	Process(message *model.WebSocketEvent) error
}

// PlatformService構造体に、Hookが使用するメモリを登録する関数
func (ps *PlatformService) makeWebSocketEventHook() map[string]WebSocketEventHook {
	return map[string]WebSocketEventHook{
		interceptorCallTimeout: &callTimeoutEventHook{rooms: make(map[string]*CallRoom)},
	}
}

// 登録されたHooksのProcessを順番に実行する関数
func (ps *PlatformService) runWebSocketEventHooks(message *model.WebSocketEvent) {
	if !ps.WebsocketEventHookEnabled {
		return
	}
	eventHooks := ps.WebSocketEventHooks

	if len(eventHooks) == 0 {
		return
	}

	for hookName, hook := range eventHooks {
		if hook == nil {
			mlog.Warn("runWebSocketEventHooks: Unable to find websocket event hook", mlog.String("hook_name", hookName))
			continue
		}

		err := hook.Process(message)
		if err != nil {
			mlog.Warn("runWebSocketEventHooks: Error processing hook", mlog.String("hook_name", hookName))
		}
	}
}

// 各通話の状態と、専用のタイマーを管理する構造体
type CallRoom struct {
	callID string
	counter int
	timer *time.Timer
	mu sync.Mutex
}

// callTimeoutEventHook	が必要とするものを格納する構造体
// PlatformService起動時に一度だけ確保され、サーバー終了時まで保持される。
type callTimeoutEventHook struct{
	mu sync.Mutex
	rooms map[string]*CallRoom // メモリ解放はdelete()ではされないため注意　CallRoomを更新するごとにマップ作り直す処理が必要
}

// WSイベントから通話プラグインのタイムアウトを判定するフック
func (h *callTimeoutEventHook) Process(message *model.WebSocketEvent) error {
	if message == nil {
		return nil
	}
	event := message.EventType()

	// プラグインイベントは右のように命名される "custom_{plugin id}_{event type}"
	eventPart := strings.SplitN(string(event), "_", 3)
	mlog.Debug("@@@@@@@@@@@@@@@@@ eventPartの長さはこれです！",
		mlog.Int("eventPart_length", len(eventPart)),
	)

// 	if len(eventPart) < 3 || !(eventPart[0] == "custom") || !(eventPart[1] == callsPluginId) {
// 		return nil
// 	}
// 	if (eventPart[0] == "posted") {
// 		mlog.Debug("***********************postedのWSイベントを受信しました")
// 	}

	mlog.Debug("@@@@@@@@@@@@@@@@@@@@ 関数内に入っています！",
		mlog.String("part1", event),
		mlog.Bool("map_initialized", h.rooms != nil),
		mlog.Int("map_len", len(h.rooms)),
		// mlog.Any("struct.mu_state", h.mu),
	)

	return nil
}


// func (ps *PlatformService) メモリ管理して解放する定期実行(){
//
// }
//





// サーバー起動後、メモリリークを防ぐための定期的なクリーンアップ
// これはMattermostのJobにSchedule登録してもいいかも

// func (ps *PlatformService) memoryManagementLoop(ctx context.Context) {
// 	ticker := time.NewTicker(1 * time.Hour)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:
// 			ps.cleanupExpiredRooms()
// 		}
// 	}
// }
//
// func (ps *PlatformService) cleanupExpiredRooms() {
// 	if len(ps.WebSocketEventHooks) == 0 {
// 		return
// 	}
//
// 	if hook, ok := ps.WebSocketEventHooks[interceptorCallTimeout]; ok {
// 		if h, ok := hook.(*callTimeoutEventHook); ok {
// 			h.mu.Lock()
// 			defer h.mu.Unlock()
//
// 			// 期限切れのルームを削除
// 			for callID, room := range h.rooms {
// 				if room.timer == nil || room.counter == 0 {
// 					delete(h.rooms, callID)
// 				}
// 			}
// 		}
// 	}
// }
