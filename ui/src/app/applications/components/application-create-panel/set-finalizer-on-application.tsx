import {Checkbox, HelpIcon} from '../../../../kit/src';
import * as React from 'react';
import {ReactForm} from '../../../../kit/src';

export const SetFinalizerOnApplication = ReactForm.FormField((props: {fieldApi: ReactForm.FieldApi}) => {
    const {
        fieldApi: {getValue, setValue}
    } = props;
    const finalizerVal = 'resources-finalizer.apps.hanzo.ai';
    const currentValue = getValue() || [];
    const index = currentValue.findIndex((item: string) => item === finalizerVal);
    const isChecked = index < 0 ? false : true;
    return (
        <div className='small-12 large-6' style={{borderBottom: '0'}}>
            <Checkbox
                id='set-finalizer'
                checked={isChecked}
                onChange={(state: boolean) => {
                    const value = getValue() || [];
                    if (!state) {
                        const i = value.findIndex((item: string) => item === finalizerVal);
                        if (i >= 0) {
                            const tmp = value.slice();
                            tmp.splice(i, 1);
                            setValue(tmp);
                        }
                    } else {
                        const tmp = value.slice();
                        tmp.push(finalizerVal);
                        setValue(tmp);
                    }
                }}
            />
            <label htmlFor={`set-finalizer`}>Set Deletion Finalizer</label>
            <HelpIcon title='If checked, the resources deletion finalizer will be set on the application. Potentially destructive, refer to the documentation for more information on the effects of the finalizer.' />
        </div>
    );
});
