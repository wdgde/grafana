resource "grafana_contact_point" "contact_point_d89fa6186b7bc22d" {
  name = "sns"

  sns {
    disable_resolve_message = true
    api_url                 = "https://sns.us-east-1.amazonaws.com"

    sigv4 {
      region     = "us-east-1"
      access_key = "[REDACTED]"
      secret_key = "[REDACTED]"
      profile    = "default"
      role_arn   = "arn:aws:iam:us-east-1:0123456789:role/my-role"
    }

    topic_arn    = "arn:aws:sns:us-east-1:0123456789:SNSTopicName"
    phone_number = "123-456-7890"
    target_arn   = "arn:aws:sns:us-east-1:0123456789:SNSTopicName"
    subject      = "subject"
    message      = "message"
    attributes = {
      attr1 = "val1"
    }
  }
}
