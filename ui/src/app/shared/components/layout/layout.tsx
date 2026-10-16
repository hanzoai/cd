import * as React from 'react';
import {Sidebar} from '../../../sidebar/sidebar';
import {GlyphName} from '../glyph';
import {ViewPreferences} from '../../services';
import {useTheme} from '../../utils';

require('./layout.scss');

export interface LayoutProps {
    navItems: Array<{path: string; glyph?: GlyphName; iconClassName?: string; title: string; tooltip?: string}>;
    onVersionClick?: () => void;
    children?: React.ReactNode;
    pref: ViewPreferences;
}

const getBGColor = (theme: string): string => (theme === 'light' ? '#f7f7f7' : '#000000');

export const ThemeWrapper = (props: {children: React.ReactNode; theme: string}) => {
    const [systemTheme] = useTheme({
        theme: props.theme
    });
    return <div className={'theme-' + systemTheme}>{props.children}</div>;
};

export const Layout = (props: LayoutProps) => {
    const [theme] = useTheme({theme: props.pref.theme});
    React.useEffect(() => {
        if (!theme) {
            return undefined;
        }
        document.body.style.background = getBGColor(theme);
        document.documentElement.dataset.theme = theme;
        // Pages outside the layout (login) draw their own dark card.
        return () => {
            delete document.documentElement.dataset.theme;
        };
    }, [theme]);

    return (
        <div className={`theme-${theme}`}>
            <div className={'cd-layout'}>
                <Sidebar onVersionClick={props.onVersionClick} navItems={props.navItems} pref={props.pref} />
                <div className={`cd-layout__content ${props.pref.hideSidebar ? 'cd-layout__content--sb-collapsed' : 'cd-layout__content--sb-expanded'} custom-styles`}>
                    {props.children}
                </div>
            </div>
        </div>
    );
};
