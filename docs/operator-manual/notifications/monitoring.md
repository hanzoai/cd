# Monitoring

The Hanzo CD Notification controller serves Prometheus metrics on port 9001.

> [!NOTE]
> The metrics port can be changed using the `--metrics-port` flag in `cd-notifications-controller` deployment.

## Metrics 
The following metrics are available:
 
### `cd_notifications_deliveries_total`
  
 Number of delivered notifications.
 Labels:

* `trigger` - trigger name
* `service` - notification service name
* `succeeded` - flag that indicates if notification was successfully sent or failed

### `cd_notifications_trigger_eval_total`
  
 Number of trigger evaluations.
 Labels:

* `name` - trigger name 
* `triggered` - flag that indicates if trigger condition returned true or false

## Examples

* Grafana Dashboard: [grafana-dashboard.json](grafana-dashboard.json)
