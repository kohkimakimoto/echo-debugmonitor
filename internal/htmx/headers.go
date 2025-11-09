package htmx

const (
	// HTTP request headers used by HTMX.
	// see: https://htmx.org/reference/#request_headers

	HeaderHXBoosted               = "HX-Boosted"
	HeaderHXCurrentURL            = "HX-Current-URL"
	HeaderHXHistoryRestoreRequest = "HX-History-Restore-Request"
	HeaderHXPrompt                = "HX-Prompt"
	HeaderHXRequest               = "HX-Request"
	HeaderHXTarget                = "HX-Target"
	HeaderHXTriggerName           = "Hx-Trigger-Name"

	// HTTP response headers used by HTMX.
	// see: https://htmx.org/reference/#response_headers

	HeaderHXLocation           = "HX-Location"
	HeaderHXPushUrl            = "HX-Push-Url"
	HeaderHXRedirect           = "HX-Redirect"
	HeaderHXRefresh            = "HX-Refresh"
	HeaderHXReplaceUrl         = "HX-Replace-Url"
	HeaderHXReswap             = "HX-Reswap"
	HeaderHXRetarget           = "HX-Retarget"
	HeaderHXReselect           = "HX-Reselect"
	HeaderHXTriggerAfterSettle = "HX-Trigger-After-Settle"
	HeaderHXTriggerAfterSwap   = "HX-Trigger-After-Swap"

	// The HX-Trigger header can be used in both requests and responses.

	HeaderHXTrigger = "HX-Trigger"
)
