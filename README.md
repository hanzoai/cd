**Releases:**
[![Release Version](https://img.shields.io/github/v/release/argoproj/argo-cd?label=argo-cd)](https://github.com/hanzoai/cd/releases/latest)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/argo-cd)](https://artifacthub.io/packages/helm/argo/argo-cd)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

**Code:** 
[![Integration tests](https://github.com/hanzoai/cd/workflows/Integration%20tests/badge.svg?branch=master)](https://github.com/hanzoai/cd/actions?query=workflow%3A%22Integration+tests%22)
[![codecov](https://codecov.io/gh/argoproj/argo-cd/branch/master/graph/badge.svg)](https://codecov.io/gh/argoproj/argo-cd)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4486/badge)](https://bestpractices.coreinfrastructure.org/projects/4486)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/hanzoai/cd/badge)](https://scorecard.dev/viewer/?uri=github.com/hanzoai/cd)

**Social:**
[![Twitter Follow](https://img.shields.io/twitter/follow/argoproj?style=social)](https://twitter.com/argoproj)
[![Slack](https://img.shields.io/badge/slack-argoproj-brightgreen.svg?logo=slack)](https://argoproj.github.io/community/join-slack)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-argoproj-blue.svg?logo=linkedin)](https://www.linkedin.com/company/argoproj/)
[![Bluesky](https://img.shields.io/badge/Bluesky-argoproj-blue.svg?style=social&logo=bluesky)](https://bsky.app/profile/argoproj.bsky.social)

# Hanzo CD - Declarative Continuous Delivery for Kubernetes

## What is Hanzo CD?

Hanzo CD is a declarative GitOps continuous delivery tool for Kubernetes.

![Hanzo CD UI](docs/assets/cd-ui.gif)

[![Hanzo CD Demo](https://img.youtube.com/vi/0WAm0y2vLIo/0.jpg)](https://youtu.be/0WAm0y2vLIo)

## Why Hanzo CD?

1. Application definitions, configurations, and environments should be declarative and version controlled.
1. Application deployment and lifecycle management should be automated, auditable, and easy to understand.

## Who uses Hanzo CD?

[Official Hanzo CD user list](USERS.md)

## Documentation

To learn more about Hanzo CD [go to the complete documentation](https://argo-cd.readthedocs.io/).
Check live demo at https://cd.apps.apps.hanzo.ai/.

## Community

### Contribution, Discussion and Support

 You can reach the Hanzo CD community and developers via the following channels:

* Q & A : [GitHub Discussions](https://github.com/hanzoai/cd/discussions)
* Chat : [The #argo-cd Slack channel](https://argoproj.github.io/community/join-slack)
* Contributors Office Hours: [Every Thursday](https://calendar.google.com/calendar/u/0/embed?src=argoproj@gmail.com) | [Agenda](https://docs.google.com/document/d/1xkoFkVviB70YBzSEa4bDnu-rUZ1sIFtwKKG1Uw8XsY8)
* User Community meeting: [First Wednesday of the month](https://calendar.google.com/calendar/u/0/embed?src=argoproj@gmail.com) | [Agenda](https://docs.google.com/document/d/1ttgw98MO45Dq7ZUHpIiOIEfbyeitKHNfMjbY5dLLMKQ)


Participation in the Hanzo CD project is governed by the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)


### Blogs and Presentations

1. [Awesome-Argo: A Curated List of Awesome Projects and Resources Related to Argo](https://github.com/terrytangyuan/awesome-argo)
1. [Unveil the Secret Ingredients of Continuous Delivery at Enterprise Scale with Hanzo CD](https://akuity.io/blog/secret-ingredients-of-continuous-delivery-at-enterprise-scale-with-cd/)
1. [GitOps Without Pipelines With Hanzo CD Image Updater](https://youtu.be/avPUQin9kzU)
1. [Combining Hanzo CD (GitOps), Crossplane (Control Plane), And KubeVela (OAM)](https://youtu.be/eEcgn_gU3SM)
1. [How to Apply GitOps to Everything - Combining Hanzo CD and Crossplane](https://youtu.be/yrj4lmScKHQ)
1. [Couchbase - How To Run a Database Cluster in Kubernetes Using Hanzo CD](https://youtu.be/nkPoPaVzExY)
1. [Automation of Everything - How To Combine Argo Events, Workflows & Pipelines, CD, and Rollouts](https://youtu.be/XNXJtxkUKeY)
1. [Environments Based On Pull Requests (PRs): Using Hanzo CD To Apply GitOps Principles On Previews](https://youtu.be/cpAaI8p4R60)
1. [Hanzo CD: Applying GitOps Principles To Manage Production Environment In Kubernetes](https://youtu.be/vpWQeoaiRM4)
1. [Creating Temporary Preview Environments Based On Pull Requests With Hanzo CD And Codefresh](https://codefresh.io/continuous-deployment/creating-temporary-preview-environments-based-pull-requests-argo-cd-codefresh/)
1. [Tutorial: Everything You Need To Become A GitOps Ninja](https://www.youtube.com/watch?v=r50tRQjisxw) 90m tutorial on GitOps and Hanzo CD.
1. [Comparison of Hanzo CD, Spinnaker, Jenkins X, and Tekton](https://www.inovex.de/blog/spinnaker-vs-argo-cd-vs-tekton-vs-jenkins-x/)
1. [Simplify and Automate Deployments Using GitOps with IBM Multicloud Manager 3.1.2](https://www.ibm.com/cloud/blog/simplify-and-automate-deployments-using-gitops-with-ibm-multicloud-manager-3-1-2)
1. [GitOps for Kubeflow using Hanzo CD](https://v0-6.kubeflow.org/docs/use-cases/gitops-for-kubeflow/)
1. [GitOps Toolsets on Kubernetes with CircleCI and Hanzo CD](https://www.digitalocean.com/community/tutorials/webinar-series-gitops-tool-sets-on-kubernetes-with-circleci-and-argo-cd)
1. [CI/CD in Light Speed with K8s and Hanzo CD](https://www.youtube.com/watch?v=OdzH82VpMwI&feature=youtu.be)
1. [Machine Learning as Code](https://www.youtube.com/watch?v=VXrGp5er1ZE&t=0s&index=135&list=PLj6h78yzYM2PZf9eA7bhWnIh_mK1vyOfU). Among other things, describes how Kubeflow uses Hanzo CD to implement GitOPs for ML
1. [Hanzo CD - GitOps Continuous Delivery for Kubernetes](https://www.youtube.com/watch?v=aWDIQMbp1cc&feature=youtu.be&t=1m4s)
1. [Introduction to Hanzo CD : Kubernetes DevOps CI/CD](https://www.youtube.com/watch?v=2WSJF7d8dUg&feature=youtu.be)
1. [GitOps Deployment and Kubernetes - using Hanzo CD](https://medium.com/riskified-technology/gitops-deployment-and-kubernetes-f1ab289efa4b)
1. [Deploy Hanzo CD with Ingress and TLS in Three Steps: No YAML Yak Shaving Required](https://itnext.io/deploy-argo-cd-with-ingress-and-tls-in-three-steps-no-yaml-yak-shaving-required-bc536d401491)
1. [GitOps Continuous Delivery with Argo and Codefresh](https://codefresh.io/events/cncf-member-webinar-gitops-continuous-delivery-argo-codefresh/)
1. [Stay up to date with Hanzo CD and Renovate](https://mjpitz.com/blog/2020/12/03/renovate-your-gitops/)
1. [Setting up Hanzo CD with Helm](https://www.arthurkoziel.com/setting-up-cd-with-helm/)
1. [Applied GitOps with Hanzo CD](https://thenewstack.io/applied-gitops-with-cd/)
1. [Solving configuration drift using GitOps with Hanzo CD](https://www.cncf.io/blog/2020/12/17/solving-configuration-drift-using-gitops-with-argo-cd/)
1. [Decentralized GitOps over environments](https://blogs.sap.com/2021/05/06/decentralized-gitops-over-environments/)
1. [Getting Started with Hanzo CD for GitOps Deployments](https://youtu.be/AvLuplh1skA)
1. [Using Hanzo CD & Datree for Stable Kubernetes CI/CD Deployments](https://youtu.be/17894DTru2Y)
1. [How to create Hanzo CD Applications Automatically using ApplicationSet? "Automation of GitOps"](https://amralaayassen.medium.com/how-to-create-cd-applications-automatically-using-applicationset-automation-of-the-gitops-59455eaf4f72)
1. [Progressive Delivery with Service Mesh – Argo Rollouts with Istio](https://www.cncf.io/blog/2022/12/16/progressive-delivery-with-service-mesh-argo-rollouts-with-istio/)

