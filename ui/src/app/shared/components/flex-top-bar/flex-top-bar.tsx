import {Toolbar, Tooltip} from '../../../../kit/src';
import * as React from 'react';
import {Observable} from 'rxjs';
import {AuthOption, DataLoader} from '../';

import './flex-top-bar.scss';

// Extend Toolbar type to support options on the right and auth control
export interface ToolbarWithOptions extends Toolbar {
    options?: React.ReactNode;
    addAuth?: boolean; // defaults to true
}

interface FlexTopBarProps {
    toolbar: ToolbarWithOptions | Observable<ToolbarWithOptions>;
}

export const FlexTopBar = (props: FlexTopBarProps) => {
    return (
        <React.Fragment>
            <div className='top-bar row flex-top-bar' key='tool-bar'>
                <DataLoader load={() => Promise.resolve(props.toolbar)}>{(toolbar: ToolbarWithOptions) => <FlexTopBarContent toolbar={toolbar} />}</DataLoader>
            </div>
            <div className='flex-top-bar__padder' />
        </React.Fragment>
    );
};

const FlexTopBarContent = (props: {toolbar: ToolbarWithOptions}) => {
    const shouldAddAuth = props.toolbar.addAuth !== false;

    return (
        <React.Fragment>
            <div className='flex-top-bar__actions'>
                {props.toolbar.actionMenu && (
                    <React.Fragment>
                        {/* The first action is the page's one primary; the rest are ghosts. */}
                        {props.toolbar.actionMenu.items.map((item, i) => (
                            <Tooltip className='custom-tooltip' content={item.title} key={item.qeId || i}>
                                <button
                                    disabled={!!item.disabled}
                                    qe-id={item.qeId}
                                    aria-label={typeof item.title === 'string' ? item.title : undefined}
                                    className={`kit-button ${i === 0 ? 'kit-button--base' : 'kit-button--base-o'}`}
                                    onClick={() => item.action()}
                                    key={i}>
                                    {item.iconClassName && <i className={item.iconClassName} />}
                                    <span className='show-for-large'>{item.title}</span>
                                </button>
                            </Tooltip>
                        ))}
                    </React.Fragment>
                )}
            </div>
            <div className='flex-top-bar__tools'>
                {props.toolbar.tools && <div className='flex-top-bar__tools-left'>{props.toolbar.tools}</div>}
                {(props.toolbar.options || shouldAddAuth) && (
                    <div className='flex-top-bar__tools-right'>
                        {props.toolbar.options}
                        {shouldAddAuth && <AuthOption />}
                    </div>
                )}
            </div>
        </React.Fragment>
    );
};
