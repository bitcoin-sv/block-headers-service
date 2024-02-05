package metrics

const appName = "pulse"

const requestMetricBaseName = "http_request"
const requestCounterName = requestMetricBaseName + "_total"
const requestDurationSecName = requestMetricBaseName + "_duration_seconds"

const domainPrefix = "bux_"

const latestBlockBlockBase = domainPrefix + "latest_block"
const latestBlockHeightName = latestBlockBlockBase + "_height"
const latestBlockTimestampName = latestBlockBlockBase + "_timestamp"
