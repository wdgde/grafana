resource "grafana_contact_point" "contact_point_350baf38f078a0cf" {
  name = "discord"

  discord {
    disable_resolve_message = true
    url                     = "[REDACTED]"
    title                   = "test-title"
    message                 = "test-message"
    avatar_url              = "http://avatar"
    use_discord_username    = true
  }
}
