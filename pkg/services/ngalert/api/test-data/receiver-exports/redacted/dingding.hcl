resource "grafana_contact_point" "contact_point_02193678342828e5" {
  name = "dingding"

  dingding {
    disable_resolve_message = true
    url                     = "[REDACTED]"
    message_type            = "actionCard"
    title                   = "Alerts firing: {{ len .Alerts.Firing }}"
    message                 = "{{ len .Alerts.Firing }} alerts are firing, {{ len .Alerts.Resolved }} are resolved"
  }
}
