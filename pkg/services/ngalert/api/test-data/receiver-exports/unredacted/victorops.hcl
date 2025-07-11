resource "grafana_contact_point" "contact_point_ad74c531be75524c" {
  name = "victorops"

  victorops {
    url          = "http://localhost"
    message_type = "test-messagetype"
    title        = "test-title"
    description  = "test-description"
  }
}
