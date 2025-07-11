resource "grafana_contact_point" "contact_point_d0526a7ec63098cf" {
  name = "line"

  line {
    disable_resolve_message = true
    token                   = "[REDACTED]"
    title                   = "test-title"
    description             = "test-description"
  }
}
