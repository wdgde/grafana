resource "grafana_contact_point" "contact_point_ecb89c76b8203085" {
  name = "opsgenie"

  opsgenie {
    api_key           = "test-api-key"
    url               = "http://localhost"
    message           = "test-message"
    description       = "test-description"
    auto_close        = false
    override_priority = false
    send_tags_as      = "both"

    responders {
      id   = "test-id"
      type = "team"
    }
    responders {
      username = "test-user"
      type     = "user"
    }
    responders {
      name = "test-schedule"
      type = "schedule"
    }
  }
}
