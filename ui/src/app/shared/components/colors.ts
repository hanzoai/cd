export const SUCCESS_COLOR = '#4ADE80';
export const WARNING_COLOR = '#EAB308';
export const FAILED_COLOR = '#EF4444';
export const RUNNING_COLOR = '#2F81F7';
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
