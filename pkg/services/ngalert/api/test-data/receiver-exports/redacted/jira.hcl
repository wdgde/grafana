resource "grafana_contact_point" "contact_point_e27ec07ed0b92981" {
  name = "jira"

  jira {
    disable_resolve_message = true
    api_url                 = "http://localhost"
    project                 = "Test Project"
    issue_type              = "Test Issue Type"
    summary                 = "Test Summary"
    description             = "Test Description"
    labels                  = ["Test Label", "Test Label 2"]
    priority                = "Test Priority"
    reopen_transition       = "Test Reopen Transition"
    resolve_transition      = "Test Resolve Transition"
    wont_fix_resolution     = "Test Won't Fix Resolution"
    reopen_duration         = "1m"
    dedup_key_field         = "10000"
    fields = {
      test-field = "test-value"
    }
    user     = "[REDACTED]"
    password = "[REDACTED]"
  }
}
