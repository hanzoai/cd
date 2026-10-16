// State hues, from @hanzo/design: state-success (green-500), state-error
// (red-500), amber for drift, and three rungs of the neutral ladder for the
// states without a verdict: in progress (the theme's second ink, so it reads
// on either ground), unknown, suspended. A var() works in an SVG fill too.
export const SUCCESS_COLOR = '#22C55E';
export const WARNING_COLOR = '#EAB308';
export const FAILED_COLOR = '#EF4444';
export const RUNNING_COLOR = 'var(--hz-ink-2)';
export const GRAY4_COLOR = '#A3A3A3';
export const GRAY6_COLOR = '#737373';
export const TERMINATING_COLOR = '#DC2626';
export const SUSPENDED_COLOR = '#737373';

export const COLORS = {
    connection_status: {
        failed: FAILED_COLOR,
        successful: SUCCESS_COLOR,
        unknown: GRAY4_COLOR
    },
    health: {
        degraded: FAILED_COLOR,
        healthy: SUCCESS_COLOR,
        missing: WARNING_COLOR,
        progressing: RUNNING_COLOR,
        suspended: SUSPENDED_COLOR,
        unknown: GRAY4_COLOR
    },
    operation: {
        error: FAILED_COLOR,
        failed: FAILED_COLOR,
        running: RUNNING_COLOR,
        success: SUCCESS_COLOR,
        terminating: TERMINATING_COLOR
    },
    sync: {
        synced: SUCCESS_COLOR,
        out_of_sync: WARNING_COLOR,
        unknown: GRAY4_COLOR
    },
    sync_result: {
        failed: FAILED_COLOR,
        synced: SUCCESS_COLOR,
        pruned: GRAY4_COLOR,
        unknown: GRAY4_COLOR
    },
    sync_window: {
        deny: FAILED_COLOR,
        allow: SUCCESS_COLOR,
        manual: WARNING_COLOR,
        inactive: GRAY4_COLOR,
        unknown: GRAY4_COLOR
    }
};
