/**
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * @flow
 * @format
 */
import type {ContextRouter} from 'react-router';
import type {NavigatableView} from '@fbcnms/ui/components/design-system/View/NavigatableViews';

import * as React from 'react';
import AppContext from '@fbcnms/ui/context/AppContext';
import NavigatableViews from '@fbcnms/ui/components/design-system/View/NavigatableViews';
import NewUserDialog from './users/NewUserDialog';
import PermissionsGroupCard from './groups/PermissionsGroupCard';
import PermissionsGroupsView, {
  PERMISSION_GROUPS_VIEW_NAME,
} from './groups/PermissionsGroupsView';
import PermissionsPoliciesView, {
  PERMISSION_POLICIES_VIEW_NAME,
} from './policies/PermissionsPoliciesView';
import Strings from '@fbcnms/strings/Strings';
import UsersView from './users/UsersView';
import fbt from 'fbt';
import {ButtonAction} from '@fbcnms/ui/components/design-system/View/ViewHeaderActions';
import {FormContextProvider} from '../../../common/FormContext';
import {NEW_DIALOG_PARAM} from './utils/UserManagementUtils';
import {UserManagementContextProvider} from './UserManagementContext';
import {useCallback, useContext, useMemo, useState} from 'react';
import {useHistory, withRouter} from 'react-router-dom';

const USERS_HEADER = fbt(
  'Users & Roles',
  'Header for view showing system users settings',
);

type Props = ContextRouter;

const UserManaementView = ({match}: Props) => {
  const history = useHistory();
  const basePath = match.path;
  const [addingNewUser, setAddingNewUser] = useState(false);
  const gotoGroupsPage = useCallback(() => history.push(`${basePath}/groups`), [
    history,
    basePath,
  ]);

  const {isFeatureEnabled} = useContext(AppContext);
  const userManagementDevMode = isFeatureEnabled('user_management_dev');

  const VIEWS: Array<NavigatableView> = useMemo(() => {
    const views = [
      {
        routingPath: 'users',
        menuItem: {
          label: USERS_HEADER,
          tooltip: `${USERS_HEADER}`,
        },
        component: {
          header: {
            title: `${USERS_HEADER}`,
            subtitle:
              'Add and manage your organization users, and set their role to control their global settings',
            actionButtons: [
              <ButtonAction action={() => setAddingNewUser(true)}>
                <fbt desc="">Add User</fbt>
              </ButtonAction>,
            ],
          },
          children: <UsersView />,
        },
      },
      {
        routingPath: 'groups',
        menuItem: {
          label: PERMISSION_GROUPS_VIEW_NAME,
          tooltip: `${PERMISSION_GROUPS_VIEW_NAME}`,
        },
        component: {
          header: {
            title: `${PERMISSION_GROUPS_VIEW_NAME}`,
            subtitle:
              'Create groups with different rules and add users to apply permissions',
            actionButtons: userManagementDevMode
              ? [
                  <ButtonAction
                    action={() => history.push(`group/${NEW_DIALOG_PARAM}`)}>
                    <fbt desc="">Create Group</fbt>
                  </ButtonAction>,
                ]
              : [],
          },
          children: <PermissionsGroupsView />,
        },
      },
      {
        routingPath: 'group/:id',
        component: {
          children: (
            <PermissionsGroupCard
              redirectToGroupsView={gotoGroupsPage}
              onClose={gotoGroupsPage}
            />
          ),
        },
        relatedMenuItemIndex: 1,
      },
    ];

    if (userManagementDevMode) {
      views.push({
        routingPath: 'policies',
        menuItem: {
          label: PERMISSION_POLICIES_VIEW_NAME,
          tooltip: `${PERMISSION_POLICIES_VIEW_NAME}`,
        },
        component: {
          header: {
            title: `${PERMISSION_POLICIES_VIEW_NAME}`,
            subtitle: 'Manage policies and apply them to groups.',
            actionButtons: [
              <ButtonAction
                action={() => history.push(`policy/${NEW_DIALOG_PARAM}`)}>
                <fbt desc="">Create Policy</fbt>
              </ButtonAction>,
            ],
          },
          children: <PermissionsPoliciesView />,
        },
      });
    }

    return views;
  }, [gotoGroupsPage, history, userManagementDevMode]);

  return (
    <UserManagementContextProvider>
      <FormContextProvider ignorePermissions={true}>
        <NavigatableViews
          header={Strings.admin.users.viewHeader}
          views={VIEWS}
          routingBasePath={basePath}
        />
      </FormContextProvider>
      {addingNewUser && (
        <NewUserDialog
          isOpened={addingNewUser}
          onClose={() => setAddingNewUser(false)}
        />
      )}
    </UserManagementContextProvider>
  );
};

export default withRouter(UserManaementView);
