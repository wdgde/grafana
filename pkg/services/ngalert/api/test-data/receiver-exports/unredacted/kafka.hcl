resource "grafana_contact_point" "contact_point_f97386721257dda9" {
  name = "kafka"

  kafka {
    rest_proxy_url = "http://localhost/"
    topic          = "test-topic"
    description    = "test-description"
    details        = "test-details"
    username       = "test-user"
    password       = "password"
    api_version    = "v2"
    cluster_id     = "12345"
  }
}
