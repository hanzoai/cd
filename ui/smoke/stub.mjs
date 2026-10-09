// A stand-in for cd-server: serves the built UI (dist/app) and answers the /v1 API
// from the fixtures below, so a browser can load every page without a cluster.
// Usage: node smoke/stub.mjs [port]. Unknown /v1 GETs answer 404 and are logged,
// which the smoke test reads as a failure to render.
import {createReadStream, existsSync, statSync} from 'node:fs';
import {createServer} from 'node:http';
import {extname, join, normalize} from 'node:path';
import {fileURLToPath} from 'node:url';

const root = fileURLToPath(new URL('../dist/app/', import.meta.url));
const port = Number(process.argv[2] || process.env.PORT || 4173);
const cookie = 'cd.token';
const now = '2026-10-09T12:00:00Z';
const ns = 'hanzo-cd';
const revision = '6689520b14a3f6291f8078394ef8586e5eea9c88';
const repo = 'https://github.com/hanzoai/universe';
const destination = {server: 'https://kubernetes.default.svc', namespace: 'hanzo'};

function resources(name, sync, health) {
    return [
        {group: '', version: 'v1', kind: 'Service', namespace: 'hanzo', name, status: 'Synced', health: {status: 'Healthy'}},
        {group: 'apps', version: 'v1', kind: 'Deployment', namespace: 'hanzo', name, status: sync, health: {status: health}},
        {group: 'monitoring.coreos.com', version: 'v1', kind: 'ServiceMonitor', namespace: 'hanzo', name, status: 'Synced'}
    ];
}

function application(name, sync, health) {
    const source = {repoURL: repo, path: `apps/${name}`, targetRevision: 'main'};
    return {
        apiVersion: 'apps.hanzo.ai/v1alpha1',
        kind: 'Application',
        metadata: {name, namespace: ns, uid: `app-${name}`, resourceVersion: '7', creationTimestamp: now, labels: {team: 'platform'}},
        spec: {project: 'default', source, destination, syncPolicy: {automated: {prune: true, selfHeal: true}}},
        status: {
            sync: {status: sync, revision, comparedTo: {source, destination}},
            health: {status: health, message: ''},
            resources: resources(name, sync, health),
            summary: {images: [`ghcr.io/hanzoai/${name}:v1.0.0`]},
            history: [{id: 1, revision, source, deployStartedAt: now, deployedAt: now, initiatedBy: {automated: true}}],
            reconciledAt: now,
            observedAt: now,
            sourceType: 'Kustomize',
            operationState: {
                phase: 'Succeeded',
                message: 'successfully synced (all tasks run)',
                startedAt: now,
                finishedAt: now,
                operation: {sync: {revision}, initiatedBy: {automated: true}},
                syncResult: {revision, source, resources: []}
            }
        }
    };
}

const apps = [application('web', 'Synced', 'Healthy'), application('api', 'OutOfSync', 'Degraded')];

