resource "grafana_contact_point" "contact_point_a630fca650c9503f" {
  name = "email"

  email {
    addresses    = ["test@grafana.com"]
    single_email = true
    message      = "test-message"
    subject      = "test-subject"
  }
}
