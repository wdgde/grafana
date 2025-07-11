resource "grafana_contact_point" "contact_point_79a686c40649722a" {
  name = "oncall"

  oncall {
    url                  = "http://localhost"
    http_method          = "PUT"
    max_alerts           = 2
    authorization_scheme = "basic"
    basic_auth_user      = "test-user"
    basic_auth_password  = "test-pass"
    title                = "test-title"
    message              = "test-message"
  }
}
