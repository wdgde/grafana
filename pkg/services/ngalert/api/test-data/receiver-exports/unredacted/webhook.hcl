resource "grafana_contact_point" "contact_point_f61592ecd3f18d98" {
  name = "webhook"

  webhook {
    url                  = "http://localhost"
    http_method          = "PUT"
    max_alerts           = 2
    authorization_scheme = "basic"
    basic_auth_user      = "test-user"
    basic_auth_password  = "test-pass"
    title                = "test-title"
    message              = "test-message"

    tlsConfig {
      insecure_skip_verify = false
      ca_certificate       = "-----BEGIN CERTIFICATE-----\nMIGrMF+gAwIBAgIBATAFBgMrZXAwADAeFw0yNDExMTYxMDI4MzNaFw0yNTExMTYx\nMDI4MzNaMAAwKjAFBgMrZXADIQCf30GvRnHbs9gukA3DLXDK6W5JVgYw6mERU/60\n2M8+rjAFBgMrZXADQQCGmeaRp/AcjeqmJrF5Yh4d7aqsMSqVZvfGNDc0ppXyUgS3\nWMQ1+3T+/pkhU612HR0vFd3vyFhmB4yqFoNV8RML\n-----END CERTIFICATE-----"
      client_certificate   = "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw\nDgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow\nEjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d\n7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B\n5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr\nBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1\nNDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l\nWf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc\n6MF9+Yw1Yy0t\n-----END CERTIFICATE-----"
      client_key           = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49\nAwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q\nEKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==\n-----END EC PRIVATE KEY-----"
    }

    hmacConfig {
      secret           = "test-hmac-secret"
      header           = "X-Grafana-Alerting-Signature"
      timestamp_header = "X-Grafana-Alerting-Timestamp"
    }

    http_config {

      oauth2 {
        client_id     = "test-client-id"
        client_secret = "test-client-secret"
        token_url     = "https://localhost/auth/token"
        scopes        = ["scope1", "scope2"]
        endpoint_params = {
          param1 = "value1"
          param2 = "value2"
        }

        tls_config {
          insecure_skip_verify = false
          ca_certificate       = "-----BEGIN CERTIFICATE-----\nMIGrMF+gAwIBAgIBATAFBgMrZXAwADAeFw0yNDExMTYxMDI4MzNaFw0yNTExMTYx\nMDI4MzNaMAAwKjAFBgMrZXADIQCf30GvRnHbs9gukA3DLXDK6W5JVgYw6mERU/60\n2M8+rjAFBgMrZXADQQCGmeaRp/AcjeqmJrF5Yh4d7aqsMSqVZvfGNDc0ppXyUgS3\nWMQ1+3T+/pkhU612HR0vFd3vyFhmB4yqFoNV8RML\n-----END CERTIFICATE-----"
          client_certificate   = "-----BEGIN CERTIFICATE-----\nMIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw\nDgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow\nEjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d\n7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B\n5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr\nBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1\nNDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l\nWf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc\n6MF9+Yw1Yy0t\n-----END CERTIFICATE-----"
          client_key           = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49\nAwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q\nEKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==\n-----END EC PRIVATE KEY-----"
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
