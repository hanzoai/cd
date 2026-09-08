import * as React from 'react';

export interface IconColumnProps {
    // Full icon class, e.g. 'kit-icon-git' or 'fa fa-object-group'. Omit for the (empty) header cell.
    icon?: string;
}

// Leading fixed-width table column that holds a single list-item icon. The shared
// `.kit-table-list__icon-column` styles size and center the glyph consistently across lists.
export const IconColumn = (props: IconColumnProps) => <div className='columns small-1 kit-table-list__icon-column'>{props.icon && <i className={`icon ${props.icon}`} />}</div>;
