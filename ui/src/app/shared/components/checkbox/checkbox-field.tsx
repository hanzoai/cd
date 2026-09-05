import {Checkbox} from '../../../../kit/src';
import * as React from 'react';
import {ReactForm} from '../../../../kit/src';

export const CheckboxField = ReactForm.FormField((props: {fieldApi: ReactForm.FieldApi; id: string; className: string; checked: boolean}) => {
    const {
        fieldApi: {getValue, setValue}
    } = props;

    return <Checkbox id={props.id} checked={!!getValue()} onChange={setValue} />;
});
