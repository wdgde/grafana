resource "grafana_contact_point" "contact_point_712bc60ce0c3f7b4" {
  name = "wecom"

  wecom {
    url      = "test-url"
    secret   = "test-secret"
    agent_id = "test-agent_id"
    corp_id  = "test-corp_id"
    message  = "test-message"
    title    = "test-title"
    msg_type = "markdown"
    to_user  = "test-touser"
  }
}
