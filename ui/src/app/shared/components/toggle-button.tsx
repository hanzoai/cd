import * as React from 'react';
import {ReactNode} from 'react';
import {Button} from './button';
import {Icon} from './icon';
import {WARNING_COLOR} from './colors';

export const ToggleButton = ({
    title,
    children,
    onToggle,
    toggled,
    beat,
    disabled,
    icon,
    rotate
}: {
    toggled: boolean;
    beat?: boolean;
    onToggle: () => void;
    children?: ReactNode;
    title: string;
    disabled?: boolean;
    icon?: Icon;
    rotate?: boolean;
}) => (
    <Button
        title={title}
        onClick={onToggle}
        icon={icon}
        rotate={rotate}
        disabled={disabled}
        beat={beat}
        style={{
            // these are the kit-button color swapped
            backgroundColor: toggled && WARNING_COLOR,
            color: toggled && 'black'
        }}>
        {children}
    </Button>
);
