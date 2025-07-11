resource "grafana_contact_point" "contact_point_6f9eebee25ba4dd3" {
  name = "pushover"

  pushover {
    disable_resolve_message = true
    user_key                = "[REDACTED]"
    api_token               = "[REDACTED]"
    priority                = 1
    ok_priority             = 2
    retry                   = 555
    expire                  = 333
    device                  = "test-device"
    sound                   = "test-sound"
    ok_sound                = "test-ok-sound"
    title                   = "test-title"
    message                 = "test-message"
    upload_image            = false
  }
}
