# Add extra Application info

You can add additional information to an Application on your Hanzo CD dashboard.
If you wish to add clickable links, see [Add external URL](external-url.md). 

This is done by providing the 'info' field a key-value in your Application manifest.

Example:
```yaml
project: cd-demo
source:
  repoURL: 'https://demo'
  path: cd-demo
destination:
  server: https://demo
  namespace: cd-demo
info:
  - name: Example:
    value: >-
      https://example.com
```
![External link](../assets/extra_info-1.png)

The additional information will be visible on the Hanzo CD Application details page.

![External link](../assets/extra_info.png)

![External link](../assets/extra_info-2.png)
