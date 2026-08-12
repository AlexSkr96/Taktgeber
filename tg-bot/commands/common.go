package commands

const (
	algoEngineURL = "http://algo-engine:9000"
	healthURL     = algoEngineURL + "/health"
	accountURL    = algoEngineURL + "/account"
	priceURL      = algoEngineURL + "/price"
)

func isHelp(args []string) bool {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		return true
	}
	return false
}
