resource "grafana_contact_point" "contact_point_c5e1497ebfc54217" {
  name = "mqtt"

  mqtt {
    disable_resolve_message = true
    broker_url              = "tcp://localhost:1883"
    client_id               = "grafana-test-client-id"
    topic                   = "grafana/alerts"
    message_format          = "json"
    username                = "test-username"
    password                = "[REDACTED]"
    qos                     = 0
    retain                  = false

    tls_config {
      insecure_skip_verify = false
      ca_certificate       = "[REDACTED]"
      client_certificate   = "[REDACTED]"
      client_key           = "[REDACTED]"
    }
  }
}
