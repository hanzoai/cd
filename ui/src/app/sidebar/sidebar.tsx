import {Tooltip} from '../../kit/src';
import {Boundary, Placement} from 'popper.js';
import {useData} from '../../kit/v2';
import * as React from 'react';
import {Glyph, GlyphName, HanzoMark} from '../shared/components/glyph';
import {Context} from '../shared/context';
import {services, ViewPreferences} from '../shared/services';

require('./sidebar.scss');

interface SidebarProps {
    onVersionClick: () => void;
    navItems: {path: string; glyph?: GlyphName; iconClassName?: string; title: string; tooltip?: string}[];
    pref: ViewPreferences;
}

export const SIDEBAR_TOOLS_ID = 'sidebar-tools';

// Phones get the collapsed rail whatever the saved preference says, and open
// the full sidebar over the page on demand. Kept in step with $phone-max in
// shared/config.scss.
const PHONE = '(max-width: 639px)';

export const useSidebarTarget = () => {
    const sidebarTarget = React.useRef(document.createElement('div'));

    React.useEffect(() => {
        const sidebar = document.getElementById(SIDEBAR_TOOLS_ID);
        sidebar.appendChild(sidebarTarget?.current);
        return () => {
            sidebarTarget.current?.remove();
        };
    }, []);

    return sidebarTarget;
};

const usePhone = () => {
    const [phone, setPhone] = React.useState(() => window.matchMedia(PHONE).matches);
    React.useEffect(() => {
        const query = window.matchMedia(PHONE);
        const change = () => setPhone(query.matches);
        query.addEventListener('change', change);
        return () => query.removeEventListener('change', change);
    }, []);
    return phone;
};

// A release reports its tag (v3.7.11); a build that was never stamped reports
// "+unknown", which says nothing worth showing.
export const shownVersion = (version?: string): string => (version && !version.includes('+unknown') ? version : '');

export const Sidebar = (props: SidebarProps) => {
    const context = React.useContext(Context);
    const [version] = useData(() => services.version.version());
    const phone = usePhone();
    const [open, setOpen] = React.useState(false);
    const locationPath = context.history.location.pathname;

    // The overlay belongs to the phone layout: leaving it closes the overlay,
    // and so does Escape.
    React.useEffect(() => {
        if (!phone) {
            setOpen(false);
        }
    }, [phone]);
    React.useEffect(() => {
        if (!open) {
            return undefined;
        }
        const close = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false);
        document.addEventListener('keydown', close);
        return () => document.removeEventListener('keydown', close);
    }, [open]);
    const collapsed = phone ? !open : props.pref.hideSidebar;
    const release = shownVersion(version?.Version);

    const toggle = () => {
        if (phone) {
            // The list filters render only while the saved preference says the
            // sidebar is shown, so opening it here opens them too.
            if (!open && props.pref.hideSidebar) {
                services.viewPreferences.updatePreferences({...props.pref, hideSidebar: false});
            }
            setOpen(!open);
        } else {
            services.viewPreferences.updatePreferences({...props.pref, hideSidebar: !props.pref.hideSidebar});
        }
    };

    const tooltipProps = {
        placement: 'right' as Placement,
        popperOptions: {
            modifiers: {
                preventOverflow: {
                    boundariesElement: 'window' as Boundary
                }
            }
        }
    };

    return (
        <>
            {phone && open && <div className='sidebar__scrim' onClick={() => setOpen(false)} />}
            <div className={`sidebar ${collapsed ? 'sidebar--collapsed' : ''} ${phone && open ? 'sidebar--over' : ''}`}>
                <div className='sidebar__container'>
                    <div className='sidebar__head'>
                        <a
                            className='sidebar__brand'
                            href={context.baseHref}
                            title='Hanzo CD'
                            onClick={e => {
                                if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) {
                                    return;
                                }
                                e.preventDefault();
                                setOpen(false);
                                context.history.push('/');
                            }}>
                            <HanzoMark size={18} />
                            {!collapsed && <span className='sidebar__name'>Hanzo CD</span>}
                        </a>
                        {!collapsed && release && (
                            <button type='button' className='sidebar__version' title='Server version' onClick={props.onVersionClick}>
                                {release}
                            </button>
                        )}
                        <button type='button' className='sidebar__toggle' aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'} onClick={toggle}>
                            <Glyph name='panel' />
                        </button>
                    </div>

                    <nav aria-label='Main'>
                        <ul className='sidebar__nav'>
                            {(props.navItems || []).map(item => {
                                // Full base-href-prefixed href so native middle-click / right-click / status-bar URL preview
                                // work; navItem paths are absolute (e.g. /applications) while baseHref carries a trailing slash.
                                const href = `${context.baseHref}${item.path.replace(/^\//, '')}`;
                                const active = locationPath === item.path || locationPath.startsWith(`${item.path}/`);
                                return (
                                    <li key={item.path} className='sidebar__nav-li'>
                                        <Tooltip content={<div className='sidebar__tooltip'>{item?.tooltip || item.title}</div>} {...tooltipProps}>
                                            <a
                                                href={href}
                                                // The title text is hidden when collapsed, so name the link explicitly for screen readers.
                                                aria-label={item.title}
                                                aria-current={active ? 'page' : undefined}
                                                className={`sidebar__nav-item ${active ? 'sidebar__nav-item--active' : ''}`}
                                                onClick={e => {
                                                    // Let modifier- and non-left-clicks fall through to the browser (open in new tab/window).
                                                    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) {
                                                        return;
                                                    }
                                                    e.preventDefault();
                                                    setOpen(false);
                                                    context.history.push(item.path);
                                                }}>
                                                <span className='sidebar__nav-icon'>{item.glyph ? <Glyph name={item.glyph} /> : <i className={item.iconClassName || ''} />}</span>
                                                {!collapsed && <span className='sidebar__nav-label'>{item.title}</span>}
                                            </a>
                                        </Tooltip>
                                    </li>
                                );
                            })}

                            {collapsed && (
                                <li className='sidebar__nav-li'>
                                    <Tooltip content='Show filters' {...tooltipProps}>
                                        <button type='button' aria-label='Show filters' onClick={toggle} className='sidebar__nav-item sidebar__filter-button'>
                                            <span className='sidebar__nav-icon'>
                                                <Glyph name='filter' />
                                            </span>
                                        </button>
                                    </Tooltip>
                                </li>
                            )}
                        </ul>
                    </nav>
                </div>
                <div id={SIDEBAR_TOOLS_ID} className='sidebar__tools' />
            </div>
        </>
    );
};
