import * as React from 'react';
import {Navigate, Route, Routes} from 'react-router';
import {KeybindingProvider} from '../../../kit/v2';
import {RouteChild} from '../../shared/components/router-compat';

import {AccountDetails} from './account-details/account-details';
import {AccountsList} from './accounts-list/accounts-list';
import {AdvancedSettings} from './advanced-settings/advanced-settings';
import {CertsList} from './certs-list/certs-list';
import {ClusterDetails} from './cluster-details/cluster-details';
import {ClustersList} from './clusters-list/clusters-list';
import {GpgKeysList} from './gpgkeys-list/gpgkeys-list';
import {ProjectDetails} from './project-details/project-details';
import {ProjectsList} from './projects-list/projects-list';
import {ReposList} from './repos-list/repos-list';
import {SettingsOverview} from './settings-overview/settings-overview';
import {AppearanceList} from './appearance-list/appearance-list';

export const SettingsContainer = () => (
    <KeybindingProvider>
        <Routes>
            <Route index={true} element={<SettingsOverview />} />
            <Route path='repos' element={<RouteChild render={props => <ReposList {...props} />} />} />
            <Route path='certs' element={<RouteChild render={props => <CertsList {...props} />} />} />
            <Route path='gpgkeys' element={<RouteChild render={props => <GpgKeysList {...props} />} />} />
            <Route path='clusters' element={<ClustersList />} />
            <Route path='clusters/:server' element={<RouteChild render={props => <ClusterDetails {...(props as any)} />} />} />
            <Route path='projects' element={<ProjectsList />} />
            <Route path='projects/:name' element={<RouteChild render={props => <ProjectDetails {...(props as any)} />} />} />
            <Route path='accounts' element={<AccountsList />} />
            <Route path='accounts/:name' element={<RouteChild render={props => <AccountDetails {...(props as any)} />} />} />
            <Route path='appearance' element={<AppearanceList />} />
            <Route path='advanced' element={<AdvancedSettings />} />
            <Route path='*' element={<Navigate to='/settings' replace={true} />} />
        </Routes>
    </KeybindingProvider>
);
