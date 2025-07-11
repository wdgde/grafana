resource "grafana_contact_point" "contact_point_712ef00ce0c67576" {
  name = "webex"

  webex {
    token   = "12345"
    api_url = "http://localhost"
    message = "test-message"
    room_id = "test-room-id"
  }
}
