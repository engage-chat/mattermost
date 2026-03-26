// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	sq "github.com/mattermost/squirrel"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/store"
)

type SqlEngageChatStore struct {
	*SqlStore
}

func newSqlEngageChatStore(sqlStore *SqlStore) store.EngageChatStore {
	s := &SqlEngageChatStore{
		SqlStore: sqlStore,
	}
	return s
}

func (s *SqlEngageChatStore) HasChannelMemberWithRoles(channelID string, options *model.EngageChatRoleSearchOptions) (bool, error) {
	if len(options.SystemRoles) == 0 && len(options.TeamRoles) == 0 {
		return false, nil
	}

	orConditions := sq.Or{}
	if len(options.SystemRoles) > 0 {
		for _, role := range options.SystemRoles {
			orConditions = append(orConditions, sq.Like{"u.Roles": "%" + role + "%"})
		}
	}
	if len(options.TeamRoles) > 0 {
		for _, role := range options.TeamRoles {
			orConditions = append(orConditions, sq.Like{"tm.Roles": "%" + role + "%"})
		}
	}

	query := s.getQueryBuilder().Select("COUNT(DISTINCT u.Id)").
		From("ChannelMembers cm").
		Join("Users u ON cm.UserId = u.Id").
		Join("Channels c ON cm.ChannelId = c.Id").
		LeftJoin("TeamMembers tm ON tm.TeamId = c.TeamId AND tm.UserId = cm.UserId").
		Where(sq.Eq{"cm.ChannelId": channelID}).
		Where(orConditions)

	queryString, args, err := query.ToSql()
	if err != nil {
		return false, errors.Wrap(err, "failed to build query")
	}

	var count int64
	if err := s.GetReplica().Get(&count, queryString, args...); err != nil {
		return false, errors.Wrap(err, "failed to query for channel member with role")
	}

	return count > 0, nil
}
