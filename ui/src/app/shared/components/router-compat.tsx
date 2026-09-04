import * as React from 'react';
import {useLocation, useParams} from 'react-router-dom';
import {history} from '../history';

// react-router v7 dropped the v5 render-prop route props and the withRouter HOC.
// The app's routed components still expect the v5 shape { history, location,
// match }, so these adapters rebuild that shape from v7 hooks in one place
// rather than rewriting every component.

export interface RouteChildProps {
    history: typeof history;
    location: ReturnType<typeof useLocation>;
    match: {params: {[k: string]: string}; url: string; path: string; isExact: boolean};
}

// Alias kept so components can go on importing the familiar name.
export type RouteComponentProps<P = {}> = RouteChildProps & {match: {params: P}};

function useRouteChildProps(): RouteChildProps {
    const location = useLocation();
    const params = useParams();
    return {
        history,
        location,
        match: {params: params as {[k: string]: string}, url: location.pathname, path: location.pathname, isExact: true}
    };
}

// Renders a routed component with the v5-shaped props injected. Used by app.tsx
// in place of the old <Route render={routeProps => ...}> callback.
export function RouteChild({render}: {render: (props: RouteChildProps) => React.ReactNode}) {
    return <>{render(useRouteChildProps())}</>;
}

export function withRouter<P extends RouteChildProps>(Component: React.ComponentType<P>): React.FC<Omit<P, keyof RouteChildProps>> {
    return function WithRouter(props) {
        const routeProps = useRouteChildProps();
        return <Component {...(props as unknown as P)} {...routeProps} />;
    };
}
