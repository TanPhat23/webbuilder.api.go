// Package seed holds development seed data (demo marketplace templates),
// ported from webbuilderv2/prisma/seed.ts. Load it with `make seed`.
package seed

import _ "embed"

//go:embed marketplace.json
var MarketplaceJSON []byte