function tree(name) {
    const deployment = {group: 'apps', kind: 'Deployment', namespace: 'hanzo', name, uid: `${name}-deploy`};
    const replicaset = {group: 'apps', kind: 'ReplicaSet', namespace: 'hanzo', name: `${name}-5d9f7c`, uid: `${name}-rs`};
    return {
        nodes: [
            {...deployment, version: 'v1', resourceVersion: '1', parentRefs: [], info: [{name: 'Revision', value: 'Rev:1'}], health: {status: 'Healthy'}, createdAt: now},
            {...replicaset, version: 'v1', resourceVersion: '1', parentRefs: [deployment], info: [{name: 'Revision', value: 'Rev:1'}], health: {status: 'Healthy'}, createdAt: now},
            {
                group: '',
                version: 'v1',
                kind: 'Pod',
                namespace: 'hanzo',
                name: `${name}-5d9f7c-x7k2p`,
                uid: `${name}-pod`,
                resourceVersion: '1',
                parentRefs: [replicaset],
                info: [
                    {name: 'Status Reason', value: 'Running'},
                    {name: 'Containers', value: '1/1'}
                ],
                images: [`ghcr.io/hanzoai/${name}:v1.0.0`],
                networkingInfo: {labels: {app: name}},
                health: {status: 'Healthy'},
                createdAt: now
            },
            {
                group: '',
                version: 'v1',
                kind: 'Service',
                namespace: 'hanzo',
                name,
                uid: `${name}-svc`,
                resourceVersion: '1',
                parentRefs: [],
                info: [],
                networkingInfo: {targetLabels: {app: name}},
                health: {status: 'Healthy'},
                createdAt: now
            },
            {group: 'monitoring.coreos.com', version: 'v1', kind: 'ServiceMonitor', namespace: 'hanzo', name, uid: `${name}-sm`, resourceVersion: '1', parentRefs: [], info: [], createdAt: now}
        ],
        orphanedNodes: [],
        hosts: []
    };
}

// The live object of a tree node, as the server returns it in a resource's manifest.
function manifest(q) {
    const group = q.get('group');
    return {
        apiVersion: group ? `${group}/${q.get('version')}` : q.get('version'),
        kind: q.get('kind'),
        metadata: {name: q.get('resourceName'), namespace: q.get('namespace'), uid: `${q.get('resourceName')}-uid`, creationTimestamp: now, labels: {app: q.get('name')}},
        spec: {},
        status: {}
    };
}

const settings = {
    url: 'https://cd.example.test',
    statusBadgeEnabled: false,
    googleAnalytics: {trackingID: '', anonymizeUsers: true},
    oidcConfig: {name: 'Hanzo'},
    help: {chatUrl: '', chatText: '', binaryUrls: {}},
    userLoginsDisabled: false,
    kustomizeVersions: [],
    uiBannerContent: '',
    execEnabled: false,
    appsInAnyNamespaceEnabled: false,
    hydratorEnabled: false,
    appLabelKey: 'app.kubernetes.io/instance',
    trackingMethod: 'annotation',
    controllerNamespace: ns
};

const version = {
    Version: 'v0.0.0-smoke',
    BuildDate: now,
    GoVersion: 'go1.27.1',
    Compiler: 'gc',
    Platform: 'linux/amd64',
    KustomizeVersion: 'v5.7.1',
    HelmVersion: 'v3.19.0',
    KubectlVersion: 'v0.34.1',
    JsonnetVersion: 'v0.21.0'
};

const project = {
    metadata: {name: 'default', namespace: ns, uid: 'project-default', resourceVersion: '1', creationTimestamp: now},
    spec: {sourceRepos: ['*'], destinations: [{server: '*', namespace: '*'}], clusterResourceWhitelist: [{group: '*', kind: '*'}]},
    status: {}
};

const cluster = {
    name: 'in-cluster',
    server: destination.server,
    config: {tlsClientConfig: {insecure: false}},
    info: {applicationsCount: 2, serverVersion: '1.34', connectionState: {status: 'Successful', message: '', attemptedAt: now}, cacheInfo: {resourcesCount: 120, apisCount: 60, lastCacheSyncTime: now}}
};

