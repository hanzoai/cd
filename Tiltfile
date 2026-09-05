load('ext://restart_process', 'docker_build_with_restart')
load('ext://uibutton', 'cmd_button', 'location')

# tilt version should be >= v0.37.0 for k8s_server_side_apply tilt-dev/tilt#6680
update_settings(
  k8s_server_side_apply="true",
)
# add ui button in web ui to run make codegen-local (top nav)
cmd_button(
    'make codegen-local',
    argv=['sh', '-c', 'make codegen-local'],
    location=location.NAV,
    icon_name='terminal',
    text='make codegen-local',
)

cmd_button(
    'make test-local',
    argv=['sh', '-c', 'make test-local'],
    location=location.NAV,
    icon_name='science',
    text='make test-local',
)

# add ui button in web ui to run make codegen-local (top nav)
cmd_button(
    'make cli-local',
    argv=['sh', '-c', 'make cli-local'],
    location=location.NAV,
    icon_name='terminal',
    text='make cli-local',
)

# detect cluster architecture for build
cluster_version = decode_yaml(local('kubectl version -o yaml'))
platform = cluster_version['serverVersion']['platform']
arch = platform.split('/')[1]

# build the cd binary on code changes
code_deps = [
    'applicationset',
    'cmd',
    'cmpserver',
    'commitserver',
    'common',
    'controller',
    'notification-controller',
    'pkg',
    'reposerver',
    'server',
    'util',
    'go.mod',
    'go.sum',
]
local_resource(
    'build',
    'CGO_ENABLED=0 GOOS=linux GOARCH=' + arch + ' go build -gcflags="all=-N -l" -mod=readonly -o .tilt-bin/cd_linux cmd/main.go',
    deps = code_deps,
    ignore = ['**/*_test.go'],
    allow_parallel=True,
)

# deploy the manifests
k8s_yaml(kustomize('manifests/dev-tilt'))

# build dev image
docker_build_with_restart(
    'ghcr.io/hanzoai/cd:latest', 
    context='.',
    dockerfile='Dockerfile.tilt',
    entrypoint=[
        "/usr/bin/tini",
        "-s",
        "--",
        "dlv",
        "exec",
        "--continue",
        "--accept-multiclient",
        "--headless",
        "--listen=:2345",
        "--api-version=2"
    ],
    platform=platform,
    live_update=[
        sync('.tilt-bin/cd_linux', '/usr/local/bin/cd'),
    ],
    only=[
        '.tilt-bin',
        'hack',
        'entrypoint.sh',
    ],
    restart_file='/tilt/.restart-proc'
)

# build image for cli jobs
docker_build(
    'hanzocd-job', 
    context='.',
    dockerfile='Dockerfile.tilt',
    platform=platform,
    only=[
        '.tilt-bin',
        'hack',
        'entrypoint.sh',
    ]
)

# track server resources and port forward
k8s_resource(
    workload='hanzocd-server',
    objects=[
        'hanzocd-server:serviceaccount',
        'hanzocd-server:role',
        'hanzocd-server:rolebinding',
        'hanzocd-cm:configmap',
        'hanzocd-cmd-params-cm:configmap',
        'hanzocd-gpg-keys-cm:configmap',
        'hanzocd-rbac-cm:configmap',
        'hanzocd-ssh-known-hosts-cm:configmap',
        'hanzocd-tls-certs-cm:configmap',
        'hanzocd-secret:secret',
        'hanzocd-server-network-policy:networkpolicy',
        'hanzocd-server:clusterrolebinding',
        'hanzocd-server:clusterrole',
    ],
    port_forwards=[
        '8080:8080',
        '9345:2345',
        '8083:8083'
    ],
    resource_deps=['build']
)

# track crds
k8s_resource(
    new_name='cluster-resources',
    objects=[
        'applications.hanzoai.io:customresourcedefinition',
        'applicationsets.hanzoai.io:customresourcedefinition',
        'appprojects.hanzoai.io:customresourcedefinition',
        'cd:namespace'
    ]
)

# track hanzocd-repo-server resources and port forward
k8s_resource(
    workload='hanzocd-repo-server',
    objects=[
        'hanzocd-repo-server:serviceaccount',
        'hanzocd-repo-server-network-policy:networkpolicy',
    ],
    port_forwards=[
        '8081:8081',
        '9346:2345',
        '8084:8084'
    ],
    resource_deps=['build']
)

