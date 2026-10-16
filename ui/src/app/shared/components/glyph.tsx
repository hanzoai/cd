import * as React from 'react';

// Line icons for the app chrome, drawn on a 24px grid with a 1.75 stroke in
// currentColor. The shapes are Lucide's (ISC, https://lucide.dev), the set the
// hanzo.ai app shell uses.
const shapes = {
    application: (
        <>
            <path d='M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z' />
            <path d='m3.3 7 8.7 5 8.7-5' />
            <path d='M12 22V12' />
        </>
    ),
    applicationset: (
        <>
            <path d='M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z' />
            <path d='m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65' />
            <path d='m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65' />
        </>
    ),
    settings: (
        <>
            <path d='M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z' />
            <circle cx='12' cy='12' r='3' />
        </>
    ),
    user: (
        <>
            <circle cx='12' cy='12' r='10' />
            <circle cx='12' cy='10' r='3' />
            <path d='M7 20.662V19a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v1.662' />
        </>
    ),
    docs: (
        <>
            <path d='M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z' />
            <path d='M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z' />
        </>
    ),
    panel: (
        <>
            <rect width='18' height='18' x='3' y='3' rx='2' />
            <path d='M9 3v18' />
        </>
    ),
    filter: <path d='M22 3H2l8 9.46V19l4 2v-8.54L22 3z' />
};

export type GlyphName = keyof typeof shapes;

export const Glyph = ({name, size = 16}: {name: GlyphName; size?: number}) => (
    <svg
        className='glyph'
        width={size}
        height={size}
        viewBox='0 0 24 24'
        fill='none'
        stroke='currentColor'
        strokeWidth={1.75}
        strokeLinecap='round'
        strokeLinejoin='round'
        aria-hidden='true'
        focusable='false'>
        {shapes[name]}
    </svg>
);

// The Hanzo mark, filled with currentColor so it follows the theme's ink.
export const HanzoMark = ({size = 20}: {size?: number}) => (
    <svg className='hanzo-mark' width={size} height={size} viewBox='0 0 67 67' fill='currentColor' aria-hidden='true' focusable='false'>
        <path d='M22.21 67V44.6369H0V67H22.21Z' />
        <path d='M66.7038 22.3184H22.2534L0.0878906 44.6367H44.4634L66.7038 22.3184Z' />
        <path d='M22.21 0H0V22.3184H22.21V0Z' />
        <path d='M66.7198 0H44.5098V22.3184H66.7198V0Z' />
        <path d='M66.7198 67V44.6369H44.5098V67H66.7198Z' />
    </svg>
);
