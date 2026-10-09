import * as React from 'react';
import {Route, Routes} from 'react-router';
import {RouteChild, RouteComponentProps} from '../../shared/components/router-compat';
import {ApplicationDetails} from './application-details/application-details';
import {ApplicationFullscreenLogs} from './application-fullscreen-logs/application-fullscreen-logs';
import {ApplicationsList} from './applications-list/applications-list';
import {ApplicationSetsList} from './applications-list/application-sets-list';

export const ApplicationsContainer = (props: RouteComponentProps<any>) => {
    // The container serves both /applications and /applicationsets; the first path segment says which.
    const objectListKind = props.location.pathname.split('/')[1] === 'applicationsets' ? 'applicationset' : 'application';
    const details = <RouteChild render={routeProps => <ApplicationDetails objectListKind={objectListKind} {...(routeProps as any)} />} />;
    const logs = <RouteChild render={routeProps => <ApplicationFullscreenLogs {...(routeProps as any)} />} />;

    return (
        <Routes>
            <Route index={true} element={objectListKind === 'application' ? <ApplicationsList {...(props as any)} /> : <ApplicationSetsList {...(props as any)} />} />
            <Route path=':name' element={details} />
            <Route path=':appnamespace/:name' element={details} />
            <Route path=':name/:namespace/:container/logs' element={logs} />
            <Route path=':appnamespace/:name/:namespace/:container/logs' element={logs} />
        </Routes>
    );
};
