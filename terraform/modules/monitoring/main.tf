resource "google_logging_metric" "accounts_created" {
  name   = "ruthless/${var.environment}/accounts_created"
  filter = "jsonPayload.event=\"AccountCreated\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_logging_metric" "sessions_created" {
  name   = "ruthless/${var.environment}/sessions_created"
  filter = "jsonPayload.event=\"SessionCreated\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_logging_metric" "cards_created" {
  name   = "ruthless/${var.environment}/cards_created"
  filter = "jsonPayload.event=\"CardCreated\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_logging_metric" "decks_created" {
  name   = "ruthless/${var.environment}/decks_created"
  filter = "jsonPayload.event=\"DeckCreated\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_logging_metric" "rounds_completed" {
  name   = "ruthless/${var.environment}/rounds_completed"
  filter = "jsonPayload.event=\"RoundCompleted\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_logging_metric" "dau" {
  name   = "ruthless/${var.environment}/dau"
  filter = "jsonPayload.event=\"UserActivity\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
    labels {
      key         = "user_id"
      value_type  = "STRING"
      description = "The user ID"
    }
  }
  label_extractors = {
    "user_id" = "EXTRACT(jsonPayload.user_id)"
  }
}

resource "google_logging_metric" "logins" {
  name   = "ruthless/${var.environment}/logins"
  filter = "jsonPayload.event=\"Login\""
  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_monitoring_dashboard" "usage_metrics" {
  dashboard_json = <<EOF
{
  "displayName": "Ruthless Usage Metrics (${var.environment})",
  "gridLayout": {
    "columns": "2",
    "widgets": [
      {
        "title": "Accounts Created",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/accounts_created\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s",
                  "crossSeriesReducer": "REDUCE_SUM"
                }
              }
            }
          }]
        }
      },
      {
        "title": "Logins",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/logins\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s",
                  "crossSeriesReducer": "REDUCE_SUM"
                }
              }
            }
          }]
        }
      },
      {
        "title": "Sessions Created",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/sessions_created\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s",
                  "crossSeriesReducer": "REDUCE_SUM"
                }
              }
            }
          }]
        }
      },
      {
        "title": "Cards Created",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/cards_created\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s",
                  "crossSeriesReducer": "REDUCE_SUM"
                }
              }
            }
          }]
        }
      },
      {
        "title": "Decks Created",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/decks_created\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s",
                  "crossSeriesReducer": "REDUCE_SUM"
                }
              }
            }
          }]
        }
      },
      {
        "title": "Rounds Completed",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/rounds_completed\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "60s",
                  "crossSeriesReducer": "REDUCE_SUM"
                }
              }
            }
          }]
        }
      },
      {
        "title": "Active Users (Hourly)",
        "xyChart": {
          "dataSets": [{
            "timeSeriesQuery": {
              "timeSeriesFilter": {
                "filter": "metric.type=\"logging.googleapis.com/user/ruthless/${var.environment}/dau\"",
                "aggregation": {
                  "perSeriesAligner": "ALIGN_SUM",
                  "alignmentPeriod": "3600s",
                  "crossSeriesReducer": "REDUCE_COUNT"
                }
              }
            }
          }]
        }
      }
    ]
  }
}
EOF
}
