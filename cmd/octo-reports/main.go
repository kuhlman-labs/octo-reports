package octoreports

import (
	"github.com/kuhlman-labs/octo-reports/pkg/org"
)

func main() {
	org.GenerateMembershipReport("enterpriseSlug", "token")
}
