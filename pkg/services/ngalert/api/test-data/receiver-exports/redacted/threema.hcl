resource "grafana_contact_point" "contact_point_687dd587c5780df1" {
  name = "threema"

  threema {
    disable_resolve_message = true
    gateway_id              = "*1234567"
    recipient_id            = "*1234567"
    api_secret              = "[REDACTED]"
    title                   = "test-title"
    description             = "test-description"
  }
}
