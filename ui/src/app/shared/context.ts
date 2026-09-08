import {AppContext as KitAppContext, NavigationApi, NotificationsApi, PopupApi} from '../../kit/src';
import {History} from 'history';
import * as React from 'react';
import * as models from './models';

export type AppContext = KitAppContext & {apis: {popup: PopupApi; notifications: NotificationsApi; navigation: NavigationApi; baseHref: string}};

export interface ContextApis {
    popup: PopupApi;
    notifications: NotificationsApi;
    navigation: NavigationApi;
    baseHref: string;
}
export const Context = React.createContext<ContextApis & {history: History}>(null);
export const {Provider, Consumer} = Context;

export const AuthSettingsCtx = React.createContext<models.AuthSettings>(null);
