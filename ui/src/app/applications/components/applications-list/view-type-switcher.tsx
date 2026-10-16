import classNames from 'classnames';
import * as React from 'react';
import {ContextApis} from '../../../shared/context';
import {AppsListPreferences, AppsListViewKey, services} from '../../../shared/services';

interface ViewTypeSwitcherProps {
    pref: AppsListPreferences & {page: number; search: string};
    ctx: ContextApis;
}

export const ViewTypeSwitcher: React.FC<ViewTypeSwitcherProps> = ({pref, ctx}) => {
    const {List, Summary, Tiles} = AppsListViewKey;
    const views: [AppsListViewKey, string, string][] = [
        [Tiles, 'Tiles', 'fa fa-th'],
        [List, 'List', 'fa fa-th-list'],
        [Summary, 'Summary', 'fa fa-chart-pie']
    ];

    return (
        <div className='applications-list__view-type' role='group' aria-label='View'>
            {views.map(([view, title, icon]) => (
                <button
                    key={view}
                    type='button'
                    title={title}
                    aria-label={title}
                    aria-pressed={pref.view === view}
                    className={classNames({selected: pref.view === view})}
                    onClick={() => {
                        ctx.navigation.goto('.', {view});
                        services.viewPreferences.updatePreferences({appList: {...pref, view}});
                    }}>
                    <i className={icon} />
                </button>
            ))}
        </div>
    );
};
