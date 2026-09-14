# Deployment

A [GitHub Action](https://github.com/satisfactorymodding/SatisfactoryModManager/blob/master/.github/workflows/push.yml)
automatically creates testing builds from commits on any branch and uploads them to the run as build artifacts.

## Development Builds

Since the build action is triggered on every push to any branch,
non developers can obtain testing builds easily from [GitHub Actions](https://github.com/satisfactorymodding/SatisfactoryModManager/actions/workflows/push.yml) artifact output.
You need to be signed into GitHub to download them, and they expire after some time.

To point development builds to ficsit.dev instead of production, see the [Readme's configuration section](./README.md#Configuration).

## Releases

To create a new published release:

1. Ensure all intended changes are merged into the `master` branch.
2. Ensure someone with Signpath access is around to approve the release in the next few minutes.
   At the time of writing, that's Vilsol and Mircea.
3. Run `task prepare-release VERSION=vX.Y.Z` (version numbers should follow [Semantic Versioning](https://semver.org/)). This will:
   - pull the latest translations and commit them if necessary
   - bump the version in `wails.json` and commit the change
   - tag the result with the specified version
4. Push the changes to the `master` branch.
5. Push the tag to trigger a [GitHub Action to make a release](https://github.com/satisfactorymodding/SatisfactoryModManager/blob/master/.github/workflows/release.yml): `git push origin tag tag_name_here`
6. The action will automatically request signing approval, sending an email to approvers and including the link in the action output.
   The action will busy-wait until approval is granted.
7. Review the draft release automatically created by Goreleaser and edit its description as needed.
8. Publish the release.

Users will automatically be prompted to update next time they launch SMM.
