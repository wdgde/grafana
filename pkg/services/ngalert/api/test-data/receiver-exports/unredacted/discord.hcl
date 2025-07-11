resource "grafana_contact_point" "contact_point_350baf38f078a0cf" {
  name = "discord"

  discord {
    url                  = "http://localhost"
    title                = "test-title"
    message              = "test-message"
    avatar_url           = "http://avatar"
    use_discord_username = true
  }
}