// Exact answers for exact paths; the patterns below cover the per-name routes.
const routes = {
    '/v1/settings': () => settings,
    '/v1/settings/plugins': () => ({plugins: []}),
    '/v1/applications': () => ({metadata: {resourceVersion: '7'}, items: apps}),
    '/v1/applicationsets': () => ({
        metadata: {},
        items: [
            {
                apiVersion: 'apps.hanzo.ai/v1alpha1',
                kind: 'ApplicationSet',
                metadata: {name: 'fleet', namespace: ns, uid: 'appset-fleet', creationTimestamp: now},
                spec: {generators: [{list: {elements: [{name: 'web'}, {name: 'api'}]}}], template: {metadata: {name: '{{name}}'}, spec: {project: 'default', source: {repoURL: repo, path: 'apps/{{name}}'}, destination}}},
                status: {conditions: [{type: 'ResourcesUpToDate', status: 'True', message: 'All applications have been generated successfully', lastTransitionTime: now}]}
            }
        ]
    }),
    '/v1/projects': () => ({items: [project]}),
    '/v1/clusters': () => ({items: [cluster]}),
    '/v1/repositories': () => ({
        items: [{repo, type: 'git', name: 'universe', project: '', connectionState: {status: 'Successful', message: '', attemptedAt: now}}]
    }),
    '/v1/repocreds': () => ({items: []}),
    '/v1/certificates': () => ({
        items: [
            {
                serverName: 'github.com',
                certType: 'ssh',
                certSubType: 'ssh-ed25519',
                certData: 'AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl',
                certInfo: 'SHA256:+DiY3wvvV6TuJJhbpZisF/zLDA0zPMSvHdkr4UvCOqU'
            }
        ]
    }),
    '/v1/gpgkeys': () => ({
        items: [{keyID: '4AEE18F83AFDEB23', fingerprint: '5DE3E0509C47EA3CF04A42D34AEE18F83AFDEB23', owner: 'GitHub <noreply@github.com>', trust: 'unknown', subType: 'rsa2048'}]
    }),
    '/v1/account': () => ({items: [{name: 'admin', enabled: true, capabilities: ['login'], tokens: []}]}),
    '/v1/notifications/services': () => ({items: []}),
    '/v1/notifications/triggers': () => ({items: []})
};

