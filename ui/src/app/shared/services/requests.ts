import * as path from 'path';
import * as agent from 'superagent';

import {BehaviorSubject, Observable, Observer} from 'rxjs';
import {filter} from 'rxjs/operators';

type Callback = (data: any) => void;

declare class EventSource {
    public onopen: Callback;
    public onmessage: Callback;
    public onerror: Callback;
    public readyState: number;
    constructor(url: string);
    public close(): void;
}

enum ReadyState {
    CONNECTING = 0,
    OPEN = 1,
    CLOSED = 2,
    DONE = 4
}

let baseHRef = '/';

const onError = new BehaviorSubject<agent.ResponseError>(null);

function toAbsURL(val: string): string {
    return path.join(baseHRef, val);
}

// Hanzo CD talks to the cloud binary's projection under /v1/deploy. The Hanzo
// surface has no `/api/` prefix and there is no separate API server behind this
// UI. One base, one place: every service here derives its paths from this root.
function apiRoot(): string {
    return toAbsURL('/v1/deploy');
}

// The CD API is SuperAdmin-gated and the session cookie is host-scoped, so a
// signed-out browser gets 401/403 on the very first bootstrap call. Without this
// the SPA dead-ends on a blank shell. Every request funnels through initHandlers,
// so this is the one place that knows a refusal happened: send the browser to the
// server's sign-in, preserving where it was headed.
//
// Deliberately client-side: the server keeps its gate closed and answers XHR with
// 403 (never a 302, which would corrupt the API contract). Deciding to navigate is
// the client's job. Guarded so one refusal can't loop or fight a redirect already
// in flight.
let signingIn = false;

function bounceToSignIn() {
    if (signingIn || typeof window === 'undefined') {
        return;
    }
    const here = window.location.pathname + window.location.search;
    if (here.startsWith('/v1/deploy/login') || here.startsWith('/v1/deploy/callback')) {
        return; // already mid-flow
    }
    signingIn = true;
    window.location.assign(toAbsURL('/v1/deploy/login') + '?returnTo=' + encodeURIComponent(here));
}

function initHandlers(req: agent.Request) {
    req.on('error', err => {
        const status = err && (err as any).status;
        if (status === 401 || status === 403) {
            bounceToSignIn();
        }
        onError.next(err);
    });
    return req;
}

export default {
    setBaseHRef(val: string) {
        baseHRef = val;
    },
    agent,
    toAbsURL,
    onError: onError.asObservable().pipe(filter(err => err != null)),
    get(url: string) {
        return initHandlers(agent.get(`${apiRoot()}${url}`));
    },

    post(url: string) {
        return initHandlers(agent.post(`${apiRoot()}${url}`)).set('Content-Type', 'application/json');
    },

    put(url: string) {
        return initHandlers(agent.put(`${apiRoot()}${url}`)).set('Content-Type', 'application/json');
    },

    patch(url: string) {
        return initHandlers(agent.patch(`${apiRoot()}${url}`)).set('Content-Type', 'application/json');
    },

    delete(url: string) {
        return initHandlers(agent.del(`${apiRoot()}${url}`)).set('Content-Type', 'application/json');
    },

    loadEventSource(url: string): Observable<string> {
        return Observable.create((observer: Observer<any>) => {
            const fullUrl = `${apiRoot()}${url}`;

            const abortController = new AbortController();
            let errored = false;

            // EventSource's error event is opaque (no status, no body). On failure,
            // issue a one-shot fetch so we can surface the server's HTTP status and
            // body. This replaces the always-on prefetch, which doubled the open
            // watch streams per subscription and held an HTTP/1.1 connection slot
            // for the lifetime of every watch — see issue #27877.
            const probeAndError = (eventSourceError: any) => {
                if (errored) {
                    return;
                }
                errored = true;
                fetch(fullUrl, {signal: abortController.signal})
                    .then(async response => {
                        if (!response.ok) {
                            const text = await response.text();
                            observer.error({status: response.status, statusText: response.statusText, body: text});
                            onError.next({status: response.status, name: response.statusText, message: text} as agent.ResponseError);
                        } else {
                            await response.body?.cancel();
                            observer.error(eventSourceError);
                            onError.next(eventSourceError);
                        }
                    })
                    .catch(fetchErr => {
                        if (fetchErr.name === 'AbortError') {
                            return;
                        }
                        observer.error(fetchErr);
                        onError.next(fetchErr);
                    });
            };

            let eventSource = new EventSource(fullUrl);
            eventSource.onmessage = msg => observer.next(msg.data);
            eventSource.onerror = e => probeAndError(e);

            // EventSource does not provide easy way to get notification when connection closed.
            // check readyState periodically instead.
            const interval = setInterval(() => {
                if (eventSource && eventSource.readyState === ReadyState.CLOSED) {
                    probeAndError('connection got closed unexpectedly');
                }
            }, 500);
            return () => {
                clearInterval(interval);
                eventSource.close();
                abortController.abort();
                eventSource = null;
            };
        });
    }
};
