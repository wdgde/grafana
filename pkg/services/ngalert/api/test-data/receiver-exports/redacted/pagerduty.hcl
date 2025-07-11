resource "grafana_contact_point" "contact_point_f173fe21f16ee67a" {
  name = "pagerduty"

  pagerduty {
    disable_resolve_message = true
    integration_key         = "[REDACTED]"
    severity                = "test-severity"
    class                   = "test-class"
    component               = "test-component"
    group                   = "test-group"
    summary                 = "test-summary"
    source                  = "test-source"
    client                  = "test-client"
    client_url              = "http://localhost/test-client-url"
    url                     = "http://localhost/test-api-url"
  }
}
