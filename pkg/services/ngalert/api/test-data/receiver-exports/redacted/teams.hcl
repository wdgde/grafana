resource "grafana_contact_point" "contact_point_c0a5382989108c0d" {
  name = "teams"

  teams {
    disable_resolve_message = true
    url                     = "http://localhost"
    message                 = "test-message"
    title                   = "test-title"
    section_title           = "test-second-title"
  }
}
