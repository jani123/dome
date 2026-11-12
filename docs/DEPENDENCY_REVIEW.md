# Dependency licensing review

Reviewed 2026-11-12 against the pinned `go.mod`/`go.sum`, for adoption of
GPL-3.0-only without a linking exception. This records engineering review scope,
not a replacement for upstream license texts or a legal opinion.

The actual `agent` build dependency graph contained 149 external modules.
The linked dependencies use Apache-2.0, MIT/ISC/BSD-style, MPL-2.0, and public-domain
terms. No incompatible license was identified in this review. MPL-2.0 components
include SOPS and HashiCorp libraries; their license's standard Exhibit B text is
not by itself a project-specific declaration of secondary-license incompatibility.
Source headers and licensing files were inspected for such declarations; a scan
of 4,065 linked external Go files found no secondary-license incompatibility
header.
`filepath-securejoin` includes BSD-3-Clause and MPL-2.0 files; retain its licensing
material. SQLite's bundled third-party licenses must be included as well.

GPLv3 is compatible with Apache-2.0 and MPL-2.0's secondary-license mechanism
where applicable; preserve upstream notices and original licensing information.
See [GNU's license compatibility list](https://www.gnu.org/licenses/license-list.html)
and [Mozilla's MPL FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/).

For each distributed binary, `scripts/dependency-notices.go` traverses actual
`go list -deps` modules for the target architecture and collects root LICENSE,
COPYING, and NOTICE material plus the Go runtime license. Missing module-level
license material fails the process. Release packaging runs this separately for
amd64 and arm64. The generator gathers evidence/attributions; it does not
automatically decide compatibility or replace a source-file licensing audit.

Rerun/review this process when dependencies change. End-user agent projects must
also review licenses/notices for their own extra modules and provide corresponding
source/build material for the complete distributed covered agent. The generated
starter includes the GPLv3 text; do not treat the initializer's notice inventory
as an inventory of a customer's separately extended agent.

Release source archives include the repository and vendored module sources so
recipients can rebuild with `-mod=vendor` without resolving Go dependencies over
the network (an appropriate Go toolchain remains a prerequisite).
