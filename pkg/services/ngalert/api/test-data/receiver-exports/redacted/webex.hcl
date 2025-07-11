resource "grafana_contact_point" "contact_point_712ef00ce0c67576" {
  name = "webex"

  webex {
    disable_resolve_message = true
    token                   = "[REDACTED]"
    api_url                 = "http://localhost"
    message                 = "test-message"
    room_id                 = "test-room-id"
  }
}
