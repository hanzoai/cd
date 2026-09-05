{
  containerPort: 80,
  image: "ghcr.io/hanzoai/cd-e2e-container:0.2",
  name: "guestbook-ui",
  replicas: 1,
  servicePort: 80,
  type: "ClusterIP",
}
