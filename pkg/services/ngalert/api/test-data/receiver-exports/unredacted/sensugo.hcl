resource "grafana_contact_point" "contact_point_36fd7e92473da42b" {
  name = "sensugo"

  sensugo {
    url       = "http://localhost"
    api_key   = "test-api-key"
    entity    = "test-entity"
    check     = "test-check"
    namespace = "test-namespace"
    handler   = "test-handler"
    message   = "test-message"
  }
}