const patterns = [
    [/^\/v1\/applications\/([^/]+)$/, m => apps.find(a => a.metadata.name === m[1])],
    [/^\/v1\/applications\/([^/]+)\/resource-tree$/, m => tree(m[1])],
    [/^\/v1\/applications\/([^/]+)\/managed-resources$/, () => ({items: []})],
    [/^\/v1\/applications\/([^/]+)\/resource$/, (m, q) => ({manifest: JSON.stringify(manifest(q))})],
    [/^\/v1\/applications\/([^/]+)\/resource\/links$/, () => ({items: []})],
    [/^\/v1\/applications\/([^/]+)\/resource\/actions$/, () => ({actions: []})],
    [/^\/v1\/applications\/([^/]+)\/events$/, () => ({items: []})],
    [/^\/v1\/applications\/([^/]+)\/syncwindows$/, () => ({activeWindows: [], assignedWindows: [], canSync: true})],
    [/^\/v1\/applications\/([^/]+)\/links$/, () => ({items: []})],
    [/^\/v1\/applications\/([^/]+)\/revisions\/[^/]+\/metadata$/, () => ({author: 'Hanzo Dev <dev@hanzo.ai>', date: now, tags: [], message: 'Ship it'})],
    [/^\/v1\/applications\/([^/]+)\/revisions\/[^/]+\/chartdetails$/, () => ({})],
    [/^\/v1\/applications\/([^/]+)\/revisions\/[^/]+\/ocimetadata$/, () => ({})],
    [/^\/v1\/applicationsets\/([^/]+)$/, () => routes['/v1/applicationsets']().items[0]],
    [/^\/v1\/projects\/([^/]+)$/, () => project],
    [/^\/v1\/projects\/([^/]+)\/detailed$/, () => ({project, globalProjects: [], repositories: [], clusters: [cluster]})],
    [/^\/v1\/projects\/([^/]+)\/syncwindows$/, () => ({windows: []})],
    [/^\/v1\/projects\/([^/]+)\/links$/, () => ({items: []})],
    [/^\/v1\/clusters\/([^/]+)$/, () => cluster],
    [/^\/v1\/account\/can-i\//, () => ({value: 'yes'})],
    [/^\/v1\/account\/([^/]+)$/, () => routes['/v1/account']().items[0]]
];

const types = {
    '.html': 'text/html; charset=utf-8',
    '.js': 'application/javascript',
    '.css': 'text/css',
    '.json': 'application/json',
    '.svg': 'image/svg+xml',
    '.png': 'image/png',
    '.ico': 'image/x-icon',
    '.woff': 'font/woff',
    '.woff2': 'font/woff2',
    '.ttf': 'font/ttf',
    '.map': 'application/json'
};

function json(res, status, body, headers = {}) {
    res.writeHead(status, {'Content-Type': 'application/json', ...headers});
    res.end(JSON.stringify(body));
}

function signedIn(req) {
    return (req.headers.cookie || '').split(/;\s*/).some(c => c.startsWith(`${cookie}=`));
}

function api(req, res, path, query) {
    if (path === '/v1/session' && req.method === 'POST') {
        return json(res, 200, {token: 'smoke'}, {'Set-Cookie': `${cookie}=smoke; Path=/; SameSite=Lax`});
    }
    if (path === '/v1/session' && req.method === 'DELETE') {
        return json(res, 200, {}, {'Set-Cookie': `${cookie}=; Path=/; Max-Age=0`});
    }
    if (path === '/v1/session/userinfo') {
        return json(res, 200, signedIn(req) ? {loggedIn: true, username: 'z', iss: 'hanzocd', groups: []} : {loggedIn: false});
    }
    if (path === '/v1/version') {
        return json(res, 200, signedIn(req) ? version : {Version: version.Version});
    }
    if (path === '/v1/settings') {
        return json(res, 200, settings);
    }
    if (!signedIn(req)) {
        return json(res, 401, {error: 'no session information', code: 16, message: 'no session information'});
    }
    if (path.startsWith('/v1/stream/')) {
        // Watches stay open, as the server's do; a resource-tree watch sends its tree first.
        res.writeHead(200, {'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache', Connection: 'keep-alive'});
        const m = path.match(/^\/v1\/stream\/applications\/([^/]+)\/resource-tree$/);
        res.write(m ? `data: ${JSON.stringify({result: tree(m[1])})}\n\n` : ': open\n\n');
        const beat = setInterval(() => res.write(': beat\n\n'), 15000);
        req.on('close', () => clearInterval(beat));
        return;
    }
    if (req.method !== 'GET') {
        return json(res, 200, {});
    }
    const exact = routes[path];
    if (exact) {
        return json(res, 200, exact(query));
    }
    for (const [re, answer] of patterns) {
        const m = path.match(re);
        if (m) {
            const body = answer(m, query);
            return body ? json(res, 200, body) : json(res, 404, {error: 'not found', code: 5, message: `${path} not found`});
        }
    }
    console.error(`stub: no fixture for ${req.method} ${path}`);
    json(res, 404, {error: 'no fixture', code: 5, message: `no fixture for ${path}`});
}

function file(res, path) {
    const target = normalize(join(root, path));
    if (target.startsWith(root) && existsSync(target) && statSync(target).isFile()) {
        res.writeHead(200, {'Content-Type': types[extname(target)] || 'application/octet-stream'});
        createReadStream(target).pipe(res);
        return;
    }
    // Every other path is a route of the single-page app.
    res.writeHead(200, {'Content-Type': types['.html']});
    createReadStream(join(root, 'index.html')).pipe(res);
}

createServer((req, res) => {
    const url = new URL(req.url, 'http://stub');
    if (url.pathname.startsWith('/v1/')) {
        return api(req, res, url.pathname, url.searchParams);
    }
    if (url.pathname === '/extensions.js') {
        // cd-server concatenates /tmp/extensions/*; with none installed the body is empty.
        res.writeHead(200, {'Content-Type': 'application/javascript'});
        return res.end();
    }
    file(res, decodeURIComponent(url.pathname));
}).listen(port, '127.0.0.1', () => console.log(`stub: http://127.0.0.1:${port}`));
