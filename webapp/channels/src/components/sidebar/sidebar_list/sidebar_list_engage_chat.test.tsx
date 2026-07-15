// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {CategorySorting} from '@mattermost/types/channel_categories';
import type {ChannelType} from '@mattermost/types/channels';
import type {TeamType} from '@mattermost/types/teams';

import {CategoryTypes} from 'mattermost-redux/constants/channel_categories';

import {shallowWithIntl} from 'tests/helpers/intl-test-helper';
import {TestHelper} from 'utils/test_helper';

import SidebarList, {type SidebarList as SidebarListComponent} from './sidebar_list';

describe('SidebarList - when component is not rendered', () => {
    const currentChannel = TestHelper.getChannelMock({
        id: 'channel_id',
        display_name: 'channel_display_name',
        create_at: 0,
        update_at: 0,
        delete_at: 0,
        team_id: '',
        type: 'O' as ChannelType,
        name: '',
        header: '',
        purpose: '',
        last_post_at: 0,
        last_root_post_at: 0,
        creator_id: '',
        scheme_id: '',
        group_constrained: false,
    });

    const baseProps = {
        currentTeam: TestHelper.getTeamMock({
            id: 'kemjcpu9bi877yegqjs18ndp4r',
            invite_id: 'ojsnudhqzbfzpk6e4n6ip1hwae',
            name: 'test',
            create_at: 123,
            update_at: 123,
            delete_at: 123,
            display_name: 'test',
            description: 'test',
            email: 'test',
            type: 'O' as TeamType,
            company_name: 'test',
            allowed_domains: 'test',
            allow_open_invite: false,
            scheme_id: 'test',
            group_constrained: false,
        }),
        currentChannelId: currentChannel.id,
        categories: [
            {
                id: 'category1',
                team_id: 'team1',
                user_id: '',
                type: CategoryTypes.CUSTOM,
                display_name: 'custom_category_1',
                sorting: CategorySorting.Alphabetical,
                channel_ids: ['channel_id'],
                muted: false,
                collapsed: false,
            },
        ],
        unreadChannelIds: [],
        displayedChannels: [currentChannel],
        newCategoryIds: [],
        multiSelectedChannelIds: [],
        isUnreadFilterEnabled: false,
        draggingState: {},
        categoryCollapsedState: {},
        handleOpenMoreDirectChannelsModal: jest.fn(),
        onDragStart: jest.fn(),
        onDragEnd: jest.fn(),
        showUnreadsCategory: false,
        collapsedThreads: true,
        hasUnreadThreads: false,
        currentStaticPageId: '',
        staticPages: [],
        actions: {
            switchToChannelById: jest.fn(),
            switchToLhsStaticPage: jest.fn(),
            close: jest.fn(),
            moveChannelsInSidebar: jest.fn(),
            moveCategory: jest.fn(),
            removeFromCategory: jest.fn(),
            setDraggingState: jest.fn(),
            stopDragging: jest.fn(),
            clearChannelSelection: jest.fn(),
            multiSelectChannelAdd: jest.fn(),
        },
    };

    test('should not throw error on team change in componentDidUpdate when scrollbar is null', () => {
        const wrapper = shallowWithIntl(
            <SidebarList {...baseProps}/>,
        );
        const instance = wrapper.instance() as SidebarListComponent;
        instance.scrollbar = {current: null};

        const newCurrentTeam = {
            ...baseProps.currentTeam,
            id: 'new_team',
        };

        expect(() => {
            wrapper.setProps({currentTeam: newCurrentTeam});
        }).not.toThrow();
    });

    test('should not throw error on scroll animation update when scrollbar is null', () => {
        const wrapper = shallowWithIntl(
            <SidebarList {...baseProps}/>,
        );
        const instance = wrapper.instance() as SidebarListComponent;
        instance.scrollbar = {current: null};

        const mockSpring = {
            getCurrentValue: jest.fn(() => 100),
        } as any;

        expect(() => {
            instance.handleScrollAnimationUpdate(mockSpring);
        }).not.toThrow();
    });

    test('should not throw error on scrolling to first unread channel when scrollbar is null', () => {
        const wrapper = shallowWithIntl(
            <SidebarList {...baseProps}/>,
        );
        const instance = wrapper.instance() as SidebarListComponent;
        instance.scrollbar = {current: null};

        expect(() => {
            instance.scrollToFirstUnreadChannel();
        }).not.toThrow();
    });

    test('should not throw error on scrolling to position when scrollbar is null', () => {
        const wrapper = shallowWithIntl(
            <SidebarList {...baseProps}/>,
        );
        const instance = wrapper.instance() as SidebarListComponent;
        instance.scrollbar = {current: null};

        expect(() => {
            instance.scrollToPosition(100);
        }).not.toThrow();
    });

    test('should not throw error on updating unread indicators when scrollbar is null', () => {
        const wrapper = shallowWithIntl(
            <SidebarList {...baseProps}/>,
        );
        const instance = wrapper.instance() as SidebarListComponent;
        instance.scrollbar = {current: null};

        expect(() => {
            instance.updateUnreadIndicators();
        }).not.toThrow();
    });

    test('should render Scrollbars by default (when shouldRender is not provided)', () => {
        const wrapper = shallowWithIntl(
            <SidebarList {...baseProps}/>,
        );
        expect(wrapper.find('Scrollbars').exists()).toBe(true);
    });

    test('should not render Scrollbars when shouldRender is false', () => {
        const wrapper = shallowWithIntl(
            <SidebarList
                {...baseProps}
                shouldRender={false}
            />,
        );
        expect(wrapper.find('Scrollbars').exists()).toBe(false);
    });
});
