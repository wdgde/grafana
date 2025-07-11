resource "grafana_contact_point" "contact_point_f61592ecd3f18d98" {
  name = "webhook"

  webhook {
    disable_resolve_message = true
    url                     = "http://localhost"
    http_method             = "PUT"
    max_alerts              = 2
    authorization_scheme    = "basic"
    basic_auth_user         = "test-user"
    basic_auth_password     = "[REDACTED]"
    title                   = "test-title"
    message                 = "test-message"

    tlsConfig {
      insecure_skip_verify = false
      ca_certificate       = "[REDACTED]"
      client_certificate   = "[REDACTED]"
      client_key           = "[REDACTED]"
    }

    hmacConfig {
      secret           = "[REDACTED]"
      header           = "X-Grafana-Alerting-Signature"
      timestamp_header = "X-Grafana-Alerting-Timestamp"
    }

    http_config {

      oauth2 {
        client_id     = "test-client-id"
        client_secret = "[REDACTED]"
        token_url     = "https://localhost/auth/token"
        scopes        = ["scope1", "scope2"]
        endpoint_params = {
          param1 = "value1"
          param2 = "value2"
        }

        tls_config {
          insecure_skip_verify = false
          ca_certificate       = "[REDACTED]"
          client_certificate   = "[REDACTED]"
          client_key           = "[REDACTED]"
        }

        proxy_config {
          proxy_url              = "http://localproxy:8080"
          no_proxy               = "localhost"
          proxy_from_environment = false
          proxy_connect_header = {
            X-Proxy-Header = "proxy-value"
          }
        }
      }
    }
  }
}
