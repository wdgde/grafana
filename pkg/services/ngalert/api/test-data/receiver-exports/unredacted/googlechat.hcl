resource "grafana_contact_point" "contact_point_19d5c872dfe99822" {
  name = "googlechat"

  googlechat {
    url     = "http://localhost"
    title   = "test-title"
    message = "test-message"
  }
}
