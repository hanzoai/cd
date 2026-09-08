import * as React from 'react';
import {CSSProperties, MouseEventHandler, ReactNode} from 'react';
import {Icon} from './icon';
import {Tooltip} from '../../../kit/src';

export const Button = ({
    onClick,
    children,
    title,
    outline,
    icon,
    className,
    style,
    disabled,
    beat,
    rotate
}: {
    onClick?: MouseEventHandler;
    children?: ReactNode;
    title?: string;
    outline?: boolean;
    icon?: Icon;
    className?: string;
    style?: CSSProperties;
    disabled?: boolean;
    beat?: boolean;
    rotate?: boolean;
}) => (
    <Tooltip content={title}>
        <button
            className={'kit-button ' + (!outline ? 'kit-button--base' : 'kit-button--base-o') + ' ' + (disabled ? 'disabled' : '') + ' ' + (className || '')}
            style={style}
            onClick={onClick}>
            {icon && <i className={'fa fa-' + icon + ' ' + (beat ? 'fa-beat' : '') + (rotate ? 'fa-rotate-180' : '')} />} {children}
        </button>
    </Tooltip>
);
