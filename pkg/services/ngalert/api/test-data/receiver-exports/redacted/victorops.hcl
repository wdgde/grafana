resource "grafana_contact_point" "contact_point_ad74c531be75524c" {
  name = "victorops"

  victorops {
    disable_resolve_message = true
    url                     = "[REDACTED]"
    message_type            = "test-messagetype"
    title                   = "test-title"
    description             = "test-description"
  }
}
