resource "grafana_contact_point" "contact_point_a630fca650c9503f" {
  name = "email"

  email {
    disable_resolve_message = true
    addresses               = ["test@grafana.com"]
    single_email            = true
    message                 = "test-message"
    subject                 = "test-subject"
  }
}
