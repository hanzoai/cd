export const ARGO_SUCCESS_COLOR = '#4ADE80';
export const ARGO_WARNING_COLOR = '#EAB308';
export const ARGO_FAILED_COLOR = '#EF4444';
export const ARGO_RUNNING_COLOR = '#2F81F7';
export const ARGO_GRAY4_COLOR = '#A3A3A3';
export const ARGO_GRAY6_COLOR = '#737373';
export const ARGO_TERMINATING_COLOR = '#DC2626';
export const ARGO_SUSPENDED_COLOR = '#737373';

export const COLORS = {
    connection_status: {
        failed: ARGO_FAILED_COLOR,
        successful: ARGO_SUCCESS_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    health: {
        degraded: ARGO_FAILED_COLOR,
        healthy: ARGO_SUCCESS_COLOR,
        missing: ARGO_WARNING_COLOR,
        progressing: ARGO_RUNNING_COLOR,
        suspended: ARGO_SUSPENDED_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    operation: {
        error: ARGO_FAILED_COLOR,
        failed: ARGO_FAILED_COLOR,
        running: ARGO_RUNNING_COLOR,
        success: ARGO_SUCCESS_COLOR,
        terminating: ARGO_TERMINATING_COLOR
    },
    sync: {
        synced: ARGO_SUCCESS_COLOR,
        out_of_sync: ARGO_WARNING_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    sync_result: {
        failed: ARGO_FAILED_COLOR,
        synced: ARGO_SUCCESS_COLOR,
        pruned: ARGO_GRAY4_COLOR,
        unknown: ARGO_GRAY4_COLOR
    },
    sync_window: {
        deny: ARGO_FAILED_COLOR,
        allow: ARGO_SUCCESS_COLOR,
        manual: ARGO_WARNING_COLOR,
        inactive: ARGO_GRAY4_COLOR,
        unknown: ARGO_GRAY4_COLOR
    }
};
