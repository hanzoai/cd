package command

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/hanzoai/cd/applicationset/progressivesync"

	"github.com/hanzoai/cd/util/vendored/stats"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/hanzoai/cd/reposerver/apiclient"
	logutils "github.com/hanzoai/cd/util/log"
	"github.com/hanzoai/cd/util/profile"
	"github.com/hanzoai/cd/util/tls"

	"github.com/hanzoai/cd/applicationset/controllers"
	"github.com/hanzoai/cd/applicationset/generators"
	"github.com/hanzoai/cd/applicationset/utils"
	"github.com/hanzoai/cd/applicationset/webhook"
	cmdutil "github.com/hanzoai/cd/cmd/util"
	"github.com/hanzoai/cd/common"
	"github.com/hanzoai/cd/util/env"
	"github.com/hanzoai/cd/util/github_app"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	appsetmetrics "github.com/hanzoai/cd/applicationset/metrics"
	"github.com/hanzoai/cd/applicationset/services"
	appv1alpha1 "github.com/hanzoai/cd/pkg/apis/application/v1alpha1"
	"github.com/hanzoai/cd/util/cli"
	"github.com/hanzoai/cd/util/db"
	"github.com/hanzoai/cd/util/errors"
	settings "github.com/hanzoai/cd/util/settings"
)

var gitSubmoduleEnabled = env.ParseBoolFromEnv(common.EnvGitSubmoduleEnabled, true)

