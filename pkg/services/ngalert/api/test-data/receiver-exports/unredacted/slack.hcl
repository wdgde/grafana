resource "grafana_contact_point" "contact_point_11ff592f6e048923" {
  name = "slack"

  slack {
    endpoint_url    = "http://localhost/endpoint_url"
    url             = "http://localhost/url"
    token           = "test-token"
    recipient       = "test-recipient"
    text            = "test-text"
    title           = "test-title"
    username        = "test-username"
    icon_emoji      = "test-icon"
    icon_url        = "http://localhost/icon_url"
    mention_channel = "channel"
    mention_users   = "test-mentionUsers"
    mention_groups  = "test-mentionGroups"
    color           = "test-color"
  }
}