# track hanzocd-redis resources and port forward
k8s_resource(
    workload='hanzocd-redis',
    objects=[
        'hanzocd-redis:serviceaccount',
        'hanzocd-redis:role',
        'hanzocd-redis:rolebinding',
        'hanzocd-redis-network-policy:networkpolicy',
    ],
    port_forwards=[
        '6379:6379',
    ],
    resource_deps=['build']
)

# track hanzocd-applicationset-controller resources
k8s_resource(
    workload='hanzocd-applicationset-controller',
    objects=[
        'hanzocd-applicationset-controller:serviceaccount',
        'hanzocd-applicationset-controller-network-policy:networkpolicy',
        'hanzocd-applicationset-controller:role',
        'hanzocd-applicationset-controller:rolebinding',
        'hanzocd-applicationset-controller:clusterrolebinding',
        'hanzocd-applicationset-controller:clusterrole',
    ],
    port_forwards=[
        '9347:2345',
        '8085:8080',
        '7000:7000'
    ],
    resource_deps=['build']
)

# track hanzocd-application-controller resources
k8s_resource(
    workload='hanzocd-application-controller',
    objects=[
        'hanzocd-application-controller:serviceaccount',
        'hanzocd-application-controller-network-policy:networkpolicy',
        'hanzocd-application-controller:role',
        'hanzocd-application-controller:rolebinding',
        'hanzocd-application-controller:clusterrolebinding',
        'hanzocd-application-controller:clusterrole',
    ],
    port_forwards=[
        '9348:2345',
        '8086:8082',
    ],
    resource_deps=['build']
)

# track hanzocd-notifications-controller resources
k8s_resource(
    workload='hanzocd-notifications-controller',
    objects=[
        'hanzocd-notifications-controller:serviceaccount',
        'hanzocd-notifications-controller-network-policy:networkpolicy',
        'hanzocd-notifications-controller:role',
        'hanzocd-notifications-controller:rolebinding',
        'hanzocd-notifications-cm:configmap',
        'hanzocd-notifications-secret:secret',
    ],
    port_forwards=[
        '9349:2345',
        '8087:9001',
    ],
    resource_deps=['build']
)

# track hanzocd-dex-server resources
k8s_resource(
    workload='hanzocd-dex-server',
    objects=[
        'hanzocd-dex-server:serviceaccount',
        'hanzocd-dex-server-network-policy:networkpolicy',
        'hanzocd-dex-server:role',
        'hanzocd-dex-server:rolebinding',
    ],
    resource_deps=['build']
)

# track hanzocd-commit-server resources
k8s_resource(
    workload='hanzocd-commit-server',
    objects=[
        'hanzocd-commit-server:serviceaccount',
        'hanzocd-commit-server-network-policy:networkpolicy',
    ],
    port_forwards=[
        '9350:2345',
        '8088:8087',
        '8089:8086',
    ],
    resource_deps=['build']
)

# ui dependencies
local_resource(
    'node-modules',
    'pnpm install',
    dir='ui',
    deps = [
        'ui/package.json',
        'ui/pnpm-lock.yaml',
    ],
    allow_parallel=True,
)

# docker for ui
docker_build(
    'hanzocd-ui',
    context='.',
    dockerfile='Dockerfile.ui.tilt',
    entrypoint=['sh', '-c', 'cd /app/ui && pnpm start'],
    only=['ui'],
    live_update=[
        sync('ui', '/app/ui'),
        run('sh -c "cd /app/ui && pnpm install --frozen-lockfile"', trigger=['/app/ui/package.json', '/app/ui/pnpm-lock.yaml']),
    ],
)

# track hanzocd-ui resources and port forward
k8s_resource(
    workload='hanzocd-ui',
    port_forwards=[
        '4000:4000',
    ],
    resource_deps=['node-modules'],
)

# linting
local_resource(
    'lint',
    'make lint-local',
    deps = code_deps,
    allow_parallel=True,
    resource_deps=['vendor']
)

local_resource(
    'lint-ui',
    'make lint-ui-local',
    deps = [
        'ui',
    ],
    allow_parallel=True,
    resource_deps=['node-modules'],
)

local_resource(
    'vendor',
    'go mod vendor',
    deps = [
        'go.mod',
        'go.sum',
    ],
    allow_parallel=True,
)

