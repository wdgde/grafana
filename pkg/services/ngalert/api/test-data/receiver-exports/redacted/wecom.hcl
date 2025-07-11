resource "grafana_contact_point" "contact_point_712bc60ce0c3f7b4" {
  name = "wecom"

  wecom {
    disable_resolve_message = true
    url                     = "[REDACTED]"
    secret                  = "[REDACTED]"
    agent_id                = "test-agent_id"
    corp_id                 = "test-corp_id"
    message                 = "test-message"
    title                   = "test-title"
    msg_type                = "markdown"
    to_user                 = "test-touser"
  }
}
