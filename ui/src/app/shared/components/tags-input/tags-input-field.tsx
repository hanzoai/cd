import * as React from 'react';
import {ReactForm} from '../../../../kit/src';
import {TagsInput} from './tags-input';

export const TagsInputField = ReactForm.FormField((props: {options: string[]; noTagsLabel?: string; fieldApi: ReactForm.FieldApi; className: string}) => {
    const {
        fieldApi: {getValue, setValue}
    } = props;
    const tags = (getValue() || []) as Array<string>;

    return (
        <div className='kit-has-value kit-field' style={{border: 'none'}}>
            <TagsInput tags={tags} autocomplete={props.options || []} onChange={vals => setValue(vals)} />
        </div>
    );
});