func NewCommand() *cobra.Command {
	var (
		clientConfig                 clientcmd.ClientConfig
		metricsAddr                  string
		probeBindAddr                string
		webhookAddr                  string
		enableLeaderElection         bool
		applicationSetNamespaces     []string
		cdRepoServer                 string
		policy                       string
		enablePolicyOverride         bool
		debugLog                     bool
		dryRun                       bool
		enableProgressiveSyncs       bool
		enableNewGitFileGlobbing     bool
		repoServerPlaintext          bool
		repoServerStrictTLS          bool
		repoServerTimeoutSeconds     int
		maxConcurrentReconciliations int
		scmRootCAPath                string
		allowedScmProviders          []string
		globalPreservedAnnotations   []string
		globalPreservedLabels        []string
		enableGitHubAPIMetrics       bool
		metricsAplicationsetLabels   []string
		enableScmProviders           bool
		webhookParallelism           int
		tokenRefStrictMode           bool
		maxResourcesStatusCount      int
		cacheSyncPeriod              time.Duration
		concurrentApplicationUpdates int
		repoServerClientTLSConfigSrc func() (tls.Configuration, error)
		scmProxyURL                  string
		scmNoProxy                   string
	)
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	command := cobra.Command{
		Use:               common.CommandApplicationSetController,
		Short:             "Starts Hanzo CD ApplicationSet controller",
		DisableAutoGenTag: true,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()

			vers := common.GetVersion()
			namespace, _, err := clientConfig.Namespace()
			applicationSetNamespaces = append(applicationSetNamespaces, namespace)

			errors.CheckError(err)
			vers.LogStartupInfo(
				"Hanzo CD ApplicationSet Controller",
				map[string]any{
					"namespace": namespace,
				},
			)

			cli.SetLogFormat(cmdutil.LogFormat)

			if debugLog {
				cli.SetLogLevel("debug")
			} else {
				cli.SetLogLevel(cmdutil.LogLevel)
			}

			ctrl.SetLogger(logutils.NewLogrusLogger(logutils.NewWithCurrentConfig()))

			// Recover from panic and log the error using the configured logger instead of the default.
			defer func() {
				if r := recover(); r != nil {
					log.WithField("trace", string(debug.Stack())).Fatal("Recovered from panic: ", r)
				}
			}()

			restConfig, err := clientConfig.ClientConfig()
			errors.CheckError(err)

			restConfig.UserAgent = fmt.Sprintf("cd-applicationset-controller/%s (%s)", vers.Version, vers.Platform)

			policyObj, exists := utils.Policies[policy]
			if !exists {
				log.Error("Policy value can be: sync, create-only, create-update, create-delete, default value: sync")
				os.Exit(1)
			}

			// By default, watch all namespaces
			var watchedNamespace string
			// If the applicationset-namespaces contains only one namespace it corresponds to the current namespace
			if len(applicationSetNamespaces) == 1 {
				watchedNamespace = (applicationSetNamespaces)[0]
			} else if enableScmProviders && len(allowedScmProviders) == 0 {
				log.Error("When enabling applicationset in any namespace using applicationset-namespaces, you must either set --enable-scm-providers=false or specify --allowed-scm-providers")
				os.Exit(1)
			}

			cacheOpt := ctrlcache.Options{SyncPeriod: &cacheSyncPeriod}

			if watchedNamespace != "" {
				cacheOpt.DefaultNamespaces = map[string]ctrlcache.Config{
					watchedNamespace: {},
				}
			}

			cfg := ctrl.GetConfigOrDie()
			err = appv1alpha1.SetK8SConfigDefaults(cfg)
			if err != nil {
				log.Error(err, "Unable to apply K8s REST config defaults")
				os.Exit(1)
			}

			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: metricsAddr,
				},
				Cache:                  cacheOpt,
				HealthProbeBindAddress: probeBindAddr,
				LeaderElection:         enableLeaderElection,
				LeaderElectionID:       "58ac56fa.applicationsets.apps.hanzo.ai",
				Client: ctrlclient.Options{
					DryRun: &dryRun,
				},
			})
			if err != nil {
				log.Error(err, "unable to start manager")
				os.Exit(1)
			}

			pprofMux := http.NewServeMux()
			profile.RegisterProfiler(pprofMux)
			// This looks a little strange. Eg, not using ctrl.Options PprofBindAddress and then adding the pprof mux
			// to the metrics server. However, it allows for the controller to dynamically expose the pprof endpoints
			// and use the existing metrics server, the same pattern that the application controller and api-server follow.
			if err = mgr.AddMetricsServerExtraHandler("/debug/pprof/", pprofMux); err != nil {
				log.Error(err, "failed to register pprof handlers")
			}
			dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
			errors.CheckError(err)
			k8sClient, err := kubernetes.NewForConfig(mgr.GetConfig())
			errors.CheckError(err)

			settingsMgr := settings.NewSettingsManager(ctx, k8sClient, namespace)
			appDB := db.NewDB(namespace, settingsMgr, k8sClient)

			clusterInformer, err := settings.NewClusterInformer(k8sClient, namespace)
			if err != nil {
				log.Error(err, "unable to create cluster informer")
				os.Exit(1)
			}
			go clusterInformer.Run(ctx.Done())

			if !cache.WaitForCacheSync(ctx.Done(), clusterInformer.HasSynced) {
				log.Error("Timed out waiting for cluster cache to sync")
				os.Exit(1)
			}

			scmConfig := generators.NewSCMConfig(
				scmRootCAPath,
				allowedScmProviders,
				enableScmProviders,
				enableGitHubAPIMetrics,
				github_app.NewAuthCredentials(appDB.(db.RepoCredsDB)),
				tokenRefStrictMode, generators.WithProxyURL(scmProxyURL),
				generators.WithNoProxyList(scmNoProxy))

			tlsConfig, err := repoServerClientTLSConfigSrc()
			errors.CheckError(err)
			tlsConfig.DisableTLS = repoServerPlaintext
			tlsConfig.StrictValidation = tlsConfig.StrictValidation || repoServerStrictTLS

			if !repoServerPlaintext && repoServerStrictTLS && tlsConfig.Certificates == nil {
				pool, err := tls.LoadX509CertPool(
					env.StringFromEnv(common.EnvAppConfigPath, common.DefaultAppConfigPath)+"/reposerver/tls/tls.crt",
					env.StringFromEnv(common.EnvAppConfigPath, common.DefaultAppConfigPath)+"/reposerver/tls/ca.crt",
				)
				errors.CheckError(err)
				tlsConfig.Certificates = pool
			}

			repoClientset := apiclient.NewRepoServerClientset(cdRepoServer, repoServerTimeoutSeconds, tlsConfig)
			repos := services.NewArgoCDService(appDB, gitSubmoduleEnabled, repoClientset, enableNewGitFileGlobbing)

			topLevelGenerators := generators.GetGenerators(ctx, mgr.GetClient(), k8sClient, namespace, repos, dynamicClient, scmConfig, clusterInformer)
			cacheSyncClient := utils.NewCacheSyncingClient(mgr.GetClient(), mgr.GetCache())

			// start a webhook server that listens to incoming webhook payloads
			webhookHandler, err := webhook.NewWebhookHandler(webhookParallelism, settingsMgr, mgr.GetClient(), topLevelGenerators)
			if err != nil {
				log.Error(err, "failed to create webhook handler")
			}
			if webhookHandler != nil {
				startWebhookServer(webhookHandler, webhookAddr)
			}

			metrics := appsetmetrics.NewApplicationsetMetrics(
				utils.NewAppsetLister(mgr.GetClient()),
				metricsAplicationsetLabels,
				func(appset *appv1alpha1.ApplicationSet) bool {
					return utils.IsNamespaceAllowed(applicationSetNamespaces, appset.Namespace)
				})
			appsetReconciler := &controllers.ApplicationSetReconciler{
				Generators: topLevelGenerators,
				Client:     cacheSyncClient,
				Scheme:     mgr.GetScheme(),
				// FIXME: record.EventRecorder -> events.EventRecorder
				// nolint:staticcheck
				Recorder:                     mgr.GetEventRecorderFor("applicationset-controller"),
				Renderer:                     &utils.Render{},
				Policy:                       policyObj,
				EnablePolicyOverride:         enablePolicyOverride,
				KubeClientset:                k8sClient,
				DB:                           appDB,
				ControllerNamespace:              namespace,
				ApplicationSetNamespaces:     applicationSetNamespaces,
				EnableProgressiveSyncs:       enableProgressiveSyncs,
				SCMRootCAPath:                scmRootCAPath,
				GlobalPreservedAnnotations:   globalPreservedAnnotations,
				GlobalPreservedLabels:        globalPreservedLabels,
				Metrics:                      &metrics,
				MaxResourcesStatusCount:      maxResourcesStatusCount,
				ClusterInformer:              clusterInformer,
				ConcurrentApplicationUpdates: concurrentApplicationUpdates,
			}
			appsetReconciler.ProgressiveSyncManager = progressivesync.NewManager(cacheSyncClient, appsetReconciler)

			if err = appsetReconciler.SetupWithManager(mgr, enableProgressiveSyncs, maxConcurrentReconciliations); err != nil {
				log.Error(err, "unable to create controller", "controller", "ApplicationSet")
				os.Exit(1)
			}

			stats.StartStatsTicker(10 * time.Minute)
			log.Info("Starting manager")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				log.Error(err, "problem running manager")
				os.Exit(1)
			}
			return nil
		},
	}
	clientConfig = cli.AddKubectlFlagsToCmd(&command)
	command.Flags().StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	command.Flags().StringVar(&probeBindAddr, "probe-addr", ":8081", "The address the probe endpoint binds to.")
	command.Flags().StringVar(&webhookAddr, "webhook-addr", ":7000", "The address the webhook endpoint binds to.")
	command.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_ENABLE_LEADER_ELECTION", false),
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	command.Flags().StringSliceVar(&applicationSetNamespaces, "applicationset-namespaces", env.StringsFromEnv("CD_APPLICATIONSET_CONTROLLER_NAMESPACES", []string{}, ","), "Hanzo CD applicationset namespaces")
	command.Flags().StringVar(&cdRepoServer, "cd-repo-server", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_REPO_SERVER", common.DefaultRepoServerAddr), "Hanzo CD repo server address")
	command.Flags().StringVar(&policy, "policy", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_POLICY", ""), "Modify how application is synced between the generator and the cluster. Default is '' (empty), which means AppSets default to 'sync', but they may override that default. Setting an explicit value prevents AppSet-level overrides, unless --allow-policy-override is enabled. Explicit options are: 'sync' (create & update & delete), 'create-only', 'create-update' (no deletion), 'create-delete' (no update)")
	command.Flags().BoolVar(&enablePolicyOverride, "enable-policy-override", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_ENABLE_POLICY_OVERRIDE", policy == ""), "For security reason if 'policy' is set, it is not possible to override it at applicationSet level. 'allow-policy-override' allows user to define their own policy")
	command.Flags().BoolVar(&debugLog, "debug", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_DEBUG", false), "Print debug logs. Takes precedence over loglevel")
	command.Flags().StringVar(&cmdutil.LogFormat, "logformat", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_LOGFORMAT", "json"), "Set the logging format. One of: json|text")
	command.Flags().StringVar(&cmdutil.LogLevel, "loglevel", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_LOGLEVEL", "info"), "Set the logging level. One of: debug|info|warn|error")
	command.Flags().StringSliceVar(&allowedScmProviders, "allowed-scm-providers", env.StringsFromEnv("CD_APPLICATIONSET_CONTROLLER_ALLOWED_SCM_PROVIDERS", []string{}, ","), "The list of allowed custom SCM provider API URLs. This restriction does not apply to SCM or PR generators which do not accept a custom API URL. (Default: Empty = all)")
	command.Flags().BoolVar(&enableScmProviders, "enable-scm-providers", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_ENABLE_SCM_PROVIDERS", true), "Enable retrieving information from SCM providers, used by the SCM and PR generators (Default: true)")
	command.Flags().BoolVar(&dryRun, "dry-run", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_DRY_RUN", false), "Enable dry run mode")
	command.Flags().BoolVar(&tokenRefStrictMode, "token-ref-strict-mode", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_TOKENREF_STRICT_MODE", false), fmt.Sprintf("Set to true to require secrets referenced by SCM providers to have the %s=%s label set (Default: false)", common.LabelKeySecretType, common.LabelValueSecretTypeSCMCreds))
	command.Flags().BoolVar(&enableProgressiveSyncs, "enable-progressive-syncs", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_ENABLE_PROGRESSIVE_SYNCS", false), "Enable use of the experimental progressive syncs feature.")
	command.Flags().BoolVar(&enableNewGitFileGlobbing, "enable-new-git-file-globbing", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_ENABLE_NEW_GIT_FILE_GLOBBING", false), "Enable new globbing in Git files generator.")
	command.Flags().BoolVar(&repoServerPlaintext, "repo-server-plaintext", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_REPO_SERVER_PLAINTEXT", false), "Disable TLS on connections to repo server")
	command.Flags().BoolVar(&repoServerStrictTLS, "repo-server-strict-tls", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_REPO_SERVER_STRICT_TLS", false), "Whether to use strict validation of the TLS cert presented by the repo server")
	errors.CheckError(command.Flags().MarkDeprecated("repo-server-strict-tls", "use --repo-server-ca-cert-path instead"))
	command.Flags().IntVar(&repoServerTimeoutSeconds, "repo-server-timeout-seconds", env.ParseNumFromEnv("CD_APPLICATIONSET_CONTROLLER_REPO_SERVER_TIMEOUT_SECONDS", 60, 0, math.MaxInt64), "Repo server RPC call timeout seconds.")
	command.Flags().IntVar(&maxConcurrentReconciliations, "concurrent-reconciliations", env.ParseNumFromEnv("CD_APPLICATIONSET_CONTROLLER_CONCURRENT_RECONCILIATIONS", 10, 1, math.MaxInt), "Max concurrent reconciliations limit for the controller")
	command.Flags().StringVar(&scmRootCAPath, "scm-root-ca-path", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_SCM_ROOT_CA_PATH", ""), "Provide Root CA Path for self-signed TLS Certificates")
	command.Flags().StringVar(&scmProxyURL, "scm-proxy-url", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_SCM_PROXY_URL", ""), "HTTP/HTTPS proxy URL for outbound SCM provider API requests (GitHub, GitLab, etc.). Does NOT affect Kubernetes API server connectivity — use --proxy-url (kubectl flag) for that.")
	command.Flags().StringVar(&scmNoProxy, "scm-no-proxy", env.StringFromEnv("CD_APPLICATIONSET_CONTROLLER_SCM_NO_PROXY", ""), "Comma-separated list of hosts that should bypass the --scm-proxy-url proxy.")
	command.Flags().StringSliceVar(&globalPreservedAnnotations, "preserved-annotations", env.StringsFromEnv("CD_APPLICATIONSET_CONTROLLER_GLOBAL_PRESERVED_ANNOTATIONS", []string{}, ","), "Sets global preserved field values for annotations")
	command.Flags().StringSliceVar(&globalPreservedLabels, "preserved-labels", env.StringsFromEnv("CD_APPLICATIONSET_CONTROLLER_GLOBAL_PRESERVED_LABELS", []string{}, ","), "Sets global preserved field values for labels")
	command.Flags().IntVar(&webhookParallelism, "webhook-parallelism-limit", env.ParseNumFromEnv("CD_APPLICATIONSET_CONTROLLER_WEBHOOK_PARALLELISM_LIMIT", 50, 1, 1000), "Number of webhook requests processed concurrently")
	command.Flags().StringSliceVar(&metricsAplicationsetLabels, "metrics-applicationset-labels", []string{}, "List of Application labels that will be added to the cd_applicationset_labels metric")
	command.Flags().BoolVar(&enableGitHubAPIMetrics, "enable-github-api-metrics", env.ParseBoolFromEnv("CD_APPLICATIONSET_CONTROLLER_ENABLE_GITHUB_API_METRICS", false), "Enable GitHub API metrics for generators that use the GitHub API")
	command.Flags().IntVar(&maxResourcesStatusCount, "max-resources-status-count", env.ParseNumFromEnv("CD_APPLICATIONSET_CONTROLLER_MAX_RESOURCES_STATUS_COUNT", 5000, 0, math.MaxInt), "Max number of resources stored in appset status.")
	command.Flags().DurationVar(&cacheSyncPeriod, "cache-sync-period", env.ParseDurationFromEnv("CD_APPLICATIONSET_CONTROLLER_CACHE_SYNC_PERIOD", time.Hour*10, 0, time.Hour*24), "Period at which the manager client cache is forcefully resynced with the Kubernetes API server. 0 disables periodic resync.")
	command.Flags().IntVar(&concurrentApplicationUpdates, "concurrent-application-updates", env.ParseNumFromEnv("CD_APPLICATIONSET_CONTROLLER_CONCURRENT_APPLICATION_UPDATES", 1, 1, 200), "Number of concurrent Application create/update/delete operations per ApplicationSet reconcile.")
	repoServerClientTLSConfigSrc = tls.AddClientTLSFlagsToCmdWithPrefix(&command, "APPLICATIONSET_CONTROLLER")
	return &command
}

func startWebhookServer(webhookHandler *webhook.WebhookHandler, webhookAddr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/webhook", webhookHandler.Handler)
	go func() {
		log.Infof("Starting webhook server %s", webhookAddr)
		err := http.ListenAndServe(webhookAddr, mux)
		if err != nil {
			log.Error(err, "failed to start webhook server")
			os.Exit(1)
		}
	}()
}
