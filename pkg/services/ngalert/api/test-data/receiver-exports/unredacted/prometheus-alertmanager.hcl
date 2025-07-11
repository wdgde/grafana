resource "grafana_contact_point" "contact_point_17a719b7f9f65571" {
  name = "prometheus-alertmanager"

  alertmanager {
    url                 = "https://alertmanager-01.com"
    basic_auth_user     = "grafana"
    basic_auth_password = "admin"
  }
}
