package agent

import "github.com/shopwarelabs/copilot-extension/copilot"

var tools []copilot.FunctionTool

func init() {
	tools = []copilot.FunctionTool{
		{
			Type: "function",
			Function: copilot.Function{
				Name:        "get_shopware_versions",
				Description: "Get all available Shopware versions",
			},
		},
	}
}
