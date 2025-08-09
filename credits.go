package smallapi

// SmallAPI Credits
var (
	Version     = "1.0.0"
	Author      = "GrandpaEJ"
	Description = "A lightweight Go web framework"
	Credits     = []string{
		"Framework by GrandpaEJ",
		"Documentation powered by GitHub Copilot",
		"Visit: https://github.com/grandpaej/smallapi",
	}
)

// GetCredits returns the framework credits
func GetCredits() []string {
	return Credits
}

// GetVersion returns the framework version
func GetVersion() string {
	return Version
}
