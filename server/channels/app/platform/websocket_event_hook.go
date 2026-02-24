// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package platform

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/v8/channels/store"
	"github.com/mattermost/mattermost/server/v8/channels/store/sqlstore"
)

const callsPluginId = "com.mattermost.calls"
const callsTimeoutDuration = time.Minute*3

const (
	callTimeoutEventHookName = "call_timeout"
)

type WebSocketEventHook interface {
	Process(message *model.WebSocketEvent) error
}

// PlatformService構造体に、Hookが使用するメモリを登録する関数
func (ps *PlatformService) makeWebSocketEventHook() map[string]WebSocketEventHook {
	return map[string]WebSocketEventHook{
		callTimeoutEventHookName: &callTimeoutEventHook{ps: ps, rooms: make(map[string]*CallRoom)},
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
	channelId string
	scheduledTask *model.ScheduledTask
	mu sync.Mutex
}

// callTimeoutEventHook	が必要とするものを格納する構造体
// PlatformService起動時に一度だけ確保され、サーバー終了時まで保持される。
type callTimeoutEventHook struct{
	ps *PlatformService // Storeアクセスのため、親を依存性注入しておく
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

	// mlog.Debug("@@@@@@@@@@@@@@@@@ eventPartの長さはこれです！",
	// 	mlog.Int("eventPart_length", len(eventPart)),
	// )

	if len(eventPart) < 3 || !(eventPart[0] == "custom") || !(eventPart[1] == callsPluginId) {
		return nil
	}

	// mlog.Debug("@@@@@@@@@@@@@@@@@@@@ 関数内に入っています！",
	// 	mlog.String("part1", event),
	// 	mlog.Bool("map_initialized", h.rooms != nil),
	// 	mlog.Int("map_len", len(h.rooms)),
	// 	// mlog.Any("struct.mu_state", h.mu),
	// )

	data :=  message.GetData()
	callId := data["id"].(string) // 各通話の識別番号
	channelId := data["channelID"].(string)

	callRoom, exist := h.rooms[callId]
	if !exist && eventPart[2] == "call_start" {
		h.mu.Lock()
		defer h.mu.Unlock()

		h.rooms[callId] = &CallRoom{
			channelId: channelId,
			scheduledTask: model.CreateTask(callId, func(){h.checkTimeoutCondition(callId)}, callsTimeoutDuration),
		}
		return nil
	}

	if exist {
		h.mu.Lock()
		defer h.mu.Unlock()

		switch eventPart[2] {
		case "user_joined":
			if callRoom.scheduledTask != nil {
				callRoom.scheduledTask.Cancel()
				callRoom.scheduledTask = nil
			}
		case "user_left":
			if callRoom.scheduledTask != nil {
				callRoom.scheduledTask.Cancel()
				callRoom.scheduledTask = nil
			}
			callRoom.scheduledTask = model.CreateTask(callId, func(){h.checkTimeoutCondition(callId)}, callsTimeoutDuration)
		case "call_end":
			if callRoom.scheduledTask != nil {
				callRoom.scheduledTask.Cancel()
				callRoom.scheduledTask = nil
			}
			delete(h.rooms, callId)
			h.rebuildMap()
		}
	}

	return nil
}

// func (h *callTimeoutEventHook) checkTimeoutCondition() {
func (h *callTimeoutEventHook) checkTimeoutCondition(callId string) {
	// タイムアウトのタイマーが途中で止められなければの関数内に入る

	callInfo, err := GetCallByID(callId)
	if err != nil {
		// エラー処理
	}

}

type callInfo struct {
	ID string
	ChannelID string
	EndAt string
	Participants string
	Stats string
	Props string
}

func (h *callTimeoutEventHook) GetCallByID(callId string) (*callInfo, error) {
	sqlStore := h.ps.Store.(*sqlstore.SqlStore)

	query := sqlStore.getQueryBuilder().
		Selec("ID", "ChanelID", "EndAt", "Participants", "Stats", "Props").
		From("Calls").
		Where(sq.Eq{"ID": callId})

	var call *callInfo
	if err := sqlStore.GetReplica().GetBuilder(&call, query); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Call", callId)
		}
		return nil,
	}
	return call, nil
}

/**
 * TODO: Server起動のどこかで登録
 */
// CallRoomのメモリ解放を行う定期処理（通常Jobと違いDBアクセスなし）
// func runCallRoomCleanup(s *Server) {
// 	model.CreateRecurringTask("taskName", func(){
// 		doCallRoomCleanup()
// 	}, time.Hour*2)
// }
//
// func doCallRoomCleanup() {
//
// }
