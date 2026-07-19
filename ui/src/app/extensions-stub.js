// Hanzo CD serves no UI extensions. ArgoCD's api-server generates extensions.js
// dynamically; on the static plane that 404s and noises up boot. This no-op keeps
// the <script defer src="extensions.js"> in index.html satisfied.
