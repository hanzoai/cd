import * as React from 'react';
import {Route, Routes} from 'react-router';

import {UserInfoOverview} from './user-info-overview/user-info-overview';

export const UserInfoContainer = () => (
    <Routes>
        <Route index={true} element={<UserInfoOverview />} />
    </Routes>
);
