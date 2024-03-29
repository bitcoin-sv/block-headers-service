package metrics

const appName = "block-header-service"

const requestMetricBaseName = "http_request"
const requestCounterName = requestMetricBaseName + "_total"
const requestDurationSecName = requestMetricBaseName + "_duration_seconds"

const domainPrefix = "bsv_"

const latestBlockBaseName = domainPrefix + "latest_block"
const latestBlockHeightName = latestBlockBaseName + "_height"
const latestBlockTimestampName = latestBlockBaseName + "_timestamp"
