module.exports = {
    platform: 'github',
    autodiscover: false,
    allowPostUpgradeCommandTemplating: true,
    allowedPostUpgradeCommands: [
        "make mockgen",
        "./hack/install.sh kustomize && make manifests-local",
        "hack/installers/checksums/add-helm-checksums.sh",
        "hack/installers/checksums/add-kustomize-checksums.sh",
        "hack/installers/checksums/add-git-lfs-checksums.sh",
    ],
    binarySource: 'install',
    extends: [
        "github>hanzoai/cd//renovate-presets/commons.json5",
        "github>hanzoai/cd//renovate-presets/custom-managers/shell.json5",
        "github>hanzoai/cd//renovate-presets/custom-managers/yaml.json5",
        "github>hanzoai/cd//renovate-presets/fix/disable-all-updates.json5",
        "github>hanzoai/cd//renovate-presets/devtool.json5",
        "github>hanzoai/cd//renovate-presets/production-binaries.json5",
        "github>hanzoai/cd//renovate-presets/dex.json5",
        "github>hanzoai/cd//renovate-presets/docs.json5",
        "group:aws-sdk-go-v2Monorepo",
        "github>hanzoai/cd//renovate-presets/fix/ignore-paths.json5"
    ],
    ignoreDeps: [
        'github.com/hanzoai/cd/gitops-engine/v3'
    ]
}