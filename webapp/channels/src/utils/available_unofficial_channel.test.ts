// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Permissions} from 'mattermost-redux/constants';
import {getChannel} from 'mattermost-redux/selectors/entities/channels';
import {haveIChannelPermission, haveICurrentTeamPermission} from 'mattermost-redux/selectors/entities/roles';

import store from 'stores/redux_store';

import {fetchChannelAccessible} from 'actions/engage_chat';

import {isAvailableUnofficialChannel, isAvailableDMGMChannel} from './available_unofficial_channel';
import {isOfficialTunagChannel} from './official_channel_utils';

jest.mock('stores/redux_store', () => ({
    getState: jest.fn(),
    dispatch: jest.fn(),
}));

jest.mock('mattermost-redux/selectors/entities/channels', () => ({
    ...jest.requireActual('mattermost-redux/selectors/entities/channels'),
    getChannel: jest.fn(),
}));

jest.mock('mattermost-redux/selectors/entities/roles', () => ({
    ...jest.requireActual('mattermost-redux/selectors/entities/roles'),
    haveIChannelPermission: jest.fn(),
    haveICurrentTeamPermission: jest.fn(),
}));

jest.mock('./official_channel_utils', () => ({
    ...jest.requireActual('./official_channel_utils'),
    isOfficialTunagChannel: jest.fn(),
}));

jest.mock('actions/engage_chat', () => ({
    fetchChannelAccessible: jest.fn().mockReturnValue(() => Promise.resolve()),
}));

// reducer_registryはモジュールロード時に呼ばれるため空実装でモック
jest.mock('mattermost-redux/store/reducer_registry', () => ({
    default: {register: jest.fn()},
}));

let mockChannel: any;
let mockIsOfficial: boolean;
let mockPermissionResult: boolean;

describe('available_unofficial_channel utils', () => {
    // Cast mock functions for easier usage
    const mockGetState = store.getState as jest.Mock;
    const mockDispatch = store.dispatch as jest.Mock;
    const mockGetChannel = getChannel as jest.Mock;
    const mockHaveIChannelPermission = haveIChannelPermission as jest.Mock;
    const mockHaveICurrentTeamPermission = haveICurrentTeamPermission as jest.Mock;
    const mockIsOfficialTunagChannelFn = isOfficialTunagChannel as jest.Mock;
    const mockFetchChannelAccessible = fetchChannelAccessible as jest.Mock;

    beforeEach(() => {
        jest.resetAllMocks();

        // Initialize variables
        mockChannel = {id: 'channel_id', team_id: 'team_id', type: 'O'};
        mockIsOfficial = false;
        mockPermissionResult = true; // Default to having permission

        // Setup mock behavior (engageChatキャッシュなしの状態)
        mockGetState.mockReturnValue({});
        mockGetChannel.mockImplementation(() => mockChannel);
        mockIsOfficialTunagChannelFn.mockImplementation(() => mockIsOfficial);
        mockFetchChannelAccessible.mockReturnValue(() => Promise.resolve());

        // Simplify permission check to return mockPermissionResult
        // (Set mockPermissionResult to false in individual test cases to simulate denial)
        mockHaveIChannelPermission.mockImplementation(() => mockPermissionResult);
        mockHaveICurrentTeamPermission.mockImplementation(() => mockPermissionResult);
    });

    describe('isAvailableUnofficialChannel', () => {
        describe('APIキャッシュ（engageChat）がある場合', () => {
            test('キャッシュがtrue → trueを返し、権限チェックを行わない', () => {
                mockGetState.mockReturnValue({
                    engageChat: {channelAccessible: {channel_id: true}},
                });

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
                expect(mockDispatch).not.toHaveBeenCalled();
            });

            test('キャッシュがfalse → falseを返し、権限チェックを行わない', () => {
                mockGetState.mockReturnValue({
                    engageChat: {channelAccessible: {channel_id: false}},
                });

                expect(isAvailableUnofficialChannel('channel_id')).toBe(false);
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
                expect(mockDispatch).not.toHaveBeenCalled();
            });
        });

        describe('APIキャッシュがない場合（フォールバック）', () => {
            test('APIフェッチをdispatchし、既存の権限ベースロジックで判定する', () => {
                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockDispatch).toHaveBeenCalledTimes(1);
            });

            test('should return false when channel does not exist', () => {
                mockGetChannel.mockReturnValue(null);
                expect(isAvailableUnofficialChannel('missing_channel')).toBe(false);
            });

            test('should return true for official Tunag channel without permission check', () => {
                mockIsOfficial = true;
                mockPermissionResult = false;

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockIsOfficialTunagChannelFn).toHaveBeenCalledWith(mockChannel);
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
            });

            test('should return true for open channel type (Type: O) or default cases', () => {
                mockChannel.type = 'O'; // Open
                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);

                mockChannel.type = 'StrangeType'; // Default case check
                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);

                // Permission check function should not be called in default case
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
            });

            test('should check CREATE_PRIVATE_CHANNEL permission for private channel (Type: P)', () => {
                mockChannel.type = 'P';
                mockPermissionResult = true;

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockHaveIChannelPermission).toHaveBeenCalledWith(
                    expect.anything(),
                    'team_id',
                    'channel_id',
                    Permissions.CREATE_PRIVATE_CHANNEL,
                );
            });

            test('should check CREATE_DIRECT_CHANNEL permission for direct message (Type: D)', () => {
                mockChannel.type = 'D';

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockHaveIChannelPermission).toHaveBeenCalledWith(
                    expect.anything(),
                    'team_id',
                    'channel_id',
                    Permissions.CREATE_DIRECT_CHANNEL,
                );
            });

            test('should check CREATE_GROUP_CHANNEL permission for group message (Type: G)', () => {
                mockChannel.type = 'G';

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockHaveIChannelPermission).toHaveBeenCalledWith(
                    expect.anything(),
                    'team_id',
                    'channel_id',
                    Permissions.CREATE_GROUP_CHANNEL,
                );
            });

            test('should return false when permission is denied', () => {
                mockChannel.type = 'P';
                mockPermissionResult = false; // Permission denied

                expect(isAvailableUnofficialChannel('channel_id')).toBe(false);
            });
        });
    });

    describe('isAvailableDMGMChannel', () => {
        test('should return true when both DM and GM creation permissions are granted', () => {
            mockPermissionResult = true;
            expect(isAvailableDMGMChannel()).toBe(true);

            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_DIRECT_CHANNEL);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_GROUP_CHANNEL);
        });

        test('should return false when DM creation permission is denied', () => {
            // Change return value for each call: 1st(DM) is False, 2nd(GM) is True
            mockHaveICurrentTeamPermission.
                mockReturnValueOnce(false).
                mockReturnValueOnce(true);

            expect(isAvailableDMGMChannel()).toBe(false);
        });

        test('should return false when GM creation permission is denied', () => {
            // Change return value for each call: 1st(DM) is True, 2nd(GM) is False
            mockHaveICurrentTeamPermission.
                mockReturnValueOnce(true).
                mockReturnValueOnce(false);

            expect(isAvailableDMGMChannel()).toBe(false);
        });
    });
});
