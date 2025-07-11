resource "grafana_contact_point" "contact_point_2f4b14dd48d2cc80" {
  name = "telegram"

  telegram {
    token                    = "test-token"
    chat_id                  = "12345678"
    message_thread_id        = "13579"
    message                  = "test-message"
    parse_mode               = "html"
    disable_web_page_preview = true
    protect_content          = true
    disable_notifications    = true
  }
}
